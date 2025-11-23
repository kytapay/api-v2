package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type APINameCheckRepository struct {
	db *sql.DB
}

func NewAPINameCheckRepository(db *sql.DB) *APINameCheckRepository {
	return &APINameCheckRepository{db: db}
}

// CreateAPINameCheck creates a new api_name_checks record
func (r *APINameCheckRepository) CreateAPINameCheck(data models.APINameCheck) error {
	query := `INSERT INTO api_name_checks 
		(merchant_id, reference_id, order_id, token, account_number, bank_code, iluma_id, status, notify_url, raw_response, account_name, is_found, is_virtual_account, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		data.MerchantID,
		data.ReferenceID,
		data.OrderID,
		data.Token,
		data.AccountNumber,
		data.BankCode,
		data.IlumaID,
		data.Status,
		data.NotifyURL,
		data.RawResponse,
		data.AccountName,
		data.IsFound,
		data.IsVirtualAccount,
		now,
		now,
	)

	return err
}

