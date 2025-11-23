package models

import "time"

// APINameCheck represents an api_name_checks record
type APINameCheck struct {
	ID              int        `json:"id" db:"id"`
	MerchantID      int        `json:"merchant_id" db:"merchant_id"`
	ReferenceID     string     `json:"reference_id" db:"reference_id"`
	OrderID         string     `json:"order_id" db:"order_id"`
	Token           string     `json:"token" db:"token"`
	AccountNumber   string     `json:"account_number" db:"account_number"`
	BankCode        string     `json:"bank_code" db:"bank_code"`
	IlumaID         string     `json:"iluma_id" db:"iluma_id"`
	Status          string     `json:"status" db:"status"`
	NotifyURL       string     `json:"notify_url" db:"notify_url"`
	RawResponse     string     `json:"raw_response" db:"raw_response"`
	AccountName     *string    `json:"account_name" db:"account_name"`
	IsFound         *int       `json:"is_found" db:"is_found"`
	IsVirtualAccount *int      `json:"is_virtual_account" db:"is_virtual_account"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

