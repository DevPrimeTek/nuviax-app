-- ═══════════════════════════════════════════════════════════════
-- Migration 002 — Layer 0 + Level 1: Structural Authority
-- Adds: goal_categories, sprint_configs, goal_metadata
-- Views: v_active_goals, v_current_sprints, v_goal_sprint_summary,
--        v_active_checkpoints
-- Functions: fn_grade_from_score
-- Triggers: trg_sprint_config_init, trg_goal_metadata_init
-- Depends on: 001_base_schema.sql
-- ═══════════════════════════════════════════════════════════════

-- ── TABLES ───────────────────────────────────────────────────────

-- Table 13: goal_categories — taxonomy for goal types
CREATE TABLE IF NOT EXISTS goal_categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT        NOT NULL UNIQUE CHECK (slug ~ '^[a-z_]+$'),
    label_ro    TEXT        NOT NULL,
    label_en    TEXT        NOT NULL,
    icon        TEXT,                         -- emoji or icon key
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE
);

-- Seed default categories (idempotent)
INSERT INTO goal_categories (slug, label_ro, label_en, icon, sort_order)
VALUES
    ('career',   'Carieră',      'Career',       '💼', 10),
    ('health',   'Sănătate',     'Health',       '🏃', 20),
    ('finance',  'Finanțe',      'Finance',      '💰', 30),
    ('learning', 'Educație',     'Learning',     '📚', 40),
    ('creative', 'Creativ',      'Creative',     '🎨', 50),
    ('personal', 'Personal',     'Personal',     '🌟', 60),
    ('other',    'Altele',       'Other',        '📌', 70)
ON CONFLICT (slug) DO NOTHING;

-- Table 14: sprint_configs — per-goal sprint configuration
CREATE TABLE IF NOT EXISTS sprint_configs (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id           UUID    NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE UNIQUE,
    sprint_days     INT     NOT NULL DEFAULT 30 CHECK (sprint_days BETWEEN 7 AND 90),
    min_tasks_daily INT     NOT NULL DEFAULT 1  CHECK (min_tasks_daily BETWEEN 1 AND 3),
    max_tasks_daily INT     NOT NULL DEFAULT 3  CHECK (max_tasks_daily BETWEEN 1 AND 5),
    checkpoint_count INT    NOT NULL DEFAULT 3  CHECK (checkpoint_count BETWEEN 1 AND 10),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_tasks_range CHECK (min_tasks_daily <= max_tasks_daily)
);

CREATE INDEX IF NOT EXISTS idx_sprint_configs_goal ON sprint_configs (go_id);

CREATE OR REPLACE TRIGGER trg_sprint_configs_updated_at
    BEFORE UPDATE ON sprint_configs
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

-- Table 15: goal_metadata — extended goal data (metrics, category)
CREATE TABLE IF NOT EXISTS goal_metadata (
    id             UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    go_id          UUID    NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE UNIQUE,
    category_id    UUID    REFERENCES goal_categories(id) ON DELETE SET NULL,
    target_value   FLOAT,                         -- ex: 100 (km), 10000 (€), etc.
    current_value  FLOAT,
    start_value    FLOAT,
    unit           TEXT    CHECK (length(unit) <= 20),  -- km, €, kg, %, etc.
    why_text       TEXT    CHECK (length(why_text) <= 1000), -- "De ce vreau asta?"
    tags           TEXT[]  DEFAULT '{}',
    is_private     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goal_metadata_goal     ON goal_metadata (go_id);
CREATE INDEX IF NOT EXISTS idx_goal_metadata_category ON goal_metadata (category_id);

CREATE OR REPLACE TRIGGER trg_goal_metadata_updated_at
    BEFORE UPDATE ON goal_metadata
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

-- ── TRIGGER: auto-create sprint_config + goal_metadata on new goal ──
CREATE OR REPLACE FUNCTION fn_init_goal_dependencies()
RETURNS TRIGGER AS $$
BEGIN
    -- Create default sprint config
    INSERT INTO sprint_configs (go_id)
    VALUES (NEW.id)
    ON CONFLICT (go_id) DO NOTHING;

    -- Create empty goal metadata
    INSERT INTO goal_metadata (go_id)
    VALUES (NEW.id)
    ON CONFLICT (go_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_goal_init_dependencies
    AFTER INSERT ON global_objectives
    FOR EACH ROW EXECUTE FUNCTION fn_init_goal_dependencies();

-- ── FUNCTION: fn_grade_from_score ────────────────────────────────
-- Same thresholds as engine/helpers.go — single source of truth
CREATE OR REPLACE FUNCTION fn_grade_from_score(score NUMERIC)
RETURNS TEXT AS $$
BEGIN
    IF score >= 0.90 THEN RETURN 'A+';
    ELSIF score >= 0.80 THEN RETURN 'A';
    ELSIF score >= 0.70 THEN RETURN 'B';
    ELSIF score >= 0.60 THEN RETURN 'C';
    ELSE RETURN 'D';
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ── VIEWS ────────────────────────────────────────────────────────

-- View 1: v_active_goals — active goals with category and metadata
CREATE OR REPLACE VIEW v_active_goals AS
SELECT
    g.id,
    g.user_id,
    g.name,
    g.description,
    g.status,
    g.start_date,
    g.end_date,
    (g.end_date - CURRENT_DATE)                AS days_left,
    EXTRACT(DAY FROM NOW() - g.start_date)::INT AS elapsed_days,
    cat.slug                                   AS category_slug,
    cat.label_ro                               AS category_label,
    cat.icon                                   AS category_icon,
    m.target_value,
    m.current_value,
    m.unit,
    m.tags
FROM global_objectives g
LEFT JOIN goal_metadata m   ON m.go_id = g.id
LEFT JOIN goal_categories cat ON cat.id = m.category_id
WHERE g.status = 'ACTIVE';

-- View 2: v_current_sprints — latest active sprint per goal
CREATE OR REPLACE VIEW v_current_sprints AS
SELECT
    s.id            AS sprint_id,
    s.go_id,
    g.user_id,
    g.name          AS goal_name,
    s.sprint_number,
    s.start_date,
    s.end_date,
    s.status,
    (s.end_date - CURRENT_DATE)::INT                         AS days_left,
    (CURRENT_DATE - s.start_date + 1)::INT                  AS day_number,
    (s.end_date - s.start_date + 1)::INT                    AS total_days
FROM sprints s
JOIN global_objectives g ON g.id = s.go_id
WHERE s.status = 'ACTIVE';

-- View 3: v_goal_sprint_summary — goal + current sprint + latest score
CREATE OR REPLACE VIEW v_goal_sprint_summary AS
SELECT
    g.id,
    g.user_id,
    g.name,
    g.status,
    g.start_date,
    g.end_date,
    cs.sprint_id,
    cs.sprint_number,
    cs.days_left,
    cs.day_number,
    COALESCE(gs.score_value, 0)                               AS score,
    COALESCE(gs.grade, 'D')                                   AS grade,
    (CAST(CURRENT_DATE - g.start_date AS FLOAT)
        / NULLIF(g.end_date - g.start_date, 0) * 100)::INT   AS progress_pct
FROM global_objectives g
LEFT JOIN v_current_sprints cs ON cs.go_id = g.id
LEFT JOIN LATERAL (
    SELECT score_value, grade FROM go_scores
    WHERE go_id = g.id ORDER BY computed_at DESC LIMIT 1
) gs ON TRUE
WHERE g.status NOT IN ('ARCHIVED');

-- View 4: v_active_checkpoints — current active checkpoint per sprint
CREATE OR REPLACE VIEW v_active_checkpoints AS
SELECT
    cp.id,
    cp.sprint_id,
    s.go_id,
    g.user_id,
    cp.name,
    cp.description,
    cp.sort_order,
    cp.status,
    cp.progress_pct,
    cp.completed_at
FROM checkpoints cp
JOIN sprints s ON s.id = cp.sprint_id
JOIN global_objectives g ON g.id = s.go_id
WHERE cp.status IN ('UPCOMING', 'IN_PROGRESS')
  AND s.status = 'ACTIVE';

-- ─────────────────────────────────────────────────────────────────
DO $$ BEGIN
    RAISE NOTICE 'Migration 002: Layer 0 + Level 1 (3 tables, 4 views, 1 fn, 3 triggers) applied';
END $$;
