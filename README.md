# NuviaX App

Platform de management obiective bazată pe NuviaX Framework rev5.6.

## 📦 Structură

\\\
nuviax-app/
├── backend/          # Go API (Fiber)
├── frontend/
│   ├── app/         # Next.js 14 - Aplicația principală
│   └── landing/     # Next.js 14 - Landing page
├── infra/           # Docker Compose, deployment
├── deploy/          # Scripturi deployment (local)
└── .github/         # CI/CD workflows
\\\

## 🚀 Links

- **App:** https://nuviax.app
- **Landing:** https://nuviaxapp.com
- **API:** https://api.nuviax.app
- **Repo:** https://github.com/DevPrimeTek/nuviax-app

## 🛠️ Tech Stack

- Frontend: Next.js 14, React, TypeScript, Tailwind
- Backend: Go 1.22, Fiber, PostgreSQL, Redis
- Infrastructure: Docker, GitHub Actions, VPS

## 📋 Deployment

\\\powershell
cd deploy
.\Deploy-NuviaX-v9.0.ps1
\\\

## 📊 Changelog
### v10.0.0 - 16.03.2026
- 🚀 Deployment automat v10.0.0
- ✅ Fix: 404 error resolution
- ✅ Git sync și merge automat
- ✅ Backup pre-deployment
- ✅ Public assets verification


### v9.0 - 16.03.2026
- ✅ FIX CRITICAL: Dockerfile public/ folder handling
- ✅ Analiză completă structură proiect
- ✅ Creare automată foldere lipsă (public/, styles/)
- ✅ Verificare și creare favicon.ico
- ✅ README structurat permanent

**Versiune:** 10.0.0 | **Status:** ✅ Production Ready

## 🗄️ Database Migrations

### Applying Migrations

```bash
# Apply all migrations (idempotent — safe to re-run)
psql -U nuviax -d nuviax_dev -f backend/migrations/apply_all.sql

# Apply individual migrations
psql -U nuviax -d nuviax_dev -f backend/migrations/001_base_schema.sql
psql -U nuviax -d nuviax_dev -f backend/migrations/002_layer0_level1.sql
psql -U nuviax -d nuviax_dev -f backend/migrations/003_level2_execution.sql
psql -U nuviax -d nuviax_dev -f backend/migrations/004_level3_adaptive.sql
psql -U nuviax -d nuviax_dev -f backend/migrations/005_level4_regulatory.sql
psql -U nuviax -d nuviax_dev -f backend/migrations/006_level5_growth.sql
```

### Expected Schema After Migrations

| Object | Count | Description |
|--------|-------|-------------|
| Tables | 28 | Core (12) + Framework (16) |
| Views | 26 | Dashboard, execution, growth, regulatory |
| Materialized Views | 1 | `mv_user_stats` — user statistics |
| Functions | 10 | Score computation, trajectory, achievements |
| Triggers | 12 | Auto-init, consistency, milestones, ceremonies |

### Migration Map

| File | Level | New Tables | New Objects |
|------|-------|-----------|-------------|
| `001_base_schema.sql` | Foundation | 12 (users→audit_log) | 2 triggers, ENUMs |
| `002_layer0_level1.sql` | Layer 0 + L1 | goal_categories, sprint_configs, goal_metadata | 4 views, 1 fn, 3 triggers |
| `003_level2_execution.sql` | Level 2 | task_executions, daily_metrics, sprint_metrics | 5 views, 2 fn, 2 triggers |
| `004_level3_adaptive.sql` | Level 3 | behavior_patterns, consistency_snapshots, adaptive_weights | 5 views, 2 fn, 1 trigger |
| `005_level4_regulatory.sql` | Level 4 | regulatory_events, goal_activation_log, resource_slots | 5 views, 2+1 fn, 2 triggers |
| `006_level5_growth.sql` | Level 5 | growth_milestones, achievement_badges, ceremonies, growth_trajectories | 7 views, 1 matview, 2 fn, 3 triggers |

### Verify Schema

```sql
-- List all tables (should show 28)
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- List views (should show 26)
SELECT table_name FROM information_schema.views
WHERE table_schema = 'public' ORDER BY table_name;

-- List functions (should show 10)
SELECT routine_name FROM information_schema.routines
WHERE routine_schema = 'public' AND routine_type = 'FUNCTION'
ORDER BY routine_name;

-- Refresh materialized view (run daily or on demand)
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_stats;
```
