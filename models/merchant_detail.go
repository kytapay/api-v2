package models

import "time"

// Merchant represents a merchant record
type Merchant struct {
	ID          int        `json:"id" db:"id"`
	UserID      *int       `json:"user_id" db:"user_id"`
	MerchantUUID string    `json:"merchant_uuid" db:"merchant_uuid"`
	BusinessName string    `json:"business_name" db:"business_name"`
	SiteURL     *string    `json:"site_url" db:"site_url"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

