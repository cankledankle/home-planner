package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func StoreRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

func ValidateRefreshToken(token string) (uuid.UUID, error) {
	// Query all non-expired tokens and compare using bcrypt
	rows, err := Pool.Query(context.Background(),
		`SELECT user_id, token_hash, expires_at FROM refresh_tokens WHERE expires_at > NOW()`)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to query refresh tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID uuid.UUID
		var storedHash string
		var expiresAt time.Time

		if err := rows.Scan(&userID, &storedHash, &expiresAt); err != nil {
			continue
		}

		// Use bcrypt to compare the incoming token with stored hash
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(token)); err == nil {
			return userID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("invalid refresh token")
}

func DeleteRefreshToken(tokenHash string) error {
	_, err := Pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func DeleteRefreshTokenByValue(token string) error {
	// Find the token by comparing with bcrypt, then delete it
	rows, err := Pool.Query(context.Background(),
		`SELECT token_hash FROM refresh_tokens WHERE expires_at > NOW()`)
	if err != nil {
		return fmt.Errorf("failed to query refresh tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var storedHash string
		if err := rows.Scan(&storedHash); err != nil {
			continue
		}

		// Use bcrypt to compare the incoming token with stored hash
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(token)); err == nil {
			// Found the matching token, delete it
			_, err = Pool.Exec(context.Background(),
				`DELETE FROM refresh_tokens WHERE token_hash = $1`,
				storedHash)
			if err != nil {
				return fmt.Errorf("failed to delete refresh token: %w", err)
			}
			return nil
		}
	}

	return nil // Token not found, but that's okay for logout
}

func DeleteAllUserRefreshTokens(userID uuid.UUID) error {
	_, err := Pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}

func CleanupExpiredRefreshTokens() error {
	result, err := Pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired refresh tokens\n", rowsAffected)
	}
	return nil
}
