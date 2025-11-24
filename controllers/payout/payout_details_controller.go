package payout

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kytapay/api-v2/helpers"
	"github.com/kytapay/api-v2/repositories"
)

type PayoutDetailsController struct {
	db              *sql.DB
	helper          *helpers.PaymentHelper
	tokenRepo       *repositories.TokenRepository
	transactionRepo *repositories.TransactionInfoRepository
}

func NewPayoutDetailsController(db *sql.DB) *PayoutDetailsController {
	return &PayoutDetailsController{
		db:              db,
		helper:          helpers.NewPaymentHelper(db),
		tokenRepo:       repositories.NewTokenRepository(db),
		transactionRepo: repositories.NewTransactionInfoRepository(db),
	}
}

// GetDetail handles POST /v2/payouts/details
func (pdc *PayoutDetailsController) GetDetail(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	paymentToken := c.GetHeader("X-AUTH-TOKEN")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011100",
			"response_message": "Unauthorized",
		})
		return
	}

	if paymentToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4001101",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011100",
			"response_message": "Unauthorized",
		})
		return
	}

	// Verify token and merchant
	_, _, merchant, _, err := pdc.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001101",
			"response_message": "Internal Server Error",
		})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011101",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		PayoutID string `json:"payout_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4001100",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	// Log request
	reqBodyJSON, _ := json.Marshal(reqBody)
	log.Printf("[PAYOUT DETAILS] Request received: payoutID=%s", reqBody.PayoutID)
	log.Printf("[PAYOUT DETAILS] Full request body: %s", string(reqBodyJSON))

	// Get transaction info
	query := `SELECT id, app_id, payment_method, amount, currency, success_url, cancel_url, notify_url, 
		grant_id, order_id, token, qris_string, bank_number, ewallet_link, bank_ewallet_name, 
		expires_in, version, status, created_at, updated_at 
		FROM app_transactions_infos 
		WHERE grant_id = ? AND token = ?`

	var detailData struct {
		ID              int
		AppID           int
		PaymentMethod   string
		Amount          int64
		Currency        string
		SuccessURL      *string
		CancelURL       *string
		NotifyURL       string
		GrantID         string
		OrderID         string
		Token           string
		QrisString      *string
		BankNumber      *string
		EwalletLink     *string
		BankEwalletName *string
		ExpiresIn       *string
		Version         *int
		Status          string
		CreatedAt       *time.Time
		UpdatedAt       *time.Time
	}

	err = pdc.db.QueryRow(query, reqBody.PayoutID, paymentToken).Scan(
		&detailData.ID,
		&detailData.AppID,
		&detailData.PaymentMethod,
		&detailData.Amount,
		&detailData.Currency,
		&detailData.SuccessURL,
		&detailData.CancelURL,
		&detailData.NotifyURL,
		&detailData.GrantID,
		&detailData.OrderID,
		&detailData.Token,
		&detailData.QrisString,
		&detailData.BankNumber,
		&detailData.EwalletLink,
		&detailData.BankEwalletName,
		&detailData.ExpiresIn,
		&detailData.Version,
		&detailData.Status,
		&detailData.CreatedAt,
		&detailData.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4041100",
			"response_message": "Transaction Not Found",
		})
		return
	}

	// Disable token
	pdc.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	// Format dates
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	var createdAtISO string

	if detailData.CreatedAt != nil {
		createdAtISO = detailData.CreatedAt.In(loc).Format(time.RFC3339)
	}

	requestTime := time.Now().In(loc).Format(time.RFC3339)

	// Format status (capitalize first letter)
	status := strings.ToLower(detailData.Status)
	if len(status) > 0 {
		status = strings.ToUpper(string(status[0])) + status[1:]
	}

	// Build payout_data
	accountNumber := ""
	accountName := ""
	if detailData.BankNumber != nil {
		accountNumber = *detailData.BankNumber
	}
	if detailData.BankEwalletName != nil {
		accountName = *detailData.BankEwalletName
	}

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2001100",
		"response_message": "Successful",
		"response_data": gin.H{
			"id":           detailData.GrantID,
			"reference_id": detailData.OrderID,
			"amount":       detailData.Amount,
			"status":       status,
			"payout_data": gin.H{
				"code":          detailData.PaymentMethod,
				"account_number": accountNumber,
				"account_name":  accountName,
			},
			"merchant_url": gin.H{
				"notify_url": detailData.NotifyURL,
			},
			"created_at":   createdAtISO,
			"request_time": requestTime,
		},
	})
}

