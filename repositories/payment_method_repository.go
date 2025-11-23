package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type PaymentMethodRepository struct {
	db *sql.DB
}

func NewPaymentMethodRepository(db *sql.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: db}
}

// GetPaymentMethodByID gets a payment method by ID
func (r *PaymentMethodRepository) GetPaymentMethodByID(id int) (*models.PaymentMethod, error) {
	var method models.PaymentMethod

	query := `SELECT id, name, code, type, status 
		FROM payment_methods 
		WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
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

