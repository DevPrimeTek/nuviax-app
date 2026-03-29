# NuviaX

Platformă SaaS de management al obiectivelor personale și profesionale, bazată pe **NUViaX Growth Framework REV 5.6** — sistem proprietar cu 40 de componente matematice (C1–C40) distribuite pe 5 niveluri.

**Versiune curentă:** `10.4.1` | **Status:** Production Ready

---

## Linkuri

| | URL |
|--|-----|
| Aplicație | https://nuviax.app |
| Landing | https://nuviaxapp.com |
| API | https://api.nuviax.app |
| Repository | https://github.com/DevPrimeTek/nuviax-app |

---

## Stack Tehnic

| Layer | Tehnologie |
|-------|-----------|
| Backend | Go 1.22 + Fiber v2.52 |
| Database | PostgreSQL 16 (Docker) |
| Cache | Redis 7 (Docker) |
| Frontend | Next.js 14 + TypeScript + Tailwind |
| Auth | JWT RS256 (RSA 4096-bit) |
| AI | Claude Haiku 4.5 (`claude-haiku-4-5-20251001`) |
| Email | Resend.com |
| CI/CD | GitHub Actions → DockerHub → VPS SSH |
| Proxy | nginx-proxy + acme-companion |

---

## NUViaX Framework REV 5.6 — 40 Componente

| Layer | Componente | Fișier engine |
|-------|-----------|--------------|
| Layer 0 | C1–C8: Drift, Chaos Index, ALI, Priority, Behavior | `engine.go` |
| Level 1 | C9–C18: Sprint Architecture, Task Generation, Relevance | `level1_structural.go` |
| Level 2 | C19–C25: Execution Rate, Sprint Score (40/25/25/10), Stagnation | `level2_execution.go` |
| Level 3 | C26–C31: Consistency, Energy, Context, Focus Rotation | `level3_adaptive.go` |
| Level 4 | C32–C36: Activation Rules, Future Vault, SRM L1/L2/L3 | `level4_regulatory.go` |
| Level 5 | C37–C40: Evolution Sprints, Ceremonies, Achievements, Visualization | `level5_growth.go` |

**Principiu:** Toate calculele rulează exclusiv server-side. Clientul primește doar rezultate opace (%, grade, liste).

---

## Structura Repo

```
nuviax-app/
├── .github/workflows/     # CI/CD: deploy.yml + deploy-frontend.yml
├── backend/
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── ai/            # Claude Haiku client
│   │   ├── api/           # Fiber server + handlers + middleware
│   │   ├── engine/        # Framework REV 5.6 (Layer 0 + Level 1-5)
│   │   ├── db/            # PostgreSQL queries
│   │   ├── email/         # Resend.com client
│   │   └── scheduler/     # 12 cron jobs
│   └── migrations/        # 010 migrări aplicate
├── frontend/
│   ├── app/               # Next.js → nuviax.app
│   └── landing/           # Next.js static → nuviaxapp.com
├── infra/                 # Docker Compose + scripts server
├── docs/                  # Documentație extinsă
│   └── archive/           # Specificații originale framework
├── CLAUDE.md              # Context master Claude Code
├── ROADMAP.md             # Plan dezvoltare
└── README.md              # Acest fișier
```

---

## Database

**10 migrări → 28 tabele, 26+ views, 1 materialized view, 10 funcții, 12 triggers**

```bash
# Aplică toate migrările (idempotent)
docker exec -i nuviax_db psql -U nuviax -d nuviax < backend/migrations/apply_all.sql
```

---

## API Endpoints principale

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password

GET    /api/v1/dashboard
GET    /api/v1/goals                    # {goals:[], waiting:[]}
POST   /api/v1/goals
GET    /api/v1/goals/:id
GET    /api/v1/goals/:id/progress
GET    /api/v1/goals/:id/visualize      # Level 5 charts

GET    /api/v1/today
POST   /api/v1/today/complete/:id
POST   /api/v1/today/personal

GET    /api/v1/ceremonies/unviewed
POST   /api/v1/ceremonies/:id/view
GET    /api/v1/achievements

GET    /api/v1/srm/status/:goalId
POST   /api/v1/srm/confirm-l3/:goalId
```

---

## Scheduler Jobs (12)

| Cron | Job |
|------|-----|
| `00:00` | Generare sarcini zilnice |
| `23:50` | Calcul scor zilnic |
| `23:55` | Verificare progres |
| `00:01` | Închidere sprinturi expirate + trimitere email Sprint Complet |
| `01:00` | Detecție Evolution Sprints |
| `01:05` | Generare Ceremonies (BRONZE→PLATINUM) |
| `02:00` | Verificare timeout SRM + refresh matview |
| `00:05` | Propunere reactivare obiective PAUSED |
| `23:58` | Detecție stagnare (≥5 zile inactive) |
| `02:00/90d` | Recalibrare relevanță anuală |

---

## Deployment

```bash
# Automat: push pe main → GitHub Actions → DockerHub → VPS
git push origin main

# Health check
curl https://api.nuviax.app/health
# {"status":"ok","db":true,"redis":true}

# Manual (pe server)
bash infra/deploy.sh
```

**GitHub Secrets necesare:** `SSH_HOST`, `SSH_PORT`, `SSH_USER`, `SSH_KEY`, `DOCKERHUB_TOKEN`, `POSTGRES_PASSWORD`, `REDIS_PASSWORD`, `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`, `ENCRYPTION_KEY`, `ANTHROPIC_API_KEY`, `RESEND_API_KEY`

Detalii complete: [`docs/deployment.md`](docs/deployment.md)

---

## Environment Variables

```env
# Database
POSTGRES_HOST=nuviax_db
POSTGRES_PASSWORD=<openssl rand -base64 32>
POSTGRES_DB=nuviax

# Redis
REDIS_HOST=nuviax_redis
REDIS_PASSWORD=<openssl rand -base64 32>

# Auth
JWT_PRIVATE_KEY=<RSA 4096-bit base64>
JWT_PUBLIC_KEY=<RSA 4096-bit public base64>
ENCRYPTION_KEY=<openssl rand -hex 32>

# Integrări (opționale — graceful degradation dacă lipsesc)
ANTHROPIC_API_KEY=sk-ant-...
RESEND_API_KEY=re_...
EMAIL_FROM=noreply@nuviax.app
```

---

## Securitate

- **Engine opac** — formulele nu ies niciodată din `internal/engine/`
- **JWT RS256** — access 15min, refresh 7 zile
- **Email criptat** — AES-256-GCM + SHA-256 hash pentru lookup
- **Admin 404** — panel returnează 404 (nu 403) pentru non-admini
- **Timing-safe** — `forgot-password` returnează mereu 200

---

## Documentație

| Document | Conținut |
|----------|---------|
| [`CLAUDE.md`](CLAUDE.md) | Context master pentru sesiuni Claude Code |
| [`ROADMAP.md`](ROADMAP.md) | Plan și priorități dezvoltare |
| [`CHANGES.md`](CHANGES.md) | Changelog detaliat |
| [`docs/project-structure.md`](docs/project-structure.md) | Structura completă repo |
| [`docs/database-reference.md`](docs/database-reference.md) | Schema DB, migrări, triggers |
| [`docs/integrations.md`](docs/integrations.md) | AI + Email implementare |
| [`docs/deployment.md`](docs/deployment.md) | VPS, Docker, CI/CD |
| [`docs/history-bugs-gaps.md`](docs/history-bugs-gaps.md) | Bug-uri rezolvate, gap-uri |
| [`CLIENT_TODO.md`](CLIENT_TODO.md) | Acțiuni necesare din partea proprietarului |

---

## Changelog

| Versiune | Data | Descriere |
|---------|------|-----------|
| v10.4.1 | 2026-03-29 | Restructurare docs/, CLAUDE.md optimizat, CLIENT_TODO |
| v10.4.0 | 2026-03-26 | P1 Gaps G-1—G-10, G-12 (10/12); migration 010 |
| v10.3.0 | 2026-03-25 | Email Resend: welcome + sprint + forgot/reset parolă |
| v10.2.0 | 2026-03-24 | Bug fixes B-2—B-11; AI Claude Haiku; upload avatar |
| v10.1.0 | 2026-03-20 | Admin Panel; P0 Gaps critice |
| v10.0.0 | 2026-03-16 | Framework REV 5.6 40/40; CI/CD complet |

> Changelog detaliat: [`CHANGES.md`](CHANGES.md)
