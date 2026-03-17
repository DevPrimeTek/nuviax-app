-- ═══════════════════════════════════════════════════════════════
-- Migration 005 — Level 4: Regulatory Authority (C32-C36)
-- Adds: regulatory_events, goal_activation_log, resource_slots
-- Views: v_goal_activation_status, v_resource_utilization,
--        v_regulatory_health, v_goal_constraints, v_conflict_matrix
-- Functions: fn_check_goal_activation, fn_close_sprint_with_score
-- Triggers: trg_goal_activation_log, trg_regulatory_check
-- Depends on: 004_level3_adaptive.sql
-- ═══════════════════════════════════════════════════════════════

-- ── TABLES ───────────────────────────────────────────────────────

-- Table 22: regulatory_events — events from regulatory rule checks
CREATE TYPE IF NOT EXISTS regulatory_event_type AS ENUM (
    'LIMIT_REACHED',        -- max active goals limit hit
    'CONFLICT_DETECTED',    -- temporal overlap between goals
    'RULE_VIOLATED',        -- business rule violated
    'ACTIVATION_BLOCKED',   -- goal activation blocked by rules
    'ACTIVATION_ALLOWED',   -- goal activation allowed (with possible warning)
    'SPRINT_CLOSED',        -- sprint closed by regulatory authority
    'GOAL_COMPLETED'        -- goal marked complete by system
);

CREATE TYPE IF NOT EXISTS regulatory_event_status AS ENUM (
    'OPEN', 'RESOLVED', 'ACKNOWLEDGED'
);

CREATE TABLE IF NOT EXISTS regulatory_events (
    id          UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID                    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    go_id       UUID                    REFERENCES global_objectives(id) ON DELETE SET NULL,
    event_type  regulatory_event_type   NOT NULL,
    status      regulatory_event_status NOT NULL DEFAULT 'OPEN',
    details     JSONB                   NOT NULL DEFAULT '{}',
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ             NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reg_events_user    ON regulatory_events (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reg_events_goal    ON regulatory_events (go_id) WHERE go_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reg_events_open    ON regulatory_events (user_id, status)
    WHERE status = 'OPEN';

-- Table 23: goal_activation_log — full history of goal status transitions
CREATE TABLE IF NOT EXISTS goal_activation_log (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id       UUID        NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_status goal_status NOT NULL,
    to_status   goal_status NOT NULL,
    reason      TEXT        CHECK (length(reason) <= 500),
    triggered_by TEXT       NOT NULL DEFAULT 'USER' CHECK (triggered_by IN ('USER','SYSTEM','SCHEDULER')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activation_log_goal ON goal_activation_log (go_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activation_log_user ON goal_activation_log (user_id, created_at DESC);

-- Table 24: resource_slots — temporal resource allocation per user/goal
-- Used for conflict detection (Level 4 C34)
CREATE TABLE IF NOT EXISTS resource_slots (
    id           UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    go_id        UUID    NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    slot_type    TEXT    NOT NULL DEFAULT 'FOCUS' CHECK (slot_type IN ('FOCUS','MAINTENANCE')),
    weekday_mask INT     NOT NULL DEFAULT 127  -- bitmask: Mon=1,Tue=2,...,Sun=64; 127=all days
        CHECK (weekday_mask BETWEEN 1 AND 127),
    weight       FLOAT   NOT NULL DEFAULT 1.0 CHECK (weight BETWEEN 0.1 AND 3.0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, go_id, slot_type)
);

CREATE INDEX IF NOT EXISTS idx_resource_slots_user ON resource_slots (user_id);
CREATE INDEX IF NOT EXISTS idx_resource_slots_goal ON resource_slots (go_id);

-- ── TRIGGER: log every goal status change ────────────────────────
CREATE OR REPLACE FUNCTION fn_log_goal_activation()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status <> OLD.status THEN
        INSERT INTO goal_activation_log
            (go_id, user_id, from_status, to_status, triggered_by)
        VALUES
            (NEW.id, NEW.user_id, OLD.status, NEW.status, 'SYSTEM');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_goal_activation_log
    AFTER UPDATE OF status ON global_objectives
    FOR EACH ROW EXECUTE FUNCTION fn_log_goal_activation();

-- ── TRIGGER: check regulatory rules on goal activation ───────────
CREATE OR REPLACE FUNCTION fn_regulatory_check_on_activate()
RETURNS TRIGGER AS $$
DECLARE
    v_active_count INT;
BEGIN
    -- Only check when transitioning TO ACTIVE
    IF NEW.status <> 'ACTIVE' OR OLD.status = 'ACTIVE' THEN
        RETURN NEW;
    END IF;

    -- Regulatory rule: max 3 active goals
    SELECT COUNT(*) INTO v_active_count
    FROM global_objectives
    WHERE user_id = NEW.user_id AND status = 'ACTIVE' AND id <> NEW.id;

    IF v_active_count >= 3 THEN
        INSERT INTO regulatory_events
            (user_id, go_id, event_type, status, details)
        VALUES
            (NEW.user_id, NEW.id, 'LIMIT_REACHED', 'OPEN',
             jsonb_build_object(
                 'active_count', v_active_count,
                 'max_allowed', 3,
                 'goal_name', NEW.name
             ));
        -- Regulatory authority blocks the activation
        RAISE EXCEPTION 'Regulatory: maximum 3 active goals allowed (currently %)', v_active_count
            USING ERRCODE = 'P0001';
    END IF;

    -- Log the allowed activation
    INSERT INTO regulatory_events
        (user_id, go_id, event_type, status, details)
    VALUES
        (NEW.user_id, NEW.id, 'ACTIVATION_ALLOWED', 'RESOLVED',
         jsonb_build_object('active_count_after', v_active_count + 1));

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_regulatory_check_on_activate
    BEFORE UPDATE OF status ON global_objectives
    FOR EACH ROW EXECUTE FUNCTION fn_regulatory_check_on_activate();

-- ── FUNCTION: fn_check_goal_activation ───────────────────────────
-- Check whether a new goal can be activated for a given user
-- Returns: (can_activate BOOL, reason TEXT)
CREATE TYPE IF NOT EXISTS activation_check_result AS (
    can_activate BOOLEAN,
    reason       TEXT
);

CREATE OR REPLACE FUNCTION fn_check_goal_activation(
    p_user_id  UUID,
    p_start    DATE,
    p_end      DATE
)
RETURNS activation_check_result AS $$
DECLARE
    v_active_count INT;
    v_result       activation_check_result;
    v_overlap_name TEXT;
BEGIN
    -- Rule 1: Max 3 active
    SELECT COUNT(*) INTO v_active_count
    FROM global_objectives
    WHERE user_id = p_user_id AND status = 'ACTIVE';

    IF v_active_count >= 3 THEN
        v_result.can_activate := FALSE;
        v_result.reason := 'Poți lucra la maxim 3 obiective în același timp.';
        RETURN v_result;
    END IF;

    -- Rule 2: Max 365 days
    IF p_end - p_start > 365 THEN
        v_result.can_activate := FALSE;
        v_result.reason := 'Un obiectiv nu poate dura mai mult de 365 de zile.';
        RETURN v_result;
    END IF;

    -- Rule 3: Temporal overlap warning (doesn't block, just warns)
    SELECT name INTO v_overlap_name
    FROM global_objectives
    WHERE user_id = p_user_id
      AND status = 'ACTIVE'
      AND start_date < p_end
      AND end_date > p_start
    LIMIT 1;

    IF v_overlap_name IS NOT NULL THEN
        v_result.can_activate := TRUE;
        v_result.reason := 'Atenție: se suprapune cu "' || v_overlap_name || '"';
        RETURN v_result;
    END IF;

    v_result.can_activate := TRUE;
    v_result.reason := '';
    RETURN v_result;
END;
$$ LANGUAGE plpgsql STABLE;

-- ── FUNCTION: fn_close_sprint_with_score ─────────────────────────
-- Closes a sprint and stores its computed score
CREATE OR REPLACE FUNCTION fn_close_sprint_with_score(p_sprint_id UUID)
RETURNS FLOAT AS $$
DECLARE
    v_score FLOAT;
    v_grade TEXT;
BEGIN
    v_score := fn_compute_sprint_score(p_sprint_id);
    v_grade := fn_grade_from_score(v_score);

    -- Store sprint result
    INSERT INTO sprint_results (sprint_id, score_value, grade)
    VALUES (p_sprint_id, v_score, v_grade)
    ON CONFLICT (sprint_id) DO UPDATE SET
        score_value = EXCLUDED.score_value,
        grade       = EXCLUDED.grade,
        computed_at = NOW();

    -- Close the sprint
    UPDATE sprints SET status = 'COMPLETED' WHERE id = p_sprint_id;

    -- Log regulatory event
    INSERT INTO regulatory_events (
        user_id, go_id, event_type, status, details
    )
    SELECT
        g.user_id, s.go_id, 'SPRINT_CLOSED', 'RESOLVED',
        jsonb_build_object('score', v_score, 'grade', v_grade, 'sprint_number', s.sprint_number)
    FROM sprints s
    JOIN global_objectives g ON g.id = s.go_id
    WHERE s.id = p_sprint_id;

    RETURN v_score;
END;
$$ LANGUAGE plpgsql;

-- ── VIEWS ────────────────────────────────────────────────────────

-- View 15: v_goal_activation_status — goal activation eligibility per user
CREATE OR REPLACE VIEW v_goal_activation_status AS
SELECT
    g.id,
    g.user_id,
    g.name,
    g.status,
    g.start_date,
    g.end_date,
    (SELECT COUNT(*) FROM global_objectives
     WHERE user_id = g.user_id AND status = 'ACTIVE')          AS user_active_count,
    3                                                           AS max_active_allowed,
    CASE WHEN (SELECT COUNT(*) FROM global_objectives
               WHERE user_id = g.user_id AND status = 'ACTIVE') < 3
         THEN TRUE ELSE FALSE END                               AS can_activate_more,
    COALESCE(
        (SELECT MAX(created_at) FROM goal_activation_log
         WHERE go_id = g.id ORDER BY created_at DESC LIMIT 1),
        g.created_at
    )                                                           AS last_status_change
FROM global_objectives g
WHERE g.status NOT IN ('ARCHIVED');

-- View 16: v_resource_utilization — resource load per user
CREATE OR REPLACE VIEW v_resource_utilization AS
SELECT
    rs.user_id,
    COUNT(DISTINCT rs.go_id)                    AS goals_with_slots,
    SUM(rs.weight)                              AS total_weight,
    AVG(rs.weight)                              AS avg_weight_per_goal,
    COUNT(*) FILTER (WHERE rs.slot_type = 'FOCUS')       AS focus_slots,
    COUNT(*) FILTER (WHERE rs.slot_type = 'MAINTENANCE') AS maintenance_slots
FROM resource_slots rs
JOIN global_objectives g ON g.id = rs.go_id
WHERE g.status = 'ACTIVE'
GROUP BY rs.user_id;

-- View 17: v_regulatory_health — regulatory status per user
CREATE OR REPLACE VIEW v_regulatory_health AS
SELECT
    u.id                                                        AS user_id,
    COUNT(re.id) FILTER (WHERE re.status = 'OPEN')              AS open_events,
    COUNT(re.id) FILTER (WHERE re.event_type = 'LIMIT_REACHED') AS limit_breaches,
    COUNT(re.id) FILTER (WHERE re.event_type = 'CONFLICT_DETECTED') AS conflicts,
    (SELECT COUNT(*) FROM global_objectives
     WHERE user_id = u.id AND status = 'ACTIVE')                AS active_goals,
    3                                                           AS max_active,
    CASE WHEN COUNT(re.id) FILTER (WHERE re.status = 'OPEN') = 0
         THEN 'HEALTHY' ELSE 'HAS_ISSUES' END                  AS health_status
FROM users u
LEFT JOIN regulatory_events re ON re.user_id = u.id
    AND re.created_at >= NOW() - INTERVAL '30 days'
WHERE u.is_active = TRUE
GROUP BY u.id;

-- View 18: v_goal_constraints — current active constraints per goal
CREATE OR REPLACE VIEW v_goal_constraints AS
SELECT
    g.id                                                   AS go_id,
    g.user_id,
    g.name,
    g.status,
    COUNT(ca.id) FILTER (
        WHERE ca.adj_type = 'PAUSE'
          AND ca.start_date <= CURRENT_DATE
          AND (ca.end_date IS NULL OR ca.end_date >= CURRENT_DATE)
    )                                                      AS active_pauses,
    COUNT(ca.id) FILTER (
        WHERE ca.adj_type = 'ENERGY_LOW'
          AND ca.start_date <= CURRENT_DATE
          AND (ca.end_date IS NULL OR ca.end_date >= CURRENT_DATE)
    )                                                      AS energy_low_flags,
    BOOL_OR(
        ca.adj_type = 'PAUSE'
        AND ca.start_date <= CURRENT_DATE
        AND (ca.end_date IS NULL OR ca.end_date >= CURRENT_DATE)
    )                                                      AS is_paused
FROM global_objectives g
LEFT JOIN context_adjustments ca ON ca.go_id = g.id
WHERE g.status NOT IN ('ARCHIVED')
GROUP BY g.id, g.user_id, g.name, g.status;

-- View 19: v_conflict_matrix — temporal overlaps between active goals per user
CREATE OR REPLACE VIEW v_conflict_matrix AS
SELECT
    g1.user_id,
    g1.id   AS goal_a_id,
    g1.name AS goal_a_name,
    g2.id   AS goal_b_id,
    g2.name AS goal_b_name,
    GREATEST(g1.start_date, g2.start_date)  AS overlap_start,
    LEAST(g1.end_date, g2.end_date)         AS overlap_end,
    (LEAST(g1.end_date, g2.end_date)
        - GREATEST(g1.start_date, g2.start_date)) AS overlap_days
FROM global_objectives g1
JOIN global_objectives g2
    ON  g1.user_id = g2.user_id
    AND g1.id < g2.id                              -- avoid duplicates
    AND g1.start_date < g2.end_date
    AND g2.start_date < g1.end_date
WHERE g1.status = 'ACTIVE'
  AND g2.status = 'ACTIVE';

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 005: Level 4 (3 tables, 5 views, 2+1 fn, 2 triggers) applied';
END $$;
