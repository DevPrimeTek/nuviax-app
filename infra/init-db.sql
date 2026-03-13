-- ============================================================
-- NUViaX — Database Schema
-- PostgreSQL 16
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_encrypted TEXT NOT NULL UNIQUE,   -- AES-256 encrypted
    email_hash      TEXT NOT NULL UNIQUE,   -- SHA-256 for lookup
    password_hash   TEXT NOT NULL,          -- bcrypt cost 14
    salt            TEXT NOT NULL,
    full_name       TEXT,
    locale          VARCHAR(5) DEFAULT 'ro',
    mfa_secret      TEXT,                   -- TOTP secret (encrypted)
    mfa_enabled     BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- USER SESSIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,   -- SHA-256 of refresh token
    device_fp       TEXT,                   -- device fingerprint hash
    ip_subnet       TEXT,
    user_agent_hash TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(token_hash);

-- ============================================================
-- GLOBAL OBJECTIVES
-- ============================================================
CREATE TABLE IF NOT EXISTS global_objectives (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    status          VARCHAR(20) DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE','PAUSED','COMPLETED','ARCHIVED','WAITING')),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
    -- NOTE: Nicio formulă, pondere sau parametru intern stocat aici
);
CREATE INDEX idx_go_user ON global_objectives(user_id);
CREATE INDEX idx_go_status ON global_objectives(status);

-- ============================================================
-- GO COMPUTED METRICS (valori opace — fără formule)
-- ============================================================
CREATE TABLE IF NOT EXISTS go_metrics (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    metric_key      VARCHAR(64) NOT NULL,   -- hash al numelui metricii
    metric_value    NUMERIC(10,6),
    computed_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_metrics_go ON go_metrics(go_id);

-- ============================================================
-- SPRINTS (Etape)
-- ============================================================
CREATE TABLE IF NOT EXISTS sprints (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    sprint_number   INTEGER NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    status          VARCHAR(20) DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE','COMPLETED','SKIPPED')),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_sprints_go ON sprints(go_id);

-- ============================================================
-- SPRINT RESULTS
-- ============================================================
CREATE TABLE IF NOT EXISTS sprint_results (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sprint_id       UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    score_value     NUMERIC(5,4),           -- număr opac 0-1
    grade           VARCHAR(2),             -- A/B/C/D
    computed_at     TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- DAILY TASKS (Activități zilnice)
-- ============================================================
CREATE TABLE IF NOT EXISTS daily_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sprint_id       UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_date       DATE NOT NULL,
    task_text       TEXT NOT NULL,
    task_type       VARCHAR(10) NOT NULL
                    CHECK (task_type IN ('MAIN','PERSONAL')),
    sort_order      INTEGER DEFAULT 0,
    completed       BOOLEAN DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_tasks_user_date ON daily_tasks(user_id, task_date);
CREATE INDEX idx_tasks_sprint ON daily_tasks(sprint_id);

-- ============================================================
-- CHECKPOINTS (Milestone-uri)
-- ============================================================
CREATE TABLE IF NOT EXISTS checkpoints (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sprint_id       UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    sort_order      INTEGER DEFAULT 0,
    status          VARCHAR(20) DEFAULT 'UPCOMING'
                    CHECK (status IN ('UPCOMING','IN_PROGRESS','COMPLETED')),
    progress_pct    INTEGER DEFAULT 0 CHECK (progress_pct BETWEEN 0 AND 100),
    completed_at    TIMESTAMPTZ
);
CREATE INDEX idx_checkpoints_sprint ON checkpoints(sprint_id);

-- ============================================================
-- GO OVERALL SCORE
-- ============================================================
CREATE TABLE IF NOT EXISTS go_scores (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    score_value     NUMERIC(5,4),           -- număr opac 0-1
    grade           VARCHAR(2),
    computed_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_scores_go ON go_scores(go_id);

-- ============================================================
-- SPRINT REFLECTIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS sprint_reflections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sprint_id       UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    q1_answer       TEXT,
    q2_answer       TEXT,
    energy_level    SMALLINT CHECK (energy_level BETWEEN 1 AND 10),
    submitted_at    TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- CONTEXT ADJUSTMENTS (Pauze + Energie)
-- ============================================================
CREATE TABLE IF NOT EXISTS context_adjustments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    go_id           UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    adj_type        VARCHAR(20) NOT NULL
                    CHECK (adj_type IN ('PAUSE','ENERGY_LOW','ENERGY_HIGH')),
    start_date      DATE NOT NULL,
    end_date        DATE,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_ctx_go ON context_adjustments(go_id);

-- ============================================================
-- AUDIT LOG
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(100) NOT NULL,
    ip_hash         TEXT,
    ua_hash         TEXT,
    risk_score      SMALLINT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_ts ON audit_log(created_at);

-- ============================================================
-- UPDATE TRIGGER
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_go_updated
    BEFORE UPDATE ON global_objectives
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
