package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kytapay/api-v2/repositories"
)

type PaymentChannelsController struct {
	db                    *sql.DB
	tokenRepo             *repositories.TokenRepository
	merchantRepo          *repositories.MerchantRepository
	merchantDetailRepo    *repositories.MerchantDetailRepository
	userDetailRepo        *repositories.UserDetailRepository
	feesLimitRepo         *repositories.FeesLimitRepository
	paymentMethodRepo     *repositories.PaymentMethodRepository
}

func NewPaymentChannelsController(db *sql.DB) *PaymentChannelsController {
	return &PaymentChannelsController{
		db:                 db,
		tokenRepo:          repositories.NewTokenRepository(db),
		merchantRepo:       repositories.NewMerchantRepository(db),
		merchantDetailRepo: repositories.NewMerchantDetailRepository(db),
		userDetailRepo:     repositories.NewUserDetailRepository(db),
		feesLimitRepo:      repositories.NewFeesLimitRepository(db),
		paymentMethodRepo:  repositories.NewPaymentMethodRepository(db),
	}
}

// PaymentChannels handles GET /v2/merchants/payment-channels
func (pcc *PaymentChannelsController) PaymentChannels(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	// Validasi format Authorization header
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010200",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010200",
			"response_message": "Unauthorized",
		})
		return
	}

	// Verifikasi token
	tokenData, err := pcc.tokenRepo.GetTokenByToken(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000201",
			"response_message": "Internal Server Error",
		})
		return
	}

	currentTime := time.Now().Unix()
	if tokenData == nil || tokenData.Status == 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010201",
			"response_message": "Invalid token",
		})
		return
	}

	// Parse expires_in as Unix timestamp
	expiresIn, err := strconv.ParseInt(tokenData.ExpiresIn, 10, 64)
	if err != nil || expiresIn < currentTime {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010201",
			"response_message": "Invalid token",
		})
		return
	}

	// Verifikasi merchant app
	merchantApp, err := pcc.merchantRepo.GetMerchantAppByID(tokenData.AppID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010201",
			"response_message": "Invalid token",
		})
		return
	}

	if merchantApp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010201",
			"response_message": "Invalid token",
		})
		return
	}

	// Verifikasi merchant
	merchant, err := pcc.merchantDetailRepo.GetMerchantByID(merchantApp.MerchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000200",
			"response_message": "General error",
		})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000200",
			"response_message": "General error",
		})
		return
	}

	if merchant.Status != "Approved" {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030200",
			"response_message": "Merchant status has not been verified",
		})
		return
	}

	// Verifikasi user
	user, err := pcc.userDetailRepo.GetUserByID(*merchant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000200",
			"response_message": "General error",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000200",
			"response_message": "General error",
		})
		return
	}

	if user.Status != "Active" {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030201",
			"response_message": "Merchant account is inactive",
		})
		return
	}

	// Ambil semua fee_limits dengan transaction_type_id = 10 (payment)
	feeLimits, err := pcc.feesLimitRepo.GetFeesLimitsByTransactionTypeID(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000201",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Ambil semua fee_limits dengan transaction_type_id = 9 (payout)
	feePayouts, err := pcc.feesLimitRepo.GetFeesLimitsByTransactionTypeID(9)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000201",
			"response_message": "Internal Server Error",
		})
		return
	}

	if len(feePayouts) == 0 && len(feeLimits) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040200",
			"response_message": "No methods available",
		})
		return
	}

	// Build payment methods array
	paymentMethods := []gin.H{}
	for _, fee := range feeLimits {
		if fee.PaymentMethodID == nil {
			continue
		}
		paymentMethod, err := pcc.paymentMethodRepo.GetPaymentMethodByID(*fee.PaymentMethodID)
		if err != nil || paymentMethod == nil {
			continue
		}

		settlementTime := "Realtime"
		if fee.ProcessingTime != "" && fee.ProcessingTime != "0" {
			processingTimeInt, err := strconv.Atoi(fee.ProcessingTime)
			if err == nil && processingTimeInt != 0 {
				settlementTime = "T+" + strconv.Itoa(processingTimeInt)
			} else {
				settlementTime = "Realtime"
			}
		}

		paymentMethods = append(paymentMethods, gin.H{
			"name":            paymentMethod.Name,
			"code":            paymentMethod.Code,
			"type":            paymentMethod.Type,
			"fee": gin.H{
				"fixed":   fee.ChargeFixed,
				"percent": fee.ChargePercentage,
			},
			"minimum":         fee.MinLimit,
			"maximum":         fee.MaxLimit,
			"settlement_time": settlementTime,
			"status":          paymentMethod.Status,
		})
	}

	// Build payout methods array
	payoutMethods := []gin.H{}
	for _, fees := range feePayouts {
		if fees.PaymentMethodID == nil {
			continue
		}
		payoutMethod, err := pcc.paymentMethodRepo.GetPaymentMethodByID(*fees.PaymentMethodID)
		if err != nil || payoutMethod == nil {
			continue
		}

		settlementTime := "Realtime"
		if fees.ProcessingTime != "" && fees.ProcessingTime != "0" {
			processingTimeInt, err := strconv.Atoi(fees.ProcessingTime)
			if err == nil && processingTimeInt != 0 {
				settlementTime = "T+" + strconv.Itoa(processingTimeInt)
			} else {
				settlementTime = "Realtime"
			}
		}

		payoutMethods = append(payoutMethods, gin.H{
			"name":            payoutMethod.Name,
			"code":            payoutMethod.Code,
			"type":            payoutMethod.Type,
			"fee": gin.H{
				"fixed":   fees.ChargeFixed,
				"percent": fees.ChargePercentage,
			},
			"minimum":         fees.MinLimit,
			"maximum":         fees.MaxLimit,
			"settlement_time": settlementTime,
			"status":          payoutMethod.Status,
		})
	}

	// Disable token setelah digunakan
	err = pcc.tokenRepo.DisableToken(apiKey)
	if err != nil {
		// Log error but continue
	}

	c.Header("X-AUTH-TOKEN", apiKey)

	// Get current time in Asia/Jakarta timezone
	loc, _ := time.LoadLocation("Asia/Jakarta")
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000200",
		"response_message": "Successful",
		"response_data": gin.H{
			"channels": gin.H{
				"payment": paymentMethods,
				"payout":  payoutMethods,
			},
			"request_time": requestTime,
		},
	})
}

