package models

// FeesLimit represents a fees limit record
type FeesLimit struct {
	ID                int     `json:"id" db:"id"`
	TransactionTypeID int     `json:"transaction_type_id" db:"transaction_type_id"`
	PaymentMethodID   int     `json:"payment_method_id" db:"payment_method_id"`
	ChargeFixed       float64 `json:"charge_fixed" db:"charge_fixed"`
	ChargePercentage  float64 `json:"charge_percentage" db:"charge_percentage"`
	MinLimit          float64 `json:"min_limit" db:"min_limit"`
	MaxLimit          float64 `json:"max_limit" db:"max_limit"`
	ProcessingTime    int     `json:"processing_time" db:"processing_time"`
}

