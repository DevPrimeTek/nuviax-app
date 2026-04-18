# NuviaX

Platformă SaaS de management al obiectivelor personale/profesionale, construită pe **NuviaX Growth Framework Rev 5.6**.

**Versiune documentație:** `1.1.0`  
**Stare produs:** MVP complet — F0–F7 verificate ✅

---

## Linkuri

| Componentă | URL |
|---|---|
| App | https://nuviax.app |
| Landing | https://nuviaxapp.com |
| API | https://api.nuviax.app |
| Repo | https://github.com/DevPrimeTek/nuviax-app |

---

## Rule #1 pentru echipă

Înainte de orice task, citește `CLAUDE.md` și urmează protocolul de acolo (analiză minimă, fișiere țintite, fără scan global inutil).

---

## Stack

- Backend: Go + Fiber v2
- DB: PostgreSQL
- Cache/Sessions: Redis
- Frontend: Next.js 14 + TypeScript
- AI: Anthropic Claude Haiku 4.5
- Email: Resend
- CI/CD: GitHub Actions → DockerHub → VPS

---

## Structură repo

```
backend/                 # API, engine, scheduler, migrations
frontend/app/            # aplicația principală
frontend/landing/        # landing site
docs/                    # documentație produs/testare/arhitectură
ROADMAP.md               # roadmap livrare (stare actuală)
CLAUDE.md                # context master de lucru
PROMPTS_MVP.md           # prompturi sesiuni Claude Code
```

---

## Endpoints API (v1)

### Auth (implementate ✅)
| Method | Endpoint | Descriere |
|---|---|---|
| POST | `/api/v1/auth/register` | Înregistrare utilizator |
| POST | `/api/v1/auth/login` | Autentificare |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/mfa/verify` | Verificare cod MFA |
| POST | `/api/v1/auth/mfa/enable` | Activare MFA |
| POST | `/api/v1/auth/forgot-password` | Inițiere reset parolă |
| POST | `/api/v1/auth/reset-password` | Resetare parolă cu token |
| POST | `/api/v1/auth/logout` | Deconectare |

### Goals / Today / Dashboard (implementate ✅ — F5a)
| Method | Endpoint | Descriere |
|---|---|---|
| POST | `/api/v1/goals/analyze` | AI validare text GO (C9/C10) |
| POST | `/api/v1/goals/suggest-category` | AI sugestie categorie |
| POST | `/api/v1/goals` | Creare GO (C3, C4, C12) |
| GET | `/api/v1/goals` | Listă GO utilizator |
| GET | `/api/v1/goals/:id` | Detaliu GO |
| GET | `/api/v1/goals/:id/visualize` | Date growth trajectories |
| GET | `/api/v1/today` | Taskuri zilnice |
| POST | `/api/v1/today/complete/:id` | Marchează task completat |
| POST | `/api/v1/today/personal` | Adaugă task personal (max 2/zi) |
| POST | `/api/v1/context/energy` | Setare nivel energie |
| GET | `/api/v1/dashboard` | Overview utilizator (Redis cache 5 min) |

### Business Logic (implementate ✅ — F5b)
| Method | Endpoint | Descriere |
|---|---|---|
| GET | `/api/v1/srm/status/:goalId` | Status SRM ✅ |
| POST | `/api/v1/srm/confirm-l2/:goalId` | Confirmare SRM L2 ✅ |
| POST | `/api/v1/srm/confirm-l3/:goalId` | Pauză GO (SRM L3) ✅ |
| GET | `/api/v1/achievements` | Lista achievements ✅ |
| GET | `/api/v1/ceremonies/:goalId` | Ultima ceremonie ✅ |
| POST | `/api/v1/ceremonies/:id/view` | Marchează ceremonie vizualizată ✅ |
| GET | `/api/v1/profile/activity` | Activitate 365 zile ✅ |
| PATCH | `/api/v1/settings` | Theme, locale ✅ |
| GET | `/api/v1/admin/stats` | Statistici admin (404 non-admin) ✅ |
| GET | `/api/v1/admin/users` | Lista utilizatori (404 non-admin) ✅ |
| POST | `/api/v1/admin/users/:id/deactivate` | Dezactivare cont (404 non-admin) ✅ |

---

## Variabile de mediu

| Variabilă | Obligatoriu | Descriere |
|---|---|---|
| `DATABASE_URL` | ✅ | PostgreSQL connection string |
| `REDIS_URL` | ✅ | Redis connection string |
| `JWT_PRIVATE_KEY` | ✅ | JWT signing key (base64 encoded) |
| `JWT_PUBLIC_KEY` | ✅ | JWT verify key (base64 encoded) |
| `ENCRYPTION_KEY` | ✅ | 32 bytes sau 64-char hex |
| `ALLOWED_ORIGINS` | ✅ | CORS whitelist (ex: `https://nuviax.app`) |
| `ANTHROPIC_API_KEY` | ❌ opțional | Claude Haiku — fallback rule-based dacă lipsește |
| `RESEND_API_KEY` | ❌ opțional | Email service — fallback log dacă lipsește |
| `EMAIL_FROM` | ❌ opțional | Sender (ex: `NuviaX <noreply@nuviax.app>`) |

---

## Scheduler Jobs (background cron — UTC)

| # | Job | Cron | Componentă |
|---|-----|------|-----------|
| 1 | `jobGenerateDailyTasks` | `00:01` | C23 Daily Stack — AI task generation (Haiku fallback) |
| 2 | `jobComputeDailyScore` | `23:50` | C24 Progress + C25 Drift computation |
| 3 | `jobCheckDailyProgress` | `23:55` | C26 Drift Engine — SRM L1 event dacă drift critic |
| 4 | `jobCloseExpiredSprints` | `00:00` | C37 Sprint Score — close + grade + ceremony + email |
| 5 | `jobStartNextSprints` | `00:05` | C19 Sprint Structuring — auto-start sprint următor |
| 6 | `jobComputeWeeklyALI` | `Sun 03:00` | C38 ALI — placeholder (post-MVP) |
| 7 | `jobRecalibrateRelevance` | `Sun 02:00` | C28 Chaos Index — SRM L2 dacă chaos ≥ 0.40 |
| 8 | `jobCheckStagnation` | `23:58` | C27 Stagnation — detectare 5+ zile fără activitate |
| 9 | `jobCheckSRMTimeouts` | `orar` | C33 SRM — L3 neconfirmate → ComputeSRMFallback |
| 10 | `jobGenerateCeremonies` | `01:05` | C37 — backfill ceremonies pentru sprints completate |
| 11 | `jobDetectEvolution` | `01:00` | C31 Behavioral Patterns — placeholder (post-MVP) |
| 12 | `jobComputeGORI` | `01:10` | C38 GORI — update go_metrics per GO |

---

## Testare

```bash
# Unit tests engine
cd backend && go test ./internal/engine/... -v

# Build check
cd backend && go build ./...

# API smoke check (după F5 complet)
TOKEN=<jwt> bash backend/scripts/test_api.sh http://localhost:8080/api/v1
```

Plan complet: `docs/testing/test-plan.md`

---

## Acces panel admin

```bash
bash scripts/setup_admin.sh <username> '<password>' '<Display Name>'
```

Email de login devine `<username>@nuviax.app`. Utilizatorii non-admin primesc 404 pe rutele admin (security by obscurity).

---

## Deployment

Push pe `main` declanșează pipeline CI/CD (build + deploy).

```bash
curl https://api.nuviax.app/health
```

---

## Reguli de securitate

- Nu expune metrici/formule interne ale engine-ului în API (drift, chaos, weights, thresholds).
- Nu comite secrete în repository.
- Endpoint-urile admin returnează 404 (nu 403) pentru non-admin.
