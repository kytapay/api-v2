package models

import "time"

// AppToken represents an app token record
type AppToken struct {
	ID        int        `json:"id" db:"id"`
	AppID     int        `json:"app_id" db:"app_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresIn string     `json:"expires_in" db:"expires_in"` // varchar(20) in DB
	Status    int        `json:"status" db:"status"`         // default: 2
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

// TokenData represents the data for creating a token
type TokenData struct {
	AppID     int    `json:"app_id" db:"app_id"`
	Token     string `json:"token" db:"token"`
	ExpiresIn string `json:"expires_in" db:"expires_in"` // varchar(20) in DB
	Status    int    `json:"status" db:"status"`         // default: 2
}

