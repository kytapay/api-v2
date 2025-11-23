package repositories

import (
	"database/sql"
	"time"

	"github.com/kytapay/api-v2/models"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// CreateToken saves a token to the app_tokens table
func (r *TokenRepository) CreateToken(tokenData models.TokenData) error {
	query := `INSERT INTO app_tokens (app_id, token, expires_in, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	status := tokenData.Status
	if status == 0 {
		status = 2 // default value
	}

	_, err := r.db.Exec(query,
		tokenData.AppID,
		tokenData.Token,
		tokenData.ExpiresIn,
		status,
		now,
		now,
	)

	return err
}

// GetTokenByToken gets a token by token string
func (r *TokenRepository) GetTokenByToken(token string) (*models.AppToken, error) {
	var appToken models.AppToken

	query := `SELECT id, app_id, token, expires_in, status, created_at, updated_at 
		FROM app_tokens 
		WHERE token = ?`

	err := r.db.QueryRow(query, token).Scan(
		&appToken.ID,
		&appToken.AppID,
		&appToken.Token,
		&appToken.ExpiresIn,
		&appToken.Status,
		&appToken.CreatedAt,
		&appToken.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &appToken, nil
}

// DisableToken disables a token by setting status to 1
func (r *TokenRepository) DisableToken(token string) error {
	query := `UPDATE app_tokens SET status = 1, updated_at = ? WHERE token = ?`
	_, err := r.db.Exec(query, time.Now(), token)
	return err
}

