package models

import "time"

// MerchantApp represents a merchant application record
type MerchantApp struct {
	ID          int        `json:"id" db:"id"`
	MerchantID  int        `json:"merchant_id" db:"merchant_id"`
	ClientID    string     `json:"client_id" db:"client_id"`
	ClientSecret string    `json:"client_secret" db:"client_secret"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

