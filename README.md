# NuviaX

Platformă SaaS de management al obiectivelor personale/profesionale, construită pe **NuviaX Growth Framework Rev 5.6**.

**Versiune documentație:** `11.0.0` (Architect Sync)  
**Stare produs:** în aliniere completă Framework (program M1–M4 activ)

---

## Linkuri

| Componentă | URL |
|---|---|
| App | https://nuviax.app |
| Landing | https://nuviaxapp.com |
| API | https://api.nuviax.app |
| Repo | https://github.com/DevPrimeTek/nuviax-app |

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
- `PLAN.md`
- `ROADMAP.md`
- `docs/framework_100_percent_implementation_playbook.md`
- `docs/framework_workflow_deviations_stress_test.md`

---

## Structură repo

```
backend/                 # API, engine, scheduler, migrations
frontend/app/            # aplicația principală
frontend/landing/        # landing site
docs/                    # documentație produs/testare/arhitectură
PLAN.md                  # plan implementare
ROADMAP.md               # roadmap livrare
CLAUDE.md                # context master de lucru
```

---

## API (high-level)

- Auth: register/login/forgot/reset
- Goals: create/list/detail/progress/visualization
- Today: list/complete/personal
- SRM: status + confirm L2 + confirm L3
- Achievements/Ceremonies
- Settings/Profile activity

> Contractele detaliate sunt în `docs/user-workflow.md`.

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
bash scripts/setup_admin.sh sbarbu_admin 'NuviaXAdmin#2026' 'Sbarbu Admin'
```

Scriptul:
1. creează contul (dacă nu există),
2. setează `is_admin=TRUE`,
3. verifică login-ul API,
4. afișează credențialele finale.

> Login se face cu **email**, nu cu username. Pentru comanda de mai sus, email-ul devine `sbarbu_admin@nuviax.app`.

---

## Deployment

Push pe `main` declanșează pipeline CI/CD (build + deploy).

Health check:

```bash
curl https://api.nuviax.app/health
```

---

## Reguli de securitate

- Nu expune metrici/formule interne ale engine-ului în API.
- Nu comite secrete în repository.
- Pentru non-admin, endpoint-urile admin trebuie mascate (404 policy).

