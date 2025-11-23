package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type MerchantTransactionRepository struct {
	db *sql.DB
}

func NewMerchantTransactionRepository(db *sql.DB) *MerchantTransactionRepository {
	return &MerchantTransactionRepository{db: db}
}

// CreateMerchantTransaction creates a new merchant payment record
func (r *MerchantTransactionRepository) CreateMerchantTransaction(transactionData models.MerchantPaymentData) error {
	query := `INSERT INTO merchant_payments 
		(merchant_id, currency_id, payment_method_id, user_id, gateway_reference, order_no, item_name, uuid, fee_bearer, percentage, charge_percentage, charge_fixed, amount, total, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		now,
		now,
	)

	return err
}

// UpdateMerchantTransaction updates a merchant payment by gateway_reference
func (r *MerchantTransactionRepository) UpdateMerchantTransaction(gatewayReference string, updateData map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE merchant_payments SET "
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

