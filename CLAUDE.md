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
**Versiune curentă:** 10.0.0
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
│   ├── migrations/
│   │   ├── 001_base_schema.sql      # Core: users, sessions, goals, sprints, tasks
│   │   ├── 002_layer0_level1.sql
│   │   ├── 003_level2_execution.sql
│   │   ├── 004_level3_adaptive.sql
│   │   ├── 005_level4_regulatory.sql
│   │   ├── 006_level5_growth.sql
│   │   └── 007_admin_fixes.sql      # Admin panel + 5 critical gap fixes (v10.1)
│   └── pkg/
│       ├── crypto/crypto.go         # AES-256-GCM, PBKDF2, SHA256
│       └── logger/logger.go         # Uber Zap structured logging
├── frontend/
│   └── app/
│       ├── app/                     # Next.js App Router
│       │   ├── dashboard/           # ✅ Funcțional
│       │   ├── goals/               # ⚠️ Bug #7: array vs obiect mismatch
│       │   ├── today/               # ⚠️ Bug #5: energy nu se salvează, Bug #6: fără add task
│       │   ├── achievements/        # ✅ Funcțional
│       │   ├── recap/               # ❌ Bug #8: endpoint /recap/current lipsă
│       │   ├── settings/            # ⚠️ Bug #9: parțial conectat
│       │   ├── profile/             # ⚠️ Bug #10: fără upload foto
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

## 4. Starea Curentă — Bug-uri Cunoscute

### 🔴 Critice (Blockers)

| # | Locație | Problemă | Fix necesar |
|---|---------|---------|------------|
| B-7 | `handlers.go:343` + `lib/api.ts:92` | Goals endpoint returnează array plat, frontend așteaptă `{goals:[], waiting:[]}` | Schimbă response handler sau schimbă api.ts |
| B-8 | `server.go` | `GET /recap/current` nu există — 404 | Adaugă endpoint `/recap/current` care returnează ultimul sprint + reflecție |
| B-3 | `handlers.go:276` | `days_left` calculat din `goal.EndDate` (90 zile) în loc de `sprint.EndDate` (30 zile) | `daysLeft := int(time.Until(sprint.EndDate).Hours() / 24)` |

### 🟠 Majore

| # | Locație | Problemă | Fix necesar |
|---|---------|---------|------------|
| B-5 | `today/page.tsx:139` | Click "Cum mă simt" nu apelează `POST /context/energy` | Conectează button la API call |
| B-6 | `today/page.tsx` | Lipsă formular/buton pentru sarcini personale | Adaugă UI pentru `POST /today/personal` |
| B-11 | `globals.css` | `--ul`, `--l2g`, `--ff-h` nedefinite; tema Light lipsă | Adaugă variabile CSS + `[data-theme="light"]` |
| B-9 | `settings/page.tsx` | Notificări/parolă/export date neconectate la API | Conectează toate acțiunile la endpoint-uri existente |
| B-4 | `level1_structural.go:72` | Sarcini zilnice din template static, ignoră contextul GO | Integrare Claude Haiku (Faza 2) |

### 🟡 Medii

| # | Locație | Problemă | Fix necesar |
|---|---------|---------|------------|
| B-2 | `handlers.go AnalyzeGO` | Analiză GO fără sugestii AI | Integrare Claude Haiku |
| B-10 | `profile/page.tsx` | Fără upload foto profil | UI + endpoint backend |

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

## 6. Integrare AI — Claude Haiku

**Model:** `claude-haiku-4-5-20251001`
**Provider:** Anthropic API
**Cost estimat:** $4-5/lună la 1.000 utilizatori activi

**Unde se folosește:**
1. **Task generation** (`level1_structural.go` → `generateTaskTexts`) — Faza 2 neimplementată
2. **GO analysis** (`handlers.go` → `AnalyzeGO`) — înlocuiește regex-ul static
3. **Semantic parsing** (clasificare BM: INCREASE/REDUCE/CREATE/EVOLVE)

**Variabilă de environment necesară:**
```env
ANTHROPIC_API_KEY=sk-ant-...
```

**Pattern de apel (Go):**
```go
// POST https://api.anthropic.com/v1/messages
// Model: claude-haiku-4-5-20251001
// Max tokens: 256 (pentru task generation)
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

**Email-uri tranzacționale necesare:**
1. Confirmare înregistrare
2. Reset parolă
3. Notificare sprint completat
4. Reminder zilnic activități (opțional, bazat pe preferințe)

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

**Cum aplici migrările:**
```bash
# Pe server în containerul nuviax_db:
docker exec -i nuviax_db psql -U nuviax -d nuviax < migrations/007_admin_fixes.sql
```

---

## 10. Frontend — Design System

**Variabile CSS definite în `frontend/app/app/globals.css`:**
```css
--bg, --bg2, --bg3         # fundal principal / secundar / terțiar
--ink, --ink2, --ink3, --ink4  # text principal → subtil
--line, --line2            # borduri
--l0, --l0l, --l0g         # Level 0 / light / glow (portocaliu)
--l2, --l2l, --l2g         # Level 2 / light / glow (verde)
--l5, --l5l                # Level 5 / light (violet)
--ul, --ug                 # urgency / urgency-glow (galben/portocaliu)  ← LIPSĂ, trebuie adăugat
--ff-d, --ff-b, --ff-m     # font display, body, mono
```

**⚠️ Variabile CSS lipsă (Bug #11):** `--ul`, `--ug`, `--l2g`, `--ff-h` — trebuie adăugate în `globals.css`

**Tema Light:** `[data-theme="light"] { ... }` — lipsă complet, trebuie adăugată

**Fonturi necesare** (import în `layout.tsx`):
- `Bricolage Grotesque` — display (`--ff-d`)
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

### Sprint curent (imediat)

1. **Fix Bug #7** — Goals API mismatch (blocker)
2. **Fix Bug #8** — Recap endpoint lipsă (blocker)
3. **Fix Bug #3** — Sprint days wrong calculation
4. **Fix Bug #11** — CSS variables lipsă + Light theme
5. **Fix Bug #5** — Energy level nu se salvează
6. **Fix Bug #6** — Personal task add
7. **Fix Bug #9** — Settings complet conectate
8. **Integrare Resend** — email service de bază
9. **Integrare Claude Haiku** — task generation + GO analysis

### Următor sprint

10. **P1 gaps** din stress test (12 gap-uri medii)
11. **Upload foto profil** (Bug #10)
12. **Light theme** CSS complet
13. **Translations** EN + RU (framework există, conținut lipsă)
14. **Onboarding** îmbunătățit cu AI suggestions

### Mai târziu

15. **Monetizare** — Stripe integration
16. **Mobile** — PWA sau React Native
17. **Analytics** — dashboard utilizator avansat

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

*Ultima actualizare: 2026-03-24 — v10.1 (admin panel + stress test gap fixes)*
