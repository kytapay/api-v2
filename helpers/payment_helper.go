package helpers

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/kytapay/api-v2/models"
	"github.com/kytapay/api-v2/repositories"
)

// PaymentHelper provides helper functions for payment processing
type PaymentHelper struct {
	db                    *sql.DB
	tokenRepo             *repositories.TokenRepository
	merchantRepo          *repositories.MerchantRepository
	merchantDetailRepo    *repositories.MerchantDetailRepository
	userDetailRepo        *repositories.UserDetailRepository
	settingRepo           *repositories.SettingRepository
	paymentMethodRepo     *repositories.PaymentMethodRepository
	feesLimitRepo         *repositories.FeesLimitRepository
	transactionInfoRepo   *repositories.TransactionInfoRepository
	merchantTransactionRepo *repositories.MerchantTransactionRepository
	callbackRepo          *repositories.CallbackRepository
}

func NewPaymentHelper(db *sql.DB) *PaymentHelper {
	return &PaymentHelper{
		db:                      db,
		tokenRepo:               repositories.NewTokenRepository(db),
		merchantRepo:            repositories.NewMerchantRepository(db),
		merchantDetailRepo:      repositories.NewMerchantDetailRepository(db),
		userDetailRepo:          repositories.NewUserDetailRepository(db),
		settingRepo:             repositories.NewSettingRepository(db),
		paymentMethodRepo:       repositories.NewPaymentMethodRepository(db),
		feesLimitRepo:           repositories.NewFeesLimitRepository(db),
		transactionInfoRepo:     repositories.NewTransactionInfoRepository(db),
		merchantTransactionRepo: repositories.NewMerchantTransactionRepository(db),
		callbackRepo:            repositories.NewCallbackRepository(db),
	}
}

// VerifyTokenAndMerchant verifies token and returns merchant info
func (ph *PaymentHelper) VerifyTokenAndMerchant(apiKey string) (*models.AppToken, *models.MerchantApp, *models.Merchant, *models.UserDetail, error) {
	// Verify token
	tokenData, err := ph.tokenRepo.GetTokenByToken(apiKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	currentTime := time.Now().Unix()
	if tokenData == nil || tokenData.Status == 1 {
		return nil, nil, nil, nil, nil // Invalid token
	}

	// Parse expires_in as Unix timestamp
	expiresIn, err := strconv.ParseInt(tokenData.ExpiresIn, 10, 64)
	if err != nil || expiresIn < currentTime {
		return nil, nil, nil, nil, nil // Expired token
	}

	// Verify merchant app
	merchantApp, err := ph.merchantRepo.GetMerchantAppByID(tokenData.AppID)
	if err != nil || merchantApp == nil {
		return nil, nil, nil, nil, err
	}

	// Verify merchant
	merchant, err := ph.merchantDetailRepo.GetMerchantByID(merchantApp.MerchantID)
	if err != nil || merchant == nil {
		return nil, nil, nil, nil, err
	}

	if merchant.Status != "Approved" {
		return nil, nil, nil, nil, nil // Merchant not approved
	}

	// Verify user
	user, err := ph.userDetailRepo.GetUserByID(*merchant.UserID)
	if err != nil || user == nil {
		return nil, nil, nil, nil, err
	}

	if user.Status != "Active" {
		return nil, nil, nil, nil, nil // User not active
	}

	return tokenData, merchantApp, merchant, user, nil
}

// GetPaymentMethodByCode gets payment method by code
func (ph *PaymentHelper) GetPaymentMethodByCode(code string) (*models.PaymentMethod, error) {
	query := `SELECT id, name, code, type, status FROM payment_methods WHERE code = ?`
	var method models.PaymentMethod
	err := ph.db.QueryRow(query, strings.ToUpper(code)).Scan(
		&method.ID,
		&method.Name,
		&method.Code,
		&method.Type,
		&method.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &method, nil
}

// GetFeesLimit gets fees limit for payment method
func (ph *PaymentHelper) GetFeesLimit(transactionTypeID, paymentMethodID int) (*models.FeesLimit, error) {
	query := `SELECT id, currency_id, transaction_type_id, payment_method_id, charge_fixed, charge_percentage, min_limit, max_limit, processing_time, has_transaction 
		FROM fees_limits 
		WHERE transaction_type_id = ? AND payment_method_id = ?`
	
	var fee models.FeesLimit
	err := ph.db.QueryRow(query, transactionTypeID, paymentMethodID).Scan(
		&fee.ID,
		&fee.CurrencyID,
		&fee.TransactionTypeID,
		&fee.PaymentMethodID,
		&fee.ChargeFixed,
		&fee.ChargePercentage,
		&fee.MinLimit,
		&fee.MaxLimit,
		&fee.ProcessingTime,
		&fee.HasTransaction,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

// GetTransactionInfoByGrantID gets transaction info by grant_id
func (ph *PaymentHelper) GetTransactionInfoByGrantID(grantID string) (*models.TransactionInfo, error) {
	query := `SELECT id, app_id, payment_method, amount, currency, success_url, cancel_url, notify_url, 
		grant_id, order_id, token, qris_string, bank_number, ewallet_link, bank_ewallet_name, 
		expires_in, version, status, created_at, updated_at 
		FROM app_transactions_infos WHERE grant_id = ?`
	
	var info models.TransactionInfo
	err := ph.db.QueryRow(query, grantID).Scan(
		&info.ID,
		&info.AppID,
		&info.PaymentMethod,
		&info.Amount,
		&info.Currency,
		&info.SuccessURL,
		&info.CancelURL,
		&info.NotifyURL,
		&info.GrantID,
		&info.OrderID,
		&info.Token,
		&info.QrisString,
		&info.BankNumber,
		&info.EwalletLink,
		&info.BankEwalletName,
		&info.ExpiresIn,
		&info.Version,
		&info.Status,
		&info.CreatedAt,
		&info.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetMerchantPaymentByGatewayRef gets merchant payment by gateway_reference
func (ph *PaymentHelper) GetMerchantPaymentByGatewayRef(gatewayRef, orderNo string) (*models.MerchantPayment, error) {
	query := `SELECT id, merchant_id, currency_id, payment_method_id, user_id, gateway_reference, 
		order_no, item_name, uuid, fee_bearer, percentage, charge_percentage, charge_fixed, 
		amount, total, status, created_at, updated_at 
		FROM merchant_payments 
		WHERE gateway_reference = ? AND order_no = ?`
	
	var payment models.MerchantPayment
	err := ph.db.QueryRow(query, gatewayRef, orderNo).Scan(
		&payment.ID,
		&payment.MerchantID,
		&payment.CurrencyID,
		&payment.PaymentMethodID,
		&payment.UserID,
		&payment.GatewayReference,
		&payment.OrderNo,
		&payment.ItemName,
		&payment.UUID,
		&payment.FeeBearer,
		&payment.Percentage,
		&payment.ChargePercentage,
		&payment.ChargeFixed,
		&payment.Amount,
		&payment.Total,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

