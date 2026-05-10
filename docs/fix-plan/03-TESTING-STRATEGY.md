# TESTING-STRATEGY — NuviaX MVP Fix Phase F8

> **Versiune:** 1.0.0
> **Data:** 2026-05-10
> **Owner:** Senior QA Tester
> **Aprobat de:** Architect, PM
> **Scop:** Fiecare gate F8.x este blocant. Fără pass green pe gate, faza următoare nu pornește.

---

## 1. Niveluri de testare

### Nivel 1 — Unit Tests (Go)
- **Locație:** `backend/internal/**/*_test.go`
- **Tool:** `go test`
- **Coverage minim:**
  - `engine/`: ≥ 80%
  - `api/handlers/`: ≥ 70%
  - `db/`: ≥ 50% (mocking pgxpool)
- **Rulare:** la fiecare commit local (pre-commit hook recomandat după F8.7)
- **CI gate:** F8.7

### Nivel 2 — Integration Tests
- **Locație:** `backend/internal/api/handlers/*_integration_test.go` (build tag `integration`)
- **Setup:** PostgreSQL în Docker (testcontainers-go), Redis fake, AI mock
- **Coverage:** end-to-end pe rutele critice (auth, goals, today, srm, achievements)
- **Rulare:** `go test -tags=integration ./...`
- **CI gate:** F8.7

### Nivel 3 — Opacity Test (custom)
- **Locație:** `backend/internal/api/handlers/opacity_test.go` (nou în F8.3)
- **Cum funcționează:**
  1. Mock JWT, mock DB cu fixtures
  2. Pentru fiecare endpoint, face request, parsează JSON, walk recursiv toate cheile
  3. Asertează: zero matches pe `{drift, chaos_index, weights, score_components, sprint_score, factors, penalties, thresholds, raw_score}`
- **Failure:** orice match → test fail cu indicarea exactă a path-ului JSON
- **CI gate:** F8.3 + F8.7

### Nivel 4 — Schema Check
- **Locație:** `backend/scripts/schema-check.sh` (nou în F8.1)
- **Cum funcționează:**
  1. Spawn PostgreSQL curat
  2. Aplică `001_schema.sql` + `002_runtime_baseline.sql`
  3. Aplică toate ALTER TABLE din `db.go::ensureGoalsTables`
  4. Verifică că toate tabelele referite de cod (extracție via `grep`) există
  5. Verifică că toate coloanele citite de queries există
- **CI gate:** F8.1 + F8.7

### Nivel 5 — Frontend Build & Type Check
- **Tool:** `npm run build` în `frontend/app/`
- **Coverage:** TypeScript strict + ESLint
- **CI gate:** F8.6 + F8.7

### Nivel 6 — E2E (Playwright recomandat)
- **Locație:** `frontend/app/__tests__/e2e/*.spec.ts` (nou în F8.7)
- **Coverage:** TS-01 (full happy path), TS-09 (personal task limit), TS-10 (theme persist)
- **Rulare:** `npx playwright test`
- **CI gate:** F8.7 (opt-in pe staging F8.8)

### Nivel 7 — Smoke Manual (staging)
- **Locație:** `docs/testing/F8.8-smoke-checklist.md` (nou în F8.8)
- **Owner:** QA + PM
- **Coverage:** TS-01..TS-12 manual + edge cases
- **CI gate:** F8.8 sign-off

### Nivel 8 — Performance Baseline
- **Tool:** `vegeta` sau `k6`
- **Targets:**
  - P50 < 200ms pentru `GET /today`, `GET /goals/:id`, `GET /dashboard`
  - P95 < 500ms pentru toate endpoint-urile autentificate
  - P99 < 1000ms
- **Rulare:** F8.8, repetabil
- **CI gate:** F8.8

### Nivel 9 — Security Scan
- **Tooling:**
  - `gosec` (Go security linter)
  - `npm audit` (frontend deps)
  - Manual: bcrypt cost, JWT alg=RS256, admin 404, forgot-password timing
- **CI gate:** F8.8

---

## 2. Mapare TS-01..TS-12 la nivele

| Test | Niveluri verificate | Faza |
|------|---------------------|------|
| TS-01 (Happy Path) | Unit + Integration + E2E + Smoke | F8.7 + F8.8 |
| TS-02 (G-10 Vault) | Integration + Smoke | F8.7 |
| TS-03 (Trajectory) | Integration + Smoke | F8.7 |
| TS-04 (SRM L1 auto) | Integration (with time mocking) + Smoke | F8.7 |
| TS-05 (SRM L2 reduce) | Integration + Smoke | F8.7 |
| TS-06 (SRM L3 freeze) | Integration + Smoke | F8.7 |
| TS-07 (Achievement) | Integration + Smoke | F8.7 |
| TS-08 (Day 1 viz) | Unit + Integration + Smoke | F8.7 |
| TS-09 (Personal task limit) | Unit + Integration + E2E | F8.7 |
| TS-10 (Theme persist) | Integration + E2E | F8.7 |
| TS-11 (AI timeout) | Integration (with AI mock + chaos) | F8.7 |
| TS-12 (API opacity) | Opacity Test (Nivel 3) | F8.3 + F8.7 |

---

## 3. Test gates per fază

### F8.1 Gate (Schema)
**Pass criteria:**
- [ ] `bash backend/scripts/schema-check.sh` → exit 0
- [ ] `psql $TEST_DB < 001_schema.sql && psql $TEST_DB < 002_*.sql` → fără erori
- [ ] Backend pornește cu `db.RunMigrations()` fără warnings critice
- [ ] BACKLOG: DEV-02, DEV-03, DEV-04, DEV-17 → `RESOLVED`

**Verifier:** DBA + QA confirmă; PM mută status; Architect aprobă.

### F8.2 Gate (Engine)
**Pass criteria:**
- [ ] `go test ./internal/engine/... -v -cover` → toate trec
- [ ] Coverage report: ≥ 80% pe engine package
- [ ] `go vet ./internal/engine/...` → zero warnings
- [ ] Toate funcțiile listate în `04-PROMPTS-FIX-SESSIONS.md §F8.2 Output` există în cod
- [ ] BACKLOG: DEV-11 → `RESOLVED`

**Verifier:** Backend Senior + QA; Architect aprobă.

### F8.3 Gate (API Security)
**Pass criteria:**
- [ ] `go test -run TestOpacity ./internal/api/handlers/...` → zero failures
- [ ] Manual scan: `grep -rn "sprint_score\|drift\|chaos_index\|weights" backend/internal/api/handlers/*.go | grep -v "_test.go"` → zero matches în query results sau JSON marshaling
- [ ] BACKLOG: DEV-01 → `RESOLVED`

**Verifier:** Security Engineer + QA; Architect aprobă.

### F8.4 Gate (Scheduler)
**Pass criteria:**
- [ ] Trigger manual fiecare job (creare scenarii fixture); toate rulează clean
- [ ] DB state după run-uri: `growth_trajectories`, `achievement_badges`, `srm_events`, `evolution_sprints`, `go_metrics` populate corect
- [ ] Integration test: `TestSchedulerEndToEnd` simulează 30 zile, verifică toate inserturile
- [ ] BACKLOG: DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19 → `RESOLVED`

**Verifier:** Backend Senior + QA; Architect + PM aprobă.

### F8.5 Gate (Handlers)
**Pass criteria:**
- [ ] Unit tests: minim 1 test per handler nou/modificat
- [ ] Integration tests: TS-04, TS-05, TS-06, TS-08, TS-11 toate pass
- [ ] Manual API test: `POST /context/energy {level: low}` → row inserat verificabil
- [ ] Performance check rapid: P95 endpoints < 500ms cu fixture realist
- [ ] BACKLOG: DEV-06, DEV-12, DEV-13, DEV-14, DEV-15, DEV-22, DEV-23, DEV-25, DEV-26, DEV-27, DEV-28 → `RESOLVED`

**Verifier:** QA lead; Architect + PM aprobă.

### F8.6 Gate (Frontend)
**Pass criteria:**
- [ ] `npm run build` → 0 erori TS, 0 warnings ESLint critice
- [ ] Manual flow onboarding: 3 GO-uri create cu durate diferite (30/90/365)
- [ ] Lighthouse audit: Accessibility ≥ 85, Performance ≥ 70 (mobile)
- [ ] Cross-browser smoke: Chrome + Firefox + Safari (manual)
- [ ] BACKLOG: DEV-20 (parte FE), DEV-21, DEV-26 (parte FE), DEV-27 (parte FE) → `RESOLVED`

**Verifier:** Frontend Senior + UX + QA; PM aprobă.

### F8.7 Gate (Test Automation)
**Pass criteria:**
- [ ] CI pipeline `.github/workflows/ci.yml` (sau echivalent) rulează: unit + integration + opacity + schema-check + frontend build
- [ ] Toate joburile CI green pe branch curent
- [ ] Coverage report `docs/testing/F8.7-coverage-report.md` cu numere reale
- [ ] Toate TS-01..TS-12 reproduse automat (script-uri sau test cases)
- [ ] Test gap report: `docs/testing/F8.7-test-gaps.md` listează edge cases ne-acoperite cu plan POST-MVP

**Verifier:** QA lead; DevOps; Architect + PM aprobă.

### F8.8 Gate (Production Validation) — FINAL
**Pass criteria:**
- [ ] Deploy staging reușit (zero erori în logs)
- [ ] Toate gate-urile precedente F8.1–F8.7 green
- [ ] Smoke manual full TS-01..TS-12 pe staging cu user real → 12/12 pass
- [ ] Performance baseline atins (P50/P95/P99 documentate)
- [ ] Security scan: `gosec` + `npm audit` → zero HIGH/CRITICAL
- [ ] Forgot-password timing test: răspunsul ≤ 5% diferență între email cunoscut și necunoscut
- [ ] Admin 404 test: regular user → exact 404 status code
- [ ] Backup & restore test reușit
- [ ] Runbook staging și production complet
- [ ] PM + Architect + Security: sign-off în writing (commit cu mesaj „F8.8 sign-off: <name> <role>")
- [ ] BACKLOG: zero items în `OPEN`/`IN_PROGRESS`/`IN_REVIEW`. Tot ce nu e RESOLVED este ACCEPTED_POST_MVP cu justificare.

**Verifier:** PM (collator); QA + Architect + Security + DevOps fiecare semnează independent.

---

## 4. Tooling necesar (instalat înainte F8.7)

| Tool | Scop | Instalare |
|------|------|-----------|
| `gotestsum` | Output curat go test | `go install gotest.tools/gotestsum@latest` |
| `gosec` | Security scan Go | `curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh \| sh -s -- -b $(go env GOPATH)/bin v2.18.0` |
| `golangci-lint` | Linter Go | `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \| sh` |
| `testcontainers-go` | Postgres în Docker pentru integration tests | `go get github.com/testcontainers/testcontainers-go/modules/postgres` |
| `playwright` | E2E browser tests | `npx playwright install` |
| `k6` sau `vegeta` | Load testing | `brew install k6` / docker image |

---

## 5. Test Data Strategy

### Fixtures (Go)
- `backend/internal/testfixtures/users.go` — utilizatori predefiniți cu password_hash bcrypt
- `backend/internal/testfixtures/goals.go` — GO-uri în diferite stări (ACTIVE, WAITING, PAUSED)
- `backend/internal/testfixtures/sprints.go` — sprints cu daily_scores parțiali
- Setup helper: `testfixtures.Setup(t, db)` clean + insert toate

### Time mocking
- Use `clock.Clock` interface în engine (introdus în F8.2)
- Tests injectează `clock.NewMock(time.Now())` și avansează cu `clock.Advance(24 * time.Hour)`
- Esențial pentru TS-04 (5 zile inactive), TS-08 (Day 1 fallback)

### AI mocking
- Interface `ai.Client` deja există
- `testfixtures/ai_mock.go`: returnează rezultate deterministe; suportă mode chaos (timeout simulat)

---

## 6. Bug Reporting Format

Când QA descoperă o problemă în testing:

1. Adaugă în BACKLOG cu prefix `NEW-XX`
2. Câmpuri obligatorii:
   - **Severitate:** CRITICAL/HIGH/MEDIUM/LOW
   - **Reproducibility:** ALWAYS / SOMETIMES (cu rate) / RARE
   - **Steps to reproduce:** numerotat clar
   - **Expected:** comportament conform spec
   - **Actual:** ce se întâmplă
   - **Logs/screenshot:** atașat sau link
   - **Environment:** local / staging / production
3. PM triază în 24h: assign owner, faza, prioritate
4. Dacă blochează un gate → notificare imediată Architect

---

## 7. Continuous Testing Mindset

**După lansarea MVP (POST F8.8):**
- Pre-commit hook: ruleaza unit + opacity + schema-check
- Pe push: full CI (toate niveluri 1-5)
- Pe PR la main: requirement minim coverage delta non-negativ
- Săptămânal: rulare automatizată E2E + perf
- Lunar: security scan complet (Snyk, Dependabot, manual review)
