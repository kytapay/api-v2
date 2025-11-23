package repositories

import (
	"database/sql"
	"errors"

	"github.com/kytapay/api-v2/models"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// GetUserWallet gets wallet by user_id
func (r *WalletRepository) GetUserWallet(userID int) (*models.Wallet, error) {
	var wallet models.Wallet

	query := `SELECT id, user_id, currency_id, balance, is_default, created_at, updated_at 
		FROM wallets 
		WHERE user_id = ?`

	err := r.db.QueryRow(query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.CurrencyID,
		&wallet.Balance,
		&wallet.IsDefault,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("wallet not found")
	}

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// DeductUserBalance deducts balance from user wallet
func (r *WalletRepository) DeductUserBalance(userID int, amount float64) error {
	// Get current wallet
	wallet, err := r.GetUserWallet(userID)
	if err != nil {
		return err
	}

	// Check if balance is sufficient
	if wallet.Balance < amount {
		return errors.New("insufficient balance")
	}

	// Update balance
	query := `UPDATE wallets SET balance = balance - ? WHERE user_id = ?`
	_, err = r.db.Exec(query, amount, userID)

	return err
}

