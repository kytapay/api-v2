package models

import "time"

// MerchantPayment represents a merchant payment record
type MerchantPayment struct {
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
}

// MerchantPaymentData represents the data for creating a merchant payment
type MerchantPaymentData struct {
	MerchantID       *int     `json:"merchant_id" db:"merchant_id"`
	CurrencyID       *int     `json:"currency_id" db:"currency_id"`
	PaymentMethodID  *int     `json:"payment_method_id" db:"payment_method_id"`
	UserID           *int     `json:"user_id" db:"user_id"`
	GatewayReference *string  `json:"gateway_reference" db:"gateway_reference"`
	OrderNo          *string  `json:"order_no" db:"order_no"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	UUID             *string  `json:"uuid" db:"uuid"`
	FeeBearer        string   `json:"fee_bearer" db:"fee_bearer"`
	Percentage       float64  `json:"percentage" db:"percentage"`
	ChargePercentage float64  `json:"charge_percentage" db:"charge_percentage"`
	ChargeFixed      float64  `json:"charge_fixed" db:"charge_fixed"`
	Amount           float64  `json:"amount" db:"amount"`
	Total            float64  `json:"total" db:"total"`
	Status           string   `json:"status" db:"status"`
}

