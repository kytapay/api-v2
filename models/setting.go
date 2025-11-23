package models

// Setting represents a setting record
type Setting struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Value *string `json:"value" db:"value"`
	Type string `json:"type" db:"type"`
}

