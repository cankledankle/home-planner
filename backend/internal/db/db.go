package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection pool limits with environment variable overrides
	// Max open connections (default: 25)
	if maxConns := os.Getenv("DB_MAX_OPEN_CONNS"); maxConns != "" {
		if n, err := strconv.Atoi(maxConns); err == nil && n > 0 {
			poolConfig.MaxConns = int32(n)
		}
	} else {
		poolConfig.MaxConns = 25
	}

	// Min idle connections (default: 5)
	if minConns := os.Getenv("DB_MAX_IDLE_CONNS"); minConns != "" {
		if n, err := strconv.Atoi(minConns); err == nil && n > 0 {
			poolConfig.MinConns = int32(n)
		}
	} else {
		poolConfig.MinConns = 5
	}

	// Connection max lifetime (default: 30 minutes)
	if maxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME"); maxLifetime != "" {
		if d, err := time.ParseDuration(maxLifetime); err == nil {
			poolConfig.MaxConnLifetime = d
		}
	} else {
		poolConfig.MaxConnLifetime = 30 * time.Minute
	}

	// Connection max idle time (default: 10 minutes)
	if maxIdleTime := os.Getenv("DB_CONN_MAX_IDLE_TIME"); maxIdleTime != "" {
		if d, err := time.ParseDuration(maxIdleTime); err == nil {
			poolConfig.MaxConnIdleTime = d
		}
	} else {
		poolConfig.MaxConnIdleTime = 10 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	Pool = pool
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

func Ping() error {
	if Pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return Pool.Ping(context.Background())
}

func RunMigrations() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.New(migrationsPath, connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
