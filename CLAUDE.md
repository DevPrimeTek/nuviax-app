# CLAUDE.md — NuviaX App: Context Master pentru Sesiuni de Dezvoltare

> **Citește acest fișier la ÎNCEPUTUL ORICĂREI sesiuni noi.**
> Conține tot contextul necesar pentru a continua dezvoltarea fără a reface munca anterioară.

---

## 1. Ce este NuviaX

**NuviaX** este o platformă SaaS de management al obiectivelor personale și profesionale, bazată pe **NUViaX Growth Framework REV 5.6** — un sistem proprietar cu 5 niveluri (Layer 0 + Level 1-5) și 40 de componente (C1–C40).

**Principiu fundamental:** Toate calculele (scoruri, sarcini, metrici) rulează exclusiv pe server. Clientul primește doar rezultate opace (%, grade A/B/C/D, liste de sarcini). Nicio formulă sau pondere nu este expusă în afara engine-ului Go.

**Produse:**
- `nuviax.app` — aplicația principală (Next.js)
- `nuviaxapp.com` — landing page (Next.js)
- `api.nuviax.app` — backend API (Go)

**Proprietar:** DevPrimeTek (`github.com/DevPrimeTek/nuviax-app`)
**Versiune curentă:** 10.4.1
**Branch de development:** `claude/*` → PR → `main`

---

## 2. Stack Tehnic

| Layer | Tehnologie | Versiune |
|-------|-----------|---------|
| Backend API | Go + Fiber v2 | Go 1.22, Fiber 2.52 |
| Database | PostgreSQL | 16 (Docker) |
| Cache/Sessions | Redis | 7 (Docker) |
| Frontend App | Next.js + TypeScript + Tailwind | Next.js 14, React 18 |
| Frontend Landing | Next.js (static) | Next.js 14 |
| Auth | JWT RS256 (RSA 4096-bit) | access: 15min, refresh: 7 zile |
| Emails | Resend.com (tranzacțional) | — |
| AI | Anthropic Claude Haiku 4.5 | model: `claude-haiku-4-5-20251001` |
| CI/CD | GitHub Actions → DockerHub → VPS SSH | — |
| Proxy | nginx-proxy + acme-companion (jwilder) | shared cu alte proiecte |
| Deploy path | `/var/www/wxr-nuviax/` pe VPS | — |

---

## 3. Structura Proiectului (v10.3 — curățată)

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
│   │   └── scheduler/scheduler.go  # 10 cron jobs: daily tasks, sprint close, ceremonies, SRM, email
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
│   │   │   ├── admin/page.tsx       # ✅ Panel admin (acces: nuviax.app/admin, link în nav doar pt admini)
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
│   │   │   ├── layout.tsx           # Root layout cu fonturi (Bricolage, DM Sans, JetBrains Mono)
│   │   │   └── page.tsx             # Redirect → /dashboard sau /auth/login
│   │   ├── components/
│   │   │   ├── layout/AppShell.tsx  # Nav + link Admin condiționat (is_admin din settings)
│   │   │   ├── CeremonyModal.tsx    # Modal ceremonie sprint (BRONZE/SILVER/GOLD/PLATINUM)
│   │   │   ├── DashboardClientLayer.tsx # Polls /ceremonies/unviewed
│   │   │   ├── GoalTabs.tsx         # Tabs Prezentare / Progres
│   │   │   ├── ProgressCharts.tsx   # LineChart + BarChart (Recharts)
│   │   │   └── SRMWarning.tsx       # Bannere SRM L1/L2/L3
│   │   ├── lib/api.ts               # API client helpers
│   │   ├── middleware.ts            # Auth middleware Next.js (redirect neautentificați)
│   │   └── styles/globals.css       # Design system: CSS vars, dark/light theme, componente
│   └── landing/                     # Landing page → nuviaxapp.com
│       └── app/page.tsx             # Pagina principală landing (statică)
│
├── infra/
│   ├── docker-compose.yml           # Prod: nuviax_db + nuviax_redis + nuviax_api
│   ├── docker-compose.frontend.yml  # Prod: nuviax_app (port 3000) + nuviax_landing (port 3001)
│   ├── .env.example                 # Template complet variabile environment
│   ├── GITHUB_SECRETS.md            # Ghid configurare GitHub Secrets CI/CD
│   ├── deploy.sh                    # Script deploy manual
│   ├── setup-server.sh              # Setup inițial VPS (Docker, nginx-proxy, etc.)
│   └── verify-deployment.sh         # Verificare health post-deploy
│
├── scripts/
│   └── test_components.sh           # Teste C1-C40 comprehensive
│
├── CLAUDE.md                        # ← Context master (citit la START oricărei sesiuni)
├── CHANGES.md                       # Changelog detaliat (v1.x → v10.3)
├── ROADMAP.md                       # Planul de dezvoltare și priorități
└── README.md                        # Documentație principală + setup
```

**Fișiere șterse în v10.3 (cleanup):**
- `NuviaX_UI_Mockup_v4.html` — mockup vechi, înlocuit de implementare reală
- `ANALYSIS_REPORT.md` — raport inițial, integrat în CHANGES.md
- `IMPLEMENTATION_CHECKLIST.md` — checklist vechi, înlocuit de această secțiune
- `TEST_REPORT.md` — generat automat, nu se mai ține în git
- `frontend/infra/` — director duplicat (conținut mutat în `infra/`)
- `frontend/.github/workflows/` — workflows duplicate (există în `.github/workflows/`)

---

## 4. Starea Curentă — Bug-uri

### ✅ Toate bug-urile rezolvate în v10.2.0 (2026-03-24)

| # | Locație | Problema | Status |
|---|---------|---------|--------|
| B-3 | `handlers.go` | `days_left` folosea `goal.EndDate` în loc de `sprint.EndDate` | ✅ Rezolvat |
| B-7 | `handlers.go` + `goals/page.tsx` | Goals returna array plat; acum returnează `{goals:[], waiting:[]}` | ✅ Rezolvat |
| B-8 | `server.go` + `handlers.go` | `GET /recap/current` lipsea — 404 | ✅ Rezolvat |
| B-5 | `today/page.tsx` | Energy nu se salva — endpoint greșit | ✅ Rezolvat |
| B-6 | `today/page.tsx` | Fără formular sarcini personale | ✅ Rezolvat |
| B-11 | `styles/globals.css` | `--ul`, `--ug`, `--l2g`, `--ff-h` lipseau; Light theme lipsă | ✅ Rezolvat |
| B-9 | `settings/page.tsx` | Parolă/export neconectate la API | ✅ Rezolvat |
| B-4 | `level1_structural.go` | Task generation static fără AI | ✅ Rezolvat (Claude Haiku cu fallback) |
| B-2 | `handlers.go AnalyzeGO` | Analiză GO fără AI | ✅ Rezolvat (Claude Haiku cu fallback) |
| B-10 | `profile/page.tsx` | Fără upload foto profil | ✅ Rezolvat |

### 🟠 În lucru / Următor

| # | Descriere | Status |
|---|-----------|--------|
| E-1 | Integrare Resend email (confirmare înregistrare, reset parolă, sprint complet) | ✅ Implementat v10.3 |
| E-2 | P1 Gaps din stress test (12 gap-uri medii) | ✅ 10/12 implementate v10.4 (G-11 rămâne) |
| E-3 | Translations EN + RU | ⏳ Neimplementat |
| E-4 | Monetizare Stripe | ⏳ Neimplementat |

---

## 5. Gap-uri Stress Test — Stare

### ✅ Rezolvate (P0 — Critical, rezolvate în v10.1)

| Gap | Problema | Fișier modificat |
|-----|---------|-----------------|
| #14 | Pauza retroactivă (max 48h) | `handlers.go`, `db/queries.go`, migration 007 |
| #15 | Regression event (valoare sub sprint start) | `level2_execution.go`, migration 007 |
| #20 | Drift loop paradox în Stabilization Mode | `level5_growth.go`, migration 007 |
| #8 | ALI total vs per-GO ambiguitate | `srm.go` |
| #13 | ALI current vs projected ambiguitate | `srm.go` |

### ⏳ Nerezolvate (P1 — Medium, 12 gap-uri)

Referință: `NUVIAX_Stress_Test_Simulation.docx` Secțiunea 9 — Gap-uri Medii.

Valorile de referință din stress test (folosite dacă nu sunt specificate altfel):
- Retroactive window: **48h**
- Reactivation stability days: **7 zile**
- Stagnation threshold: **5 zile consecutive fără progres**
- Chaos Index threshold pentru L2: **0.40**
- ALI Ambition Buffer: **1.0 – 1.15**
- Evolution delta threshold: **≥5%**

---

## 6. Integrare AI — Claude Haiku ✅ Implementat în v10.2

**Model:** `claude-haiku-4-5-20251001`
**Provider:** Anthropic API
**Fișier:** `backend/internal/ai/ai.go` — client HTTP direct (stdlib net/http, fără SDK)
**Cost estimat:** $4-5/lună la 1.000 utilizatori activi

**Implementat:**
1. **Task generation** (`level1_structural.go` → `generateTaskTexts`) — Claude Haiku cu fallback pe template-uri statice
2. **GO analysis** (`handlers.go` → `AnalyzeGO`) — Claude Haiku cu fallback pe rule-based
3. **Graceful degradation:** dacă `ANTHROPIC_API_KEY` lipsește → fallback automat, fără erori

**Variabilă de environment:**
```env
ANTHROPIC_API_KEY=sk-ant-...
```

---

## 7. Integrare Email — Resend.com

**Provider:** Resend.com
**Domeniu trimitere:** `noemail@nuviax.app` (requires DNS setup pe name.com)
**M365:** Rămâne pentru email business al proprietarului, nu pentru tranzacțional

**DNS records necesare pe name.com pentru Resend:**
```
TXT  @         "v=spf1 include:spf.resend.com ~all"
TXT  resend._domainkey   [DKIM key de la Resend dashboard]
CNAME send               [tracking domain de la Resend]
```

**Variabile de environment necesare:**
```env
RESEND_API_KEY=re_...
EMAIL_FROM=noreply@nuviax.app
```

**Email-uri tranzacționale necesare (⏳ E-1 — neimplementat):**
1. Confirmare înregistrare (la Register)
2. Reset parolă (endpoint `POST /auth/forgot-password`)
3. Notificare sprint completat (declanșat din scheduler)
4. Reminder zilnic activități (opțional, bazat pe preferințe user)

**Implementat în v10.3 (E-1 ✅):**
- `backend/internal/email/email.go` — client HTTP Resend API direct (fără SDK)
- Email welcome trimis în `Register` handler (goroutine fire-and-forget)
- Email sprint complet trimis în scheduler `jobCloseExpiredSprints`
- `POST /api/v1/auth/forgot-password` — generează token, trimite email (timing-safe)
- `POST /api/v1/auth/reset-password` — validează token, actualizează parola
- Frontend: `/auth/forgot-password` + `/auth/reset-password` pages
- Migration 009: tabelă `password_reset_tokens`

---

## 8. Deployment — Infrastructură

**VPS:** `83.143.69.103` (SSH user: `sbarbu`, path: `/var/www/wxr-nuviax/`)
**DNS registrar:** name.com
**Docker:** nginx-proxy + acme-companion (shared cu alte proiecte pe același VPS)
**Deploy flow:** GitHub `main` branch push → Actions → DockerHub → SSH → docker compose up

**GitHub Secrets necesare:**
```
SSH_HOST          = 83.143.69.103
SSH_PORT          = 22
SSH_USER          = sbarbu
SSH_KEY           = [RSA private key]
DOCKERHUB_TOKEN   = [DockerHub access token]
POSTGRES_PASSWORD = [generat cu openssl rand -base64 32]
REDIS_PASSWORD    = [generat cu openssl rand -base64 32]
JWT_PRIVATE_KEY   = [RSA 4096-bit, base64 encoded]
JWT_PUBLIC_KEY    = [RSA 4096-bit public, base64 encoded]
ENCRYPTION_KEY    = [openssl rand -hex 32]
ANTHROPIC_API_KEY = [sk-ant-...]  # Nou în v10.1
RESEND_API_KEY    = [re_...]      # Nou în v10.1
```

**Domenii configurate:**
- `nuviax.app` → container `nuviax_app` (port 3000)
- `www.nuviax.app` → același container
- `api.nuviax.app` → container `nuviax_api` (port 8080)
- `nuviaxapp.com` → container `nuviax_landing` (port 3001)

---

## 9. Baza de Date — Referință Rapidă

**28 tabele + 26 views + 1 materialized view + 10 funcții + 12 triggers**

**Tabele core (migration 001):**
- `users` — cu `is_admin` boolean (adăugat migration 007)
- `user_sessions` — token_hash, device_fp, expires_at
- `global_objectives` — status: ACTIVE/PAUSED/COMPLETED/ARCHIVED/WAITING
- `sprints` — cu `expected_pct_frozen` + `frozen_expected_pct` (migration 007)
- `daily_tasks` — MAIN sau PERSONAL
- `checkpoints` — milestone-uri per sprint
- `context_adjustments` — cu `retroactive` bool (migration 007)
- `audit_log` — toate evenimentele de securitate

**Tabele noi (migration 007):**
- `regression_events` — detecție valori sub sprint start
- `ali_snapshots` — ALI current vs projected per zi
- Views admin: `v_admin_platform_stats`, `v_admin_user_list`
- Funcție: `fn_dev_reset_data(admin_id)` — dev only

**Migration 008 (v10.2):**
- `users.avatar_url VARCHAR(500)` — URL avatar profil

**Migration 009 (v10.3):**
- `password_reset_tokens` — token_hash, user_id, expires_at, used_at (1 oră TTL)

**Migration 010 (v10.4):**
- `srm_events` — audit trail complet SRM L1/L2/L3 per obiectiv
- `reactivation_protocols` — tracking 7-day stability per obiectiv PAUSED
- `stagnation_events` — log zile consecutive inactive per GO

**Cum aplici migrările:**
```bash
# Pe server în containerul nuviax_db:
docker exec -i nuviax_db psql -U nuviax -d nuviax < migrations/apply_all.sql
# SAU individual:
docker exec -i nuviax_db psql -U nuviax -d nuviax < migrations/010_p1_gaps.sql
```

---

## 10. Frontend — Design System

**Fișier CSS:** `frontend/app/styles/globals.css` (NU `app/app/globals.css`)

**Variabile CSS definite (toate prezente din v10.2):**
```css
--bg, --bg2, --bg3         # fundal principal / secundar / terțiar
--ink, --ink2, --ink3, --ink4  # text principal → subtil
--line, --line2            # borduri
--l0, --l0l, --l0g         # Level 0 / light / glow (portocaliu)
--l2, --l2l, --l2g         # Level 2 / light / glow (verde)  ✅ adăugat v10.2
--l5, --l5l                # Level 5 / light (violet)
--u, --ul, --ug            # urgency / urgency-light / urgency-glow  ✅ adăugat v10.2
--ff-d, --ff-b, --ff-m, --ff-h  # fonturi display, body, mono, heading  ✅ --ff-h adăugat v10.2
```

**Tema Light:** `[data-theme="light"] { ... }` — ✅ prezentă în v10.2

**Fonturi** (import în `layout.tsx`):
- `Bricolage Grotesque` — display (`--ff-d`, `--ff-h`)
- `DM Sans` — body (`--ff-b`)
- `JetBrains Mono` — mono (`--ff-m`)

---

## 11. Workflow de Dezvoltare

### Pattern pentru sesiuni noi

```bash
# 1. Verifică branch-ul curent
git branch -a
git log --oneline -10

# 2. Creează branch pentru sesiunea curentă
git checkout -b claude/feature-name-XXXXX

# 3. La final: commit + push + PR spre main
git add [fișiere specifice]
git commit -m "feat/fix/docs: descriere clară"
git push -u origin claude/feature-name-XXXXX
```

### Convenții commit

```
feat:  funcționalitate nouă
fix:   corectare bug
docs:  documentație
refactor: restructurare cod fără funcționalitate nouă
chore: configurare, dependențe
```

### Nu comite niciodată

- `.env`, `.env.*` (nu sunt în git — `.gitignore`)
- `.keys/` (chei private)
- `node_modules/`, `vendor/`

---

## 12. Regula README.md — Actualizare Permanentă

> **OBLIGATORIU:** `README.md` trebuie actualizat la FIECARE sesiune care modifică oricare dintre:

| Eveniment | Ce actualizezi în README.md |
|-----------|---------------------------|
| Versiune nouă (bump) | Linia `**Versiune curentă:**`, tabelul Changelog |
| Funcționalitate nouă | Secțiunea API Endpoints și/sau Framework Components |
| Migration nouă | Secțiunea Database (număr migrări, tabele) |
| Job scheduler nou/modificat | Tabelul Scheduler Jobs |
| Endpoint nou/eliminat | Secțiunea API Endpoints |
| Fișier structură modificat | Secțiunea Structura Proiectului |
| Variabilă de environment nouă | Secțiunea Environment Variables |
| Procedură deploy modificată | Secțiunea Deployment |

**Locație fișier:** `README.md` (rădăcina proiectului)

**Verificare rapidă la sfârșitul sesiunii:**
```bash
# Verifică că versiunea din README coincide cu cea din CLAUDE.md
grep "Versiune curentă" README.md
grep "Versiune curentă" CLAUDE.md
```

---

## 14. Roadmap — Priorități

### ✅ Sprinturi completate

| Sprint | Versiune | Conținut principal |
|--------|---------|-------------------|
| Sprint 0 | v10.0.0 | Framework REV 5.6: 40/40 componente |
| Sprint 1a | v10.1.0 | Admin Panel + 5 P0 Gaps critice |
| Sprint 1b | v10.2.0 | 10 Bug Fixes (B-2—B-11) + Claude Haiku AI |
| Sprint 1c | v10.3.0 | Email Resend (E-1) + Forgot/Reset Password |
| Sprint 1d | v10.3.1 | Admin nav fix + cleanup fișiere duplicate |
| Sprint 2 | v10.4.0 | P1 Gaps 10/12 (G-1—G-10, G-12) + migration 010 |

### 🎯 Sprint curent — Sprint 3 (UX + Traduceri + G-11)

1. **G-11** — Behavior Model dominance (EVOLVE override pentru GO hibride) — migration 011
2. **Traduceri EN** — framework i18n, `useTranslation()` hook, toate textele UI
3. **Traduceri RU** — același framework i18n
4. **Onboarding AI** — Claude Haiku sugestii la clasificare GO nouă

### Mai târziu — Sprint 4 (Monetizare)

5. **Stripe integration** — subscripție Pro + Free tier limits + Trial 14 zile
6. **PWA + Notificări push** — service worker + web push opt-in din settings
7. **Export PDF** — raport lunar din `/recap`
8. **Statistici avansate** — heatmap activitate în `/profile`

---

## 15. Resurse și Referințe

| Resursă | Locație |
|---------|---------|
| Stress Test (120 zile, 38 componente) | `NUVIAX_Stress_Test_Simulation.docx` |
| GitHub Secrets Guide | `infra/GITHUB_SECRETS.md` |
| Env Template | `infra/.env.example` |
| Changelog detaliat | `CHANGES.md` |
| Plan de dezvoltare | `ROADMAP.md` |
| Setup cont admin (prima rulare) | `scripts/setup_admin.sh` |

**Fișiere șterse în v10.3.1 (nu mai există în repo):**
- `NuviaX_UI_Mockup_v4.html` — înlocuit de implementare reală
- `ANALYSIS_REPORT.md` — conținut integrat în `CHANGES.md`
- `IMPLEMENTATION_CHECKLIST.md` — înlocuit de `ROADMAP.md`
- `TEST_REPORT.md` — generat automat, nu se ține în git

---

*Ultima actualizare: 2026-03-26 — v10.4.1 (admin page standalone, setup_admin.sh, docs sync complet)*
