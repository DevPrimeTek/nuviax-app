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
**Versiune curentă:** 10.2.0
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

## 3. Structura Proiectului

```
nuviax-app/
├── backend/
│   ├── cmd/server/main.go           # Entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── server.go            # Toate rutele + middleware
│   │   │   ├── handlers/
│   │   │   │   ├── handlers.go      # Auth, Goals, Tasks, Sprint, Context, Settings
│   │   │   │   ├── admin.go         # Admin panel handlers
│   │   │   │   ├── srm.go           # SRM (Strategic Reset Management)
│   │   │   │   ├── ceremonies.go    # Level 5 ceremonies
│   │   │   │   ├── achievements.go  # Level 5 badges
│   │   │   │   └── visualization.go # Progress charts
│   │   │   └── middleware/
│   │   │       ├── jwt.go           # JWT auth middleware
│   │   │       └── admin.go         # Admin-only middleware (is_admin check)
│   │   ├── engine/
│   │   │   ├── engine.go            # Layer 0 + public API
│   │   │   ├── level1_structural.go # C9-C18: Sprint, checkpoints, tasks
│   │   │   ├── level2_execution.go  # C19-C25: Execution rate, regression events
│   │   │   ├── level3_adaptive.go   # C26-C31: Consistency, energy, context
│   │   │   ├── level4_regulatory.go # C32-C36: Activation rules, SRM
│   │   │   └── level5_growth.go     # C37-C40: Evolution, ceremonies, visualization
│   │   ├── models/models.go         # 50+ structs (User.IsAdmin adăugat în v10.1)
│   │   ├── db/queries.go            # Toate query-urile DB (CreateRetroactivePause adăugat)
│   │   ├── auth/auth.go             # JWT service
│   │   ├── cache/cache.go           # Redis helpers
│   │   └── scheduler/scheduler.go  # 10 background jobs (cron)
│   │   ├── ai/ai.go                 # Claude Haiku HTTP client (adăugat v10.2)
│   ├── migrations/
│   │   ├── 001_base_schema.sql      # Core: users, sessions, goals, sprints, tasks
│   │   ├── 002_layer0_level1.sql
│   │   ├── 003_level2_execution.sql
│   │   ├── 004_level3_adaptive.sql
│   │   ├── 005_level4_regulatory.sql
│   │   ├── 006_level5_growth.sql
│   │   ├── 007_admin_fixes.sql      # Admin panel + 5 critical gap fixes (v10.1)
│   │   └── 008_avatar.sql           # avatar_url pe users (v10.2)
│   └── pkg/
│       ├── crypto/crypto.go         # AES-256-GCM, PBKDF2, SHA256
│       └── logger/logger.go         # Uber Zap structured logging
├── frontend/
│   └── app/
│       ├── app/                     # Next.js App Router
│       │   ├── dashboard/           # ✅ Funcțional
│       │   ├── goals/               # ✅ Fix v10.2: {goals:[], waiting:[]} response
│       │   ├── today/               # ✅ Fix v10.2: energy salvată + personal tasks
│       │   ├── achievements/        # ✅ Funcțional
│       │   ├── recap/               # ✅ Fix v10.2: GET /recap/current implementat
│       │   ├── settings/            # ✅ Fix v10.2: parolă + export date conectate
│       │   ├── profile/             # ✅ Fix v10.2: upload avatar implementat
│       │   ├── admin/               # ✅ Nou în v10.1 (panel administrare)
│       │   ├── auth/                # ✅ Login + Register
│       │   └── onboarding/          # ✅ Funcțional
│       ├── components/
│       │   ├── layout/AppShell.tsx
│       │   ├── DashboardClientLayer.tsx  # Polls ceremonies/unviewed
│       │   ├── SRMWarning.tsx           # L1/L2/L3 warning banners
│       │   ├── CeremonyModal.tsx        # Sprint ceremony modal
│       │   ├── ProgressCharts.tsx       # Recharts LineChart + BarChart
│       │   └── GoalTabs.tsx             # Prezentare/Progres tabs
│       ├── lib/api.ts               # API client utilities
│       └── api/
│           ├── auth/                # Login/register/logout routes
│           └── proxy/[...path]/     # Generic backend proxy cu auto-refresh token
└── infra/
    ├── docker-compose.yml           # Prod: DB + Redis + API + App + Landing
    ├── .env.example                 # Template variabile
    └── GITHUB_SECRETS.md            # SSH_HOST=83.143.69.103, SSH_USER=sbarbu
```

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
| E-2 | P1 Gaps din stress test (12 gap-uri medii) | ⏳ Neimplementat |
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

**Cum aplici migrările:**
```bash
# Pe server în containerul nuviax_db:
docker exec -i nuviax_db psql -U nuviax -d nuviax < migrations/apply_all.sql
# SAU individual:
docker exec -i nuviax_db psql -U nuviax -d nuviax < migrations/009_password_reset.sql
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

## 12. Roadmap — Priorități

### ✅ Sprint anterior — COMPLET (v10.2.0, 2026-03-24)

1. ✅ Fix Bug #7 — Goals API mismatch
2. ✅ Fix Bug #8 — Recap endpoint implementat
3. ✅ Fix Bug #3 — Sprint days calcul corect
4. ✅ Fix Bug #11 — CSS variables + Light theme
5. ✅ Fix Bug #5 — Energy salvată corect
6. ✅ Fix Bug #6 — Personal task add UI
7. ✅ Fix Bug #9 — Settings complet conectate
8. ✅ Fix Bug #2 — Analiză GO cu Claude Haiku
9. ✅ Fix Bug #4 — Task generation cu Claude Haiku
10. ✅ Fix Bug #10 — Upload avatar profil

### 🎯 Sprint curent (imediat)

1. ✅ **Integrare Resend** — email service complet (E-1)
2. **P1 gaps** din stress test — 12 gap-uri medii (E-2)
3. **Translations** — framework EN/RU (E-3)

### Mai târziu

4. **Monetizare** — Stripe integration (E-4)
5. **Mobile** — PWA sau React Native
6. **Analytics** — dashboard utilizator avansat
7. **Onboarding** îmbunătățit cu AI suggestions

---

## 13. Resurse și Referințe

| Resursă | Locație |
|---------|---------|
| Stress Test (120 zile, 38 componente) | `NUVIAX_Stress_Test_Simulation.docx` |
| Mockup UI v4 | `NuviaX_UI_Mockup_v4.html` |
| API Documentation | `backend/API.md` |
| Bug Analysis | `ANALYSIS_REPORT.md` |
| Implementation Status | `IMPLEMENTATION_CHECKLIST.md` |
| GitHub Secrets Guide | `infra/GITHUB_SECRETS.md` |
| Env Template | `infra/.env.example` |

---

*Ultima actualizare: 2026-03-25 — v10.3 (Resend email integration + forgot/reset password flow)*
