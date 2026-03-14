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

// RunMigrations checks schema is up to date (init-db.sql runs at container start)
func RunMigrations(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify required tables exist
	tables := []string{
		"users", "user_sessions", "global_objectives",
		"go_metrics", "sprints", "sprint_results",
		"daily_tasks", "checkpoints", "go_scores",
		"sprint_reflections", "context_adjustments", "audit_log",
	}
	for _, t := range tables {
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
			return fmt.Errorf("table %q not found — run init-db.sql first", t)
		}
	}

	logger.Info("Database schema verified", zap.Int("tables", len(tables)))
	return nil
}

// Healthcheck for /health endpoint
func Healthcheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
