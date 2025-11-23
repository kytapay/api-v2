package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type FeesExpressRepository struct {
	db *sql.DB
}

func NewFeesExpressRepository(db *sql.DB) *FeesExpressRepository {
	return &FeesExpressRepository{db: db}
}

// GetFeesExpressByTransactionTypeID gets fees express by transaction type ID
func (r *FeesExpressRepository) GetFeesExpressByTransactionTypeID(transactionTypeID int) (*models.FeesExpress, error) {
	query := `SELECT id, transaction_type_id, charge_fixed, charge_percentage 
		FROM fees_express 
		WHERE transaction_type_id = ?`

	var fee models.FeesExpress
	err := r.db.QueryRow(query, transactionTypeID).Scan(
		&fee.ID,
		&fee.TransactionTypeID,
		&fee.ChargeFixed,
		&fee.ChargePercentage,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &fee, nil
}

