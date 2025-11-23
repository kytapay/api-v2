package models

import "time"

// CallbackStatus represents a callback status record
type CallbackStatus struct {
	ID                int       `json:"id" db:"id"`
	TransactionInfoID int       `json:"transaction_info_id" db:"transaction_info_id"`
	MerchantID        int       `json:"merchant_id" db:"merchant_id"`
	NotifyURL         string    `json:"notify_url" db:"notify_url"`
	ErrorMessage      *string   `json:"error_message" db:"error_message"`
	ResponseBody      *string   `json:"response_body" db:"response_body"`
	Payload           *string   `json:"payload" db:"payload"`
	Status            string    `json:"status" db:"status"` // enum: Success, Pending, Failed
	RetryCount        int       `json:"retry_count" db:"retry_count"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CallbackData represents the data for creating a callback
type CallbackData struct {
	TransactionInfoID int     `json:"transaction_info_id" db:"transaction_info_id"`
	MerchantID        int     `json:"merchant_id" db:"merchant_id"`
	NotifyURL         string  `json:"notify_url" db:"notify_url"`
	ErrorMessage      *string `json:"error_message" db:"error_message"`
	ResponseBody      *string `json:"response_body" db:"response_body"`
	Payload           *string `json:"payload" db:"payload"`
	Status            string  `json:"status" db:"status"` // enum: Success, Pending, Failed
	RetryCount        int     `json:"retry_count" db:"retry_count"`
}

