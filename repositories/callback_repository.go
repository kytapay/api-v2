package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type CallbackRepository struct {
	db *sql.DB
}

func NewCallbackRepository(db *sql.DB) *CallbackRepository {
	return &CallbackRepository{db: db}
}

// CreateCallback creates a new callback status record
func (r *CallbackRepository) CreateCallback(callbackData models.CallbackData) error {
	query := `INSERT INTO callback_status 
		(transaction_info_id, merchant_id, notify_url, error_message, response_body, payload, status, retry_count, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		callbackData.TransactionInfoID,
		callbackData.MerchantID,
		callbackData.NotifyURL,
		callbackData.ErrorMessage,
		callbackData.ResponseBody,
		callbackData.Payload,
		callbackData.Status,
		callbackData.RetryCount,
		now,
		now,
	)

	return err
}

// UpdateCallback updates a callback status by transaction_info_id
func (r *CallbackRepository) UpdateCallback(transactionInfoID int, updateData map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE callback_status SET "
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

	query += ", updated_at = ? WHERE transaction_info_id = ?"
	args = append(args, time.Now(), transactionInfoID)

	_, err := r.db.Exec(query, args...)
	return err
}

