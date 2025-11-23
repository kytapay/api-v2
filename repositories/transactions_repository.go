package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type TransactionsRepository struct {
	db *sql.DB
}

func NewTransactionsRepository(db *sql.DB) *TransactionsRepository {
	return &TransactionsRepository{db: db}
}

// CreateTransactions creates a new transaction record in the transactions table
func (r *TransactionsRepository) CreateTransactions(transactionData models.TransactionsData) error {
	query := `INSERT INTO transactions 
		(user_id, end_user_id, currency_id, payment_method_id, merchant_id, bank_id, file_id, uuid, grant_id, refund_reference, transaction_reference_id, transaction_type_id, user_type, email, phone, subtotal, percentage, charge_percentage, charge_fixed, total, note, payment_status, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		transactionData.UserID,
		transactionData.EndUserID,
		transactionData.CurrencyID,
		transactionData.PaymentMethodID,
		transactionData.MerchantID,
		transactionData.BankID,
		transactionData.FileID,
		transactionData.UUID,
		transactionData.GrantID,
		transactionData.RefundReference,
		transactionData.TransactionReferenceID,
		transactionData.TransactionTypeID,
		transactionData.UserType,
		transactionData.Email,
		transactionData.Phone,
		transactionData.Subtotal,
		transactionData.Percentage,
		transactionData.ChargePercentage,
		transactionData.ChargeFixed,
		transactionData.Total,
		transactionData.Note,
		transactionData.PaymentStatus,
		transactionData.Status,
		now,
		now,
	)

	return err
}

// UpdateTransactions updates a transaction by uuid
func (r *TransactionsRepository) UpdateTransactions(uuid string, updateData map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE transactions SET "
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

	query += ", updated_at = ? WHERE uuid = ?"
	args = append(args, time.Now(), uuid)

	_, err := r.db.Exec(query, args...)
	return err
}

