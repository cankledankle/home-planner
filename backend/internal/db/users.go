package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRow struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func GetUserByID(ctx context.Context, userID string) (*UserRow, error) {
	var user UserRow

	err := Pool.QueryRow(ctx,
		"SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1",
		userID).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(ctx context.Context, email string) (*UserRow, error) {
	var user UserRow

	err := Pool.QueryRow(ctx,
		"SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func GetAllUsers(ctx context.Context) ([]UserRow, error) {
	rows, err := Pool.Query(ctx,
		"SELECT id, name, email, password_hash, role, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserRow
	for rows.Next() {
		var user UserRow
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func CreateUser(ctx context.Context, name, email, password, role string) (*UserRow, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user UserRow
	err = Pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, role, created_at, updated_at`,
		name, email, string(hashedPassword), role).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(ctx context.Context, userID, name, email, role string) (*UserRow, error) {
	var user UserRow
	err := Pool.QueryRow(ctx,
		`UPDATE users SET name = $2, email = $3, role = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password_hash, role, created_at, updated_at`,
		userID, name, email, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func UpdateUserPassword(ctx context.Context, userID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := Pool.Exec(ctx,
		"UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1",
		userID, string(hashedPassword))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func DeleteUser(ctx context.Context, userID string) error {
	result, err := Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func IsEmailTaken(ctx context.Context, email string, excludeUserID *string) (bool, error) {
	var query string
	var args []interface{}

	if excludeUserID != nil {
		query = "SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2"
		args = []interface{}{email, *excludeUserID}
	} else {
		query = "SELECT COUNT(*) FROM users WHERE email = $1"
		args = []interface{}{email}
	}

	var count int
	err := Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
