-- ═══════════════════════════════════════════════════════════════
-- Migration 001 — Base Schema (Layer -1: Foundation)
-- Creates the 12 core tables used by the application.
-- Safe to run multiple times (IF NOT EXISTS everywhere).
-- ═══════════════════════════════════════════════════════════════

-- Extensions required by the framework
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ── Base trigger function ────────────────────────────────────────
CREATE OR REPLACE FUNCTION fn_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ── 1. users ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email_encrypted  TEXT        NOT NULL,
    email_hash       TEXT        NOT NULL UNIQUE,
    password_hash    TEXT        NOT NULL,
    salt             TEXT        NOT NULL,
    full_name        TEXT,
    locale           TEXT        NOT NULL DEFAULT 'ro' CHECK (locale IN ('ro','en','ru')),
    mfa_secret       TEXT,
    mfa_enabled      BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email_hash  ON users (email_hash);
CREATE INDEX IF NOT EXISTS idx_users_is_active   ON users (is_active) WHERE is_active = TRUE;

CREATE OR REPLACE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

-- ── 2. user_sessions ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT        NOT NULL UNIQUE,
    device_fp       TEXT,
    ip_subnet       TEXT,
    user_agent_hash TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash  ON user_sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON user_sessions (user_id)
    WHERE revoked = FALSE;
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at  ON user_sessions (expires_at);

-- ── 3. global_objectives (goals) ─────────────────────────────────
CREATE TYPE IF NOT EXISTS goal_status AS ENUM (
    'ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED', 'WAITING'
);

CREATE TABLE IF NOT EXISTS global_objectives (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL CHECK (length(name) BETWEEN 2 AND 200),
    description TEXT,
    status      goal_status NOT NULL DEFAULT 'ACTIVE',
    start_date  DATE        NOT NULL,
    end_date    DATE        NOT NULL CHECK (end_date > start_date),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goals_user_status  ON global_objectives (user_id, status);
CREATE INDEX IF NOT EXISTS idx_goals_status       ON global_objectives (status);
CREATE INDEX IF NOT EXISTS idx_goals_user_active  ON global_objectives (user_id)
    WHERE status = 'ACTIVE';

CREATE OR REPLACE TRIGGER trg_goals_updated_at
    BEFORE UPDATE ON global_objectives
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

-- ── 4. go_metrics ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS go_metrics (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id        UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    metric_key   TEXT        NOT NULL,
    metric_value FLOAT       NOT NULL,
    recorded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_go_metrics_goal   ON go_metrics (go_id);
CREATE INDEX IF NOT EXISTS idx_go_metrics_key    ON go_metrics (go_id, metric_key);

-- ── 5. sprints ────────────────────────────────────────────────────
CREATE TYPE IF NOT EXISTS sprint_status AS ENUM (
    'ACTIVE', 'COMPLETED', 'SKIPPED'
);

CREATE TABLE IF NOT EXISTS sprints (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id         UUID         NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    sprint_number INT          NOT NULL CHECK (sprint_number >= 1),
    start_date    DATE         NOT NULL,
    end_date      DATE         NOT NULL CHECK (end_date > start_date),
    status        sprint_status NOT NULL DEFAULT 'ACTIVE',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (go_id, sprint_number)
);

CREATE INDEX IF NOT EXISTS idx_sprints_goal_active ON sprints (go_id)
    WHERE status = 'ACTIVE';
CREATE INDEX IF NOT EXISTS idx_sprints_status      ON sprints (status);

-- ── 6. sprint_results ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sprint_results (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id   UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE UNIQUE,
    score_value FLOAT       NOT NULL CHECK (score_value BETWEEN 0 AND 1),
    grade       TEXT        NOT NULL CHECK (grade IN ('A+','A','B','C','D')),
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sprint_results_sprint ON sprint_results (sprint_id);

-- ── 7. daily_tasks ────────────────────────────────────────────────
CREATE TYPE IF NOT EXISTS task_type AS ENUM ('MAIN', 'PERSONAL');

CREATE TABLE IF NOT EXISTS daily_tasks (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id    UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    go_id        UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_date    DATE        NOT NULL,
    task_text    TEXT        NOT NULL CHECK (length(task_text) BETWEEN 2 AND 500),
    task_type    task_type   NOT NULL DEFAULT 'MAIN',
    sort_order   INT         NOT NULL DEFAULT 0,
    completed    BOOLEAN     NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_date   ON daily_tasks (user_id, task_date);
CREATE INDEX IF NOT EXISTS idx_tasks_goal_date   ON daily_tasks (go_id, task_date);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint      ON daily_tasks (sprint_id);
CREATE INDEX IF NOT EXISTS idx_tasks_completed   ON daily_tasks (user_id, task_date, completed);

-- ── 8. checkpoints ────────────────────────────────────────────────
CREATE TYPE IF NOT EXISTS checkpoint_status AS ENUM (
    'UPCOMING', 'IN_PROGRESS', 'COMPLETED'
);

CREATE TABLE IF NOT EXISTS checkpoints (
    id           UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id    UUID              NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    name         TEXT              NOT NULL CHECK (length(name) BETWEEN 2 AND 200),
    description  TEXT,
    sort_order   INT               NOT NULL DEFAULT 0,
    status       checkpoint_status NOT NULL DEFAULT 'UPCOMING',
    progress_pct INT               NOT NULL DEFAULT 0 CHECK (progress_pct BETWEEN 0 AND 100),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_sprint  ON checkpoints (sprint_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_checkpoints_status  ON checkpoints (sprint_id, status);

-- ── 9. go_scores ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS go_scores (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id       UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    score_value FLOAT       NOT NULL CHECK (score_value BETWEEN 0 AND 1),
    grade       TEXT        NOT NULL CHECK (grade IN ('A+','A','B','C','D')),
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_go_scores_goal    ON go_scores (go_id, computed_at DESC);

-- ── 10. sprint_reflections ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sprint_reflections (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id    UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    q1_answer    TEXT,
    q2_answer    TEXT,
    energy_level INT         CHECK (energy_level BETWEEN 1 AND 5),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sprint_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_reflections_sprint ON sprint_reflections (sprint_id);
CREATE INDEX IF NOT EXISTS idx_reflections_user   ON sprint_reflections (user_id);

-- ── 11. context_adjustments ──────────────────────────────────────
CREATE TYPE IF NOT EXISTS adj_type AS ENUM (
    'PAUSE', 'ENERGY_LOW', 'ENERGY_HIGH'
);

CREATE TABLE IF NOT EXISTS context_adjustments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id      UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    adj_type   adj_type    NOT NULL,
    start_date DATE        NOT NULL,
    end_date   DATE,
    note       TEXT        CHECK (length(note) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_adjustments_goal   ON context_adjustments (go_id);
CREATE INDEX IF NOT EXISTS idx_adjustments_active ON context_adjustments (go_id, start_date, end_date);

-- ── 12. audit_log ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audit_log (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT        NOT NULL,
    ip_hash    TEXT,
    ua_hash    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_user       ON audit_log (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action     ON audit_log (action);

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 001: Base schema (12 tables) applied successfully';
END $$;
