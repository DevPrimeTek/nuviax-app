-- ═══════════════════════════════════════════════════════════════
-- NUVIAX FRAMEWORK REV 5.6 — Apply All Migrations
--
-- Usage:
--   psql -U nuviax -d nuviax_dev -f backend/migrations/apply_all.sql
--
-- Note: Run migrations in order. Each migration is idempotent
-- (IF NOT EXISTS, ON CONFLICT DO NOTHING) — safe to re-run.
-- ═══════════════════════════════════════════════════════════════

\set ON_ERROR_STOP on

\echo ''
\echo '══════════════════════════════════════════════════════'
\echo '  NUViaX Framework REV 5.6 — Database Migrations'
\echo '══════════════════════════════════════════════════════'
\echo ''

-- ── Migration 001: Base Schema (12 tables) ────────────────────
\echo '[001/006] Applying base schema (Layer -1)...'
\i 001_base_schema.sql

-- ── Migration 002: Layer 0 + Level 1 (3 tables) ──────────────
\echo '[002/006] Applying Layer 0 + Level 1 (Structural Authority)...'
\i 002_layer0_level1.sql

-- ── Migration 003: Level 2 (3 tables) ────────────────────────
\echo '[003/006] Applying Level 2 (Execution Engine)...'
\i 003_level2_execution.sql

-- ── Migration 004: Level 3 (3 tables) ────────────────────────
\echo '[004/006] Applying Level 3 (Adaptive Intelligence)...'
\i 004_level3_adaptive.sql

-- ── Migration 005: Level 4 (3 tables) ────────────────────────
\echo '[005/006] Applying Level 4 (Regulatory Authority)...'
\i 005_level4_regulatory.sql

-- ── Migration 006: Level 5 (4 tables) ────────────────────────
\echo '[006/006] Applying Level 5 (Growth Orchestration)...'
\i 006_level5_growth.sql

\echo ''
\echo '══════════════════════════════════════════════════════'
\echo '  All migrations applied successfully!'
\echo '══════════════════════════════════════════════════════'
\echo ''

-- ── Schema verification ───────────────────────────────────────
\echo 'Schema verification:'
\echo ''

SELECT
    COUNT(*) AS total_tables
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE';

SELECT
    COUNT(*) AS total_views
FROM information_schema.views
WHERE table_schema = 'public';

SELECT
    COUNT(*) AS total_materialized_views
FROM pg_matviews
WHERE schemaname = 'public';

SELECT
    COUNT(*) AS total_functions
FROM information_schema.routines
WHERE routine_schema = 'public'
  AND routine_type = 'FUNCTION';

SELECT
    COUNT(*) AS total_triggers
FROM information_schema.triggers
WHERE trigger_schema = 'public';

\echo ''
\echo 'Expected: 28 tables, 26 views, 1 materialized view,'
\echo '          10 functions, 12 triggers'
\echo ''

-- ── Detailed table list ───────────────────────────────────────
\echo 'Tables created:'

SELECT
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns
     WHERE table_schema = 'public' AND table_name = t.table_name) AS columns
FROM information_schema.tables t
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY table_name;
