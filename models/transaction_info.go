package models

import "time"

// TransactionInfo represents an app transaction info record
type TransactionInfo struct {
	ID              int        `json:"id" db:"id"`
	AppID           int        `json:"app_id" db:"app_id"`
	PaymentMethod   string     `json:"payment_method" db:"payment_method"`
	Amount          int64      `json:"amount" db:"amount"` // decimal(20,0) - using int64 for precision
	Currency        string     `json:"currency" db:"currency"`
	SuccessURL      *string    `json:"success_url" db:"success_url"`
	CancelURL       *string    `json:"cancel_url" db:"cancel_url"`
	NotifyURL       string     `json:"notify_url" db:"notify_url"`
	GrantID         string     `json:"grant_id" db:"grant_id"`
	OrderID         string     `json:"order_id" db:"order_id"`
	Token           string     `json:"token" db:"token"`
	QrisString      *string    `json:"qris_string" db:"qris_string"`
	BankNumber      *string    `json:"bank_number" db:"bank_number"`
	EwalletLink     *string    `json:"ewallet_link" db:"ewallet_link"`
	BankEwalletName *string    `json:"bank_ewallet_name" db:"bank_ewallet_name"`
	ExpiresIn       *string    `json:"expires_in" db:"expires_in"`
	Version         *int       `json:"version" db:"version"`
	Status          string     `json:"status" db:"status"` // pending, success, cancel
	CreatedAt       *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionInfoData represents the data for creating a transaction info
type TransactionInfoData struct {
	AppID           int     `json:"app_id" db:"app_id"`
	PaymentMethod   string  `json:"payment_method" db:"payment_method"`
	Amount          int64   `json:"amount" db:"amount"`
	Currency        string  `json:"currency" db:"currency"`
	SuccessURL      *string `json:"success_url" db:"success_url"`
	CancelURL       *string `json:"cancel_url" db:"cancel_url"`
	NotifyURL       string  `json:"notify_url" db:"notify_url"`
	GrantID         string  `json:"grant_id" db:"grant_id"`
	OrderID         string  `json:"order_id" db:"order_id"`
	Token           string  `json:"token" db:"token"`
	QrisString      *string `json:"qris_string" db:"qris_string"`
	BankNumber      *string `json:"bank_number" db:"bank_number"`
	EwalletLink     *string `json:"ewallet_link" db:"ewallet_link"`
	BankEwalletName *string `json:"bank_ewallet_name" db:"bank_ewallet_name"`
	ExpiresIn       *string `json:"expires_in" db:"expires_in"`
	Version         *int   `json:"version" db:"version"`
	Status          string  `json:"status" db:"status"`
}

