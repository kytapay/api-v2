package payout

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

type PayoutTransfersController struct {
	db                 *sql.DB
	helper             *helpers.PaymentHelper
	tokenRepo          *repositories.TokenRepository
	settingRepo        *repositories.SettingRepository
	walletRepo         *repositories.WalletRepository
	feesExpressRepo    *repositories.FeesExpressRepository
	transactionRepo    *repositories.TransactionInfoRepository
	transactionsRepo   *repositories.TransactionsRepository
	merchantPayoutRepo *repositories.MerchantPayoutRepository
	callbackRepo       *repositories.CallbackRepository
	pakaiLinkService   *services.PakaiLinkService
	linkQuService      *services.LinkQuService
}

func NewPayoutTransfersController(db *sql.DB) *PayoutTransfersController {
	return &PayoutTransfersController{
		db:                 db,
		helper:             helpers.NewPaymentHelper(db),
		tokenRepo:          repositories.NewTokenRepository(db),
		settingRepo:        repositories.NewSettingRepository(db),
		walletRepo:         repositories.NewWalletRepository(db),
		feesExpressRepo:    repositories.NewFeesExpressRepository(db),
		transactionRepo:    repositories.NewTransactionInfoRepository(db),
		transactionsRepo:   repositories.NewTransactionsRepository(db),
		merchantPayoutRepo: repositories.NewMerchantPayoutRepository(db),
		callbackRepo:       repositories.NewCallbackRepository(db),
		pakaiLinkService:   services.NewPakaiLinkService(),
		linkQuService:      services.NewLinkQuService(),
	}
}

// ProcessPayout handles POST /v2/payouts/transfers
func (ptc *PayoutTransfersController) ProcessPayout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011000",
			"response_message": "Unauthorized",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011000",
			"response_message": "Unauthorized",
		})
		return
	}

	// Check maintenance
	isMaintenancePayout, err := ptc.settingRepo.GetMaintenancePayout()
	if err == nil && isMaintenancePayout {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"response_code":    "5031000",
			"response_message": "The Payout service is currently under maintenance. Please try again later.",
		})
		return
	}

	// Verify token and merchant
	tokenData, merchantApp, merchant, user, err := ptc.helper.VerifyTokenAndMerchant(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	if tokenData == nil || merchant == nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011001",
			"response_message": "Invalid token",
		})
		return
	}

	// Validate request body
	var reqBody struct {
		ReferenceID string  `json:"reference_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Destination struct {
			Code          string `json:"code" binding:"required"`
			AccountNumber string `json:"account_number" binding:"required"`
			AccountName   string `json:"account_name" binding:"required"`
		} `json:"destination" binding:"required"`
		NotifyURL string `json:"notify_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4001000",
			"response_message": "Invalid Mandatory Field",
		})
		return
	}

	if reqBody.Destination.Code == "" || reqBody.Destination.AccountNumber == "" || reqBody.Destination.AccountName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"response_code":    "4001001",
			"response_message": "Invalid Field Format",
		})
		return
	}

	if reqBody.Amount < 10000 || reqBody.Amount > 100000000 {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4031002",
			"response_message": "Exceeds Transaction Amount Limit",
		})
		return
	}

	// Get payment method
	paymentMethod, err := ptc.helper.GetPaymentMethodByCode(strings.ToUpper(reqBody.Destination.Code))
	if err != nil || paymentMethod == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response_code":    "4041000",
			"response_message": "No methods available",
		})
		return
	}

	// Get fees express
	feeExpress, err := ptc.feesExpressRepo.GetFeesExpressByTransactionTypeID(9)
	if err != nil || feeExpress == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001000",
			"response_message": "General error",
		})
		return
	}

	// Calculate charges
	chargePercentage := feeExpress.ChargePercentage
	chargeFixed := feeExpress.ChargeFixed
	amountPercentage := (reqBody.Amount * chargePercentage) / 100
	totalAdmin := amountPercentage + chargeFixed
	totalAmount := reqBody.Amount + totalAdmin

	// Check wallet balance
	wallet, err := ptc.walletRepo.GetUserWallet(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	if wallet.Balance < totalAmount {
		c.JSON(http.StatusForbidden, gin.H{
			"response_code":    "4031003",
			"response_message": "Insufficient Funds",
		})
		return
	}

	// Generate payout ID
	payoutID := uuid.New().String()
	log.Printf("[PAYOUT] Starting payout transfer: payoutID=%s, referenceID=%s, amount=%.2f, destination=%s (%s)",
		payoutID, reqBody.ReferenceID, reqBody.Amount, reqBody.Destination.Code, reqBody.Destination.AccountNumber)

	// Fee untuk transfer = 2000
	transferFee := 2000.0
	totalNeeded := reqBody.Amount + transferFee

	// Cek saldo PakaiLink (jika error, lanjut ke LinkQu)
	pakaiLinkBalance := 0.0
	pakaiLinkConfig := config.GetPakaiLinkConfig()
	partnerRefBalance := uuid.New().String()
	pakaiLinkBalanceResponse, pakaiLinkBalanceErr := ptc.pakaiLinkService.GetBalance(partnerRefBalance, pakaiLinkConfig.AccountNo)
	if pakaiLinkBalanceErr != nil {
		log.Printf("[PAYOUT] PakaiLink balance check failed for payout %s: %v, will try LinkQu", payoutID, pakaiLinkBalanceErr)
	} else {
		responseCode, _ := pakaiLinkBalanceResponse["responseCode"].(string)
		if responseCode == "2001100" {
			if accountInfo, ok := pakaiLinkBalanceResponse["accountInfo"].([]interface{}); ok && len(accountInfo) > 0 {
				if firstAccount, ok := accountInfo[0].(map[string]interface{}); ok {
					if activeBalance, ok := firstAccount["activeBalance"].(map[string]interface{}); ok {
						if valueStr, ok := activeBalance["value"].(string); ok {
							if valueFloat, err := strconv.ParseFloat(valueStr, 64); err == nil {
								pakaiLinkBalance = valueFloat
							}
						}
					}
				}
			}
		}
	}

	// Cek saldo LinkQu (jika error, log saja)
	linkQuBalance := 0.0
	linkQuBalanceResponse, linkQuBalanceErr := ptc.linkQuService.GetBalance()
	if linkQuBalanceErr != nil {
		log.Printf("[PAYOUT] LinkQu balance check failed for payout %s: %v", payoutID, linkQuBalanceErr)
	} else {
		rc, _ := linkQuBalanceResponse["rc"].(string)
		if rc == "00" {
			if balance, ok := linkQuBalanceResponse["balance"].(float64); ok {
				linkQuBalance = balance
			} else if balanceStr, ok := linkQuBalanceResponse["balance"].(string); ok {
				if balanceFloat, err := strconv.ParseFloat(balanceStr, 64); err == nil {
					linkQuBalance = balanceFloat
				}
			}
		}
	}

	// Tentukan provider berdasarkan saldo (prioritas PakaiLink jika saldo cukup)
	usePakaiLink := pakaiLinkBalanceErr == nil && pakaiLinkBalance >= totalNeeded
	useLinkQu := linkQuBalanceErr == nil && linkQuBalance >= totalNeeded

	log.Printf("[PAYOUT] Balance check for payout %s: PakaiLink=%.2f (error: %v), LinkQu=%.2f (error: %v), needed=%.2f, using=%s",
		payoutID, pakaiLinkBalance, pakaiLinkBalanceErr != nil, linkQuBalance, linkQuBalanceErr != nil, totalNeeded,
		map[bool]string{true: "PakaiLink", false: "LinkQu"}[usePakaiLink])

	if !usePakaiLink && !useLinkQu {
		log.Printf("[PAYOUT] Both providers unavailable for payout %s: PakaiLink error=%v, LinkQu error=%v, PakaiLink balance=%.2f, LinkQu balance=%.2f, needed=%.2f",
			payoutID, pakaiLinkBalanceErr != nil, linkQuBalanceErr != nil, pakaiLinkBalance, linkQuBalance, totalNeeded)
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Deteksi apakah bank atau e-wallet berdasarkan code
	destinationCodeUpper := strings.ToUpper(reqBody.Destination.Code)
	isEWallet := destinationCodeUpper == "OVO" ||
		destinationCodeUpper == "DANA" ||
		destinationCodeUpper == "GOPAY" ||
		destinationCodeUpper == "LINKAJA" ||
		destinationCodeUpper == "SHOPEEPAY" ||
		strings.ToUpper(paymentMethod.Type) == "EWALLET"

	// CRITICAL: Create all database records BEFORE calling payment API
	// This ensures transaction exists even if API call fails
	// Create transaction info
	transactionData := models.TransactionInfoData{
		AppID:           tokenData.AppID,
		PaymentMethod:   paymentMethod.Code,
		Amount:          int64(reqBody.Amount),
		Currency:        "IDR",
		NotifyURL:       reqBody.NotifyURL,
		GrantID:         payoutID,
		OrderID:         reqBody.ReferenceID,
		Token:           apiKey,
		BankNumber:      &reqBody.Destination.AccountNumber,
		BankEwalletName: &reqBody.Destination.AccountName,
		Status:          "pending",
	}

	err = ptc.transactionRepo.CreateTransaction(transactionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Create transactions
	currencyID := 1
	paymentMethodID := paymentMethod.ID
	merchantID := merchantApp.MerchantID
	userID := user.ID
	transactionsData := models.TransactionsData{
		UserID:                 &userID,
		CurrencyID:             &currencyID,
		PaymentMethodID:        &paymentMethodID,
		MerchantID:             &merchantID,
		UUID:                   &reqBody.ReferenceID,
		GrantID:                &payoutID,
		TransactionReferenceID: 1,
		TransactionTypeID:      func() *int { v := 9; return &v }(),
		UserType:               "registered",
		Subtotal:               reqBody.Amount,
		Percentage:             chargePercentage,
		ChargePercentage:       amountPercentage,
		ChargeFixed:            chargeFixed,
		Total:                  totalAmount,
		Note:                   &reqBody.Description,
		PaymentStatus:          func() *string { v := "Pending"; return &v }(),
		Status:                 "Pending",
	}

	err = ptc.transactionsRepo.CreateTransactions(transactionsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Create merchant payout
	merchantPayoutData := models.MerchantPayoutData{
		MerchantID:       &merchantID,
		PaymentMethodID:  &paymentMethodID,
		GatewayReference: &payoutID,
		OrderNo:          &reqBody.ReferenceID,
		UUID:             &reqBody.ReferenceID,
		FeeBearer:        "Merchant",
		Percentage:       chargePercentage,
		ChargePercentage: amountPercentage,
		ChargeFixed:      chargeFixed,
		Amount:           reqBody.Amount,
		Total:            totalAmount,
		Status:           "Pending",
		BankName:         &paymentMethod.Name,
		AccountName:      &reqBody.Destination.AccountName,
		AccountNumber:    &reqBody.Destination.AccountNumber,
	}

	err = ptc.merchantPayoutRepo.CreateMerchantPayout(merchantPayoutData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Get transaction info ID
	transactionInfo, err := ptc.helper.GetTransactionInfoByGrantID(payoutID)
	if err != nil || transactionInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
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

	err = ptc.callbackRepo.CreateCallback(callbackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Now proceed with API calls - transaction already exists in DB
	// Try PakaiLink first if selected, fallback to LinkQu if PakaiLink fails
	var responseData map[string]interface{}
	var sessionId string
	var inquiryRef interface{}
	var providerUsed string
	var pakaiLinkFailed bool

	if usePakaiLink {
		// PakaiLink: 2 fase - inquiry dulu
		if isEWallet {
			// E-Wallet inquiry
			productCode := helpers.GetPakaiLinkProductCode(reqBody.Destination.Code)
			log.Printf("[PAKAILINK E-WALLET] Starting inquiry for payout %s: account=%s, product=%s, amount=%.2f",
				payoutID, reqBody.Destination.AccountNumber, productCode, reqBody.Amount)

			inquiryResponse, err := ptc.pakaiLinkService.EWalletAccountInquiry(
				payoutID,
				reqBody.Destination.AccountNumber,
				productCode,
				reqBody.Amount,
			)
			if err != nil {
				log.Printf("[PAKAILINK E-WALLET] Inquiry failed for payout %s: %v, will try LinkQu", payoutID, err)
				pakaiLinkFailed = true
			} else {
				inquiryResponseJSON, _ := json.Marshal(inquiryResponse)
				log.Printf("[PAKAILINK E-WALLET] Inquiry response for payout %s: %s", payoutID, string(inquiryResponseJSON))

				responseCode, _ := inquiryResponse["responseCode"].(string)
				if responseCode != "2003700" {
					log.Printf("[PAKAILINK E-WALLET] Inquiry response code not success for payout %s: %s, will try LinkQu", payoutID, responseCode)
					pakaiLinkFailed = true
				} else {
					sessionId, _ = inquiryResponse["sessionId"].(string)
					if sessionId == "" {
						log.Printf("[PAKAILINK E-WALLET] Missing sessionId in inquiry response for payout %s, will try LinkQu", payoutID)
						pakaiLinkFailed = true
					} else {
						log.Printf("[PAKAILINK E-WALLET] Inquiry success for payout %s, sessionId=%s, starting payment", payoutID, sessionId)

						// E-Wallet transfer
						pakaiLinkConfig := config.GetPakaiLinkConfig()
						callbackURL := fmt.Sprintf("%s/payouts/pakailink/ewallet", pakaiLinkConfig.CallbackURL)
						log.Printf("[PAKAILINK E-WALLET] Starting payment for payout %s: sessionId=%s, callbackURL=%s",
							payoutID, sessionId, callbackURL)

						responseData, err = ptc.pakaiLinkService.TopupEWallet(
							payoutID,
							reqBody.Destination.AccountNumber,
							productCode,
							sessionId,
							reqBody.Amount,
							callbackURL,
						)
						if err != nil {
							log.Printf("[PAKAILINK E-WALLET] Payment failed for payout %s: %v, will try LinkQu", payoutID, err)
							pakaiLinkFailed = true
						} else {
							responseDataJSON, _ := json.Marshal(responseData)
							log.Printf("[PAKAILINK E-WALLET] Payment response for payout %s: %s", payoutID, string(responseDataJSON))

							responseCode, _ := responseData["responseCode"].(string)
							if responseCode != "2003800" {
								log.Printf("[PAKAILINK E-WALLET] Payment response code not success for payout %s: %s (expected 2003800), will try LinkQu", payoutID, responseCode)
								pakaiLinkFailed = true
							} else {
								log.Printf("[PAKAILINK E-WALLET] Payment success for payout %s", payoutID)
								providerUsed = "PakaiLink"
							}
						}
					}
				}
			}
		} else {
			// Bank inquiry
			bankCode := helpers.GetPakaiLinkBankCode(reqBody.Destination.Code)
			log.Printf("[PAKAILINK BANK] Starting inquiry for payout %s: account=%s, bank=%s, amount=%.2f",
				payoutID, reqBody.Destination.AccountNumber, bankCode, reqBody.Amount)

			inquiryResponse, err := ptc.pakaiLinkService.BankAccountInquiry(
				payoutID,
				reqBody.Destination.AccountNumber,
				bankCode,
				reqBody.Amount,
			)
			if err != nil {
				log.Printf("[PAKAILINK BANK] Inquiry failed for payout %s: %v, will try LinkQu", payoutID, err)
				pakaiLinkFailed = true
			} else {
				inquiryResponseJSON, _ := json.Marshal(inquiryResponse)
				log.Printf("[PAKAILINK BANK] Inquiry response for payout %s: %s", payoutID, string(inquiryResponseJSON))

				responseCode, _ := inquiryResponse["responseCode"].(string)
				if responseCode != "2004200" {
					log.Printf("[PAKAILINK BANK] Inquiry response code not success for payout %s: %s, will try LinkQu", payoutID, responseCode)
					pakaiLinkFailed = true
				} else {
					sessionId, _ = inquiryResponse["sessionId"].(string)
					if sessionId == "" {
						log.Printf("[PAKAILINK BANK] Missing sessionId in inquiry response for payout %s, will try LinkQu", payoutID)
						pakaiLinkFailed = true
					} else {
						log.Printf("[PAKAILINK BANK] Inquiry success for payout %s, sessionId=%s, starting payment", payoutID, sessionId)

						// Bank transfer
						pakaiLinkConfig := config.GetPakaiLinkConfig()
						callbackURL := fmt.Sprintf("%s/payouts/pakailink/bank", pakaiLinkConfig.CallbackURL)
						log.Printf("[PAKAILINK BANK] Starting payment for payout %s: sessionId=%s, callbackURL=%s",
							payoutID, sessionId, callbackURL)

						responseData, err = ptc.pakaiLinkService.TransferBank(
							payoutID,
							reqBody.Destination.AccountNumber,
							bankCode,
							sessionId,
							reqBody.Amount,
							callbackURL,
							reqBody.Description,
						)
						if err != nil {
							log.Printf("[PAKAILINK BANK] Payment failed for payout %s: %v, will try LinkQu", payoutID, err)
							pakaiLinkFailed = true
						} else {
							responseDataJSON, _ := json.Marshal(responseData)
							log.Printf("[PAKAILINK BANK] Payment response for payout %s: %s", payoutID, string(responseDataJSON))

							responseCode, _ := responseData["responseCode"].(string)
							if responseCode != "2004300" {
								log.Printf("[PAKAILINK BANK] Payment response code not success for payout %s: %s (expected 2004300), will try LinkQu", payoutID, responseCode)
								pakaiLinkFailed = true
							} else {
								log.Printf("[PAKAILINK BANK] Payment success for payout %s", payoutID)
								providerUsed = "PakaiLink"
							}
						}
					}
				}
			}
		}
	}

	// Fallback to LinkQu if PakaiLink failed or not selected
	if pakaiLinkFailed || !usePakaiLink {
		// LinkQu: 2 fase - inquiry dulu
		linkQuFailed := false
		if isEWallet {
			// E-Wallet inquiry
			ewalletCode := helpers.GetLinkQuEWalletCode(reqBody.Destination.Code)
			log.Printf("[LINKQU E-WALLET] Starting inquiry for payout %s: account=%s, ewallet=%s, amount=%d",
				payoutID, reqBody.Destination.AccountNumber, ewalletCode, int64(reqBody.Amount))

			inquiryResponse, err := ptc.linkQuService.EWalletReloadInquiry(
				ewalletCode,
				reqBody.Destination.AccountNumber,
				int64(reqBody.Amount),
			)
			if err != nil {
				log.Printf("[LINKQU E-WALLET] Inquiry failed for payout %s: %v", payoutID, err)
				linkQuFailed = true
			} else {
				inquiryResponseJSON, _ := json.Marshal(inquiryResponse)
				log.Printf("[LINKQU E-WALLET] Inquiry response for payout %s: %s", payoutID, string(inquiryResponseJSON))

				status, _ := inquiryResponse["status"].(string)
				responseCode, _ := inquiryResponse["response_code"].(string)
				if status != "SUCCESS" || responseCode != "00" {
					log.Printf("[LINKQU E-WALLET] Inquiry response not success for payout %s: status=%s, response_code=%s",
						payoutID, status, responseCode)
					linkQuFailed = true
				} else {
					inquiryRef = inquiryResponse["inquiry_reff"]
					if inquiryRef == nil {
						log.Printf("[LINKQU E-WALLET] Missing inquiry_reff in response for payout %s", payoutID)
						linkQuFailed = true
					} else {
						log.Printf("[LINKQU E-WALLET] Inquiry success for payout %s, inquiry_reff=%v, starting payment", payoutID, inquiryRef)

						// E-Wallet transfer
						linkQuConfig := config.GetLinkQuConfig()
						callbackURL := fmt.Sprintf("%s/payouts/linkqu/ewallet", linkQuConfig.CallbackURL)
						inquiryRefStr := fmt.Sprintf("%v", inquiryRef)
						log.Printf("[LINKQU E-WALLET] Starting payment for payout %s: inquiry_reff=%s, callbackURL=%s",
							payoutID, inquiryRefStr, callbackURL)

						responseData, err = ptc.linkQuService.EWalletReloadPayment(
							ewalletCode,
							reqBody.Destination.AccountNumber,
							int64(reqBody.Amount),
							payoutID,
							inquiryRefStr,
							callbackURL,
						)
						if err != nil {
							log.Printf("[LINKQU E-WALLET] Payment failed for payout %s: %v", payoutID, err)
							linkQuFailed = true
						} else {
							responseDataJSON, _ := json.Marshal(responseData)
							log.Printf("[LINKQU E-WALLET] Payment response for payout %s: %s", payoutID, string(responseDataJSON))

							status, _ := responseData["status"].(string)
							responseCode, _ := responseData["response_code"].(string)
							if status != "SUCCESS" || responseCode != "00" {
								log.Printf("[LINKQU E-WALLET] Payment response not success for payout %s: status=%s, response_code=%s (expected SUCCESS/00)",
									payoutID, status, responseCode)
								linkQuFailed = true
							} else {
								log.Printf("[LINKQU E-WALLET] Payment success for payout %s", payoutID)
								providerUsed = "LinkQu"
							}
						}
					}
				}
			}
		} else {
			// Bank inquiry
			bankCode := helpers.GetLinkQuBankCode(reqBody.Destination.Code)
			log.Printf("[LINKQU BANK] Starting inquiry for payout %s: account=%s, bank=%s, amount=%d",
				payoutID, reqBody.Destination.AccountNumber, bankCode, int64(reqBody.Amount))

			inquiryResponse, err := ptc.linkQuService.BankWithdrawInquiry(
				bankCode,
				reqBody.Destination.AccountNumber,
				int64(reqBody.Amount),
				payoutID,
			)
			if err != nil {
				log.Printf("[LINKQU BANK] Inquiry failed for payout %s: %v", payoutID, err)
				linkQuFailed = true
			} else {
				inquiryResponseJSON, _ := json.Marshal(inquiryResponse)
				log.Printf("[LINKQU BANK] Inquiry response for payout %s: %s", payoutID, string(inquiryResponseJSON))

				status, _ := inquiryResponse["status"].(string)
				responseCode, _ := inquiryResponse["response_code"].(string)
				if status != "SUCCESS" || responseCode != "00" {
					log.Printf("[LINKQU BANK] Inquiry response not success for payout %s: status=%s, response_code=%s",
						payoutID, status, responseCode)
					linkQuFailed = true
				} else {
					inquiryRef = inquiryResponse["inquiry_reff"]
					if inquiryRef == nil {
						log.Printf("[LINKQU BANK] Missing inquiry_reff in response for payout %s", payoutID)
						linkQuFailed = true
					} else {
						log.Printf("[LINKQU BANK] Inquiry success for payout %s, inquiry_reff=%v, starting payment", payoutID, inquiryRef)

						// Bank transfer
						linkQuConfig := config.GetLinkQuConfig()
						callbackURL := fmt.Sprintf("%s/payouts/linkqu/bank", linkQuConfig.CallbackURL)
						inquiryRefStr := fmt.Sprintf("%v", inquiryRef)
						log.Printf("[LINKQU BANK] Starting payment for payout %s: inquiry_reff=%s, callbackURL=%s",
							payoutID, inquiryRefStr, callbackURL)

						responseData, err = ptc.linkQuService.BankWithdrawPayment(
							bankCode,
							reqBody.Destination.AccountNumber,
							int64(reqBody.Amount),
							payoutID,
							inquiryRefStr,
							callbackURL,
						)
						if err != nil {
							log.Printf("[LINKQU BANK] Payment failed for payout %s: %v", payoutID, err)
							linkQuFailed = true
						} else {
							responseDataJSON, _ := json.Marshal(responseData)
							log.Printf("[LINKQU BANK] Payment response for payout %s: %s", payoutID, string(responseDataJSON))

							status, _ := responseData["status"].(string)
							responseCode, _ := responseData["response_code"].(string)
							if status != "SUCCESS" || responseCode != "00" {
								log.Printf("[LINKQU BANK] Payment response not success for payout %s: status=%s, response_code=%s (expected SUCCESS/00)",
									payoutID, status, responseCode)
								linkQuFailed = true
							} else {
								log.Printf("[LINKQU BANK] Payment success for payout %s", payoutID)
								providerUsed = "LinkQu"
							}
						}
					}
				}
			}
		}

		// If LinkQu also failed and we already tried PakaiLink, return error
		// Otherwise, if LinkQu failed but we didn't try PakaiLink (LinkQu was primary), return error
		if linkQuFailed {
			if pakaiLinkFailed {
				// Both providers failed
				log.Printf("[PAYOUT] Both providers failed for payout %s: PakaiLink failed, LinkQu also failed", payoutID)
				c.JSON(http.StatusInternalServerError, gin.H{
					"response_code":    "5001001",
					"response_message": "Internal Server Error",
				})
				return
			} else if !usePakaiLink {
				// LinkQu was primary and failed, but PakaiLink wasn't available
				log.Printf("[PAYOUT] LinkQu failed for payout %s and PakaiLink not available", payoutID)
				c.JSON(http.StatusInternalServerError, gin.H{
					"response_code":    "5001001",
					"response_message": "Internal Server Error",
				})
				return
			}
		} else {
			providerUsed = "LinkQu"
		}
	}

	// Get transaction info for response (already created above)
	transactionInfo, err = ptc.helper.GetTransactionInfoByGrantID(payoutID)
	if err != nil || transactionInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001001",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Disable token
	ptc.tokenRepo.DisableToken(apiKey)
	c.Header("X-AUTH-TOKEN", apiKey)

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	// Determine final provider used
	if providerUsed == "" {
		// If no provider succeeded but we didn't return error, it means we're relying on webhook
		// Use the originally selected provider for logging
		if usePakaiLink {
			providerUsed = "PakaiLink"
		} else {
			providerUsed = "LinkQu"
		}
	}

	log.Printf("[PAYOUT] Successfully processed payout %s: referenceID=%s, amount=%.2f, provider=%s",
		payoutID, reqBody.ReferenceID, reqBody.Amount, providerUsed)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2001000",
		"response_message": "Successful",
		"response_data": gin.H{
			"id":           payoutID,
			"reference_id": reqBody.ReferenceID,
			"amount":       reqBody.Amount,
			"payout_data": gin.H{
				"code":           reqBody.Destination.Code,
				"account_number": reqBody.Destination.AccountNumber,
				"account_name":   reqBody.Destination.AccountName,
			},
			"merchant_url": gin.H{
				"notify_url": reqBody.NotifyURL,
			},
			"request_time": requestTime,
		},
	})
}
