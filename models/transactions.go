package models

import "time"

// Transactions represents a transactions table record (different from Transaction)
type Transactions struct {
	ID                    int        `json:"id" db:"id"`
	UserID                *int       `json:"user_id" db:"user_id"`
	EndUserID             *int       `json:"end_user_id" db:"end_user_id"`
	CurrencyID            *int       `json:"currency_id" db:"currency_id"`
	PaymentMethodID       *int       `json:"payment_method_id" db:"payment_method_id"`
	MerchantID            *int       `json:"merchant_id" db:"merchant_id"`
	BankID                *int       `json:"bank_id" db:"bank_id"`
	FileID                *int       `json:"file_id" db:"file_id"`
	UUID                  *string    `json:"uuid" db:"uuid"` // varchar(13)
	GrantID               *string    `json:"grant_id" db:"grant_id"`
	RefundReference       *string    `json:"refund_reference" db:"refund_reference"` // varchar(13)
	TransactionReferenceID int       `json:"transaction_reference_id" db:"transaction_reference_id"`
	TransactionTypeID     *int       `json:"transaction_type_id" db:"transaction_type_id"`
	UserType              string     `json:"user_type" db:"user_type"` // registered, unregistered
	Email                 *string    `json:"email" db:"email"`
	Phone                 *string    `json:"phone" db:"phone"`
	Subtotal              float64    `json:"subtotal" db:"subtotal"`
	Percentage            float64    `json:"percentage" db:"percentage"`
	ChargePercentage      float64    `json:"charge_percentage" db:"charge_percentage"`
	ChargeFixed           float64    `json:"charge_fixed" db:"charge_fixed"`
	Total                 float64    `json:"total" db:"total"`
	Note                  *string    `json:"note" db:"note"`
	PaymentStatus         *string    `json:"payment_status" db:"payment_status"` // Pending, Success, Blocked
	Status                string     `json:"status" db:"status"` // Pending, Success, Refund, Blocked, Pending_Settlement
	CreatedAt             *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionsData represents the data for creating a transaction
type TransactionsData struct {
	UserID                *int     `json:"user_id" db:"user_id"`
	EndUserID             *int     `json:"end_user_id" db:"end_user_id"`
	CurrencyID            *int     `json:"currency_id" db:"currency_id"`
	PaymentMethodID       *int     `json:"payment_method_id" db:"payment_method_id"`
	MerchantID            *int     `json:"merchant_id" db:"merchant_id"`
	BankID                *int     `json:"bank_id" db:"bank_id"`
	FileID                *int     `json:"file_id" db:"file_id"`
	UUID                  *string  `json:"uuid" db:"uuid"`
	GrantID               *string  `json:"grant_id" db:"grant_id"`
	RefundReference       *string  `json:"refund_reference" db:"refund_reference"`
	TransactionReferenceID int     `json:"transaction_reference_id" db:"transaction_reference_id"`
	TransactionTypeID     *int     `json:"transaction_type_id" db:"transaction_type_id"`
	UserType              string   `json:"user_type" db:"user_type"`
	Email                 *string  `json:"email" db:"email"`
	Phone                 *string  `json:"phone" db:"phone"`
	Subtotal              float64  `json:"subtotal" db:"subtotal"`
	Percentage            float64  `json:"percentage" db:"percentage"`
	ChargePercentage      float64  `json:"charge_percentage" db:"charge_percentage"`
	ChargeFixed           float64  `json:"charge_fixed" db:"charge_fixed"`
	Total                 float64  `json:"total" db:"total"`
	Note                  *string  `json:"note" db:"note"`
	PaymentStatus         *string  `json:"payment_status" db:"payment_status"`
	Status                string   `json:"status" db:"status"`
}

