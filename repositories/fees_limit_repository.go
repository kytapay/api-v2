package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type FeesLimitRepository struct {
	db *sql.DB
}

func NewFeesLimitRepository(db *sql.DB) *FeesLimitRepository {
	return &FeesLimitRepository{db: db}
}

// GetFeesLimitsByTransactionTypeID gets fees limits by transaction type ID
func (r *FeesLimitRepository) GetFeesLimitsByTransactionTypeID(transactionTypeID int) ([]models.FeesLimit, error) {
	query := `SELECT id, currency_id, transaction_type_id, payment_method_id, charge_fixed, charge_percentage, min_limit, max_limit, processing_time, has_transaction 
		FROM fees_limits 
		WHERE transaction_type_id = ?`

	rows, err := r.db.Query(query, transactionTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feesLimits []models.FeesLimit
	for rows.Next() {
		var fee models.FeesLimit
		err := rows.Scan(
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
		if err != nil {
			continue
		}
		feesLimits = append(feesLimits, fee)
	}

	return feesLimits, nil
}

