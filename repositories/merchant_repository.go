package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type MerchantRepository struct {
	db *sql.DB
}

func NewMerchantRepository(db *sql.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

// GetMerchantAccess checks if merchant has access based on client_id & client_secret
func (r *MerchantRepository) GetMerchantAccess(clientID, clientSecret string) (*models.MerchantApp, error) {
	var merchant models.MerchantApp

	query := `SELECT id, merchant_id, client_id, client_secret, created_at, updated_at 
		FROM merchant_apps 
		WHERE client_id = ? AND client_secret = ?`

	err := r.db.QueryRow(query, clientID, clientSecret).Scan(
		&merchant.ID,
		&merchant.MerchantID,
		&merchant.ClientID,
		&merchant.ClientSecret,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &merchant, nil
}

// GetMerchantAppByID gets a merchant app by ID
func (r *MerchantRepository) GetMerchantAppByID(appID int) (*models.MerchantApp, error) {
	var merchant models.MerchantApp

	query := `SELECT id, merchant_id, client_id, client_secret, created_at, updated_at 
		FROM merchant_apps 
		WHERE id = ?`

	err := r.db.QueryRow(query, appID).Scan(
		&merchant.ID,
		&merchant.MerchantID,
		&merchant.ClientID,
		&merchant.ClientSecret,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &merchant, nil
}

