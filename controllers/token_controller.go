package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kytapay/api-v2/models"
	"github.com/kytapay/api-v2/repositories"
)

type TokenController struct {
	db                    *sql.DB
	merchantRepo          *repositories.MerchantRepository
	merchantDetailRepo    *repositories.MerchantDetailRepository
	tokenRepo             *repositories.TokenRepository
}

func NewTokenController(db *sql.DB) *TokenController {
	return &TokenController{
		db:                 db,
		merchantRepo:       repositories.NewMerchantRepository(db),
		merchantDetailRepo: repositories.NewMerchantDetailRepository(db),
		tokenRepo:          repositories.NewTokenRepository(db),
	}
}

// VerifyClientCredentials handles POST /v2/access-token
func (tc *TokenController) VerifyClientCredentials(c *gin.Context) {
	// Ambil credentials dari Authorization header (Basic Auth)
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Basic ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010100",
			"response_message": "Unauthorized",
		})
		return
	}

	// Decode Base64 credentials
	base64Credentials := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(base64Credentials)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010100",
			"response_message": "Unauthorized",
		})
		return
	}

	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010101",
			"response_message": "Client ID and Client Secret are required in Basic Auth",
		})
		return
	}

	clientID := credentials[0]
	clientSecret := credentials[1]

	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010101",
			"response_message": "Client ID and Client Secret are required in Basic Auth",
		})
		return
	}

	// Validasi grant_type dari request body
	var reqBody struct {
		GrantType string `json:"grant_type"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000100",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	if reqBody.GrantType != "client_credentials" {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000101",
			"response_message": "Invalid Field Format",
		})
		return
	}

	// Cek merchant berdasarkan client_id dan client_secret
	merchantApp, err := tc.merchantRepo.GetMerchantAccess(clientID, clientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000101",
			"response_message": "Internal Server Error",
		})
		return
	}

	if merchantApp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010102",
			"response_message": "Can not verify the client. Please check Client ID and Client Secret",
		})
		return
	}

	// Cek status merchant
	merchant, err := tc.merchantDetailRepo.GetMerchantByID(merchantApp.MerchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000100",
			"response_message": "General Error",
		})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000100",
			"response_message": "General Error",
		})
		return
	}

	// Buat Access Token
	accessToken, err := tc.createAccessToken(merchantApp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000101",
			"response_message": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, accessToken)
}

// createAccessToken creates an access token and saves it to database
func (tc *TokenController) createAccessToken(app *models.MerchantApp) (gin.H, error) {
	// Generate random token (26 karakter = 13 bytes hex)
	tokenBytes := make([]byte, 13)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return nil, err
	}
	token := make([]byte, 26)
	for i, b := range tokenBytes {
		token[i*2] = "0123456789abcdef"[b>>4]
		token[i*2+1] = "0123456789abcdef"[b&0x0f]
	}

	// Expire dalam 1 jam (Unix timestamp)
	expiresIn := time.Now().Unix() + 3600

	// Save token to database
	tokenData := models.TokenData{
		AppID:     app.ID,
		Token:     string(token),
		ExpiresIn: strconv.FormatInt(expiresIn, 10),
		Status:    2,
	}

	err = tc.tokenRepo.CreateToken(tokenData)
	if err != nil {
		return nil, err
	}

	// Get current time in Asia/Jakarta timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	return gin.H{
		"response_code":    "2000100",
		"response_message": "Successful",
		"response_data": gin.H{
			"access_token": string(token),
			"token_type":   "Bearer",
			"expires_in":   900, // 15 minutes in seconds
			"request_time": requestTime,
		},
	}, nil
}

