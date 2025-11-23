package models

import "time"

// MerchantPayout represents a merchant payout record
type MerchantPayout struct {
	ID              int        `json:"id" db:"id"`
	MerchantID      *int       `json:"merchant_id" db:"merchant_id"`
	CurrencyID      *int       `json:"currency_id" db:"currency_id"`
	PaymentMethodID *int       `json:"payment_method_id" db:"payment_method_id"`
	UserID          *int       `json:"user_id" db:"user_id"`
	GatewayReference *string   `json:"gateway_reference" db:"gateway_reference"`
	OrderNo         *string   `json:"order_no" db:"order_no"`
	ItemName        *string   `json:"item_name" db:"item_name"`
	UUID            *string   `json:"uuid" db:"uuid"`
	FeeBearer       string    `json:"fee_bearer" db:"fee_bearer"` // Merchant, User
	Percentage      float64   `json:"percentage" db:"percentage"`
	ChargePercentage float64  `json:"charge_percentage" db:"charge_percentage"`
	ChargeFixed     float64   `json:"charge_fixed" db:"charge_fixed"`
	Amount          float64   `json:"amount" db:"amount"`
	Total           float64   `json:"total" db:"total"`
	Status          string    `json:"status" db:"status"` // Pending, Success, Refund, Blocked
	CreatedAt       *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
	BankName        *string   `json:"bank_name" db:"bank_name"`
	AccountName     *string   `json:"account_name" db:"account_name"`
	AccountNumber   *string   `json:"account_number" db:"account_number"`
}

// MerchantPayoutData represents the data for creating a merchant payout
type MerchantPayoutData struct {
	MerchantID       *int     `json:"merchant_id" db:"merchant_id"`
	CurrencyID       *int     `json:"currency_id" db:"currency_id"`
	PaymentMethodID *int     `json:"payment_method_id" db:"payment_method_id"`
	UserID           *int     `json:"user_id" db:"user_id"`
	GatewayReference *string `json:"gateway_reference" db:"gateway_reference"`
	OrderNo          *string `json:"order_no" db:"order_no"`
	ItemName         *string `json:"item_name" db:"item_name"`
	UUID             *string `json:"uuid" db:"uuid"`
	FeeBearer        string  `json:"fee_bearer" db:"fee_bearer"`
	Percentage       float64 `json:"percentage" db:"percentage"`
	ChargePercentage float64 `json:"charge_percentage" db:"charge_percentage"`
	ChargeFixed      float64 `json:"charge_fixed" db:"charge_fixed"`
	Amount           float64 `json:"amount" db:"amount"`
	Total            float64 `json:"total" db:"total"`
	Status           string  `json:"status" db:"status"`
	BankName         *string `json:"bank_name" db:"bank_name"`
	AccountName      *string `json:"account_name" db:"account_name"`
	AccountNumber    *string `json:"account_number" db:"account_number"`
}

