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

type MerchantInformationController struct {
	db                 *sql.DB
	tokenRepo          *repositories.TokenRepository
	merchantRepo       *repositories.MerchantRepository
	merchantDetailRepo *repositories.MerchantDetailRepository
	userDetailRepo     *repositories.UserDetailRepository
	walletRepo         *repositories.WalletRepository
}

func NewMerchantInformationController(db *sql.DB) *MerchantInformationController {
	return &MerchantInformationController{
		db:                 db,
		tokenRepo:          repositories.NewTokenRepository(db),
		merchantRepo:       repositories.NewMerchantRepository(db),
		merchantDetailRepo: repositories.NewMerchantDetailRepository(db),
		userDetailRepo:     repositories.NewUserDetailRepository(db),
		walletRepo:         repositories.NewWalletRepository(db),
	}
}

// Informations handles GET /v2/merchants/information
func (mic *MerchantInformationController) Informations(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	// Validasi format Authorization header
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010300",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010300",
			"response_message": "Unauthorized",
		})
		return
	}

	// Verifikasi token
	tokenData, err := mic.tokenRepo.GetTokenByToken(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000301",
			"response_message": "Internal Server Error",
		})
		return
	}

	currentTime := time.Now().Unix()
	if tokenData == nil || tokenData.Status == 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010301",
			"response_message": "Invalid token",
		})
		return
	}

	// Parse expires_in as Unix timestamp
	expiresIn, err := strconv.ParseInt(tokenData.ExpiresIn, 10, 64)
	if err != nil || expiresIn < currentTime {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010301",
			"response_message": "Invalid token",
		})
		return
	}

	// Verifikasi merchant app
	merchantApp, err := mic.merchantRepo.GetMerchantAppByID(tokenData.AppID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010301",
			"response_message": "Invalid token",
		})
		return
	}

	if merchantApp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010301",
			"response_message": "Invalid token",
		})
		return
	}

	// Verifikasi merchant
	merchant, err := mic.merchantDetailRepo.GetMerchantByID(merchantApp.MerchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000300",
			"response_message": "General error",
		})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000300",
			"response_message": "General error",
		})
		return
	}

	if merchant.Status != "Approved" {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030300",
			"response_message": "Merchant status has not been verified",
		})
		return
	}

	// Verifikasi user
	user, err := mic.userDetailRepo.GetUserByID(*merchant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000300",
			"response_message": "General error",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000300",
			"response_message": "General error",
		})
		return
	}

	if user.Status != "Active" {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030301",
			"response_message": "Merchant account is inactive",
		})
		return
	}

	// Get wallet
	wallet, err := mic.walletRepo.GetUserWallet(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000300",
			"response_message": "General error",
		})
		return
	}

	// Disable token setelah digunakan
	err = mic.tokenRepo.DisableToken(apiKey)
	if err != nil {
		// Log error but continue
	}

	c.Header("X-AUTH-TOKEN", apiKey)

	// Build user name
	userName := ""
	if user.FirstName != nil {
		userName = *user.FirstName
	}
	if user.LastName != nil {
		if userName != "" {
			userName += " "
		}
		userName += *user.LastName
	}

	// Get current time in Asia/Jakarta timezone
	loc, _ := time.LoadLocation("Asia/Jakarta")
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2000300",
		"response_message": "Successful",
		"response_data": gin.H{
			"user_information": gin.H{
				"name":    userName,
				"email":   user.Email,
				"phone":   user.FormattedPhone,
				"balance": wallet.Balance,
			},
			"merchant_information": gin.H{
				"id":           merchant.MerchantUUID,
				"business_name": merchant.BusinessName,
				"url":          merchant.SiteURL,
			},
			"request_time": requestTime,
		},
	})
}

