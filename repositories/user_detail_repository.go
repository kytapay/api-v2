package repositories

import (
	"database/sql"

	"github.com/kytapay/api-v2/models"
)

type UserDetailRepository struct {
	db *sql.DB
}

func NewUserDetailRepository(db *sql.DB) *UserDetailRepository {
	return &UserDetailRepository{db: db}
}

// GetUserByID gets a user by ID
func (r *UserDetailRepository) GetUserByID(id int) (*models.UserDetail, error) {
	var user models.UserDetail

	query := `SELECT id, first_name, last_name, email, formattedPhone, status, created_at, updated_at 
		FROM users 
		WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.FormattedPhone,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

