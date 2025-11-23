package models

import "time"

// Wallet represents a wallet record
type Wallet struct {
	ID        int        `json:"id" db:"id"`
	UserID    *int       `json:"user_id" db:"user_id"`
	CurrencyID *int      `json:"currency_id" db:"currency_id"`
	Balance   float64    `json:"balance" db:"balance"` // decimal(20,8)
	IsDefault string     `json:"is_default" db:"is_default"` // Yes, No
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

