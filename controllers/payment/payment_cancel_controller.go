package payment

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kytapay/api-v2/helpers"
	"github.com/kytapay/api-v2/repositories"
)

type PaymentCancelController struct {
	db              *sql.DB
	helper          *helpers.PaymentHelper
	tokenRepo       *repositories.TokenRepository
	transactionRepo *repositories.TransactionInfoRepository
	merchantTxRepo  *repositories.MerchantTransactionRepository
}

func NewPaymentCancelController(db *sql.DB) *PaymentCancelController {
	return &PaymentCancelController{
		db:              db,
		helper:          helpers.NewPaymentHelper(db),
		tokenRepo:       repositories.NewTokenRepository(db),
		transactionRepo: repositories.NewTransactionInfoRepository(db),
		merchantTxRepo:  repositories.NewMerchantTransactionRepository(db),
	}
}

// CancelPayment handles POST /v2/payments/cancel
func (pcc *PaymentCancelController) CancelPayment(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	paymentToken := c.GetHeader("X-AUTH-TOKEN")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010800",
			"response_message": "Unauthorized",
		})
		return
	}

	if paymentToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000801",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010800",
			"response_message": "Unauthorized",
		})
		return
	}

	// Verify token and merchant
	_, _, merchant, _, err := pcc.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000801",
			"response_message": "Internal Server Error",
		})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010801",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		PaymentID string `json:"payment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000800",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	// Get transaction info
	query := `SELECT id, grant_id, order_id, status FROM app_transactions_infos WHERE grant_id = ? AND token = ?`
	var appTransaction struct {
		ID      int
		GrantID string
		OrderID string
		Status  string
	}

	err = pcc.db.QueryRow(query, reqBody.PaymentID, paymentToken).Scan(
		&appTransaction.ID,
		&appTransaction.GrantID,
		&appTransaction.OrderID,
		&appTransaction.Status,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040800",
			"response_message": "Transaction Not Found",
		})
		return
	}

	// Get merchant payment
	merchantPayment, err := pcc.helper.GetMerchantPaymentByGatewayRef(reqBody.PaymentID, appTransaction.OrderID)
	if err != nil || merchantPayment == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040800",
			"response_message": "Transaction Not Found",
		})
		return
	}

	// Check if transaction can be cancelled
	if appTransaction.Status != "pending" || merchantPayment.Status != "Pending" {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030802",
			"response_message": "Request Not Permitted",
		})
		return
	}

	// Update transaction status
	updateData := map[string]interface{}{
		"status": "cancel",
	}
	err = pcc.transactionRepo.UpdateTransaction(reqBody.PaymentID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000801",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Update merchant transaction status
	merchantUpdateData := map[string]interface{}{
		"status": "Blocked",
	}
	err = pcc.merchantTxRepo.UpdateMerchantTransaction(reqBody.PaymentID, merchantUpdateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000801",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Disable token
	pcc.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000800",
		"response_message": "Successful",
		"response_data": gin.H{
			"request_time": requestTime,
		},
	})
}

