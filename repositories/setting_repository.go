package repositories

import (
	"database/sql"
	"strconv"

	"github.com/kytapay/api-v2/models"
)

type SettingRepository struct {
	db *sql.DB
}

func NewSettingRepository(db *sql.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

// GetMaintenancePayment gets the maintenance_payment setting value
func (r *SettingRepository) GetMaintenancePayment() (bool, error) {
	return r.getMaintenanceSetting("maintenance_payment")
}

// GetMaintenancePayout gets the maintenance_payout setting value
func (r *SettingRepository) GetMaintenancePayout() (bool, error) {
	return r.getMaintenanceSetting("maintenance_payout")
}

// GetMaintenanceQris gets the maintenance_qris setting value
func (r *SettingRepository) GetMaintenanceQris() (bool, error) {
	return r.getMaintenanceSetting("maintenance_qris")
}

// GetMaintenanceVa gets the maintenance_va setting value
func (r *SettingRepository) GetMaintenanceVa() (bool, error) {
	return r.getMaintenanceSetting("maintenance_va")
}

// GetMaintenanceEwallet gets the maintenance_ewallet setting value
func (r *SettingRepository) GetMaintenanceEwallet() (bool, error) {
	return r.getMaintenanceSetting("maintenance_ewallet")
}

// getMaintenanceSetting is a helper function to get maintenance settings
func (r *SettingRepository) getMaintenanceSetting(name string) (bool, error) {
	var setting models.Setting

	query := `SELECT id, name, value, type 
		FROM settings 
		WHERE name = ?`

	err := r.db.QueryRow(query, name).Scan(
		&setting.ID,
		&setting.Name,
		&setting.Value,
		&setting.Type,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	// Check if value is null or empty
	if setting.Value == nil || *setting.Value == "" {
		return false, nil
	}

	// Convert string value to boolean
	result, err := strconv.ParseBool(*setting.Value)
	if err != nil {
		return false, err
	}

	return result, nil
}

