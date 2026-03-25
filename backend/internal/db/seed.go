package db

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func (s *Store) SeedAdminUser() error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminName := os.Getenv("ADMIN_NAME")

	if adminEmail == "" || adminPassword == "" || adminName == "" {
		return nil
	}

	var count int
	err := s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check user count: %w", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = s.pool.Exec(context.Background(),
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, 'admin')`,
		adminName, adminEmail, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

func (s *Store) CheckAdminUserExists() (bool, error) {
	var count int
	err := s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check admin user: %w", err)
	}
	return count > 0, nil
}
