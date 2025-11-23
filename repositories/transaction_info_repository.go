package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type TransactionInfoRepository struct {
	db *sql.DB
}

func NewTransactionInfoRepository(db *sql.DB) *TransactionInfoRepository {
	return &TransactionInfoRepository{db: db}
}

// CreateTransaction creates a new transaction info record
func (r *TransactionInfoRepository) CreateTransaction(transactionData models.TransactionInfoData) error {
	query := `INSERT INTO app_transactions_infos 
		(app_id, payment_method, amount, currency, success_url, cancel_url, notify_url, grant_id, order_id, token, qris_string, bank_number, ewallet_link, bank_ewallet_name, expires_in, version, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		transactionData.AppID,
		transactionData.PaymentMethod,
		transactionData.Amount,
		transactionData.Currency,
		transactionData.SuccessURL,
		transactionData.CancelURL,
		transactionData.NotifyURL,
		transactionData.GrantID,
		transactionData.OrderID,
		transactionData.Token,
		transactionData.QrisString,
		transactionData.BankNumber,
		transactionData.EwalletLink,
		transactionData.BankEwalletName,
		transactionData.ExpiresIn,
		transactionData.Version,
		transactionData.Status,
		now,
		now,
	)

	return err
}

// UpdateTransaction updates a transaction info by grant_id
func (r *TransactionInfoRepository) UpdateTransaction(grantID string, updateData map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE app_transactions_infos SET "
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

	query += ", updated_at = ? WHERE grant_id = ?"
	args = append(args, time.Now(), grantID)

	_, err := r.db.Exec(query, args...)
	return err
}

