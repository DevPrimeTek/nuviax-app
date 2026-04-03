-- 001_schema.sql — NuviaX MVP Database Schema
-- Framework: NuviaX Growth Framework Rev 5.6
-- Idempotent: safe to re-run (all IF NOT EXISTS)

-- ============================================================
-- 1. users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    email_hash      TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT,
    is_admin        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 2. sessions
-- ============================================================
CREATE TABLE IF NOT EXISTS sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash  TEXT UNIQUE NOT NULL,
    fingerprint         TEXT,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 3. password_reset_tokens
-- ============================================================
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT UNIQUE NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 4. audit_log
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    ip_hash     TEXT,
    ua_hash     TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 5. ENUM behavior_model (C2 — 5 Behavior Models)
-- ============================================================
DO $$ BEGIN
    CREATE TYPE behavior_model AS ENUM (
        'CREATE', 'INCREASE', 'REDUCE', 'MAINTAIN', 'EVOLVE'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- 6. ENUM go_status
-- ============================================================
DO $$ BEGIN
    CREATE TYPE go_status AS ENUM (
        'DRAFT', 'ACTIVE', 'WAITING', 'PAUSED', 'COMPLETED', 'ARCHIVED'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- 7. global_objectives (C3 max 3 active, C4 max 365 days)
-- ============================================================
CREATE TABLE IF NOT EXISTS global_objectives (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                TEXT NOT NULL CHECK (length(name) >= 5 AND length(name) <= 200),
    description         TEXT,
    behavior_model      behavior_model NOT NULL,
    domain              TEXT NOT NULL,
    metric              TEXT NOT NULL,
    target_value        NUMERIC(10,2),
    unit                TEXT,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    status              go_status NOT NULL DEFAULT 'DRAFT',
    relevance_score     NUMERIC(3,2) CHECK (relevance_score >= 0 AND relevance_score <= 1),
    ai_confidence       NUMERIC(3,2),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    CHECK (end_date > start_date),
    CHECK (end_date - start_date <= 365)
);

-- ============================================================
-- 8. ENUM sprint_status (C19 — Sprint statuses)
-- ============================================================
DO $$ BEGIN
    CREATE TYPE sprint_status AS ENUM (
        'PENDING', 'ACTIVE', 'COMPLETED', 'SUSPENDED'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- 9. sprints (C5 fixed 30-day, C19 Sprint Structuring)
-- ============================================================
CREATE TABLE IF NOT EXISTS sprints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sprint_number   INT NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    target_value    NUMERIC(10,2),
    status          sprint_status NOT NULL DEFAULT 'PENDING',
    sprint_score    NUMERIC(4,3),
    grade           CHAR(2),
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (go_id, sprint_number)
);

-- ============================================================
-- 10. ENUM task_type (C23 — Core/Optional Stack)
-- ============================================================
DO $$ BEGIN
    CREATE TYPE task_type AS ENUM (
        'MAIN', 'SUPPORT', 'OPTIONAL'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- 11. ENUM task_status
-- ============================================================
DO $$ BEGIN
    CREATE TYPE task_status AS ENUM (
        'PENDING', 'DONE', 'SKIPPED'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- 12. daily_tasks (C23 Daily Stack Generator)
-- ============================================================
CREATE TABLE IF NOT EXISTS daily_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    sprint_id       UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_date       DATE NOT NULL,
    title           TEXT NOT NULL,
    task_type       task_type NOT NULL DEFAULT 'MAIN',
    status          task_status NOT NULL DEFAULT 'PENDING',
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_daily_tasks_user_date ON daily_tasks (user_id, task_date);
CREATE INDEX IF NOT EXISTS idx_daily_tasks_go_date   ON daily_tasks (go_id, task_date);

-- ============================================================
-- 13. daily_scores (C24 Progress Computation, C25 Variance)
-- ============================================================
CREATE TABLE IF NOT EXISTS daily_scores (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id               UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sprint_id           UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    score_date          DATE NOT NULL,
    real_progress       NUMERIC(5,4) DEFAULT 0 CHECK (real_progress >= 0 AND real_progress <= 1),
    expected_progress   NUMERIC(5,4) DEFAULT 0,
    drift               NUMERIC(6,4),
    tasks_done          INT DEFAULT 0,
    tasks_total         INT DEFAULT 0,
    computed_at         TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (go_id, score_date)
);

-- ============================================================
-- 14. go_ai_analysis (C9 Semantic Parsing, C10 BM Classification)
-- ============================================================
CREATE TABLE IF NOT EXISTS go_ai_analysis (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id                       UUID REFERENCES global_objectives(id) ON DELETE SET NULL,
    user_id                     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    raw_input                   TEXT NOT NULL,
    parsed_domain               TEXT,
    parsed_direction            TEXT,
    parsed_metric               TEXT,
    suggested_behavior_model    behavior_model,
    confidence                  NUMERIC(3,2),
    ai_feedback                 TEXT,
    needs_reformulation         BOOLEAN DEFAULT FALSE,
    created_at                  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Additional indexes for query performance
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_sessions_user_id           ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_user_id     ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id          ON audit_log (user_id);
CREATE INDEX IF NOT EXISTS idx_global_objectives_user_id  ON global_objectives (user_id);
CREATE INDEX IF NOT EXISTS idx_sprints_go_id              ON sprints (go_id);
CREATE INDEX IF NOT EXISTS idx_sprints_user_id            ON sprints (user_id);
CREATE INDEX IF NOT EXISTS idx_daily_scores_go_id         ON daily_scores (go_id);
CREATE INDEX IF NOT EXISTS idx_daily_scores_user_id       ON daily_scores (user_id);
CREATE INDEX IF NOT EXISTS idx_go_ai_analysis_user_id     ON go_ai_analysis (user_id);
CREATE INDEX IF NOT EXISTS idx_go_ai_analysis_go_id       ON go_ai_analysis (go_id);
