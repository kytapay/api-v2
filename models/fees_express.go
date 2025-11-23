package models

// FeesExpress represents a fees_express record
type FeesExpress struct {
	ID                int     `json:"id" db:"id"`
	TransactionTypeID int     `json:"transaction_type_id" db:"transaction_type_id"`
	ChargeFixed       float64 `json:"charge_fixed" db:"charge_fixed"`
	ChargePercentage  float64 `json:"charge_percentage" db:"charge_percentage"`
}

