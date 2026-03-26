-- ═══════════════════════════════════════════════════════════════
-- Migration 010 — P1 Gap Fixes (Sprint 2)
-- Adds tables required by P1 stress test gap implementations:
--   • srm_events           — SRM L1/L2/L3 event log (G-3, G-12)
--   • reactivation_protocols — Post-SRM L3 recovery ramp (G-7)
--   • stagnation_events    — Consecutive-days stagnation log (G-5)
-- Depends on: 009_password_reset.sql
-- ═══════════════════════════════════════════════════════════════

-- ── 1. srm_events — Strategic Reset Management event log (G-3, G-12) ─────
CREATE TYPE IF NOT EXISTS srm_level_type AS ENUM ('L1', 'L2', 'L3');

CREATE TABLE IF NOT EXISTS srm_events (
    id             UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id          UUID           NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    srm_level      srm_level_type NOT NULL,
    trigger_reason TEXT           NOT NULL DEFAULT 'system',
    triggered_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    -- Confirmation tracking for L2 (single confirm) and L3 (double confirm)
    confirmed_at   TIMESTAMPTZ,
    confirmed_by   UUID           REFERENCES users(id) ON DELETE SET NULL,
    -- Revocation: when the SRM event is resolved or overridden
    revoked_at     TIMESTAMPTZ,
    revoke_reason  TEXT
);

CREATE INDEX IF NOT EXISTS idx_srm_events_goal ON srm_events (go_id, triggered_at DESC);
CREATE INDEX IF NOT EXISTS idx_srm_events_active ON srm_events (go_id)
    WHERE revoked_at IS NULL;

-- ── 2. reactivation_protocols — Post-SRM L3 recovery ramp (G-7, C36) ─────
-- After SRM L3 stabilization, intensity ramps up over ~8 days (0.2 → 1.0).
-- Each day the scheduler increments current_day and current_intensity.
CREATE TABLE IF NOT EXISTS reactivation_protocols (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id                   UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    current_day             INT         NOT NULL DEFAULT 1 CHECK (current_day >= 1),
    current_intensity       FLOAT       NOT NULL DEFAULT 0.2 CHECK (current_intensity BETWEEN 0.1 AND 1.5),
    -- SRM adjustments during reactivation
    srm1_disabled           BOOLEAN     NOT NULL DEFAULT TRUE,   -- L1 suppressed during ramp
    srm2_threshold_adjusted BOOLEAN     NOT NULL DEFAULT TRUE,   -- L2 threshold raised during ramp
    started_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ,                         -- NULL = still ramping
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (go_id)  -- one active protocol per goal at a time
);

CREATE INDEX IF NOT EXISTS idx_reactivation_active ON reactivation_protocols (go_id)
    WHERE completed_at IS NULL;

-- ── 3. stagnation_events — Stagnation detection log (G-5) ─────────────────
-- Recorded when ConsecutiveInactiveDays >= 5 for a goal.
-- Used by Focus Rotation (G-2) to prioritize task generation.
CREATE TABLE IF NOT EXISTS stagnation_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id           UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inactive_days   INT         NOT NULL CHECK (inactive_days >= 1),
    resolved_at     TIMESTAMPTZ,  -- set when a task is completed for this goal
    UNIQUE (go_id, detected_at::date)
);

CREATE INDEX IF NOT EXISTS idx_stagnation_goal ON stagnation_events (go_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_stagnation_open ON stagnation_events (go_id)
    WHERE resolved_at IS NULL;

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 010: P1 gaps (srm_events, reactivation_protocols, stagnation_events) applied';
END $$;
