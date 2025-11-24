package payout

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kytapay/api-v2/helpers"
	"github.com/kytapay/api-v2/models"
	"github.com/kytapay/api-v2/repositories"
	"github.com/kytapay/api-v2/services"
)

type PayoutInquiryController struct {
	db                *sql.DB
	helper            *helpers.PaymentHelper
	tokenRepo         *repositories.TokenRepository
	walletRepo        *repositories.WalletRepository
	transactionRepo   *repositories.TransactionInfoRepository
	callbackRepo      *repositories.CallbackRepository
	nameCheckRepo     *repositories.APINameCheckRepository
	ilumaService      *services.IlumaService
}

func NewPayoutInquiryController(db *sql.DB) *PayoutInquiryController {
	return &PayoutInquiryController{
		db:                db,
		helper:            helpers.NewPaymentHelper(db),
		tokenRepo:         repositories.NewTokenRepository(db),
		walletRepo:        repositories.NewWalletRepository(db),
		transactionRepo:   repositories.NewTransactionInfoRepository(db),
		callbackRepo:      repositories.NewCallbackRepository(db),
		nameCheckRepo:     repositories.NewAPINameCheckRepository(db),
		ilumaService:      services.NewIlumaService(),
	}
}

// ProcessInquiry handles POST /v2/payouts/inquiry
func (pic *PayoutInquiryController) ProcessInquiry(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010900",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010900",
			"response_message": "Unauthorized",
		})
		return
	}

	// Verify token and merchant
	tokenData, merchantApp, merchant, user, err := pic.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000901",
			"response_message": "Internal Server Error",
		})
		return
	}

	if tokenData == nil || merchant == nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4010901",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		ReferenceID string `json:"reference_id" binding:"required"`
		Destination struct {
			Code          string `json:"code" binding:"required"`
			AccountNumber string `json:"account_number" binding:"required"`
		} `json:"destination" binding:"required"`
		NotifyURL string `json:"notify_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000900",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	// Log request
	reqBodyJSON, _ := json.Marshal(reqBody)
	log.Printf("[PAYOUT INQUIRY] Request received: referenceID=%s, destinationCode=%s, accountNumber=%s", 
		reqBody.ReferenceID, reqBody.Destination.Code, reqBody.Destination.AccountNumber)
	log.Printf("[PAYOUT INQUIRY] Full request body: %s", string(reqBodyJSON))

	if reqBody.Destination.Code == "" || reqBody.Destination.AccountNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4000901",
			"response_message": "Invalid Field Format",
		})
		return
	}

	// Get payment method
	paymentMethod, err := pic.helper.GetPaymentMethodByCode(strings.ToUpper(reqBody.Destination.Code))
	if err != nil || paymentMethod == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4040900",
			"response_message": "No methods available",
		})
		return
	}

	// Check wallet balance (min 150)
	wallet, err := pic.walletRepo.GetUserWallet(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000901",
			"response_message": "Internal Server Error",
		})
		return
	}

	if wallet.Balance < 150 {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4030902",
			"response_message": "Insufficient Funds",
		})
		return
	}

	// Generate reference ID
	referenceID := uuid.New().String()

	// Call Iluma API
	responseData, err := pic.ilumaService.ValidateBankAccount(
		reqBody.Destination.AccountNumber,
		reqBody.Destination.Code,
		referenceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5000901",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Parse response
	status, _ := responseData["status"].(string)
	ilumaID, _ := responseData["id"].(string)
	rawResponse, _ := json.Marshal(responseData)

	// Prepare insert data
	insertData := models.APINameCheck{
		MerchantID:    merchantApp.MerchantID,
		ReferenceID:   referenceID,
		OrderID:       reqBody.ReferenceID,
		Token:         apiKey,
		AccountNumber: reqBody.Destination.AccountNumber,
		BankCode:      reqBody.Destination.Code,
		IlumaID:       ilumaID,
		Status:        status,
		NotifyURL:     reqBody.NotifyURL,
		RawResponse:   string(rawResponse),
	}

	// Prepare transaction data
	transactionData := models.TransactionInfoData{
		AppID:         tokenData.AppID,
		PaymentMethod: paymentMethod.Code,
		Amount:        150,
		Currency:      "IDR",
		NotifyURL:     reqBody.NotifyURL,
		GrantID:       referenceID,
		OrderID:       reqBody.ReferenceID,
		Token:         apiKey,
		BankNumber:    &reqBody.Destination.AccountNumber,
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	now := time.Now().In(loc)
	requestTime := now.Format(time.RFC3339)

	// Handle based on status
	if status == "COMPLETED" {
		result, ok := responseData["result"].(map[string]interface{})
		if ok {
			isFound, _ := result["is_found"].(bool)
			if isFound {
				// Account found
				accountName, _ := result["account_holder_name"].(string)
				isVirtualAccount, _ := result["is_virtual_account"].(bool)

				insertData.AccountName = &accountName
				isFoundInt := 1
				insertData.IsFound = &isFoundInt
				if isVirtualAccount {
					isVA := 1
					insertData.IsVirtualAccount = &isVA
				} else {
					isVA := 0
					insertData.IsVirtualAccount = &isVA
				}

				transactionData.Status = "success"

				// Deduct balance
				err = pic.walletRepo.DeductUserBalance(user.ID, 150)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				// Insert data
				err = pic.nameCheckRepo.CreateAPINameCheck(insertData)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				err = pic.transactionRepo.CreateTransaction(transactionData)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				// Get transaction info ID
				transactionInfo, err := pic.helper.GetTransactionInfoByGrantID(referenceID)
				if err != nil || transactionInfo == nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				// Create callback
				callbackData := models.CallbackData{
					TransactionInfoID: transactionInfo.ID,
					MerchantID:        merchantApp.MerchantID,
					NotifyURL:         reqBody.NotifyURL,
					Status:            "Pending",
					RetryCount:        0,
				}

				err = pic.callbackRepo.CreateCallback(callbackData)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				// Disable token
				pic.tokenRepo.DisableToken(apiKey)
				c.Header("X-AUTH-TOKEN", apiKey)

				c.JSON(http.StatusOK, gin.H{
					"response_code":    "2000900",
					"response_message": "Successful",
					"response_data": gin.H{
						"id":           referenceID,
						"reference_id": reqBody.ReferenceID,
						"status":       "Success",
						"inquiry_data": gin.H{
							"code":          reqBody.Destination.Code,
							"account_number": reqBody.Destination.AccountNumber,
							"account_name":  accountName,
						},
						"merchant_url": gin.H{
							"notify_url": reqBody.NotifyURL,
						},
						"request_time": requestTime,
					},
				})
				return
			} else {
				// Account not found
				isFoundInt := 0
				insertData.IsFound = &isFoundInt
				transactionData.Status = "cancel"

				// Insert data
				err = pic.nameCheckRepo.CreateAPINameCheck(insertData)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				err = pic.transactionRepo.CreateTransaction(transactionData)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"response_code":    "5000901",
						"response_message": "Internal Server Error",
					})
					return
				}

				// Disable token
				pic.tokenRepo.DisableToken(apiKey)
				c.Header("X-AUTH-TOKEN", apiKey)

				c.JSON(http.StatusNotFound, gin.H{
					"response_code":    "4040901",
					"response_message": "Recipient not found",
				})
				return
			}
		}
	} else {
		// Status is PENDING
		transactionData.Status = "pending"

		// Insert data
		err = pic.nameCheckRepo.CreateAPINameCheck(insertData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response_code":    "5000901",
				"response_message": "Internal Server Error",
			})
			return
		}

		err = pic.transactionRepo.CreateTransaction(transactionData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response_code":    "5000901",
				"response_message": "Internal Server Error",
			})
			return
		}

		// Get transaction info ID
		transactionInfo, err := pic.helper.GetTransactionInfoByGrantID(referenceID)
		if err != nil || transactionInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response_code":    "5000901",
				"response_message": "Internal Server Error",
			})
			return
		}

		// Create callback
		callbackData := models.CallbackData{
			TransactionInfoID: transactionInfo.ID,
			MerchantID:        merchantApp.MerchantID,
			NotifyURL:         reqBody.NotifyURL,
			Status:            "Pending",
			RetryCount:        0,
		}

		err = pic.callbackRepo.CreateCallback(callbackData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response_code":    "5000901",
				"response_message": "Internal Server Error",
			})
			return
		}

		// Disable token
		pic.tokenRepo.DisableToken(apiKey)
		c.Header("X-AUTH-TOKEN", apiKey)

		c.JSON(http.StatusAccepted, gin.H{
			"response_code":    "2020900",
			"response_message": "Successful",
			"response_data": gin.H{
				"id":           referenceID,
				"reference_id": reqBody.ReferenceID,
				"status":       "Pending",
				"inquiry_data": gin.H{
					"code":          reqBody.Destination.Code,
					"account_number": reqBody.Destination.AccountNumber,
				},
				"merchant_url": gin.H{
					"notify_url": reqBody.NotifyURL,
				},
				"request_time": requestTime,
			},
		})
		return
	}
}

