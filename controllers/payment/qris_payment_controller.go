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

type QRISPaymentController struct {
	db              *sql.DB
	helper          *helpers.PaymentHelper
	linkQuService   *services.LinkQuService
	tokenRepo       *repositories.TokenRepository
	settingRepo     *repositories.SettingRepository
	transactionRepo *repositories.TransactionInfoRepository
	merchantTxRepo  *repositories.MerchantTransactionRepository
	callbackRepo    *repositories.CallbackRepository
}

func NewQRISPaymentController(db *sql.DB) *QRISPaymentController {
	return &QRISPaymentController{
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

// CreateQRIS handles POST /v2/payments/create/qris
func (qpc *QRISPaymentController) CreateQRIS(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010500",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010500",
			"response_message": "Unauthorized",
		})
		return
	}

	// Check maintenance
	isMaintenancePayment, err := qpc.settingRepo.GetMaintenancePayment()
	if err == nil && isMaintenancePayment {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030500",
			"response_message": "The Payment service is currently under maintenance. Please try again later.",
		})
		return
	}

	isMaintenanceQRIS, err := qpc.settingRepo.GetMaintenanceQris()
	if err == nil && isMaintenanceQRIS {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5030501",
			"response_message": "The QRIS payment service is currently unavailable. Please use other payment methods.",
		})
		return
	}

	// Verify token and merchant
	tokenData, merchantApp, merchant, user, err := qpc.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
			"response_message": "Internal Server Error",
		})
		return
	}

	if tokenData == nil || merchant == nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010501",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		ReferenceID string  `json:"reference_id" binding:"required"`
		Amount       float64 `json:"amount" binding:"required"`
		NotifyURL    string  `json:"notify_url" binding:"required"`
		SuccessURL   string  `json:"success_url" binding:"required"`
		FailedURL    string  `json:"failed_url" binding:"required"`
		ExpiresTime  *int    `json:"expires_time"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000500",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	if reqBody.Amount < 500 || reqBody.Amount > 10000000 {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030502",
			"response_message": "Exceeds Transaction Amount Limit",
		})
		return
	}

	// Get payment method
	paymentMethod, err := qpc.helper.GetPaymentMethodByCode("QR")
	if err != nil || paymentMethod == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040500",
			"response_message": "No methods available",
		})
		return
	}

	// Get fees limit
	feeLimit, err := qpc.helper.GetFeesLimit(10, paymentMethod.ID)
	if err != nil || feeLimit == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000500",
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
	expiresAtLinkQu := expiresAt.Format("20060102150405")

	// Generate customer ID (unique)
	customerID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Get callback URL from config
	linkQuConfig := config.GetLinkQuConfig()
	callbackURL := fmt.Sprintf("%s/payments/linkqu/qris", linkQuConfig.CallbackURL)

	// Call LinkQu service
	responseData, err := qpc.linkQuService.CreateQRIS(
		grantID,
		customerID,
		merchant.BusinessName,
		int64(reqBody.Amount),
		expiresAtLinkQu,
		callbackURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Check response
	status, _ := responseData["status"].(string)
	responseCode, _ := responseData["response_code"].(string)
	qrString, _ := responseData["qris_text"].(string)

	if status != "SUCCESS" || responseCode != "00" || qrString == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Create transaction info
	transactionData := models.TransactionInfoData{
		AppID:         tokenData.AppID,
		PaymentMethod: paymentMethod.Name,
		Amount:        int64(reqBody.Amount),
		Currency:      "IDR",
		SuccessURL:    &reqBody.SuccessURL,
		CancelURL:     &reqBody.FailedURL,
		NotifyURL:     reqBody.NotifyURL,
		GrantID:       grantID,
		OrderID:       reqBody.ReferenceID,
		Token:         apiKey,
		QrisString:    &qrString,
		ExpiresIn:     &expiresAtStr,
		Status:        "pending",
	}

	err = qpc.transactionRepo.CreateTransaction(transactionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
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

	err = qpc.merchantTxRepo.CreateMerchantTransaction(merchantTxData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Get transaction info ID
	transactionInfo, err := qpc.helper.GetTransactionInfoByGrantID(grantID)
	if err != nil || transactionInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
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

	err = qpc.callbackRepo.CreateCallback(callbackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000501",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Disable token
	qpc.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	linkQuConfig := config.GetLinkQuConfig()
	checkoutURL := fmt.Sprintf("%s/web/v1/%s", linkQuConfig.BaseURL, grantID)

	requestTime := now.Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000500",
		"response_message": "Successful",
		"response_data": gin.H{
			"id":           grantID,
			"reference_id": reqBody.ReferenceID,
			"amount":       reqBody.Amount,
			"payment_data": gin.H{
				"qr_string": qrString,
			},
			"merchant_url": gin.H{
				"notify_url":  reqBody.NotifyURL,
				"success_url": reqBody.SuccessURL,
				"failed_url":  reqBody.FailedURL,
			},
			"checkout_url": checkoutURL,
			"expires_at":   expiresAt.Format(time.RFC3339),
			"request_time": requestTime,
		},
	})
}

