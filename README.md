# NuviaX

Platformă SaaS de management al obiectivelor personale/profesionale, construită pe **NuviaX Growth Framework Rev 5.6**.

**Versiune documentație:** `1.0.0` (MVP Reset)  
**Stare produs:** reconstrucție pe faze F0–F7, framework Rev 5.6

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

- Backend: Go + Fiber
- DB: PostgreSQL
- Cache/Sessions: Redis
- Frontend: Next.js + TypeScript
- AI: Anthropic Claude Haiku
- Email: Resend
- CI/CD: GitHub Actions → DockerHub → VPS

---

## Framework alignment program (important)

În prezent se execută programul de aliniere completă Rev 5.6, organizat pe 4 milestones:

1. **M1:** Behavior Model canonic + SRM single-active-level
2. **M2:** `execution_windows` + `SEASONAL_PAUSE`
3. **M3:** Regression pipeline + Temporal validity (A3)
4. **M4:** verification final (docs + tests + compliance matrix)

Detalii:
- `ROADMAP.md`

---

## Structură repo

```
backend/                 # API, engine, scheduler, migrations
frontend/app/            # aplicația principală
frontend/landing/        # landing site
docs/                    # documentație produs/testare/arhitectură
ROADMAP.md               # roadmap livrare
CLAUDE.md                # context master de lucru
```

- Auth: register/login/forgot/reset
- Goals: create/list/detail/progress/visualization
- Today: list/complete/personal
- SRM: status + confirm L2 + confirm L3
- Achievements/Ceremonies
- Settings/Profile activity

> Contractele detaliate sunt în `docs/user-workflow.md`.

---

## Scheduler Jobs (background cron — UTC)

| # | Job | Cron | Componentă |
|---|-----|------|-----------|
| 1 | `jobGenerateDailyTasks` | `00:01` | C23 Daily Stack — AI task generation (Haiku fallback) |
| 2 | `jobComputeDailyScore` | `23:50` | C24 Progress + C25 Drift computation → daily_scores |
| 3 | `jobCheckDailyProgress` | `23:55` | C26 Drift Engine — SRM L1 event if drift critical |
| 4 | `jobCloseExpiredSprints` | `00:00` | C37 Sprint Score — close + grade + ceremony + email |
| 5 | `jobStartNextSprints` | `00:05` | C19 Sprint Structuring — auto-start next sprint |
| 6 | `jobComputeWeeklyALI` | `Sun 03:00` | C38 ALI — placeholder (post-MVP) |
| 7 | `jobRecalibrateRelevance` | `Sun 02:00` | C28 Chaos Index — SRM L2 if chaos ≥ 0.40 |
| 8 | `jobCheckStagnation` | `23:58` | C27 Stagnation — detect 5+ days no activity |
| 9 | `jobCheckSRMTimeouts` | `hourly` | C33 SRM — L3 unconfirmed → ComputeSRMFallback |
| 10 | `jobGenerateCeremonies` | `01:05` | C37 — backfill ceremonies for completed sprints |
| 11 | `jobDetectEvolution` | `01:00` | C31 Behavioral Patterns — placeholder (post-MVP) |
| 12 | `jobComputeGORI` | `01:10` | C38 GORI — update go_metrics per GO |

---

## Testare (Unit + Integration)

### Backend validation quick run

```bash
bash backend/scripts/test_all.sh
```

### API smoke checks

```bash
TOKEN=<jwt> bash backend/scripts/test_api.sh http://localhost:8080/api/v1
```

### Plan complet de testare

- `docs/testing/test-plan.md`
- `docs/testing/flows/*.md`
- `docs/testing/scenarios/*.md`

---


## Acces rapid panel admin

Dacă un utilizator vede "Acces restricționat" pe `/admin`, cauza este aproape întotdeauna `is_admin = FALSE` în DB.

Bootstrap automat cont admin:

```bash
bash scripts/setup_admin.sh <username> '<password>' '<Display Name>'
```

Scriptul:
1. creează contul (dacă nu există),
2. setează `is_admin=TRUE`,
3. verifică login-ul API,
4. afișează credențialele finale.

> Login se face cu **email**, nu cu username. Email-ul devine `<username>@nuviax.app`.

Health check:

## Deployment

Push pe `main` declanșează pipeline CI/CD (build + deploy).

Health check:

```bash
curl https://api.nuviax.app/health
```

Health check:

## Reguli de securitate

- Nu expune metrici/formule interne ale engine-ului în API.
- Nu comite secrete în repository.
- Pentru non-admin, endpoint-urile admin trebuie mascate (404 policy).

