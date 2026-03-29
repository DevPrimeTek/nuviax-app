-- ═══════════════════════════════════════════════════════════════
-- Migration 011 — G-11 Behavior Model dominance
-- Adds dominant_behavior_model field to global_objectives
-- for EVOLVE override handling in hybrid goals
-- ═══════════════════════════════════════════════════════════════

-- Add dominant_behavior_model column to global_objectives
-- Values: 'ANALYTIC', 'STRATEGIC', 'TACTICAL', 'REACTIVE', or NULL (auto-detected)
ALTER TABLE global_objectives
ADD COLUMN IF NOT EXISTS dominant_behavior_model VARCHAR(20)
    DEFAULT NULL
    CHECK (dominant_behavior_model IS NULL OR
           dominant_behavior_model IN ('ANALYTIC', 'STRATEGIC', 'TACTICAL', 'REACTIVE'));

-- Index for behavior model lookups
CREATE INDEX IF NOT EXISTS idx_goals_behavior_model
    ON global_objectives (dominant_behavior_model)
    WHERE dominant_behavior_model IS NOT NULL;

-- Constraint: composite index for user + behavior model filtering
CREATE INDEX IF NOT EXISTS idx_goals_user_behavior_model
    ON global_objectives (user_id, dominant_behavior_model)
    WHERE status = 'ACTIVE' AND dominant_behavior_model IS NOT NULL;
