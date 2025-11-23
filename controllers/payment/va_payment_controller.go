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

type VAPaymentController struct {
	db              *sql.DB
	helper          *helpers.PaymentHelper
	pakaiLinkService *services.PakaiLinkService
	tokenRepo       *repositories.TokenRepository
	settingRepo     *repositories.SettingRepository
	transactionRepo *repositories.TransactionInfoRepository
	merchantTxRepo  *repositories.MerchantTransactionRepository
	callbackRepo    *repositories.CallbackRepository
}

func NewVAPaymentController(db *sql.DB) *VAPaymentController {
	return &VAPaymentController{
		db:              db,
		helper:          helpers.NewPaymentHelper(db),
		pakaiLinkService: services.NewPakaiLinkService(),
		tokenRepo:       repositories.NewTokenRepository(db),
		settingRepo:     repositories.NewSettingRepository(db),
		transactionRepo: repositories.NewTransactionInfoRepository(db),
		merchantTxRepo:  repositories.NewMerchantTransactionRepository(db),
		callbackRepo:    repositories.NewCallbackRepository(db),
	}
}

var supportedBanks = []string{"BCA", "BRI", "MANDIRI", "BNI", "DANAMON", "PERMATA", "MAYBANK", "PANIN", "CIMB", "OCBC", "MUAMALAT", "SINARMAS", "BSI", "BNC"}

// CreateVA handles POST /v2/payments/create/va
func (vac *VAPaymentController) CreateVA(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010400",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010400",
			"response_message": "Unauthorized",
		})
		return
	}

	// Check maintenance
	isMaintenancePayment, err := vac.settingRepo.GetMaintenancePayment()
	if err == nil && isMaintenancePayment {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030400",
			"response_message": "The Payment service is currently under maintenance. Please try again later.",
		})
		return
	}

	isMaintenanceVA, err := vac.settingRepo.GetMaintenanceVa()
	if err == nil && isMaintenanceVA {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030401",
			"response_message": "The Virtual Account payment service is currently unavailable. Please use other payment methods.",
		})
		return
	}

	// Verify token and merchant
	tokenData, merchantApp, merchant, user, err := vac.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	if tokenData == nil || merchant == nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010401",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		ReferenceID string  `json:"reference_id" binding:"required"`
		Amount       float64 `json:"amount" binding:"required"`
		BankCode     string  `json:"bank_code" binding:"required"`
		NotifyURL    string  `json:"notify_url" binding:"required"`
		SuccessURL   string  `json:"success_url" binding:"required"`
		FailedURL    string  `json:"failed_url" binding:"required"`
		ExpiresTime  *int    `json:"expires_time"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000400",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	if reqBody.Amount < 10000 || reqBody.Amount > 500000000 {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030402",
			"response_message": "Exceeds Transaction Amount Limit",
		})
		return
	}

	bankCodeUpper := strings.ToUpper(reqBody.BankCode)
	isSupported := false
	for _, bank := range supportedBanks {
		if bank == bankCodeUpper {
			isSupported = true
			break
		}
	}

	if !isSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000401",
			"response_message": "Invalid Field Format",
		})
		return
	}

	// Get payment method
	paymentMethod, err := vac.helper.GetPaymentMethodByCode(bankCodeUpper)
	if err != nil || paymentMethod == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040400",
			"response_message": "No methods available",
		})
		return
	}

	// Get fees limit
	feeLimit, err := vac.helper.GetFeesLimit(10, paymentMethod.ID) // transaction_type_id = 10
	if err != nil || feeLimit == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000400",
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
	expiresTime := 86400 // default 24 hours
	if reqBody.ExpiresTime != nil {
		expiresTime = *reqBody.ExpiresTime
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	now := time.Now().In(loc)
	expiresAt := now.Add(time.Duration(expiresTime) * time.Second)
	expiresAtStr := expiresAt.Format("2006-01-02 15:04:05")
	expiresAtISO := expiresAt.Format(time.RFC3339)

	// Convert bank code to PakaiLink format
	bankCode := helpers.GetBankCode(bankCodeUpper)

	// Generate customer number (unique)
	customerNo := fmt.Sprintf("%d", time.Now().UnixNano())

	// Call PakaiLink service
	// Get callback URL from config
	pakaiLinkConfig := config.GetPakaiLinkConfig()
	callbackURL := fmt.Sprintf("%s/payments/pakailink/va", pakaiLinkConfig.CallbackURL)

	responseData, err := vac.pakaiLinkService.CreateVA(
		grantID,
		customerNo,
		fmt.Sprintf("%s", merchant.BusinessName),
		reqBody.Amount,
		expiresAtISO,
		bankCode,
		callbackURL,
	)
	if err != nil {
		// Log error for debugging
		fmt.Printf("[ERROR] PakaiLink CreateVA failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Check response code
	responseCode, _ := responseData["responseCode"].(string)
	responseMessage, _ := responseData["responseMessage"].(string)
	if responseCode != "2002700" {
		// Log error for debugging
		fmt.Printf("[ERROR] PakaiLink CreateVA response error: code=%s, message=%s, full_response=%+v\n", responseCode, responseMessage, responseData)
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Extract virtual account data
	virtualAccountData, ok := responseData["virtualAccountData"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	virtualAccountNo, _ := virtualAccountData["virtualAccountNo"].(string)
	if virtualAccountNo == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	accountNumber := virtualAccountNo
	accountName := fmt.Sprintf("%s", merchant.BusinessName)
	bankCodeResp := bankCodeUpper

	// Create transaction info
	transactionData := models.TransactionInfoData{
		AppID:         tokenData.AppID,
		PaymentMethod: "VA",
		Amount:        int64(reqBody.Amount),
		Currency:      "IDR",
		SuccessURL:    &reqBody.SuccessURL,
		CancelURL:     &reqBody.FailedURL,
		NotifyURL:     reqBody.NotifyURL,
		GrantID:       grantID,
		OrderID:       reqBody.ReferenceID,
		Token:         apiKey,
		BankNumber:    &accountNumber,
		BankEwalletName: &bankCodeResp,
		ExpiresIn:     &expiresAtStr,
		Status:        "pending",
	}

	err = vac.transactionRepo.CreateTransaction(transactionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
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
		ChargeFixed:      chargeFixed,
		Amount:           reqBody.Amount,
		Total:            total,
		Status:           "Pending",
	}

	err = vac.merchantTxRepo.CreateMerchantTransaction(merchantTxData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Get transaction info ID
	transactionInfo, err := vac.helper.GetTransactionInfoByGrantID(grantID)
	if err != nil || transactionInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
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

	err = vac.callbackRepo.CreateCallback(callbackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000401",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Disable token
	vac.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	// Get checkout URL from config
	pakaiLinkConfigFinal := config.GetPakaiLinkConfig()
	checkoutURL := fmt.Sprintf("%s/%s", pakaiLinkConfigFinal.CheckoutURL, grantID)

	requestTime := now.Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000400",
		"response_message": "Successful",
		"response_data": gin.H{
			"id":           grantID,
			"reference_id": reqBody.ReferenceID,
			"amount":       reqBody.Amount,
			"payment_data": gin.H{
				"bank_code":      bankCodeResp,
				"account_number": accountNumber,
				"account_name":   accountName,
			},
			"merchant_url": gin.H{
				"notify_url":  reqBody.NotifyURL,
				"success_url": reqBody.SuccessURL,
				"failed_url":  reqBody.FailedURL,
			},
			"checkout_url": checkoutURL,
			"expires_at":   expiresAtISO,
			"request_time": requestTime,
		},
	})
}

