package models

import "time"

// UserDetail represents a user detail record
type UserDetail struct {
	ID           int        `json:"id" db:"id"`
	FirstName    *string    `json:"first_name" db:"first_name"`
	LastName     *string    `json:"last_name" db:"last_name"`
	Email        string     `json:"email" db:"email"`
	FormattedPhone *string   `json:"formattedPhone" db:"formattedPhone"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
}

