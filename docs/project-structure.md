# Structura Proiectului NuviaX

> Referință completă a structurii repo și design system frontend.
> Actualizează la orice modificare structurală majoră.

---

## Structura Repo (v10.5.0)

```
nuviax-app/
├── .github/workflows/
│   ├── deploy.yml                   # CI/CD backend: push main → DockerHub → VPS
│   └── deploy-frontend.yml          # CI/CD frontend: push main → DockerHub → VPS
│
├── backend/
│   ├── cmd/server/main.go           # Entry point: config, DB, Redis, email, scheduler, HTTP
│   ├── internal/
│   │   ├── ai/ai.go                 # Claude Haiku 4.5 HTTP client (fără SDK)
│   │   ├── api/
│   │   │   ├── server.go            # Toate rutele + middleware Fiber
│   │   │   ├── handlers/
│   │   │   │   ├── handlers.go      # Auth, Goals, Tasks, Sprint, Context, Settings, Recap
│   │   │   │   ├── admin.go         # Admin panel: stats, users, audit, health, dev-reset
│   │   │   │   ├── srm.go           # SRM (Strategic Reset Management) L1/L2/L3
│   │   │   │   ├── ceremonies.go    # Level 5: GetLatestCeremony, MarkViewed
│   │   │   │   ├── achievements.go  # Level 5: badge grid
│   │   │   │   └── visualization.go # Level 5: progress charts data
│   │   │   └── middleware/
│   │   │       ├── jwt.go           # JWT RS256 auth middleware
│   │   │       └── admin.go         # AdminOnly: verifică is_admin=TRUE, returnează 404 altfel
│   │   ├── auth/auth.go             # JWT service (RS256, access 15min, refresh 7 zile)
│   │   ├── cache/cache.go           # Redis helpers (sessions, dashboard cache)
│   │   ├── db/
│   │   │   ├── db.go                # pgxpool connect + RunMigrations
│   │   │   └── queries.go           # Toate query-urile (users, goals, sprints, tasks, SRM, admin, email reset)
│   │   ├── email/email.go           # Resend.com HTTP client: Welcome + PasswordReset + SprintComplete
│   │   ├── engine/
│   │   │   ├── engine.go            # Layer 0 (C1-C8) + API publică
│   │   │   ├── helpers.go           # Funcții interne reutilizabile
│   │   │   ├── level1_structural.go # C9-C18: Sprint, checkpoints, task generation (Claude Haiku)
│   │   │   ├── level2_execution.go  # C19-C25: Execution rate, regression events
│   │   │   ├── level3_adaptive.go   # C26-C31: Consistency, energy, context
│   │   │   ├── level4_regulatory.go # C32-C36: Activation rules, SRM
│   │   │   └── level5_growth.go     # C37-C40: Evolution, ceremonies, visualization
│   │   ├── models/models.go         # 50+ structuri Go (UserSettings include is_admin din v10.3)
│   │   └── scheduler/scheduler.go  # 12 cron jobs: daily tasks, sprint close, ceremonies, SRM, email
│   ├── migrations/
│   │   ├── 001_base_schema.sql      # Core tables: users, sessions, goals, sprints, tasks, audit
│   │   ├── 002_layer0_level1.sql    # Layer 0 + Level 1 tables
│   │   ├── 003_level2_execution.sql # Level 2 tables
│   │   ├── 004_level3_adaptive.sql  # Level 3 tables
│   │   ├── 005_level4_regulatory.sql # Level 4 + SRM tables
│   │   ├── 006_level5_growth.sql    # Level 5: ceremonies, achievements, trajectories
│   │   ├── 007_admin_fixes.sql      # Admin + 5 P0 gap fixes: regression, ALI, retroactive pause
│   │   ├── 008_avatar.sql           # users.avatar_url
│   │   ├── 009_password_reset.sql   # password_reset_tokens (forgot-password flow)
│   │   ├── 010_p1_gaps.sql          # srm_events, reactivation_protocols, stagnation_events
│   │   ├── 011_behavior_model.sql   # dominant_behavior_model on global_objectives (G-11)
│   │   ├── 012_theme.sql            # users.theme (dark/light preference persistence)
│   │   └── apply_all.sql            # Script aplicare toate migrările (idempotent)
│   ├── pkg/
│   │   ├── crypto/crypto.go         # AES-256-GCM, PBKDF2, bcrypt, SHA256, RandomHex
│   │   └── logger/logger.go         # Uber Zap structured logging
│   └── scripts/
│       ├── test_all.sh              # Build validation + gofmt check
│       ├── test_api.sh              # Teste curl pe endpoint-uri API
│       ├── verify_db.sql            # Verificare integritate schema DB
│       ├── performance_check.sql    # View timing + index stats
│       └── integration_test.md     # Ghid E2E test manual (10 scenarii)
│
├── frontend/
│   ├── app/                         # Aplicația principală → nuviax.app
│   │   ├── app/                     # Next.js App Router pages
│   │   │   ├── admin/page.tsx       # Panel admin (acces: nuviax.app/admin)
│   │   │   ├── achievements/page.tsx
│   │   │   ├── api/
│   │   │   │   ├── auth/            # login, register, logout, forgot-password, reset-password, set
│   │   │   │   └── proxy/[...path]/ # Proxy JWT auto-refresh → backend
│   │   │   ├── auth/                # login, register, forgot-password, reset-password pages
│   │   │   ├── dashboard/page.tsx
│   │   │   ├── goals/               # list + [id]/page.tsx (detalii + charts)
│   │   │   ├── onboarding/page.tsx
│   │   │   ├── profile/page.tsx     # Upload avatar
│   │   │   ├── recap/page.tsx
│   │   │   ├── settings/page.tsx    # Schimbare parolă + export date
│   │   │   ├── today/page.tsx       # Energy + sarcini principale + sarcini personale
│   │   │   ├── layout.tsx           # Root layout cu fonturi
│   │   │   └── page.tsx             # Redirect → /dashboard sau /auth/login
│   │   ├── components/
│   │   │   ├── layout/AppShell.tsx  # Nav + link Admin condiționat (is_admin) + dark/light toggle
│   │   │   ├── ActivityHeatmap.tsx  # GitHub-style 52-week activity grid (Sprint 3)
│   │   │   ├── CeremonyModal.tsx    # Modal ceremonie sprint (BRONZE/SILVER/GOLD/PLATINUM)
│   │   │   ├── DashboardClientLayer.tsx
│   │   │   ├── GoalTabs.tsx         # Tabs Prezentare / Progres
│   │   │   ├── ProgressCharts.tsx   # LineChart + BarChart (Recharts)
│   │   │   └── SRMWarning.tsx       # Bannere SRM L1/L2/L3
│   │   ├── lib/
│   │   │   ├── api.ts               # API client helpers
│   │   │   └── i18n.ts              # useTranslation() hook (EN/RU/RO, no external lib)
│   │   ├── public/locales/
│   │   │   ├── ro.json              # Traduceri română (sursă de adevăr)
│   │   │   ├── en.json              # Traduceri English
│   │   │   └── ru.json              # Переводы Русский
│   │   ├── middleware.ts            # Auth middleware Next.js
│   │   └── styles/globals.css       # Design system: CSS vars, dark/light theme
│   └── landing/                     # Landing page → nuviaxapp.com
│       └── app/page.tsx
│
├── infra/
│   ├── docker-compose.yml           # Prod: nuviax_db + nuviax_redis + nuviax_api
│   ├── docker-compose.frontend.yml  # Prod: nuviax_app (port 3000) + nuviax_landing (port 3001)
│   ├── .env.example                 # Template complet variabile environment
│   ├── GITHUB_SECRETS.md            # Ghid configurare GitHub Secrets CI/CD
│   ├── deploy.sh                    # Script deploy manual
│   ├── setup-server.sh              # Setup inițial VPS
│   └── verify-deployment.sh         # Verificare health post-deploy
│
├── scripts/
│   └── test_components.sh
│
├── docs/                            # ← Documentație extinsă (nu se încarcă în Project Knowledge)
│   ├── project-structure.md         # Acest fișier
│   ├── frontend-design-system.md    # CSS vars, fonturi, tema
│   ├── database-reference.md        # Schema DB completă, migrări
│   ├── integrations.md              # AI (Haiku), Email (Resend), detalii implementare
│   ├── deployment.md                # VPS, Docker, CI/CD, secrets
│   └── history-bugs-gaps.md         # Bug-uri rezolvate, gap-uri stress test
│
├── CLAUDE.md                        # Context master sesiuni (CITIT LA START)
├── CHANGES.md                       # Changelog detaliat
├── ROADMAP.md                       # Plan și priorități
└── README.md                        # Documentație principală + setup
```

---

## Frontend Design System

**Fișier CSS:** `frontend/app/styles/globals.css`

### Variabile CSS

```css
--bg, --bg2, --bg3         /* fundal principal / secundar / terțiar */
--ink, --ink2, --ink3, --ink4  /* text principal → subtil */
--line, --line2            /* borduri */
--l0, --l0l, --l0g         /* Level 0 / light / glow (portocaliu) */
--l2, --l2l, --l2g         /* Level 2 / light / glow (verde) */
--l5, --l5l                /* Level 5 / light (violet) */
--u, --ul, --ug            /* urgency / urgency-light / urgency-glow */
--ff-d, --ff-b, --ff-m, --ff-h  /* fonturi display, body, mono, heading */
```

**Tema Light:** `[data-theme="light"] { ... }` — prezentă în v10.2

### Fonturi (import în `layout.tsx`)

| Variabilă | Font | Utilizare |
|-----------|------|-----------|
| `--ff-d`, `--ff-h` | Bricolage Grotesque | Titluri, display |
| `--ff-b` | DM Sans | Body text |
| `--ff-m` | JetBrains Mono | Badge-uri, cod |

### Persistență temă

```javascript
localStorage.getItem('nv_theme')   // dark / light
localStorage.getItem('nv_lang')    // ro / en / ru
// Rehydrat pe: document.documentElement.dataset.theme
```

---

## Fișiere șterse (istoric)

| Fișier | Versiune | Motiv |
|--------|---------|-------|
| `NuviaX_UI_Mockup_v4.html` | v10.3.1 | Înlocuit de implementare reală |
| `ANALYSIS_REPORT.md` | v10.3.1 | Integrat în `CHANGES.md` |
| `IMPLEMENTATION_CHECKLIST.md` | v10.3.1 | Înlocuit de `ROADMAP.md` |
| `TEST_REPORT.md` | v10.3.1 | Generat automat, nu se ține în git |
| `frontend/infra/` | v10.3.1 | Director duplicat |
| `frontend/.github/workflows/` | v10.3.1 | Workflows duplicate |

---

*Actualizat: v10.5.0 — 2026-03-29 — Sprint 3 complete: i18n EN/RU, AI onboarding, activity heatmap, theme persistence, migration 012*
