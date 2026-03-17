-- NUViaX Framework - Performance Check
\echo '========================================'
\echo 'View & Query Performance Check'
\echo '========================================'
\echo ''

\timing on

\echo '1. latest_ceremonies view:'
SELECT * FROM latest_ceremonies LIMIT 10;

\echo ''
\echo '2. active_srm_status view:'
SELECT * FROM active_srm_status LIMIT 10;

\echo ''
\echo '3. user_achievement_stats view:'
SELECT * FROM user_achievement_stats LIMIT 10;

\echo ''
\echo '4. unviewed_ceremonies view:'
SELECT * FROM unviewed_ceremonies LIMIT 10;

\echo ''
\echo '5. user_progress_overview materialized view:'
SELECT * FROM user_progress_overview LIMIT 10;

\echo ''
\echo '6. growth_trajectories view:'
SELECT * FROM growth_trajectories LIMIT 10;

\timing off

\echo ''
\echo '7. Index usage check (top tables):'
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan   AS scans,
  idx_tup_read AS tuples_read
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC
LIMIT 20;

\echo ''
\echo '8. Table sizes:'
SELECT
  relname AS table_name,
  pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
  n_live_tup AS live_rows
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 15;

\echo ''
\echo '✅ Performance check complete!'
