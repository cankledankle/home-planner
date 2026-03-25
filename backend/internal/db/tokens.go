package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// HashToken returns a SHA-256 hex digest of the token.
// Refresh tokens are already high-entropy random UUIDs, so SHA-256 is appropriate
// and allows O(1) indexed lookup instead of O(N) bcrypt comparison.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Store) StoreRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

func (s *Store) ValidateRefreshToken(token string) (uuid.UUID, error) {
	tokenHash := HashToken(token)
	var userID uuid.UUID
	err := s.pool.QueryRow(context.Background(),
		`SELECT user_id FROM refresh_tokens WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash).Scan(&userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid refresh token")
	}
	return userID, nil
}

func (s *Store) DeleteRefreshToken(tokenHash string) error {
	_, err := s.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (s *Store) DeleteRefreshTokenByValue(token string) error {
	tokenHash := HashToken(token)
	_, err := s.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil // Token not found is not an error (idempotent logout)
}

func (s *Store) DeleteAllUserRefreshTokens(userID uuid.UUID) error {
	_, err := s.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}

func (s *Store) CleanupExpiredRefreshTokens() error {
	result, err := s.pool.Exec(context.Background(),
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
