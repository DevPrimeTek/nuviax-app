# NuviaX App

Platformă SaaS de management al obiectivelor personale și profesionale, bazată pe **NUViaX Growth Framework REV 5.6** — un sistem proprietar cu 5 niveluri (Layer 0 + Level 1-5) și 40 de componente (C1–C40).

**Versiune curentă:** `10.4.1` | **Status:** ✅ Production Ready

---

## Linkuri

| Serviciu | URL |
|---------|-----|
| Aplicație | https://nuviax.app |
| Landing Page | https://nuviaxapp.com |
| API | https://api.nuviax.app |
| Repository | https://github.com/DevPrimeTek/nuviax-app |

---

## Stack Tehnic

| Layer | Tehnologie | Versiune |
|-------|-----------|---------|
| Backend API | Go + Fiber v2 | Go 1.22, Fiber 2.52 |
| Database | PostgreSQL | 16 (Docker) |
| Cache / Sessions | Redis | 7 (Docker) |
| Frontend App | Next.js + TypeScript + Tailwind | Next.js 14, React 18 |
| Frontend Landing | Next.js (static) | Next.js 14 |
| Auth | JWT RS256 (RSA 4096-bit) | access: 15min, refresh: 7 zile |
| Email | Resend.com (tranzacțional) | — |
| AI | Anthropic Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| CI/CD | GitHub Actions → DockerHub → VPS SSH | — |
| Proxy | nginx-proxy + acme-companion | shared VPS |

---

## Structura Proiectului

```
nuviax-app/
├── .github/workflows/
│   ├── deploy.yml                   # CI/CD backend
│   └── deploy-frontend.yml          # CI/CD frontend
│
├── backend/
│   ├── cmd/server/main.go           # Entry point
│   ├── internal/
│   │   ├── ai/ai.go                 # Claude Haiku client (stdlib HTTP)
│   │   ├── api/
│   │   │   ├── server.go            # Rute + middleware Fiber
│   │   │   └── handlers/            # handlers.go, admin.go, srm.go,
│   │   │                            # ceremonies.go, achievements.go, visualization.go
│   │   ├── auth/auth.go             # JWT RS256 service
│   │   ├── cache/cache.go           # Redis helpers
│   │   ├── db/
│   │   │   ├── db.go                # pgxpool connect + RunMigrations
│   │   │   └── queries.go           # Toate query-urile
│   │   ├── email/email.go           # Resend.com client (stdlib HTTP)
│   │   ├── engine/
│   │   │   ├── engine.go            # Layer 0 — orchestrator + API publică
│   │   │   ├── helpers.go           # Utilitare interne
│   │   │   ├── level1_structural.go # C9-C18: Sprint, checkpoints, task gen
│   │   │   ├── level2_execution.go  # C19-C25: Execuție, Chaos Index, stagnare
│   │   │   ├── level3_adaptive.go   # C26-C31: Consistență, energie, context
│   │   │   ├── level4_regulatory.go # C32-C36: SRM, Vault, reactivare
│   │   │   └── level5_growth.go     # C37-C40: Evoluție, ceremonies, vizualizare
│   │   ├── models/models.go         # 50+ structuri Go
│   │   └── scheduler/scheduler.go  # 12 cron jobs
│   ├── migrations/
│   │   ├── 001_base_schema.sql
│   │   ├── 002_layer0_level1.sql
│   │   ├── 003_level2_execution.sql
│   │   ├── 004_level3_adaptive.sql
│   │   ├── 005_level4_regulatory.sql
│   │   ├── 006_level5_growth.sql
│   │   ├── 007_admin_fixes.sql
│   │   ├── 008_avatar.sql
│   │   ├── 009_password_reset.sql
│   │   ├── 010_p1_gaps.sql          # srm_events, reactivation_protocols, stagnation_events
│   │   └── apply_all.sql
│   └── scripts/
│       ├── test_all.sh              # Build validation + gofmt
│       ├── test_api.sh              # Teste curl endpoint-uri
│       ├── verify_db.sql            # Verificare schema DB
│       └── performance_check.sql   # Timing + index stats
│
├── frontend/
│   ├── app/                         # nuviax.app (Next.js App Router)
│   │   ├── app/
│   │   │   ├── admin/               # Panel admin (doar pt is_admin=true)
│   │   │   ├── achievements/
│   │   │   ├── api/                 # Route handlers (auth + proxy JWT)
│   │   │   ├── auth/                # login, register, forgot/reset-password
│   │   │   ├── dashboard/
│   │   │   ├── goals/               # list + [id] (detalii + charts)
│   │   │   ├── onboarding/
│   │   │   ├── profile/             # Upload avatar
│   │   │   ├── recap/
│   │   │   ├── settings/            # Parolă + export date
│   │   │   └── today/               # Energy + sarcini zilnice
│   │   ├── components/
│   │   │   ├── layout/AppShell.tsx  # Nav + link Admin condiționat
│   │   │   ├── CeremonyModal.tsx
│   │   │   ├── DashboardClientLayer.tsx
│   │   │   ├── GoalTabs.tsx
│   │   │   ├── ProgressCharts.tsx   # Recharts
│   │   │   └── SRMWarning.tsx
│   │   └── styles/globals.css       # Design system CSS vars + dark/light
│   └── landing/                     # nuviaxapp.com (static)
│
├── infra/
│   ├── docker-compose.yml           # Prod: DB + Redis + API
│   ├── docker-compose.frontend.yml  # Prod: App + Landing
│   ├── .env.example                 # Template variabile environment
│   ├── GITHUB_SECRETS.md            # Ghid configurare GitHub Secrets
│   ├── deploy.sh                    # Deploy manual
│   └── setup-server.sh             # Setup inițial VPS
│
├── CLAUDE.md                        # Context master sesiuni de dezvoltare
├── CHANGES.md                       # Changelog detaliat v1→v10.4
├── ROADMAP.md                       # Plan și priorități
└── README.md                        # ← acest fișier
```

---

## NUViaX Framework REV 5.6 — 40 Componente

| Layer | Componente | Implementare |
|-------|-----------|-------------|
| Layer 0 | C1-C8: Drift, Chaos Index, Continuity, GORI, Visibility, Priority, ALI, Behavior | `engine.go` |
| Level 1 | C9-C18: Sprint Architecture, Checkpoints, Task Gen (AI), Intensity, Velocity Control | `level1_structural.go` |
| Level 2 | C19-C25: Completion Rate, Sprint Score (40/25/25/10), Chaos Index, Stagnation, Focus Rotation | `level2_execution.go` |
| Level 3 | C26-C31: Consistency, Context Factors, Energy Bonus/Penalty, Pause | `level3_adaptive.go` |
| Level 4 | C32-C36: Activation Rules, Future Vault, SRM L1/L2/L3, Stabilization, Reactivation | `level4_regulatory.go` |
| Level 5 | C37-C40: Evolution Sprints, Ceremonies (BRONZE→PLATINUM), Achievements, Visualization | `level5_growth.go` |

**Total: 40/40 componente implementate**

---

## Scheduler Jobs (12 total)

| # | Cron | Job | Descriere |
|---|------|-----|-----------|
| 1 | `0 0 * * *` | GenerateDailyTasks | Generează sarcini zilnice pentru toți userii activi |
| 2 | `50 23 * * *` | ComputeDailyScore | Calculează scorul zilnic per obiectiv |
| 3 | `55 23 * * *` | CheckDailyProgress | Verifică progres + actualizează checkpoints |
| 4 | `1 0 * * *` | CloseExpiredSprints | Închide sprinturi expirate + trimite email |
| 5 | `0 2 */90 * *` | RecalibrateRelevance | Recalibrare 90 zile + Chaos Index storage |
| 6 | `0 1 * * *` | DetectEvolutionSprints | Detectează sprinturi cu evoluție (Δ ≥5%) |
| 7 | `5 1 * * *` | GenerateCeremonies | Generează ceremonies pentru sprinturi completate |
| 8 | `5 0 * * *` | ProgressReactivation | Avansează protocolul de reactivare (ramp 0.2→1.0) |
| 9 | `0 * * * *` | CheckSRMTimeouts | Verifică SRM L3 neconfirmat la 24/72/168h |
| 10 | `0 * * * *` | RefreshProgressOverview | Refresh materialized view analytics |
| 11 | `58 23 * * *` | DetectStagnation | Detectează GO cu ≥5 zile inactive → stagnation_events |
| 12 | `10 0 * * *` | ProposeReactivation | Propune reactivare GO PAUSED cu ≥7 zile stabilitate |

---

## API Endpoints

### Auth (public)
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
```

### Protected (JWT required)
```
GET  /api/v1/dashboard
GET  /api/v1/goals               POST /api/v1/goals
GET  /api/v1/goals/:id           PATCH/DELETE /api/v1/goals/:id
POST /api/v1/goals/:id/activate
GET  /api/v1/today               POST /api/v1/today/complete/:id
POST /api/v1/today/personal
POST /api/v1/context/pause       POST /api/v1/context/energy
GET  /api/v1/srm/status/:goalId
POST /api/v1/srm/confirm-l2/:goalId    # SRM L2 — single confirm
POST /api/v1/srm/confirm-l3/:goalId    # SRM L3 — double confirm + pause
GET  /api/v1/settings            PATCH /api/v1/settings
GET  /api/v1/ceremonies/unviewed
GET  /api/v1/achievements
GET  /api/v1/goals/:id/visualize
```

### Admin (JWT + is_admin = TRUE)
```
GET  /api/v1/admin/stats
GET  /api/v1/admin/users
GET  /api/v1/admin/audit
POST /api/v1/admin/users/:id/promote
```

---

## Database

**10 migrații → 33+ tabele, 26+ views, 1 materialized view, 10 funcții, 12 triggers**

| Migrație | Conținut |
|----------|---------|
| `001_base_schema.sql` | users, sessions, goals, sprints, tasks, checkpoints, audit_log |
| `002_layer0_level1.sql` | goal_categories, sprint_configs, goal_metadata, views Layer 0 |
| `003_level2_execution.sql` | task_executions, daily_metrics, sprint_metrics |
| `004_level3_adaptive.sql` | behavior_patterns, consistency_snapshots, adaptive_weights |
| `005_level4_regulatory.sql` | regulatory_events, goal_activation_log, resource_slots |
| `006_level5_growth.sql` | growth_milestones, achievement_badges, completion_ceremonies, evolution_sprints |
| `007_admin_fixes.sql` | is_admin col, regression_events, ali_snapshots, freeze cols pe sprints |
| `008_avatar.sql` | users.avatar_url |
| `009_password_reset.sql` | password_reset_tokens (1h TTL, single-use) |
| `010_p1_gaps.sql` | srm_events, reactivation_protocols, stagnation_events |

```bash
# Aplică toate migrările (idempotent)
docker exec -i nuviax_db psql -U nuviax -d nuviax < backend/migrations/apply_all.sql
```

---

## Environment Variables

```env
# Database
POSTGRES_HOST=nuviax_db
POSTGRES_PORT=5432
POSTGRES_USER=nuviax
POSTGRES_PASSWORD=<generate: openssl rand -base64 32>
POSTGRES_DB=nuviax

# Redis
REDIS_HOST=nuviax_redis
REDIS_PORT=6379
REDIS_PASSWORD=<generate: openssl rand -base64 32>

# Auth
JWT_PRIVATE_KEY=<RSA 4096-bit, base64 encoded>
JWT_PUBLIC_KEY=<RSA 4096-bit public, base64 encoded>
ENCRYPTION_KEY=<openssl rand -hex 32>

# AI (optional — graceful degradation if absent)
ANTHROPIC_API_KEY=sk-ant-...

# Email (optional — graceful degradation if absent)
RESEND_API_KEY=re_...
EMAIL_FROM=NuviaX <noreply@nuviax.app>

# Server
PORT=8080
ALLOWED_ORIGINS=https://nuviax.app,https://www.nuviax.app
```

---

## Deployment

### Automat (recomandat)

Push pe `main` → GitHub Actions → DockerHub → VPS SSH:
```
.github/workflows/deploy.yml          # backend
.github/workflows/deploy-frontend.yml # frontend
```

GitHub Secrets necesare: `SSH_HOST`, `SSH_PORT`, `SSH_USER`, `SSH_KEY`, `DOCKERHUB_TOKEN`, `POSTGRES_PASSWORD`, `REDIS_PASSWORD`, `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`, `ENCRYPTION_KEY`, `ANTHROPIC_API_KEY`, `RESEND_API_KEY`

### Manual

```bash
cd infra
cp .env.example .env      # completează variabilele
./deploy.sh               # sau: docker compose up -d
```

---

## Testare

```bash
# Build validation + gofmt
cd backend && ./scripts/test_all.sh

# API endpoint tests (necesită server pornit + JWT token)
TOKEN=<jwt> ./backend/scripts/test_api.sh

# Verificare integritate schema DB
docker exec -i nuviax_db psql -U nuviax -d nuviax \
  -f backend/scripts/verify_db.sql
```

---

## Principii de Securitate

- **Engine opac** — formulele și ponderile nu ies niciodată din `internal/engine/`
- **JWT RS256** — access token 15min, refresh 7 zile
- **Email hash** — adresele email sunt stocate criptat (AES-256-GCM) + hash SHA-256 pentru lookup
- **Admin 404** — panel-ul admin returnează 404 (nu 403) pentru non-admini
- **Timing-safe** — `forgot-password` returnează mereu 200 (previne user enumeration)
- **Graceful degradation** — AI și Email funcționează fără cheile respective (fallback automat)

---

## Changelog Rapid

| Versiune | Data | Descriere |
|---------|------|-----------|
| v10.4.1 | 2026-03-26 | Admin page standalone (fără AppShell); setup_admin.sh script |
| v10.4.0 | 2026-03-26 | P1 Gaps: G-1—G-10, G-12 (10/12) implementate; migration 010 |
| v10.3.1 | 2026-03-26 | Admin fix: is_admin în nav; cleanup fișiere duplicate |
| v10.3.0 | 2026-03-25 | Email Resend: welcome + sprint complet + forgot/reset parolă |
| v10.2.0 | 2026-03-24 | Fix toate bug-urile B-2—B-11; AI integration Claude Haiku; upload avatar |
| v10.1.0 | 2026-03-20 | P0 Gaps: regresie, pauză retroactivă, drift paradox, ALI disambiguation |
| v10.0.0 | 2026-03-16 | Restructurare completă, deploy automat CI/CD |

> Changelog complet: [CHANGES.md](./CHANGES.md)
