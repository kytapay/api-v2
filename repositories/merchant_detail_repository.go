package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type MerchantDetailRepository struct {
	db *sql.DB
}

func NewMerchantDetailRepository(db *sql.DB) *MerchantDetailRepository {
	return &MerchantDetailRepository{db: db}
}

// GetMerchantByID gets a merchant by ID
func (r *MerchantDetailRepository) GetMerchantByID(id int) (*models.Merchant, error) {
	var merchant models.Merchant

	query := `SELECT id, user_id, merchant_uuid, business_name, site_url, status, created_at, updated_at 
		FROM merchants 
		WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&merchant.ID,
		&merchant.UserID,
		&merchant.MerchantUUID,
		&merchant.BusinessName,
		&merchant.SiteURL,
		&merchant.Status,
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

