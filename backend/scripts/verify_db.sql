-- NUViaX Framework - Database Verification
\echo '========================================'
\echo 'Database Structure Verification'
\echo '========================================'
\echo ''

\echo '1. Tables Count:'
SELECT COUNT(*) AS total_tables
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
\echo 'Expected: 28'
\echo ''

\echo '2. Views Count (regular + materialized):'
SELECT
  (SELECT COUNT(*) FROM information_schema.views        WHERE table_schema = 'public')              AS regular_views,
  (SELECT COUNT(*) FROM pg_matviews                     WHERE schemaname   = 'public')              AS materialized_views;
\echo 'Expected: regular_views ~26, materialized_views ~1'
\echo ''

\echo '3. Functions Count:'
SELECT COUNT(*) AS total_functions
FROM information_schema.routines
WHERE routine_schema = 'public' AND routine_type = 'FUNCTION';
\echo 'Expected: 10'
\echo ''

\echo '4. Triggers Count:'
SELECT COUNT(*) AS total_triggers
FROM information_schema.triggers
WHERE trigger_schema = 'public';
\echo 'Expected: 12'
\echo ''

\echo '5. Table List:'
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;
\echo ''

\echo '6. View List (first 15):'
SELECT table_name
FROM information_schema.views
WHERE table_schema = 'public'
ORDER BY table_name
LIMIT 15;
\echo ''

\echo '7. Key L4/L5 Tables Present:'
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'srm_events', 'suspension_events', 'reactivation_protocols',
    'evolution_sprints', 'completion_ceremonies', 'user_achievements',
    'achievement_badges', 'growth_trajectories', 'progress_snapshots'
  )
ORDER BY table_name;
\echo ''

\echo '8. Test check_pause_limit function (null user):'
SELECT * FROM check_pause_limit('00000000-0000-0000-0000-000000000000'::uuid);
\echo ''

\echo '✅ Database verification complete!'
