package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Connect creates and validates a PostgreSQL connection pool
func Connect(databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Pool settings
	cfg.MaxConns = 20
	cfg.MinConns = 3
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	logger.Info("PostgreSQL connected",
		zap.Int32("max_conns", cfg.MaxConns),
	)
	return pool, nil
}

// RunMigrations checks schema is up to date.
// Tables are created by running backend/migrations/apply_all.sql before starting the server.
func RunMigrations(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Auth-only core tables (reset foundation)
	coreTables := []string{
		"users", "user_sessions", "audit_log", "password_reset_tokens",
	}

	missing := make([]string, 0)
	for _, t := range coreTables {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, t).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check table %s: %w", t, err)
		}
		if !exists {
			missing = append(missing, t)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("core table(s) missing: %v — run backend/migrations/apply_all.sql", missing)
	}

	logger.Info("Database schema verified",
		zap.Int("core_tables", len(coreTables)),
	)
	return nil
}

// Healthcheck for /health endpoint
func Healthcheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
