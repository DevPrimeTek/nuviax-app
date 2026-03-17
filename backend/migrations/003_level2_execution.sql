-- ═══════════════════════════════════════════════════════════════
-- Migration 003 — Level 2: Execution Engine (C19-C25)
-- Adds: task_executions, daily_metrics, sprint_metrics
-- Views: v_today_tasks, v_sprint_execution_stats,
--        v_daily_completion_rate, v_task_streaks, v_sprint_progress
-- Functions: fn_compute_sprint_score, fn_compute_goal_score
-- Triggers: trg_daily_metrics_snapshot, trg_sprint_metrics_update
-- Depends on: 002_layer0_level1.sql
-- ═══════════════════════════════════════════════════════════════

-- ── TABLES ───────────────────────────────────────────────────────

-- Table 16: task_executions — fine-grained execution log per task
-- Opaque from the API — used only by the engine for quality signals
CREATE TABLE IF NOT EXISTS task_executions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id       UUID        NOT NULL REFERENCES daily_tasks(id) ON DELETE CASCADE UNIQUE,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quality_score FLOAT       CHECK (quality_score BETWEEN 0 AND 1), -- NULL = not rated
    duration_min  INT         CHECK (duration_min >= 0),
    notes         TEXT        CHECK (length(notes) <= 500),
    executed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_exec_user    ON task_executions (user_id);
CREATE INDEX IF NOT EXISTS idx_task_exec_task    ON task_executions (task_id);

-- Table 17: daily_metrics — daily completion snapshot per goal (opaque)
CREATE TABLE IF NOT EXISTS daily_metrics (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id            UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metric_date      DATE        NOT NULL,
    tasks_total      INT         NOT NULL DEFAULT 0,
    tasks_done       INT         NOT NULL DEFAULT 0,
    completion_rate  FLOAT       NOT NULL DEFAULT 0 CHECK (completion_rate BETWEEN 0 AND 1),
    intensity_used   FLOAT       NOT NULL DEFAULT 1.0,
    recorded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (go_id, metric_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_metrics_goal  ON daily_metrics (go_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_daily_metrics_user  ON daily_metrics (user_id, metric_date DESC);

-- Table 18: sprint_metrics — per-sprint engine computation snapshot (opaque)
CREATE TABLE IF NOT EXISTS sprint_metrics (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id        UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE UNIQUE,
    completion_rate  FLOAT       NOT NULL DEFAULT 0 CHECK (completion_rate BETWEEN 0 AND 1),
    consistency_score FLOAT      NOT NULL DEFAULT 0 CHECK (consistency_score BETWEEN 0 AND 1),
    context_penalty  FLOAT       NOT NULL DEFAULT 0 CHECK (context_penalty BETWEEN 0 AND 1),
    energy_bonus     FLOAT       NOT NULL DEFAULT 0 CHECK (energy_bonus BETWEEN 0 AND 1),
    final_score      FLOAT       NOT NULL DEFAULT 0 CHECK (final_score BETWEEN 0 AND 1),
    computed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sprint_metrics_sprint ON sprint_metrics (sprint_id);

-- ── FUNCTION: fn_compute_sprint_score ────────────────────────────
-- Mirrors engine/level2_execution.go: computeSprintInternal
-- Returns opaque 0-1 score
CREATE OR REPLACE FUNCTION fn_compute_sprint_score(p_sprint_id UUID)
RETURNS FLOAT AS $$
DECLARE
    v_total     INT;
    v_completed INT;
BEGIN
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE completed = TRUE)
    INTO v_total, v_completed
    FROM daily_tasks
    WHERE sprint_id = p_sprint_id AND task_type = 'MAIN';

    IF v_total = 0 THEN RETURN 0; END IF;
    RETURN LEAST(GREATEST(v_completed::FLOAT / v_total, 0), 1);
END;
$$ LANGUAGE plpgsql STABLE;

-- ── FUNCTION: fn_compute_goal_score ──────────────────────────────
-- Mirrors engine/engine.go: computeInternalMetrics (simplified SQL version)
-- Full computation stays in Go engine — this is for DB-side reporting
CREATE OR REPLACE FUNCTION fn_compute_goal_score(p_go_id UUID)
RETURNS FLOAT AS $$
DECLARE
    v_total          INT;
    v_completed      INT;
    v_active_days    INT;
    v_total_days     INT;
    v_completion     FLOAT;
    v_consistency    FLOAT;
    v_score          FLOAT;
BEGIN
    -- Completion rate
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE dt.completed = TRUE)
    INTO v_total, v_completed
    FROM daily_tasks dt
    JOIN sprints s ON s.id = dt.sprint_id
    WHERE s.go_id = p_go_id
      AND dt.task_type = 'MAIN'
      AND dt.task_date <= CURRENT_DATE;

    IF v_total = 0 THEN RETURN 0; END IF;
    v_completion := v_completed::FLOAT / v_total;

    -- Consistency (days with at least one completed task)
    SELECT
        COUNT(DISTINCT task_date) FILTER (WHERE completed = TRUE),
        COUNT(DISTINCT task_date)
    INTO v_active_days, v_total_days
    FROM daily_tasks
    WHERE go_id = p_go_id
      AND task_type = 'MAIN'
      AND task_date <= CURRENT_DATE;

    v_consistency := CASE WHEN v_total_days > 0
        THEN v_active_days::FLOAT / v_total_days ELSE 0 END;

    -- Composite score (same weights as engine — opaque)
    v_score := LEAST(GREATEST(v_completion * 0.65 + v_consistency * 0.35, 0), 1);
    RETURN v_score;
END;
$$ LANGUAGE plpgsql STABLE;

-- ── TRIGGER: snapshot daily_metrics after task completion ─────────
CREATE OR REPLACE FUNCTION fn_snapshot_daily_metrics()
RETURNS TRIGGER AS $$
DECLARE
    v_total    INT;
    v_done     INT;
    v_rate     FLOAT;
    v_goal_id  UUID;
BEGIN
    -- Only trigger on task completion
    IF NEW.completed = FALSE OR OLD.completed = TRUE THEN
        RETURN NEW;
    END IF;

    -- Get goal_id for this task
    v_goal_id := NEW.go_id;

    SELECT COUNT(*), COUNT(*) FILTER (WHERE completed = TRUE)
    INTO v_total, v_done
    FROM daily_tasks
    WHERE go_id = v_goal_id
      AND task_date = NEW.task_date
      AND task_type = 'MAIN';

    v_rate := CASE WHEN v_total > 0 THEN v_done::FLOAT / v_total ELSE 0 END;

    INSERT INTO daily_metrics
        (go_id, user_id, metric_date, tasks_total, tasks_done, completion_rate)
    VALUES
        (v_goal_id, NEW.user_id, NEW.task_date, v_total, v_done, v_rate)
    ON CONFLICT (go_id, metric_date) DO UPDATE SET
        tasks_total     = EXCLUDED.tasks_total,
        tasks_done      = EXCLUDED.tasks_done,
        completion_rate = EXCLUDED.completion_rate,
        recorded_at     = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_daily_metrics_snapshot
    AFTER UPDATE OF completed ON daily_tasks
    FOR EACH ROW EXECUTE FUNCTION fn_snapshot_daily_metrics();

-- ── TRIGGER: update sprint_metrics after task completion ──────────
CREATE OR REPLACE FUNCTION fn_update_sprint_metrics()
RETURNS TRIGGER AS $$
DECLARE
    v_score FLOAT;
BEGIN
    IF NEW.completed = FALSE OR OLD.completed = TRUE THEN
        RETURN NEW;
    END IF;

    v_score := fn_compute_sprint_score(NEW.sprint_id);

    INSERT INTO sprint_metrics
        (sprint_id, completion_rate, final_score)
    VALUES
        (NEW.sprint_id, v_score, v_score)
    ON CONFLICT (sprint_id) DO UPDATE SET
        completion_rate = EXCLUDED.completion_rate,
        final_score     = EXCLUDED.final_score,
        computed_at     = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_sprint_metrics_update
    AFTER UPDATE OF completed ON daily_tasks
    FOR EACH ROW EXECUTE FUNCTION fn_update_sprint_metrics();

-- ── VIEWS ────────────────────────────────────────────────────────

-- View 5: v_today_tasks — today's tasks with goal/sprint context
CREATE OR REPLACE VIEW v_today_tasks AS
SELECT
    dt.id,
    dt.user_id,
    dt.go_id            AS goal_id,
    g.name              AS goal_name,
    dt.sprint_id,
    dt.task_date,
    dt.task_text        AS text,
    dt.task_type        AS type,
    dt.sort_order,
    dt.completed,
    dt.completed_at
FROM daily_tasks dt
JOIN global_objectives g ON g.id = dt.go_id
WHERE dt.task_date = CURRENT_DATE;

-- View 6: v_sprint_execution_stats — per-sprint execution breakdown
CREATE OR REPLACE VIEW v_sprint_execution_stats AS
SELECT
    s.id                                                          AS sprint_id,
    s.go_id,
    g.user_id,
    s.sprint_number,
    COUNT(dt.id)                                                  AS total_tasks,
    COUNT(dt.id) FILTER (WHERE dt.completed = TRUE)               AS completed_tasks,
    ROUND(
        (COUNT(dt.id) FILTER (WHERE dt.completed = TRUE))::NUMERIC
        / NULLIF(COUNT(dt.id), 0) * 100, 1
    )                                                             AS completion_pct,
    COUNT(DISTINCT dt.task_date)                                  AS active_days,
    COUNT(DISTINCT dt.task_date) FILTER (WHERE dt.completed = TRUE) AS productive_days,
    COALESCE(sm.final_score, fn_compute_sprint_score(s.id))      AS score,
    fn_grade_from_score(
        COALESCE(sm.final_score, fn_compute_sprint_score(s.id))
    )                                                             AS grade
FROM sprints s
JOIN global_objectives g ON g.id = s.go_id
LEFT JOIN daily_tasks dt ON dt.sprint_id = s.id AND dt.task_type = 'MAIN'
LEFT JOIN sprint_metrics sm ON sm.sprint_id = s.id
GROUP BY s.id, s.go_id, g.user_id, s.sprint_number, sm.final_score;

-- View 7: v_daily_completion_rate — rolling 30d completion rate per goal
CREATE OR REPLACE VIEW v_daily_completion_rate AS
SELECT
    dm.go_id,
    g.user_id,
    dm.metric_date,
    dm.completion_rate,
    dm.tasks_total,
    dm.tasks_done,
    AVG(dm.completion_rate) OVER (
        PARTITION BY dm.go_id
        ORDER BY dm.metric_date
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    )                      AS rolling_7d_avg,
    AVG(dm.completion_rate) OVER (
        PARTITION BY dm.go_id
        ORDER BY dm.metric_date
        ROWS BETWEEN 29 PRECEDING AND CURRENT ROW
    )                      AS rolling_30d_avg
FROM daily_metrics dm
JOIN global_objectives g ON g.id = dm.go_id
WHERE dm.metric_date >= CURRENT_DATE - 30;

-- View 8: v_task_streaks — consecutive productive days per user
CREATE OR REPLACE VIEW v_task_streaks AS
WITH daily_completion AS (
    SELECT
        user_id,
        task_date,
        BOOL_OR(completed)   AS had_completion
    FROM daily_tasks
    WHERE task_type = 'MAIN'
      AND task_date <= CURRENT_DATE
    GROUP BY user_id, task_date
),
ordered AS (
    SELECT
        user_id,
        task_date,
        had_completion,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY task_date DESC) AS rn,
        task_date - (CURRENT_DATE - (ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY task_date DESC) - 1))::INT AS grp
    FROM daily_completion
    WHERE had_completion = TRUE
)
SELECT
    user_id,
    COUNT(*) AS streak_days,
    MIN(task_date) AS streak_start
FROM ordered
WHERE grp = 0
GROUP BY user_id;

-- View 9: v_sprint_progress — sprint progress vs expected
CREATE OR REPLACE VIEW v_sprint_progress AS
SELECT
    s.id                                                          AS sprint_id,
    s.go_id,
    g.user_id,
    s.sprint_number,
    s.start_date,
    s.end_date,
    s.status,
    (CURRENT_DATE - s.start_date + 1)::INT                       AS day_number,
    (s.end_date - s.start_date + 1)::INT                         AS total_days,
    ROUND(
        (CURRENT_DATE - s.start_date + 1)::NUMERIC
        / NULLIF(s.end_date - s.start_date + 1, 0) * 100, 1
    )                                                             AS time_elapsed_pct,
    COUNT(dt.id) FILTER (WHERE dt.completed = TRUE)               AS tasks_completed,
    COUNT(dt.id)                                                  AS tasks_total,
    ROUND(
        COUNT(dt.id) FILTER (WHERE dt.completed = TRUE)::NUMERIC
        / NULLIF(COUNT(dt.id), 0) * 100, 1
    )                                                             AS actual_completion_pct
FROM sprints s
JOIN global_objectives g ON g.id = s.go_id
LEFT JOIN daily_tasks dt ON dt.sprint_id = s.id
    AND dt.task_type = 'MAIN'
    AND dt.task_date <= CURRENT_DATE
WHERE s.status = 'ACTIVE'
GROUP BY s.id, s.go_id, g.user_id;

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 003: Level 2 (3 tables, 5 views, 2 fn, 2 triggers) applied';
END $$;
