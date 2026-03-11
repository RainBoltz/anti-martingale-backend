package database

import (
	"database/sql"
	"fmt"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetBalance retrieves a user's balance
func (r *UserRepository) GetBalance(userID string) (float64, error) {
	var balance float64
	err := r.db.QueryRow(
		"SELECT balance FROM users WHERE user_id = $1",
		userID,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// UpdateBalance updates a user's balance
func (r *UserRepository) UpdateBalance(userID string, newBalance float64) error {
	result, err := r.db.Exec(
		"UPDATE users SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2",
		newBalance, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// GetNickname retrieves a user's nickname
func (r *UserRepository) GetNickname(userID string) (string, error) {
	var nickname string
	err := r.db.QueryRow(
		"SELECT nickname FROM users WHERE user_id = $1",
		userID,
	).Scan(&nickname)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get nickname: %w", err)
	}

	return nickname, nil
}

// CreateUser creates a new user with default balance
func (r *UserRepository) CreateUser(userID, nickname string, balance float64) error {
	_, err := r.db.Exec(
		`INSERT INTO users (user_id, nickname, balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO NOTHING`,
		userID, nickname, balance,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UserExists checks if a user exists
func (r *UserRepository) UserExists(userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)",
		userID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// GetOrCreateUser retrieves a user or creates them if they don't exist
func (r *UserRepository) GetOrCreateUser(userID, nickname string, defaultBalance float64) (balance float64, finalNickname string, err error) {
	// Try to get existing user
	err = r.db.QueryRow(
		"SELECT balance, nickname FROM users WHERE user_id = $1",
		userID,
	).Scan(&balance, &finalNickname)

	if err == sql.ErrNoRows {
		// User doesn't exist, create them
		err = r.CreateUser(userID, nickname, defaultBalance)
		if err != nil {
			return 0, "", err
		}
		return defaultBalance, nickname, nil
	}

	if err != nil {
		return 0, "", fmt.Errorf("failed to get user: %w", err)
	}

	return balance, finalNickname, nil
}
