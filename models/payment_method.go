package models

// PaymentMethod represents a payment method record
type PaymentMethod struct {
	ID     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Code   string `json:"code" db:"code"`
	Type   string `json:"type" db:"type"`
	Status string `json:"status" db:"status"`
}

