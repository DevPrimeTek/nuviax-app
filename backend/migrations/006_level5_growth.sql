-- ═══════════════════════════════════════════════════════════════
-- Migration 006 — Level 5: Growth Orchestration (C37-C40)
-- Adds: growth_milestones, achievement_badges, ceremonies,
--       growth_trajectories
-- Views: v_growth_overview, v_achievement_summary,
--        v_ceremony_schedule, v_milestone_progress,
--        v_grade_history, v_trajectory_analysis, v_progress_chart
-- Materialized: mv_user_stats
-- Functions: fn_compute_growth_trajectory, fn_award_achievement_if_earned
-- Triggers: trg_milestone_check, trg_achievement_award, trg_trajectory_snapshot
-- Depends on: 005_level4_regulatory.sql
-- ═══════════════════════════════════════════════════════════════

-- ── TABLES ───────────────────────────────────────────────────────

-- Table 25: growth_milestones — significant framework milestones
CREATE TYPE IF NOT EXISTS milestone_type AS ENUM (
    'FIRST_TASK',           -- first task ever completed
    'FIRST_SPRINT',         -- first sprint completed
    'FIRST_GOAL',           -- first goal completed
    'STREAK_3',             -- 3-day streak
    'STREAK_7',             -- 7-day streak
    'STREAK_14',            -- 14-day streak
    'STREAK_30',            -- 30-day streak
    'GRADE_A_FIRST',        -- first A or A+ grade on a sprint
    'GRADE_A_PLUS_FIRST',   -- first A+ grade
    'CONSISTENCY_90',       -- 90%+ consistency for a full sprint
    'PERFECT_SPRINT',       -- 100% completion rate on a sprint
    'GOAL_100_DAYS',        -- goal active for 100 days
    'MULTI_GOAL_ACTIVE'     -- 2+ goals active simultaneously
);

CREATE TABLE IF NOT EXISTS growth_milestones (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    go_id           UUID          REFERENCES global_objectives(id) ON DELETE SET NULL,
    sprint_id       UUID          REFERENCES sprints(id) ON DELETE SET NULL,
    milestone_type  milestone_type NOT NULL,
    achieved_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    details         JSONB         NOT NULL DEFAULT '{}',
    UNIQUE (user_id, milestone_type, go_id)   -- one per goal per milestone type
);

CREATE INDEX IF NOT EXISTS idx_milestones_user    ON growth_milestones (user_id, achieved_at DESC);
CREATE INDEX IF NOT EXISTS idx_milestones_goal    ON growth_milestones (go_id) WHERE go_id IS NOT NULL;

-- Table 26: achievement_badges — gamification badges
CREATE TYPE IF NOT EXISTS badge_type AS ENUM (
    'STARTER',              -- registered and created first goal
    'CONSISTENT_WEEK',      -- 7 days consistent
    'CONSISTENT_MONTH',     -- 30 days consistent
    'GRADE_HUNTER',         -- 3 A-grade sprints
    'PERFECTIONIST',        -- 1 perfect sprint (100%)
    'GOAL_SLAYER',          -- completed 1 full goal
    'MULTI_TASKER',         -- 2+ active goals
    'COMEBACK_KID',         -- resumed after 7+ day break
    'EARLY_BIRD',           -- completed tasks before 9:00 for 5 days
    'MARATHON_RUNNER'       -- goal running 6+ months
);

CREATE TABLE IF NOT EXISTS achievement_badges (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_type  badge_type  NOT NULL,
    go_id       UUID        REFERENCES global_objectives(id) ON DELETE SET NULL,
    sprint_id   UUID        REFERENCES sprints(id) ON DELETE SET NULL,
    awarded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, badge_type)   -- each badge awarded once per user
);

CREATE INDEX IF NOT EXISTS idx_badges_user ON achievement_badges (user_id, awarded_at DESC);

-- Table 27: ceremonies — sprint lifecycle ceremonies
CREATE TYPE IF NOT EXISTS ceremony_type AS ENUM (
    'KICKOFF',              -- sprint start ceremony
    'MIDPOINT',             -- halfway check-in
    'RETROSPECTIVE',        -- sprint end retrospective
    'GOAL_COMPLETION'       -- goal completed ceremony
);

CREATE TYPE IF NOT EXISTS ceremony_status AS ENUM (
    'SCHEDULED', 'COMPLETED', 'SKIPPED'
);

CREATE TABLE IF NOT EXISTS ceremonies (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    go_id           UUID            NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    sprint_id       UUID            REFERENCES sprints(id) ON DELETE SET NULL,
    ceremony_type   ceremony_type   NOT NULL,
    status          ceremony_status NOT NULL DEFAULT 'SCHEDULED',
    scheduled_at    TIMESTAMPTZ     NOT NULL,
    completed_at    TIMESTAMPTZ,
    notes           TEXT            CHECK (length(notes) <= 1000)
);

CREATE INDEX IF NOT EXISTS idx_ceremonies_user     ON ceremonies (user_id, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_ceremonies_sprint   ON ceremonies (sprint_id) WHERE sprint_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ceremonies_upcoming ON ceremonies (user_id, status, scheduled_at)
    WHERE status = 'SCHEDULED';

-- Table 28: growth_trajectories — trajectory snapshots per goal
CREATE TYPE IF NOT EXISTS trajectory_trend AS ENUM (
    'AHEAD', 'ON_TRACK', 'SLIGHTLY_BEHIND', 'BEHIND', 'AT_RISK'
);

CREATE TABLE IF NOT EXISTS growth_trajectories (
    id            UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id         UUID             NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id       UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    snapshot_date DATE             NOT NULL,
    actual_pct    FLOAT            NOT NULL CHECK (actual_pct BETWEEN 0 AND 100),
    expected_pct  FLOAT            NOT NULL CHECK (expected_pct BETWEEN 0 AND 100),
    delta         FLOAT            NOT NULL,   -- actual - expected
    trend         trajectory_trend NOT NULL DEFAULT 'ON_TRACK',
    score         FLOAT            CHECK (score BETWEEN 0 AND 1),
    recorded_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (go_id, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_trajectories_goal ON growth_trajectories (go_id, snapshot_date DESC);
CREATE INDEX IF NOT EXISTS idx_trajectories_user ON growth_trajectories (user_id, snapshot_date DESC);

-- ── TRIGGER: auto-create KICKOFF ceremony when sprint starts ──────
CREATE OR REPLACE FUNCTION fn_create_sprint_ceremony()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id UUID;
    v_midpoint TIMESTAMPTZ;
BEGIN
    SELECT user_id INTO v_user_id
    FROM global_objectives WHERE id = NEW.go_id;

    -- Kickoff ceremony at sprint start
    INSERT INTO ceremonies (user_id, go_id, sprint_id, ceremony_type, scheduled_at)
    VALUES (v_user_id, NEW.go_id, NEW.id, 'KICKOFF',
            (NEW.start_date::TIMESTAMP AT TIME ZONE 'UTC'))
    ON CONFLICT DO NOTHING;

    -- Midpoint ceremony
    v_midpoint := ((NEW.start_date + (NEW.end_date - NEW.start_date) / 2)::TIMESTAMP AT TIME ZONE 'UTC');
    INSERT INTO ceremonies (user_id, go_id, sprint_id, ceremony_type, scheduled_at)
    VALUES (v_user_id, NEW.go_id, NEW.id, 'MIDPOINT', v_midpoint)
    ON CONFLICT DO NOTHING;

    -- Retrospective at sprint end
    INSERT INTO ceremonies (user_id, go_id, sprint_id, ceremony_type, scheduled_at)
    VALUES (v_user_id, NEW.go_id, NEW.id, 'RETROSPECTIVE',
            (NEW.end_date::TIMESTAMP AT TIME ZONE 'UTC') + INTERVAL '1 hour')
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_create_sprint_ceremonies
    AFTER INSERT ON sprints
    FOR EACH ROW EXECUTE FUNCTION fn_create_sprint_ceremony();

-- ── TRIGGER: close KICKOFF ceremony when sprint starts ───────────
CREATE OR REPLACE FUNCTION fn_complete_kickoff_ceremony()
RETURNS TRIGGER AS $$
BEGIN
    -- When a sprint goes ACTIVE, mark its KICKOFF as COMPLETED
    IF NEW.status = 'ACTIVE' AND OLD.status <> 'ACTIVE' THEN
        UPDATE ceremonies
        SET status = 'COMPLETED', completed_at = NOW()
        WHERE sprint_id = NEW.id AND ceremony_type = 'KICKOFF' AND status = 'SCHEDULED';
    END IF;
    -- When sprint completes, mark RETROSPECTIVE ceremony as scheduled-now
    IF NEW.status = 'COMPLETED' AND OLD.status = 'ACTIVE' THEN
        UPDATE ceremonies
        SET scheduled_at = NOW()
        WHERE sprint_id = NEW.id AND ceremony_type = 'RETROSPECTIVE' AND status = 'SCHEDULED';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_ceremony_sprint_lifecycle
    AFTER UPDATE OF status ON sprints
    FOR EACH ROW EXECUTE FUNCTION fn_complete_kickoff_ceremony();

-- ── TRIGGER: check for new milestones after sprint closes ─────────
CREATE OR REPLACE FUNCTION fn_check_milestones_on_sprint_close()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id       UUID;
    v_go_id         UUID;
    v_score         FLOAT;
    v_sprint_count  INT;
    v_streak        INT;
BEGIN
    -- Only on COMPLETED transition
    IF NEW.status <> 'COMPLETED' OR OLD.status = 'COMPLETED' THEN
        RETURN NEW;
    END IF;

    SELECT g.user_id, g.id INTO v_user_id, v_go_id
    FROM global_objectives g WHERE g.id = NEW.go_id;

    -- Milestone: FIRST_SPRINT
    INSERT INTO growth_milestones (user_id, go_id, sprint_id, milestone_type)
    SELECT v_user_id, v_go_id, NEW.id, 'FIRST_SPRINT'
    WHERE NOT EXISTS (
        SELECT 1 FROM growth_milestones
        WHERE user_id = v_user_id AND milestone_type = 'FIRST_SPRINT' AND go_id = v_go_id
    )
    ON CONFLICT DO NOTHING;

    -- Milestone: PERFECT_SPRINT (score = 1.0)
    SELECT score_value INTO v_score FROM sprint_results WHERE sprint_id = NEW.id;
    IF v_score >= 1.0 THEN
        INSERT INTO growth_milestones (user_id, go_id, sprint_id, milestone_type,
            details)
        VALUES (v_user_id, v_go_id, NEW.id, 'PERFECT_SPRINT',
            jsonb_build_object('score', v_score, 'sprint_number', NEW.sprint_number))
        ON CONFLICT DO NOTHING;
    END IF;

    -- Milestone: GRADE_A_FIRST
    IF v_score >= 0.80 THEN
        INSERT INTO growth_milestones (user_id, go_id, sprint_id, milestone_type)
        SELECT v_user_id, v_go_id, NEW.id, 'GRADE_A_FIRST'
        WHERE NOT EXISTS (
            SELECT 1 FROM growth_milestones
            WHERE user_id = v_user_id AND milestone_type = 'GRADE_A_FIRST' AND go_id = v_go_id
        )
        ON CONFLICT DO NOTHING;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_milestone_check_on_sprint_close
    AFTER UPDATE OF status ON sprints
    FOR EACH ROW EXECUTE FUNCTION fn_check_milestones_on_sprint_close();

-- ── FUNCTION: fn_compute_growth_trajectory ───────────────────────
-- Mirrors engine/level5_growth.go: computeProgressVsExpected
-- Stores a trajectory snapshot for the current date
CREATE OR REPLACE FUNCTION fn_compute_growth_trajectory(p_go_id UUID)
RETURNS FLOAT AS $$
DECLARE
    v_user_id     UUID;
    v_start       DATE;
    v_end         DATE;
    v_total_days  FLOAT;
    v_elapsed     FLOAT;
    v_expected    FLOAT;
    v_completed   INT;
    v_total_cp    INT;
    v_actual      FLOAT;
    v_delta       FLOAT;
    v_trend       trajectory_trend;
    v_score       FLOAT;
    v_sprint_id   UUID;
BEGIN
    SELECT user_id, start_date, end_date INTO v_user_id, v_start, v_end
    FROM global_objectives WHERE id = p_go_id;

    v_total_days := v_end - v_start;
    v_elapsed    := CURRENT_DATE - v_start;
    IF v_total_days <= 0 THEN RETURN 0; END IF;

    v_expected := ROUND((v_elapsed / v_total_days) * 100, 1);

    -- Current sprint
    SELECT id INTO v_sprint_id FROM sprints
    WHERE go_id = p_go_id AND status = 'ACTIVE'
    ORDER BY sprint_number DESC LIMIT 1;

    -- Checkpoint completion as proxy for actual progress
    SELECT
        COUNT(*) FILTER (WHERE status = 'COMPLETED'),
        COUNT(*)
    INTO v_completed, v_total_cp
    FROM checkpoints
    WHERE sprint_id = v_sprint_id;

    v_actual := CASE WHEN v_total_cp > 0
        THEN ROUND((v_completed::FLOAT / v_total_cp) * 100, 1)
        ELSE v_expected END;

    v_delta := v_actual - v_expected;

    v_trend := CASE
        WHEN v_delta >= 10  THEN 'AHEAD'
        WHEN v_delta >= -5  THEN 'ON_TRACK'
        WHEN v_delta >= -15 THEN 'SLIGHTLY_BEHIND'
        WHEN v_delta >= -30 THEN 'BEHIND'
        ELSE 'AT_RISK'
    END;

    v_score := fn_compute_goal_score(p_go_id);

    INSERT INTO growth_trajectories
        (go_id, user_id, snapshot_date, actual_pct, expected_pct, delta, trend, score)
    VALUES
        (p_go_id, v_user_id, CURRENT_DATE, v_actual, v_expected, v_delta, v_trend, v_score)
    ON CONFLICT (go_id, snapshot_date) DO UPDATE SET
        actual_pct  = EXCLUDED.actual_pct,
        expected_pct = EXCLUDED.expected_pct,
        delta       = EXCLUDED.delta,
        trend       = EXCLUDED.trend,
        score       = EXCLUDED.score,
        recorded_at = NOW();

    RETURN v_score;
END;
$$ LANGUAGE plpgsql;

-- ── FUNCTION: fn_award_achievement_if_earned ─────────────────────
-- Checks and awards achievement badges based on current user state
CREATE OR REPLACE FUNCTION fn_award_achievement_if_earned(
    p_user_id UUID,
    p_go_id   UUID DEFAULT NULL
)
RETURNS INT AS $$  -- returns count of new badges awarded
DECLARE
    v_streak      INT;
    v_awarded     INT := 0;
    v_goal_count  INT;
    v_a_sprints   INT;
BEGIN
    -- Badge: STARTER (has at least 1 goal)
    IF NOT EXISTS (SELECT 1 FROM achievement_badges WHERE user_id = p_user_id AND badge_type = 'STARTER') THEN
        IF EXISTS (SELECT 1 FROM global_objectives WHERE user_id = p_user_id LIMIT 1) THEN
            INSERT INTO achievement_badges (user_id, badge_type, go_id) VALUES (p_user_id, 'STARTER', p_go_id) ON CONFLICT DO NOTHING;
            v_awarded := v_awarded + 1;
        END IF;
    END IF;

    -- Badge: CONSISTENT_WEEK (streak >= 7)
    SELECT COALESCE(streak_days, 0) INTO v_streak FROM v_task_streaks WHERE user_id = p_user_id;
    IF v_streak >= 7 AND NOT EXISTS (SELECT 1 FROM achievement_badges WHERE user_id = p_user_id AND badge_type = 'CONSISTENT_WEEK') THEN
        INSERT INTO achievement_badges (user_id, badge_type) VALUES (p_user_id, 'CONSISTENT_WEEK') ON CONFLICT DO NOTHING;
        v_awarded := v_awarded + 1;
    END IF;

    -- Badge: CONSISTENT_MONTH (streak >= 30)
    IF v_streak >= 30 AND NOT EXISTS (SELECT 1 FROM achievement_badges WHERE user_id = p_user_id AND badge_type = 'CONSISTENT_MONTH') THEN
        INSERT INTO achievement_badges (user_id, badge_type) VALUES (p_user_id, 'CONSISTENT_MONTH') ON CONFLICT DO NOTHING;
        v_awarded := v_awarded + 1;
    END IF;

    -- Badge: MULTI_TASKER (2+ active goals)
    SELECT COUNT(*) INTO v_goal_count FROM global_objectives WHERE user_id = p_user_id AND status = 'ACTIVE';
    IF v_goal_count >= 2 AND NOT EXISTS (SELECT 1 FROM achievement_badges WHERE user_id = p_user_id AND badge_type = 'MULTI_TASKER') THEN
        INSERT INTO achievement_badges (user_id, badge_type) VALUES (p_user_id, 'MULTI_TASKER') ON CONFLICT DO NOTHING;
        v_awarded := v_awarded + 1;
    END IF;

    -- Badge: GRADE_HUNTER (3+ A-grade sprints)
    SELECT COUNT(*) INTO v_a_sprints
    FROM sprint_results sr
    JOIN sprints s ON s.id = sr.sprint_id
    JOIN global_objectives g ON g.id = s.go_id
    WHERE g.user_id = p_user_id AND sr.grade IN ('A', 'A+');

    IF v_a_sprints >= 3 AND NOT EXISTS (SELECT 1 FROM achievement_badges WHERE user_id = p_user_id AND badge_type = 'GRADE_HUNTER') THEN
        INSERT INTO achievement_badges (user_id, badge_type) VALUES (p_user_id, 'GRADE_HUNTER') ON CONFLICT DO NOTHING;
        v_awarded := v_awarded + 1;
    END IF;

    RETURN v_awarded;
END;
$$ LANGUAGE plpgsql;

-- ── MATERIALIZED VIEW: mv_user_stats ─────────────────────────────
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_user_stats AS
SELECT
    u.id                                                          AS user_id,
    COUNT(DISTINCT g.id)                                          AS total_goals,
    COUNT(DISTINCT g.id) FILTER (WHERE g.status = 'ACTIVE')       AS active_goals,
    COUNT(DISTINCT g.id) FILTER (WHERE g.status = 'COMPLETED')    AS completed_goals,
    COUNT(DISTINCT s.id)                                          AS total_sprints,
    COUNT(DISTINCT s.id) FILTER (WHERE s.status = 'COMPLETED')    AS completed_sprints,
    COUNT(DISTINCT dt.id)                                         AS total_tasks,
    COUNT(DISTINCT dt.id) FILTER (WHERE dt.completed = TRUE)      AS completed_tasks,
    ROUND(AVG(sr.score_value)::NUMERIC, 3)                        AS avg_sprint_score,
    COUNT(DISTINCT ab.id)                                         AS badge_count,
    COALESCE(ts.streak_days, 0)                                   AS current_streak,
    MAX(dt.task_date) FILTER (WHERE dt.completed = TRUE)          AS last_active_date,
    NOW()                                                         AS refreshed_at
FROM users u
LEFT JOIN global_objectives g  ON g.user_id = u.id
LEFT JOIN sprints s            ON s.go_id = g.id
LEFT JOIN sprint_results sr    ON sr.sprint_id = s.id
LEFT JOIN daily_tasks dt       ON dt.user_id = u.id AND dt.task_type = 'MAIN'
LEFT JOIN achievement_badges ab ON ab.user_id = u.id
LEFT JOIN v_task_streaks ts    ON ts.user_id = u.id
WHERE u.is_active = TRUE
GROUP BY u.id, ts.streak_days;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_user_stats_user ON mv_user_stats (user_id);

-- ── VIEWS ────────────────────────────────────────────────────────

-- View 20: v_growth_overview — overall growth metrics per user
CREATE OR REPLACE VIEW v_growth_overview AS
SELECT
    ms.user_id,
    ms.total_goals,
    ms.active_goals,
    ms.completed_goals,
    ms.total_sprints,
    ms.completed_sprints,
    ms.total_tasks,
    ms.completed_tasks,
    ROUND(
        ms.completed_tasks::NUMERIC / NULLIF(ms.total_tasks, 0) * 100, 1
    )                         AS overall_completion_pct,
    ms.avg_sprint_score,
    fn_grade_from_score(COALESCE(ms.avg_sprint_score, 0)) AS overall_grade,
    ms.badge_count,
    ms.current_streak,
    ms.last_active_date
FROM mv_user_stats ms;

-- View 21: v_achievement_summary — badges earned per user
CREATE OR REPLACE VIEW v_achievement_summary AS
SELECT
    ab.user_id,
    ab.badge_type,
    ab.go_id,
    ab.sprint_id,
    ab.awarded_at,
    g.name AS goal_name
FROM achievement_badges ab
LEFT JOIN global_objectives g ON g.id = ab.go_id
ORDER BY ab.awarded_at DESC;

-- View 22: v_ceremony_schedule — upcoming ceremonies
CREATE OR REPLACE VIEW v_ceremony_schedule AS
SELECT
    c.id,
    c.user_id,
    c.go_id,
    g.name          AS goal_name,
    c.sprint_id,
    s.sprint_number,
    c.ceremony_type,
    c.status,
    c.scheduled_at,
    c.completed_at,
    c.notes,
    (c.scheduled_at - NOW()) AS time_until
FROM ceremonies c
JOIN global_objectives g ON g.id = c.go_id
LEFT JOIN sprints s ON s.id = c.sprint_id
WHERE c.status = 'SCHEDULED'
  AND c.scheduled_at >= NOW() - INTERVAL '1 day'
ORDER BY c.scheduled_at ASC;

-- View 23: v_milestone_progress — milestones achieved per user
CREATE OR REPLACE VIEW v_milestone_progress AS
SELECT
    gm.user_id,
    gm.go_id,
    g.name          AS goal_name,
    gm.milestone_type,
    gm.achieved_at,
    gm.details
FROM growth_milestones gm
LEFT JOIN global_objectives g ON g.id = gm.go_id
ORDER BY gm.achieved_at DESC;

-- View 24: v_grade_history — grade evolution per goal over sprints
CREATE OR REPLACE VIEW v_grade_history AS
SELECT
    sr.sprint_id,
    s.go_id,
    g.user_id,
    g.name                  AS goal_name,
    s.sprint_number,
    s.start_date,
    s.end_date,
    sr.score_value          AS score,
    sr.grade,
    sr.computed_at,
    LAG(sr.grade) OVER (
        PARTITION BY s.go_id ORDER BY s.sprint_number
    )                       AS prev_grade,
    CASE
        WHEN sr.score_value > LAG(sr.score_value) OVER (
            PARTITION BY s.go_id ORDER BY s.sprint_number) THEN 'UP'
        WHEN sr.score_value < LAG(sr.score_value) OVER (
            PARTITION BY s.go_id ORDER BY s.sprint_number) THEN 'DOWN'
        ELSE 'SAME'
    END                     AS trend
FROM sprint_results sr
JOIN sprints s ON s.id = sr.sprint_id
JOIN global_objectives g ON g.id = s.go_id;

-- View 25: v_trajectory_analysis — trajectory analysis per goal
CREATE OR REPLACE VIEW v_trajectory_analysis AS
SELECT
    gt.go_id,
    gt.user_id,
    g.name              AS goal_name,
    gt.snapshot_date,
    gt.actual_pct,
    gt.expected_pct,
    gt.delta,
    gt.trend,
    gt.score,
    fn_grade_from_score(COALESCE(gt.score, 0)) AS grade
FROM growth_trajectories gt
JOIN global_objectives g ON g.id = gt.go_id
WHERE gt.snapshot_date >= CURRENT_DATE - 30
ORDER BY gt.go_id, gt.snapshot_date DESC;

-- View 26: v_progress_chart — chart data for progress visualization
CREATE OR REPLACE VIEW v_progress_chart AS
SELECT
    dm.go_id,
    dm.user_id,
    g.name              AS goal_name,
    dm.metric_date      AS date,
    dm.completion_rate,
    dm.tasks_done,
    dm.tasks_total,
    COALESCE(gt.actual_pct, NULL) AS trajectory_actual_pct,
    COALESCE(gt.expected_pct, NULL) AS trajectory_expected_pct,
    COALESCE(gt.trend, NULL)       AS trajectory_trend
FROM daily_metrics dm
JOIN global_objectives g ON g.id = dm.go_id
LEFT JOIN growth_trajectories gt ON gt.go_id = dm.go_id
    AND gt.snapshot_date = dm.metric_date
WHERE dm.metric_date >= CURRENT_DATE - 90
ORDER BY dm.go_id, dm.metric_date DESC;

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 006: Level 5 (4 tables, 7 views, 1 matview, 2 fn, 3 triggers) applied';
    RAISE NOTICE 'Schema totals: 28 tables, 26 views + 1 materialized, 10 functions, 12 triggers';
END $$;
