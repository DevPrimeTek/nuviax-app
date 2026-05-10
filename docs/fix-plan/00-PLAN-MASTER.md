# PLAN-MASTER — NuviaX MVP Fix Plan (Phase F8)

> **Versiune:** 1.0.0
> **Data:** 2026-05-10
> **Owner:** Solution Architect (rol Claude Opus)
> **Sursă:** `docs/audit/AUDIT-01-deviation-report.md`
> **Status:** READY FOR EXECUTION
> **Branch root:** `claude/fix-mvp-*`

---

## 1. Misiune

Adu codebase-ul NuviaX la calitate de **lansare MVP**: zero devieri critice față de Framework Rev 5.6, contractul API opac complet respectat, scenariile TS-01–TS-12 toate verzi, schema DB versionată local, suite de teste automatizate care să blocheze regresii.

**Țintă livrare:** F8 complet (toate sub-fazele F8.1–F8.8) → MVP gata de pilot/launch.

**Non-goals (rămân POST-MVP):** componente C15–C18, C29, C31, C34–C36, C39, C40 conform `MVP_SCOPE.md`. Stripe, i18n complet, PWA, advanced analytics — nu intră în F8.

---

## 2. Principii de execuție

1. **Foundation first.** Schema DB se reconciliază înainte de orice fix de cod. Cod care depinde de tabele inexistente nu se atinge până la F8.1 ✅.
2. **Engine-first contracts.** Orice handler/scheduler care trebuie să apeleze o funcție engine, așteaptă ca acea funcție să fie definită în F8.2.
3. **Strict gating.** Fiecare sub-fază F8.x are un **gate de testare** — fără gate verde, faza următoare nu pornește. Owner-ul gate-ului este Senior QA + Senior PM.
4. **Backlog driven.** Fiecare DEV-XX din audit este un item de backlog. Status-ul backlog-ului este sursa de adevăr — nu progres ne-tracked.
5. **One session = one branch = one PR.** Branch convention: `claude/fix-<phase>-<slug>` (ex: `claude/fix-01-schema-reconciliation`). PR-uri draft, merge după gate verde.
6. **Romanian summaries, English prompts.** Prompturile pentru sesiuni Claude Code sunt în engleză (consistență cu API). Rapoarte, status, comentarii PR — în română.
7. **No silent regressions.** Fiecare PR rulează: unit tests, integration tests (după F8.7), build frontend, opacity scan, schema diff.

---

## 3. Plan de faze F8.1 → F8.8

| # | Sub-fază | Scop | Owner principal | Backlog acoperit | Estimare | Pre-condiții |
|---|----------|------|-----------------|------------------|----------|--------------|
| F8.1 | **Schema Reconciliation** | Definește local toate cele ~17 tabele referite de cod; aliniază users/sessions; idempotent migrations | DBA + Backend Senior | DEV-02, DEV-03, DEV-04, DEV-17 | 60 min | AUDIT-01 mergere ✅ |
| F8.2 | **Engine Restructure** | Adaugă funcțiile engine lipsă (`GenerateProgressVisualization`, `FreezeExpectedTrajectory`, `MarkEvolutionSprint`, `ApplySRMFallback`, `GenerateCompletionCeremony`, `ApplyEvolveOverride`); tests | Backend Senior + Architect | DEV-11 (split), pre-rec pentru F8.4/F8.5 | 90 min | F8.1 ✅ |
| F8.3 | **API Security Hardening** | Eliminare `sprint_score` raw din `GetCeremony`; opacity scan complet pe toate handler-ele | Security Engineer + Backend | DEV-01 | 30 min | F8.1 ✅ (poate rula în paralel cu F8.2) |
| F8.4 | **Scheduler Wiring** | Apeluri reale către `growth_trajectories`, `achievements`, `srm_events`; evolution detection; ALI weekly | Backend Senior | DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19 | 90 min | F8.1 ✅, F8.2 ✅ |
| F8.5 | **Handler Hardening** | Day-1 visualization fallback; SRM L3 freeze; energy DB write; ALI breakdown în SRM status; AI timeouts; consistency fixes | Backend Senior | DEV-06, DEV-12, DEV-13, DEV-14, DEV-15, DEV-22, DEV-23, DEV-25, DEV-26, DEV-27, DEV-28 | 90 min | F8.2 ✅, F8.4 ✅ |
| F8.6 | **Frontend Polish** | Onboarding selecție durată GO; sprint cap; cleanup labels energy; verificare ceremony envelope | Frontend Senior + UX | DEV-20, DEV-21 | 60 min | F8.5 ✅ |
| F8.7 | **Integration & E2E Tests** | Automatizare TS-01..TS-12; smoke scripts; opacity test ca CI gate; coverage minim 70% pe handlers | Senior QA | Toate DEV-XX (regression coverage) | 120 min | F8.1–F8.6 ✅ |
| F8.8 | **Staging + Production Validation** | Deploy staging, full TS run, perf check, security scan, sign-off | DevOps + QA + PM | — (gate final) | 90 min | F8.7 ✅ |

**Total estimat:** ~10–12 ore execuție, distribuite în 8 sesiuni (max 60–90 min fiecare conform CLAUDE.md §4).

**Drum critic:** F8.1 → F8.2 → F8.4 → F8.5 → F8.7 → F8.8. F8.3 și F8.6 pot rula în paralel cu F8.2/F8.5 dacă specialiștii sunt disponibili.

---

## 4. Diagrama dependențelor

```
F8.1 Schema ───┬──> F8.2 Engine ──┬──> F8.4 Scheduler ──┐
               │                  │                     ├──> F8.7 Tests ──> F8.8 Staging
               └──> F8.3 Security │                     │
                                  └──> F8.5 Handlers ──┬┘
                                                       │
                                       F8.6 Frontend ──┘
```

---

## 5. Gates (criterii de trecere)

Niciun gate nu este "self-attest" — fiecare are un verificator extern (Senior PM sau QA).

### Gate F8.1 — Schema verde
- [ ] `psql` poate aplica `001_schema.sql` + `002_runtime_baseline.sql` pe DB curată fără erori
- [ ] `db.RunMigrations()` la startup nu produce nicio eroare
- [ ] Toate cele 17 tabele referite de cod există local
- [ ] Schema check script (nou: `backend/scripts/schema-check.sh`) → exit 0
- [ ] Backlog: DEV-02, DEV-03, DEV-04, DEV-17 → status `RESOLVED`

### Gate F8.2 — Engine verde
- [ ] Funcțiile listate în `04-PROMPTS-FIX-SESSIONS.md §F8.2` toate există și sunt apelabile
- [ ] `go test ./internal/engine/... -v` → toate trec
- [ ] Coverage engine ≥ 80%
- [ ] Backlog: DEV-11 → status `RESOLVED` (split în sub-tasks)

### Gate F8.3 — API opacity verde
- [ ] `grep -rn "drift\|chaos_index\|weights\|score_components\|sprint_score" backend/internal/api/handlers/` → zero matches în răspunsuri JSON
- [ ] Test nou `opacity_test.go` rulează toate endpoint-urile cu user mock și verifică structura răspunsului
- [ ] Backlog: DEV-01 → status `RESOLVED`

### Gate F8.4 — Scheduler verde
- [ ] Trigger manual fiecare job (12 jobs) — toate rulează fără eroare
- [ ] După un sprint simulat: `growth_trajectories` populat, `achievements` populate, `srm_events` corect generate
- [ ] Backlog: DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19 → `RESOLVED`

### Gate F8.5 — Handlers verzi
- [ ] TS-08 manual: `GET /goals/:id/visualize` Day 1 → exact 1 entry, nu null
- [ ] TS-06: L3 confirm → `frozen_expected` în răspuns
- [ ] TS-04 indirect: `POST /context/energy low` → DB INSERT verificat
- [ ] Backlog: DEV-06, DEV-12, DEV-13, DEV-14, DEV-15, DEV-22, DEV-23, DEV-25, DEV-26, DEV-27, DEV-28 → `RESOLVED`

### Gate F8.6 — Frontend verde
- [ ] `npm run build` → 0 erori TypeScript
- [ ] Onboarding manual: durata GO se poate alege
- [ ] Sprint cap respectat (verificare DB)
- [ ] Backlog: DEV-20, DEV-21 → `RESOLVED`

### Gate F8.7 — Tests verzi
- [ ] CI pipeline rulează: unit + integration + opacity + schema-check
- [ ] Coverage handlers ≥ 70%
- [ ] TS-01..TS-12 toate automatizate, toate verzi
- [ ] Test gate raport în `docs/testing/F8.7-test-coverage-report.md`

### Gate F8.8 — Production-ready
- [ ] Deploy staging reușit
- [ ] Full TS-01..TS-12 walkthrough manual pe staging — toate verzi
- [ ] Security scan: 0 vuln HIGH/CRITICAL
- [ ] Performance: P95 < 500ms pentru endpoint-uri principale
- [ ] Sign-off de la Senior PM + Architect
- [ ] Raport final: `docs/testing/F8.8-staging-validation-report.md`

---

## 6. Definiție „MVP gata de lansare"

MVP NuviaX este lansabil când TOATE sunt true:

1. ✅ TS-01..TS-12 toate verzi (manual + automat)
2. ✅ AUDIT-01: 28 devieri toate `RESOLVED` sau `ACCEPTED_POST_MVP` (cu justificare)
3. ✅ Schema DB versionată complet (zero `runtime-only` tables)
4. ✅ Opacity API: zero internal fields în răspunsuri (verificat automatizat)
5. ✅ CI verde pe `main` (unit + integration + build + opacity + schema-check)
6. ✅ Performance baseline documentat (P50/P95/P99 pentru top 10 endpoints)
7. ✅ Smoke E2E manual reușit pe staging cu user real
8. ✅ Security: bcrypt cost ≥ 12, JWT RS256, admin returns 404, forgot-password timing-safe
9. ✅ Observability: logs structurate, error tracking minimal (Sentry sau echivalent)
10. ✅ Documentație up-to-date: README, CLAUDE.md, ROADMAP.md, API docs

---

## 7. Rolurile de echipă (vezi `01-TEAM-ROSTER.md`)

Solution Architect (Opus), Senior Backend Engineer (Sonnet), Senior Frontend Engineer (Sonnet), DBA / DB Engineer (Sonnet), Senior QA Tester (Sonnet), Security Engineer (Opus), Senior Project Manager (Sonnet), DevOps Engineer (Sonnet), Product Manager / Framework Owner (Opus), UX Lead (Sonnet) — 10 roluri, fiecare cu responsabilități clare.

---

## 8. Cum se folosește planul

1. **Senior PM** deschide `02-BACKLOG.md`, alege următorul item conform priorității și fazei curente.
2. **Architect** validează că pre-condițiile fazei sunt îndeplinite (gate-ul precedent verde).
3. Specialistul desemnat (per `01-TEAM-ROSTER.md`) deschide o nouă sesiune Claude Code cu prompt-ul din `04-PROMPTS-FIX-SESSIONS.md` corespunzător sesiunii.
4. La final: PR draft, **Senior QA** verifică gate-ul, **PM** mută backlog-ul în `IN_REVIEW` apoi `RESOLVED`.
5. Architect actualizează `ROADMAP.md` cu data de finalizare a fazei.
6. Trecere la următoarea fază.

---

## 9. Risk register (toplevel)

| Risc | Impact | Probabilitate | Mitigare |
|------|--------|---------------|----------|
| DB de pe server are schema diferită de cea reconstituită | HIGH (downtime la deploy) | MEDIUM | F8.1 include export schema reală de pe server; comparație diff înainte de migrare |
| Funcții engine noi rup unit tests existente | MEDIUM | LOW | F8.2 menține contractul vechi; doar adăugări |
| Opacity scan găsește scurgeri suplimentare neidentificate | MEDIUM | MEDIUM | F8.3 include scan automatizat care va prinde toate cazurile |
| Tabele lipsă (vezi DEV-02) descoperite târziu | HIGH | LOW (după F8.1) | F8.1 face audit complet de table refs |
| Frontend rupe contractul după ce backend răspunde nou (envelope changes) | MEDIUM | MEDIUM | F8.6 sincronizează frontend cu backend; integration tests în F8.7 |
| Time overrun (>12h total) | MEDIUM | MEDIUM | Splitting agresiv pe sesiuni 60-90min; PM ajustează priorități |

---

## 10. Versionare plan

| Versiune | Data | Schimbări |
|----------|------|-----------|
| 1.0.0 | 2026-05-10 | Plan inițial, 8 sub-faze, 28 backlog items |

Modificări la plan necesită aprobare PM + Architect.
