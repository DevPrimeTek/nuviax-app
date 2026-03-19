-- ═══════════════════════════════════════════════════════════════
-- Migration 007 — Admin Panel + Critical Gap Fixes
-- Addresses:
--   • GAP #14 — Retroactive pause support (retroactive_pause flag)
--   • GAP #15 — Regression event tracking (value below sprint start)
--   • GAP #20 — Stabilization mode freeze flag on sprints
--   • GAP #8/#13 — ALI current vs projected disambiguation table
--   • Admin infrastructure: is_admin column, admin_stats view,
--     dev reset function, regression_events table
-- Depends on: 006_level5_growth.sql
-- ═══════════════════════════════════════════════════════════════

-- ── 1. Admin flag on users ────────────────────────────────────────
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_users_admin ON users (is_admin) WHERE is_admin = TRUE;

-- ── 2. GAP #14 — retroactive_pause flag on context_adjustments ───
-- context_adjustments may not exist yet under that name; use safe add
ALTER TABLE context_adjustments
    ADD COLUMN IF NOT EXISTS retroactive BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE context_adjustments
    ADD COLUMN IF NOT EXISTS retroactive_reason TEXT;

-- Constraint: retroactive pauses cannot start more than 48h in the past
-- Enforced at application level (retroactive window = 48h).

-- ── 3. GAP #15 — regression_events table ──────────────────────────
CREATE TABLE IF NOT EXISTS regression_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id           UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    sprint_id       UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Value metrics at point of detection
    value_at_detection  NUMERIC(10,4) NOT NULL,
    value_at_sprint_start NUMERIC(10,4) NOT NULL,
    regression_delta    NUMERIC(10,4) NOT NULL, -- always negative
    -- Resolution
    resolved_at     TIMESTAMPTZ,
    resolution_note TEXT,
    UNIQUE (go_id, sprint_id, detected_at::date)
);

CREATE INDEX IF NOT EXISTS idx_regression_go     ON regression_events (go_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_regression_sprint ON regression_events (sprint_id);
CREATE INDEX IF NOT EXISTS idx_regression_user   ON regression_events (user_id, detected_at DESC);

-- ── 4. GAP #20 — stabilization_freeze flag on sprints ─────────────
-- When SRM L3 activates, expected_pct is frozen at the current value
-- to prevent the "drift loop paradox" (expected keeps advancing while
-- user is in stabilization mode).
ALTER TABLE sprints
    ADD COLUMN IF NOT EXISTS expected_pct_frozen  BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE sprints
    ADD COLUMN IF NOT EXISTS frozen_expected_pct  NUMERIC(5,4);  -- null = not frozen

-- ── 5. GAP #8/#13 — ALI snapshot table ────────────────────────────
-- Stores both current and projected ALI values at each computation.
-- Eliminates ambiguity between "what the user has done" vs "what
-- the system projects they'll do" over the rest of the sprint.
CREATE TABLE IF NOT EXISTS ali_snapshots (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    snapshot_date       DATE        NOT NULL,
    -- Per-goal ALI components (JSONB array of {go_id, ali_current, ali_projected})
    goal_ali_breakdown  JSONB       NOT NULL DEFAULT '[]',
    -- Totals
    ali_current         NUMERIC(5,4) NOT NULL,  -- actual ambition load so far
    ali_projected       NUMERIC(5,4) NOT NULL,  -- projected load if pace continues
    -- Ambition Buffer state
    in_ambition_buffer  BOOLEAN     NOT NULL DEFAULT FALSE,
    velocity_control_on BOOLEAN     NOT NULL DEFAULT FALSE,
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_ali_user_date ON ali_snapshots (user_id, snapshot_date DESC);

-- ── 6. Admin statistics view ───────────────────────────────────────
CREATE OR REPLACE VIEW v_admin_platform_stats AS
SELECT
    -- User counts
    (SELECT COUNT(*) FROM users WHERE is_active = TRUE)         AS total_users,
    (SELECT COUNT(*) FROM users WHERE is_admin = TRUE)          AS admin_users,
    (SELECT COUNT(*) FROM users
     WHERE is_active = TRUE
       AND created_at >= NOW() - INTERVAL '7 days')             AS new_users_7d,
    (SELECT COUNT(*) FROM users
     WHERE is_active = TRUE
       AND created_at >= NOW() - INTERVAL '30 days')            AS new_users_30d,
    -- Goal counts
    (SELECT COUNT(*) FROM global_objectives WHERE status = 'ACTIVE')    AS active_goals,
    (SELECT COUNT(*) FROM global_objectives WHERE status = 'COMPLETED') AS completed_goals,
    (SELECT COUNT(*) FROM global_objectives WHERE status = 'PAUSED')    AS paused_goals,
    (SELECT COUNT(*) FROM global_objectives)                             AS total_goals,
    -- Sprint counts
    (SELECT COUNT(*) FROM sprints WHERE status = 'ACTIVE')      AS active_sprints,
    (SELECT COUNT(*) FROM sprints WHERE status = 'COMPLETED')   AS completed_sprints,
    -- Daily activity (last 24h)
    (SELECT COUNT(*) FROM daily_tasks
     WHERE task_date = CURRENT_DATE)                            AS tasks_today,
    (SELECT COUNT(*) FROM daily_tasks
     WHERE task_date = CURRENT_DATE AND completed = TRUE)       AS tasks_completed_today,
    -- SRM events
    (SELECT COUNT(*) FROM srm_events
     WHERE triggered_at >= NOW() - INTERVAL '30 days')         AS srm_events_30d,
    (SELECT COUNT(*) FROM srm_events
     WHERE srm_level = 'L3'
       AND triggered_at >= NOW() - INTERVAL '30 days')         AS srm_l3_events_30d,
    -- Regression events (GAP #15)
    (SELECT COUNT(*) FROM regression_events
     WHERE detected_at >= NOW() - INTERVAL '30 days')          AS regression_events_30d,
    -- Ceremony & achievement stats
    (SELECT COUNT(*) FROM completion_ceremonies
     WHERE generated_at >= NOW() - INTERVAL '30 days')         AS ceremonies_30d,
    (SELECT COUNT(*) FROM achievement_badges
     WHERE awarded_at >= NOW() - INTERVAL '30 days')           AS badges_awarded_30d,
    -- Computed at
    NOW()                                                       AS computed_at;

-- ── 7. Admin user list view ────────────────────────────────────────
CREATE OR REPLACE VIEW v_admin_user_list AS
SELECT
    u.id,
    u.full_name,
    u.locale,
    u.is_active,
    u.is_admin,
    u.mfa_enabled,
    u.created_at,
    u.updated_at,
    -- Goals
    COUNT(DISTINCT g.id) FILTER (WHERE g.status = 'ACTIVE')    AS active_goals,
    COUNT(DISTINCT g.id) FILTER (WHERE g.status = 'COMPLETED') AS completed_goals,
    COUNT(DISTINCT g.id)                                        AS total_goals,
    -- Sprints
    COUNT(DISTINCT s.id) FILTER (WHERE s.status = 'COMPLETED') AS completed_sprints,
    -- Tasks (last 30 days)
    COUNT(DISTINCT dt.id) FILTER (
        WHERE dt.completed = TRUE
          AND dt.task_date >= CURRENT_DATE - 30
    )                                                           AS tasks_last_30d,
    -- Last activity
    MAX(dt.completed_at)                                        AS last_active_at,
    -- Session count
    COUNT(DISTINCT sess.id) FILTER (
        WHERE sess.revoked = FALSE AND sess.expires_at > NOW()
    )                                                           AS active_sessions
FROM users u
LEFT JOIN global_objectives g  ON g.user_id = u.id
LEFT JOIN sprints s            ON s.go_id = g.id
LEFT JOIN daily_tasks dt       ON dt.user_id = u.id
LEFT JOIN user_sessions sess   ON sess.user_id = u.id
GROUP BY u.id, u.full_name, u.locale, u.is_active, u.is_admin,
         u.mfa_enabled, u.created_at, u.updated_at
ORDER BY u.created_at DESC;

-- ── 8. Dev reset function ──────────────────────────────────────────
-- CAUTION: Only callable when APP_ENV = 'development'.
-- Application layer MUST check env before calling this function.
-- Deletes all non-admin user data. Schema is preserved.
CREATE OR REPLACE FUNCTION fn_dev_reset_data(requesting_admin_id UUID)
RETURNS TABLE(deleted_users INT, deleted_goals INT, deleted_tasks INT) AS $$
DECLARE
    v_deleted_users INT;
    v_deleted_goals INT;
    v_deleted_tasks INT;
BEGIN
    -- Verify requester is admin
    IF NOT EXISTS (
        SELECT 1 FROM users WHERE id = requesting_admin_id AND is_admin = TRUE
    ) THEN
        RAISE EXCEPTION 'Access denied: requester is not an admin';
    END IF;

    -- Count before deletion
    SELECT COUNT(*) INTO v_deleted_tasks  FROM daily_tasks
        WHERE user_id NOT IN (SELECT id FROM users WHERE is_admin = TRUE);
    SELECT COUNT(*) INTO v_deleted_goals  FROM global_objectives
        WHERE user_id NOT IN (SELECT id FROM users WHERE is_admin = TRUE);
    SELECT COUNT(*) INTO v_deleted_users  FROM users WHERE is_admin = FALSE;

    -- Delete non-admin data (cascade handles child tables)
    DELETE FROM users WHERE is_admin = FALSE;

    -- Reset sequences and clean orphans
    DELETE FROM regression_events WHERE user_id NOT IN (SELECT id FROM users);
    DELETE FROM ali_snapshots      WHERE user_id NOT IN (SELECT id FROM users);

    RETURN QUERY SELECT v_deleted_users, v_deleted_goals, v_deleted_tasks;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ── 9. Daily ALI snapshot trigger ─────────────────────────────────
-- Logs ALI snapshots automatically when context_adjustments change
-- (application engine inserts directly for initial implementation).

-- ── 10. Audit log entries for admin actions ────────────────────────
-- Reuse existing audit_log table; admin actions use action prefix 'ADMIN_'
-- (e.g., 'ADMIN_DEV_RESET', 'ADMIN_DEACTIVATE_USER', 'ADMIN_PROMOTE_ADMIN')

COMMENT ON TABLE regression_events IS
    'GAP #15 fix: tracks when goal progress value drops below sprint start value';

COMMENT ON COLUMN sprints.expected_pct_frozen IS
    'GAP #20 fix: TRUE when SRM L3 stabilization freezes the expected progress trajectory';

COMMENT ON COLUMN sprints.frozen_expected_pct IS
    'GAP #20 fix: the expected_pct value at the moment of L3 freeze';

COMMENT ON TABLE ali_snapshots IS
    'GAP #8/#13 fix: stores ALI_current vs ALI_projected to eliminate metric ambiguity';

COMMENT ON COLUMN context_adjustments.retroactive IS
    'GAP #14 fix: TRUE when pause was logged after it started (max 48h retroactive window)';
