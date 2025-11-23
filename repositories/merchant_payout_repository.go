package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type MerchantPayoutRepository struct {
	db *sql.DB
}

func NewMerchantPayoutRepository(db *sql.DB) *MerchantPayoutRepository {
	return &MerchantPayoutRepository{db: db}
}

// CreateMerchantPayout creates a new merchant payout record
func (r *MerchantPayoutRepository) CreateMerchantPayout(transactionData models.MerchantPayoutData) error {
	query := `INSERT INTO merchant_payouts 
		(merchant_id, currency_id, payment_method_id, user_id, gateway_reference, order_no, item_name, uuid, fee_bearer, percentage, charge_percentage, charge_fixed, amount, total, status, bank_name, account_name, account_number, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		transactionData.MerchantID,
		transactionData.CurrencyID,
		transactionData.PaymentMethodID,
		transactionData.UserID,
		transactionData.GatewayReference,
		transactionData.OrderNo,
		transactionData.ItemName,
		transactionData.UUID,
		transactionData.FeeBearer,
		transactionData.Percentage,
		transactionData.ChargePercentage,
		transactionData.ChargeFixed,
		transactionData.Amount,
		transactionData.Total,
		transactionData.Status,
		transactionData.BankName,
		transactionData.AccountName,
		transactionData.AccountNumber,
		now,
		now,
	)

	return err
}

// UpdateMerchantPayout updates a merchant payout by gateway_reference
func (r *MerchantPayoutRepository) UpdateMerchantPayout(gatewayReference string, updateData map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE merchant_payouts SET "
	args := []interface{}{}
	first := true

	for key, value := range updateData {
		if !first {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		first = false
	}

	query += ", updated_at = ? WHERE gateway_reference = ?"
	args = append(args, time.Now(), gatewayReference)

	_, err := r.db.Exec(query, args...)
	return err
}

