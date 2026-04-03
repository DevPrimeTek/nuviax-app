-- ═══════════════════════════════════════════════════════════════
-- Migration 013 — Completion Ceremonies Table
-- Fixes missing table required by:
--   - Level 5 engine (GenerateCompletionCeremony)
--   - Admin stats view (v_admin_platform_stats)
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS completion_ceremonies (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id      UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    go_id          UUID NOT NULL REFERENCES global_objectives(id) ON DELETE CASCADE,
    ceremony_tier  TEXT NOT NULL CHECK (ceremony_tier IN ('BRONZE', 'SILVER', 'GOLD', 'PLATINUM')),
    ceremony_data  JSONB NOT NULL DEFAULT '{}'::jsonb,
    viewed         BOOLEAN NOT NULL DEFAULT FALSE,
    viewed_at      TIMESTAMPTZ,
    generated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sprint_id)
);

CREATE INDEX IF NOT EXISTS idx_completion_ceremonies_goal
    ON completion_ceremonies (go_id, generated_at DESC);

CREATE INDEX IF NOT EXISTS idx_completion_ceremonies_unviewed
    ON completion_ceremonies (go_id, viewed)
    WHERE viewed = FALSE;

DO $$
BEGIN
    RAISE NOTICE 'Migration 013: completion_ceremonies table ensured';
END $$;
