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

// RunMigrations checks core auth tables exist and auto-creates goals tables.
// Auth tables must be created by running backend/migrations/apply_all.sql before first start.
// Goals tables are created idempotently on every startup (IF NOT EXISTS).
func RunMigrations(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Verify core auth tables exist (fail fast if missing — auth tables need manual migration)
	coreTables := []string{
		"users", "user_sessions", "audit_log", "password_reset_tokens",
	}
	for _, t := range coreTables {
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)`,
			t).Scan(&exists); err != nil {
			return fmt.Errorf("check table %s: %w", t, err)
		}
		if !exists {
			return fmt.Errorf("core table %q missing — run backend/migrations/apply_all.sql", t)
		}
	}

	// Auto-create goals tables (idempotent — safe to run on every startup)
	if err := ensureGoalsTables(ctx, pool); err != nil {
		return fmt.Errorf("ensure goals tables: %w", err)
	}

	logger.Info("Database schema verified and ready")
	return nil
}

// ensureGoalsTables creates goals-related tables and types if they don't exist.
// All statements are idempotent (IF NOT EXISTS / EXCEPTION WHEN duplicate_object).
func ensureGoalsTables(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		// Enums
		`DO $$ BEGIN CREATE TYPE behavior_model AS ENUM ('CREATE','INCREASE','REDUCE','MAINTAIN','EVOLVE'); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN CREATE TYPE go_status AS ENUM ('DRAFT','ACTIVE','WAITING','PAUSED','COMPLETED','ARCHIVED'); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN CREATE TYPE sprint_status AS ENUM ('PENDING','ACTIVE','COMPLETED','SUSPENDED'); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN CREATE TYPE task_type AS ENUM ('MAIN','SUPPORT','OPTIONAL'); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN CREATE TYPE task_status AS ENUM ('PENDING','DONE','SKIPPED'); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,

		// global_objectives (C3 max 3 active, C4 max 365 days)
		`CREATE TABLE IF NOT EXISTS global_objectives (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name              TEXT NOT NULL CHECK (length(name) >= 5 AND length(name) <= 200),
			description       TEXT,
			behavior_model    behavior_model NOT NULL,
			domain            TEXT NOT NULL,
			metric            TEXT NOT NULL,
			target_value      NUMERIC(10,2),
			unit              TEXT,
			start_date        DATE NOT NULL,
			end_date          DATE NOT NULL,
			status            go_status NOT NULL DEFAULT 'DRAFT',
			relevance_score   NUMERIC(3,2) CHECK (relevance_score >= 0 AND relevance_score <= 1),
			ai_confidence     NUMERIC(3,2),
			created_at        TIMESTAMPTZ DEFAULT NOW(),
			updated_at        TIMESTAMPTZ DEFAULT NOW(),
			CHECK (end_date > start_date),
			CHECK (end_date - start_date <= 365)
		)`,

		// sprints (C5 fixed 30-day)
		`CREATE TABLE IF NOT EXISTS sprints (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			go_id         UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
			user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			sprint_number INT NOT NULL,
			start_date    DATE NOT NULL,
			end_date      DATE NOT NULL,
			target_value  NUMERIC(10,2),
			status        sprint_status NOT NULL DEFAULT 'PENDING',
			sprint_score  NUMERIC(4,3),
			grade         CHAR(2),
			completed_at  TIMESTAMPTZ,
			created_at    TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (go_id, sprint_number)
		)`,

		// daily_tasks (C23)
		`CREATE TABLE IF NOT EXISTS daily_tasks (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			go_id        UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
			sprint_id    UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			task_date    DATE NOT NULL,
			title        TEXT NOT NULL,
			task_type    task_type NOT NULL DEFAULT 'MAIN',
			status       task_status NOT NULL DEFAULT 'PENDING',
			completed_at TIMESTAMPTZ,
			created_at   TIMESTAMPTZ DEFAULT NOW()
		)`,

		// daily_scores (C24, C25)
		`CREATE TABLE IF NOT EXISTS daily_scores (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			go_id             UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
			user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			sprint_id         UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			score_date        DATE NOT NULL,
			real_progress     NUMERIC(5,4) DEFAULT 0 CHECK (real_progress >= 0 AND real_progress <= 1),
			expected_progress NUMERIC(5,4) DEFAULT 0,
			drift             NUMERIC(6,4),
			tasks_done        INT DEFAULT 0,
			tasks_total       INT DEFAULT 0,
			computed_at       TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (go_id, score_date)
		)`,

		// go_ai_analysis (C9, C10)
		`CREATE TABLE IF NOT EXISTS go_ai_analysis (
			id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			go_id                    UUID REFERENCES global_objectives(id) ON DELETE SET NULL,
			user_id                  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			raw_input                TEXT NOT NULL,
			parsed_domain            TEXT,
			parsed_direction         TEXT,
			parsed_metric            TEXT,
			suggested_behavior_model behavior_model,
			confidence               NUMERIC(3,2),
			ai_feedback              TEXT,
			needs_reformulation      BOOLEAN DEFAULT FALSE,
			created_at               TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Schema migrations for existing tables (idempotent — ADD COLUMN IF NOT EXISTS)
		`ALTER TABLE sprints ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_global_objectives_user_id ON global_objectives (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sprints_go_id             ON sprints (go_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sprints_user_id           ON sprints (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_tasks_user_date     ON daily_tasks (user_id, task_date)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_tasks_go_date       ON daily_tasks (go_id, task_date)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_scores_go_id        ON daily_scores (go_id)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_scores_user_id      ON daily_scores (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_go_ai_analysis_user_id    ON go_ai_analysis (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_go_ai_analysis_go_id      ON go_ai_analysis (go_id)`,
	}

	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("exec statement: %w\nSQL: %.120s", err, stmt)
		}
	}
	logger.Info("Goals tables ready", zap.Int("statements", len(stmts)))
	return nil
}

// Healthcheck for /health endpoint
func Healthcheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
