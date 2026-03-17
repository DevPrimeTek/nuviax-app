-- ═══════════════════════════════════════════════════════════════
-- Migration 004 — Level 3: Adaptive Intelligence (C26-C31)
-- Adds: behavior_patterns, consistency_snapshots, adaptive_weights
-- Views: v_user_consistency, v_behavior_summary, v_context_impact,
--        v_weekly_summary, v_consistency_trend
-- Functions: fn_get_user_consistency, fn_detect_behavior_pattern
-- Triggers: trg_consistency_snapshot_weekly, trg_behavior_detect
-- Depends on: 003_level2_execution.sql
-- ═══════════════════════════════════════════════════════════════

-- ── TABLES ───────────────────────────────────────────────────────

-- Table 19: behavior_patterns — detected behavioral patterns per user/goal
CREATE TYPE IF NOT EXISTS behavior_pattern_type AS ENUM (
    'MORNING_PERSON',       -- completes tasks before 12:00
    'EVENING_PERSON',       -- completes tasks after 18:00
    'WEEKEND_WARRIOR',      -- more productive on weekends
    'WEEKDAY_FOCUSED',      -- more productive on weekdays
    'SPRINT_STARTER',       -- high output at sprint start, fades
    'SPRINT_CLOSER',        -- low output start, high at end
    'CONSISTENT',           -- stable day-to-day output
    'BURST_WORKER'          -- alternating high/low activity
);

CREATE TABLE IF NOT EXISTS behavior_patterns (
    id              UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID                  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    go_id           UUID                  REFERENCES global_objectives(id) ON DELETE CASCADE,
    pattern_type    behavior_pattern_type NOT NULL,
    strength        FLOAT                 NOT NULL DEFAULT 0.5 CHECK (strength BETWEEN 0 AND 1),
    sample_days     INT                   NOT NULL DEFAULT 0,  -- days of data used
    detected_at     TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,                               -- NULL = permanent
    UNIQUE (user_id, go_id, pattern_type)
);

CREATE INDEX IF NOT EXISTS idx_behavior_user    ON behavior_patterns (user_id);
CREATE INDEX IF NOT EXISTS idx_behavior_goal    ON behavior_patterns (go_id);
CREATE INDEX IF NOT EXISTS idx_behavior_active  ON behavior_patterns (user_id)
    WHERE expires_at IS NULL OR expires_at > NOW();

-- Table 20: consistency_snapshots — weekly consistency readings per goal
CREATE TABLE IF NOT EXISTS consistency_snapshots (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id             UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start        DATE        NOT NULL,
    active_days       INT         NOT NULL DEFAULT 0 CHECK (active_days BETWEEN 0 AND 7),
    total_days        INT         NOT NULL DEFAULT 7,
    consistency_score FLOAT       NOT NULL DEFAULT 0 CHECK (consistency_score BETWEEN 0 AND 1),
    tasks_completed   INT         NOT NULL DEFAULT 0,
    recorded_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (go_id, week_start)
);

CREATE INDEX IF NOT EXISTS idx_consistency_goal   ON consistency_snapshots (go_id, week_start DESC);
CREATE INDEX IF NOT EXISTS idx_consistency_user   ON consistency_snapshots (user_id, week_start DESC);

-- Table 21: adaptive_weights — per-user weight customization (future use)
-- Stored but not yet applied — framework reserves these for personalization
CREATE TABLE IF NOT EXISTS adaptive_weights (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight_key   TEXT        NOT NULL,   -- e.g. 'completion_rate', 'consistency'
    weight_value FLOAT       NOT NULL CHECK (weight_value BETWEEN 0 AND 1),
    reason       TEXT        CHECK (length(reason) <= 200),
    applied_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_to   TIMESTAMPTZ,            -- NULL = current
    UNIQUE (user_id, weight_key, applied_from)
);

CREATE INDEX IF NOT EXISTS idx_adaptive_weights_user ON adaptive_weights (user_id, weight_key);

-- ── FUNCTION: fn_get_user_consistency ────────────────────────────
-- Returns consistency score (0-1) for a user over N days
CREATE OR REPLACE FUNCTION fn_get_user_consistency(
    p_user_id UUID,
    p_days    INT DEFAULT 30
)
RETURNS FLOAT AS $$
DECLARE
    v_active_days INT;
    v_total_days  INT;
BEGIN
    SELECT
        COUNT(DISTINCT task_date) FILTER (WHERE BOOL_OR(completed) = TRUE),
        COUNT(DISTINCT task_date)
    INTO v_active_days, v_total_days
    FROM daily_tasks
    WHERE user_id = p_user_id
      AND task_type = 'MAIN'
      AND task_date >= CURRENT_DATE - p_days
      AND task_date <= CURRENT_DATE;

    IF v_total_days = 0 THEN RETURN 0; END IF;
    RETURN v_active_days::FLOAT / v_total_days;
END;
$$ LANGUAGE plpgsql STABLE;

-- ── FUNCTION: fn_detect_behavior_pattern ─────────────────────────
-- Analyzes recent task completion times and updates behavior_patterns
-- Runs as part of the 90-day recalibration job
CREATE OR REPLACE FUNCTION fn_detect_behavior_pattern(
    p_user_id UUID,
    p_go_id   UUID DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_total       INT;
    v_weekdays    INT;
    v_weekends    INT;
    v_strength    FLOAT;
    v_pattern     behavior_pattern_type;
BEGIN
    -- Count task completions by weekday vs weekend (last 30 days)
    SELECT
        COUNT(*) FILTER (WHERE EXTRACT(DOW FROM task_date) NOT IN (0, 6)),
        COUNT(*) FILTER (WHERE EXTRACT(DOW FROM task_date) IN (0, 6)),
        COUNT(*)
    INTO v_weekdays, v_weekends, v_total
    FROM daily_tasks
    WHERE user_id = p_user_id
      AND (p_go_id IS NULL OR go_id = p_go_id)
      AND task_type = 'MAIN'
      AND completed = TRUE
      AND task_date >= CURRENT_DATE - 30;

    IF v_total < 5 THEN RETURN; END IF;  -- Not enough data

    IF v_weekdays::FLOAT / NULLIF(v_total, 0) >= 0.80 THEN
        v_pattern  := 'WEEKDAY_FOCUSED';
        v_strength := v_weekdays::FLOAT / v_total;
    ELSIF v_weekends::FLOAT / NULLIF(v_total, 0) >= 0.50 THEN
        v_pattern  := 'WEEKEND_WARRIOR';
        v_strength := v_weekends::FLOAT / v_total;
    ELSE
        v_pattern  := 'CONSISTENT';
        v_strength := fn_get_user_consistency(p_user_id, 30);
    END IF;

    INSERT INTO behavior_patterns (user_id, go_id, pattern_type, strength, sample_days)
    VALUES (p_user_id, p_go_id, v_pattern, v_strength, v_total)
    ON CONFLICT (user_id, go_id, pattern_type) DO UPDATE SET
        strength    = EXCLUDED.strength,
        sample_days = EXCLUDED.sample_days,
        detected_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- ── TRIGGER: weekly consistency snapshot (runs via scheduler job) ──
-- This function is called by fn_snapshot_daily_metrics at week boundary
CREATE OR REPLACE FUNCTION fn_snapshot_weekly_consistency()
RETURNS TRIGGER AS $$
DECLARE
    v_week_start    DATE;
    v_active_days   INT;
    v_tasks_done    INT;
    v_score         FLOAT;
BEGIN
    -- Only snapshot at week boundary (Monday)
    v_week_start := DATE_TRUNC('week', NEW.metric_date)::DATE;

    SELECT
        COUNT(*) FILTER (WHERE completion_rate > 0),
        COALESCE(SUM(tasks_done), 0)
    INTO v_active_days, v_tasks_done
    FROM daily_metrics
    WHERE go_id = NEW.go_id
      AND metric_date >= v_week_start
      AND metric_date <= NEW.metric_date;

    v_score := LEAST(v_active_days::FLOAT / 7, 1);

    INSERT INTO consistency_snapshots
        (go_id, user_id, week_start, active_days, consistency_score, tasks_completed)
    VALUES
        (NEW.go_id, NEW.user_id, v_week_start, v_active_days, v_score, v_tasks_done)
    ON CONFLICT (go_id, week_start) DO UPDATE SET
        active_days       = EXCLUDED.active_days,
        consistency_score = EXCLUDED.consistency_score,
        tasks_completed   = EXCLUDED.tasks_completed,
        recorded_at       = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_consistency_snapshot
    AFTER INSERT OR UPDATE ON daily_metrics
    FOR EACH ROW EXECUTE FUNCTION fn_snapshot_weekly_consistency();

-- ── VIEWS ────────────────────────────────────────────────────────

-- View 10: v_user_consistency — consistency metrics per user (7d, 30d)
CREATE OR REPLACE VIEW v_user_consistency AS
SELECT
    u.id                                           AS user_id,
    fn_get_user_consistency(u.id, 7)               AS consistency_7d,
    fn_get_user_consistency(u.id, 30)              AS consistency_30d,
    COALESCE(st.streak_days, 0)                    AS current_streak,
    COALESCE(st.streak_start, CURRENT_DATE)        AS streak_since
FROM users u
LEFT JOIN v_task_streaks st ON st.user_id = u.id
WHERE u.is_active = TRUE;

-- View 11: v_behavior_summary — strongest pattern per user/goal
CREATE OR REPLACE VIEW v_behavior_summary AS
SELECT DISTINCT ON (bp.user_id, bp.go_id)
    bp.user_id,
    bp.go_id,
    g.name    AS goal_name,
    bp.pattern_type,
    bp.strength,
    bp.detected_at
FROM behavior_patterns bp
LEFT JOIN global_objectives g ON g.id = bp.go_id
WHERE (bp.expires_at IS NULL OR bp.expires_at > NOW())
ORDER BY bp.user_id, bp.go_id, bp.strength DESC;

-- View 12: v_context_impact — how context adjustments affect scoring
CREATE OR REPLACE VIEW v_context_impact AS
SELECT
    ca.go_id,
    g.user_id,
    ca.adj_type,
    ca.start_date,
    ca.end_date,
    COUNT(dm.metric_date)                                             AS days_in_adjustment,
    AVG(dm.completion_rate)                                           AS avg_completion_during,
    AVG(dm.completion_rate) FILTER (
        WHERE dm.metric_date < ca.start_date
          AND dm.metric_date >= ca.start_date - 14
    )                                                                 AS avg_completion_before
FROM context_adjustments ca
JOIN global_objectives g ON g.id = ca.go_id
LEFT JOIN daily_metrics dm ON dm.go_id = ca.go_id
    AND dm.metric_date BETWEEN ca.start_date
        AND COALESCE(ca.end_date, CURRENT_DATE)
GROUP BY ca.go_id, g.user_id, ca.adj_type, ca.start_date, ca.end_date;

-- View 13: v_weekly_summary — weekly performance per goal (last 4 weeks)
CREATE OR REPLACE VIEW v_weekly_summary AS
SELECT
    cs.go_id,
    cs.user_id,
    g.name                          AS goal_name,
    cs.week_start,
    cs.week_start + 6               AS week_end,
    cs.active_days,
    cs.consistency_score,
    cs.tasks_completed,
    fn_grade_from_score(cs.consistency_score) AS weekly_grade
FROM consistency_snapshots cs
JOIN global_objectives g ON g.id = cs.go_id
WHERE cs.week_start >= CURRENT_DATE - 28
ORDER BY cs.go_id, cs.week_start DESC;

-- View 14: v_consistency_trend — 4-week consistency trend per goal
CREATE OR REPLACE VIEW v_consistency_trend AS
SELECT
    go_id,
    user_id,
    AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 7)  AS last_week,
    AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 14
                                     AND week_start < CURRENT_DATE - 7)   AS week_minus_1,
    AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 21
                                     AND week_start < CURRENT_DATE - 14)  AS week_minus_2,
    AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 28
                                     AND week_start < CURRENT_DATE - 21)  AS week_minus_3,
    CASE
        WHEN AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 7)
             > AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 28
                                                AND week_start < CURRENT_DATE - 7)
        THEN 'IMPROVING'
        WHEN AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 7)
             < AVG(consistency_score) FILTER (WHERE week_start >= CURRENT_DATE - 28
                                                AND week_start < CURRENT_DATE - 7)
        THEN 'DECLINING'
        ELSE 'STABLE'
    END                                                                   AS trend
FROM consistency_snapshots
WHERE week_start >= CURRENT_DATE - 28
GROUP BY go_id, user_id;

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 004: Level 3 (3 tables, 5 views, 2 fn, 1 trigger) applied';
END $$;
