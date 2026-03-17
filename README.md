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

---

## NUViaX Framework REV 5.6 — Complete Implementation

### Features Implemented

| Layer | Components | Status |
|-------|-----------|--------|
| Layer 0 | Drift, Chaos, Continuity, GORI, Visibility, Priority (C1–C8) | ✅ |
| Level 1 | Sprint Architecture, Checkpoints, Smart Reactivation (C9–C18) | ✅ |
| Level 2 | Execution Matrix, Daily Stack, Progress, Velocity (C19–C25) | ✅ |
| Level 3 | Context, Energy, Pause, Rhythm, Consistency, Behavioral (C26–C31) | ✅ |
| Level 4 | Adaptive Context, SRM, Stabilization, Reactivation (C32–C36) | ✅ |
| Level 5 | Evolution, Ceremonies, Achievements, Visualization (C37–C40) | ✅ |

**Total: 40/40 components implemented**

### Backend — Go (Fiber v2 + pgx/v5)

```
backend/internal/
├── models/models.go          # 50+ structs, 15 enum types
├── engine/
│   ├── engine.go             # Core orchestrator
│   ├── level1_sprints.go     # Sprint + checkpoint logic
│   ├── level2_execution.go   # Execution matrix, velocity
│   ├── level3_adaptive.go    # Behavior + consistency
│   ├── level4_regulatory.go  # SRM state machine
│   └── level5_growth.go      # Evolution, ceremonies, achievements
├── api/
│   ├── server.go             # 40+ routes
│   └── handlers/             # ceremonies, achievements, srm, visualization
└── scheduler/scheduler.go    # 14 background jobs
```

### Scheduler Jobs (14 total)

| Schedule | Job | Description |
|----------|-----|-------------|
| `0 0 * * *` | Daily tasks | Generate tasks for all active goals |
| `0 1 * * *` | Sprint close | Close finished sprints |
| `5 1 * * *` | Ceremonies | Generate completion ceremonies |
| `0 2 * * *` | Checkpoints | Evaluate checkpoint progress |
| `0 3 * * *` | SRM evaluation | Assess strategic reset triggers |
| `0 1 * * *` | Evolution detection | Detect evolution sprints (Δ ≥ 5%) |
| `5 1 * * *` | Ceremony gen | Create ceremonies for evolution |
| `5 0 * * *` | Reactivation progress | Advance reactivation protocol days |
| `0 * * * *` | SRM timeouts | Apply fallback labels at 24/72/168h |
| `0 * * * *` | Progress overview | Refresh materialized view |
| *(+ 4 others)* | | Drift, chaos, continuity, GORI |

### Frontend — Next.js 14 App Router

```
frontend/app/
├── app/
│   ├── dashboard/page.tsx    # SRMWarning + DashboardClientLayer
│   ├── achievements/page.tsx # Badge grid by category
│   ├── goals/[id]/page.tsx   # Goal detail with tabs
│   └── today/page.tsx        # Daily tasks
└── components/
    ├── CeremonyModal.tsx      # Tier-colored completion modal
    ├── DashboardClientLayer.tsx # Ceremony auto-check on mount
    ├── SRMWarning.tsx         # L1/L2/L3 warning banners
    ├── ProgressCharts.tsx     # recharts line + bar trajectory
    └── GoalTabs.tsx           # Overview / Progress tab switcher
```

### Testing

```bash
# Backend validation (offline-safe)
./backend/scripts/test_all.sh

# Database verification (requires PostgreSQL + nuviax_dev)
psql -U nuviax -d nuviax_dev -f backend/scripts/verify_db.sql

# API endpoint tests (requires running server + JWT token)
TOKEN=<jwt> ./backend/scripts/test_api.sh

# Performance check
psql -U nuviax -d nuviax_dev -f backend/scripts/performance_check.sql

# Integration test guide
cat backend/scripts/integration_test.md
```

### Deployment

```bash
# Apply all migrations
psql -U nuviax -d nuviax_prod -f backend/migrations/apply_all.sql

# Build & start backend
cd backend && go build -o server ./cmd/server && ./server

# Build & start frontend
cd frontend/app && npm run build && npm start
```

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
