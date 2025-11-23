package payment

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kytapay/api-v2/config"
	"github.com/kytapay/api-v2/helpers"
	"github.com/kytapay/api-v2/models"
	"github.com/kytapay/api-v2/repositories"
	"github.com/kytapay/api-v2/services"
)

type EWalletPaymentController struct {
	db              *sql.DB
	helper          *helpers.PaymentHelper
	linkQuService   *services.LinkQuService
	tokenRepo       *repositories.TokenRepository
	settingRepo     *repositories.SettingRepository
	transactionRepo *repositories.TransactionInfoRepository
	merchantTxRepo  *repositories.MerchantTransactionRepository
	callbackRepo    *repositories.CallbackRepository
}

func NewEWalletPaymentController(db *sql.DB) *EWalletPaymentController {
	return &EWalletPaymentController{
		db:              db,
		helper:          helpers.NewPaymentHelper(db),
		linkQuService:   services.NewLinkQuService(),
		tokenRepo:       repositories.NewTokenRepository(db),
		settingRepo:     repositories.NewSettingRepository(db),
		transactionRepo: repositories.NewTransactionInfoRepository(db),
		merchantTxRepo:  repositories.NewMerchantTransactionRepository(db),
		callbackRepo:    repositories.NewCallbackRepository(db),
	}
}

var supportedEWallets = []string{"GOPAY", "DANA", "LINKAJA", "OVO", "SHOPEEPAY", "ASTRAPAY"}

// CreateEWallet handles POST /v2/payments/create/ewallet
func (ewc *EWalletPaymentController) CreateEWallet(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010600",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010600",
			"response_message": "Unauthorized",
		})
		return
	}

	// Check maintenance
	isMaintenancePayment, err := ewc.settingRepo.GetMaintenancePayment()
	if err == nil && isMaintenancePayment {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030600",
			"response_message": "The Payment service is currently under maintenance. Please try again later.",
		})
		return
	}

	isMaintenanceEWallet, err := ewc.settingRepo.GetMaintenanceEwallet()
	if err == nil && isMaintenanceEWallet {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030601",
			"response_message": "The E-Wallet payment service is currently unavailable. Please use other payment methods.",
		})
		return
	}

	// Verify token and merchant
	tokenData, merchantApp, merchant, user, err := ewc.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	if tokenData == nil || merchant == nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010601",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		ReferenceID string  `json:"reference_id" binding:"required"`
		Amount       float64 `json:"amount" binding:"required"`
		ChannelCode  string  `json:"channel_code" binding:"required"`
		NotifyURL    string  `json:"notify_url" binding:"required"`
		SuccessURL   string  `json:"success_url" binding:"required"`
		FailedURL    string  `json:"failed_url" binding:"required"`
		ExpiresTime  *int    `json:"expires_time"`
		PhoneNumber  *string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000600",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	if reqBody.Amount < 10000 || reqBody.Amount > 500000000 {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030602",
			"response_message": "Exceeds Transaction Amount Limit",
		})
		return
	}

	channelCodeUpper := strings.ToUpper(reqBody.ChannelCode)
	isSupported := false
	for _, ewallet := range supportedEWallets {
		if ewallet == channelCodeUpper {
			isSupported = true
			break
		}
	}

	if !isSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000601",
			"response_message": "Invalid Field Format",
		})
		return
	}

	// Get payment method
	paymentMethod, err := ewc.helper.GetPaymentMethodByCode(channelCodeUpper)
	if err != nil || paymentMethod == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040600",
			"response_message": "No methods available",
		})
		return
	}

	// Get fees limit
	feeLimit, err := ewc.helper.GetFeesLimit(10, paymentMethod.ID)
	if err != nil || feeLimit == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000600",
			"response_message": "General error",
		})
		return
	}

	// Calculate charges
	chargePercentage := feeLimit.ChargePercentage
	chargeFixed := feeLimit.ChargeFixed
	amountPercentage := (reqBody.Amount * chargePercentage) / 100
	total := reqBody.Amount - amountPercentage - chargeFixed

	// Generate grant ID
	grantID := uuid.New().String()

	// Calculate expiration
	expiresTime := 86400
	if reqBody.ExpiresTime != nil {
		expiresTime = *reqBody.ExpiresTime
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	expiresAt := now.Add(time.Duration(expiresTime) * time.Second)
	expiresAtStr := expiresAt.Format("2006-01-02 15:04:05")
	expiresAtISO := expiresAt.Format(time.RFC3339)
	expiresAtLinkQu := expiresAt.Format("20060102150405")

	// Generate customer ID (unique)
	customerID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Map channel code to retail code
	retailCodeMap := map[string]string{
		"DANA":      "PAYDANA",
		"LINKAJA":   "PAYLINKAJA",
		"SHOPEEPAY": "PAYSHOPEE",
		"OVO":       "PAYOVO",
	}

	retailCode, exists := retailCodeMap[channelCodeUpper]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000601",
			"response_message": "Invalid Field Format",
		})
		return
	}

	// Get callback URL from config
	linkQuConfig := config.GetLinkQuConfig()
	callbackURL := fmt.Sprintf("%s/payments/linkqu/ewallet", linkQuConfig.CallbackURL)

	// Call LinkQu service
	responseData, err := ewc.linkQuService.CreateEWallet(
		grantID,
		customerID,
		merchant.BusinessName,
		int64(reqBody.Amount),
		expiresAtLinkQu,
		retailCode,
		callbackURL,
		reqBody.PhoneNumber,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Check response
	status, _ := responseData["status"].(string)
	responseCode, _ := responseData["response_code"].(string)
	urlPayment, _ := responseData["url_payment"].(string)

	if status != "SUCCESS" || responseCode != "00" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	var checkoutURL string
	if channelCodeUpper == "OVO" {
		// OVO uses push notification, no URL
		linkQuConfig := config.GetLinkQuConfig()
		checkoutURL = fmt.Sprintf("%s/web/v1/%s", linkQuConfig.BaseURL, grantID)
	} else {
		checkoutURL = urlPayment
		if checkoutURL == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response_code":    "5000601",
				"response_message": "Internal Server Error",
			})
			return
		}
	}

	// Create transaction info
	transactionData := models.TransactionInfoData{
		AppID:         tokenData.AppID,
		PaymentMethod: "EWALLET",
		Amount:        int64(reqBody.Amount),
		Currency:      "IDR",
		SuccessURL:    &reqBody.SuccessURL,
		CancelURL:     &reqBody.FailedURL,
		NotifyURL:     reqBody.NotifyURL,
		GrantID:       grantID,
		OrderID:       reqBody.ReferenceID,
		Token:         apiKey,
		EwalletLink:   &checkoutURL,
		BankEwalletName: &channelCodeUpper,
		ExpiresIn:     &expiresAtStr,
		Status:        "pending",
	}

	err = ewc.transactionRepo.CreateTransaction(transactionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Create merchant transaction
	merchantID := merchantApp.MerchantID
	merchantTxData := models.MerchantPaymentData{
		MerchantID:       &merchantID,
		PaymentMethodID:  &paymentMethod.ID,
		GatewayReference: &grantID,
		OrderNo:          &reqBody.ReferenceID,
		UUID:             &reqBody.ReferenceID,
		FeeBearer:        "Merchant",
		Percentage:       chargePercentage,
		ChargePercentage: amountPercentage,
		ChargeFixed:      chargeFixed,
		Amount:           reqBody.Amount,
		Total:            total,
		Status:           "Pending",
	}

	err = ewc.merchantTxRepo.CreateMerchantTransaction(merchantTxData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Get transaction info ID
	transactionInfo, err := ewc.helper.GetTransactionInfoByGrantID(grantID)
	if err != nil || transactionInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Create callback
	callbackData := models.CallbackData{
		TransactionInfoID: transactionInfo.ID,
		MerchantID:        merchantID,
		NotifyURL:         reqBody.NotifyURL,
		Status:            "Pending",
		RetryCount:        0,
	}

	err = ewc.callbackRepo.CreateCallback(callbackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000601",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Disable token
	ewc.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	linkQuConfig := config.GetLinkQuConfig()
	checkoutURLFinal := fmt.Sprintf("%s/web/v1/%s", linkQuConfig.BaseURL, grantID)

	requestTime := now.Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000600",
		"response_message": "Successful",
		"response_data": gin.H{
			"id":           grantID,
			"reference_id": reqBody.ReferenceID,
			"amount":       reqBody.Amount,
			"payment_data": gin.H{
				"channel_code":  channelCodeUpper,
				"redirect_url": checkoutURL,
			},
			"merchant_url": gin.H{
				"notify_url":  reqBody.NotifyURL,
				"success_url": reqBody.SuccessURL,
				"failed_url":  reqBody.FailedURL,
			},
			"checkout_url": checkoutURLFinal,
			"expires_at":   expiresAtISO,
			"request_time": requestTime,
		},
	})
}

