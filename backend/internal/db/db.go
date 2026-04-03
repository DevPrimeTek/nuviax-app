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

	// Core tables (migration 001)
	coreTables := []string{
		"users", "user_sessions", "global_objectives",
		"go_metrics", "sprints", "sprint_results",
		"daily_tasks", "checkpoints", "go_scores",
		"sprint_reflections", "context_adjustments", "audit_log",
	}

	// Framework tables (migrations 002–006)
	frameworkTables := []string{
		"goal_categories", "sprint_configs", "goal_metadata",         // L1
		"task_executions", "daily_metrics", "sprint_metrics",          // L2
		"behavior_patterns", "consistency_snapshots", "adaptive_weights", // L3
		"regulatory_events", "goal_activation_log", "resource_slots",  // L4
		"growth_milestones", "achievement_badges", "ceremonies", "growth_trajectories", "completion_ceremonies", // L5
	}

	allTables := append(coreTables, frameworkTables...)

	missing := make([]string, 0)
	for _, t := range allTables {
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
		logger.Warn("Missing framework tables — run migrations",
			zap.Strings("tables", missing),
			zap.String("cmd", "psql -f backend/migrations/apply_all.sql"),
		)
		// Only fail if core tables are missing
		for _, t := range coreTables {
			for _, m := range missing {
				if t == m {
					return fmt.Errorf("core table %q missing — run backend/migrations/apply_all.sql", t)
				}
			}
		}
	}

	logger.Info("Database schema verified",
		zap.Int("core_tables", len(coreTables)),
		zap.Int("framework_tables", len(frameworkTables)-len(missing)),
		zap.Int("missing", len(missing)),
	)
	return nil
}

// Healthcheck for /health endpoint
func Healthcheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
