# BACKLOG — NuviaX MVP Fix Phase F8

> **Versiune:** 1.0.0
> **Data:** 2026-05-10
> **Owner:** Senior Project Manager
> **Sursă:** `docs/audit/AUDIT-01-deviation-report.md`
> **Format status:** `OPEN` → `IN_PROGRESS` → `IN_REVIEW` → `RESOLVED` / `BLOCKED` / `ACCEPTED_POST_MVP`

---

## Cum se folosește

1. PM mută item-ul în `IN_PROGRESS` când se deschide sesiunea
2. Specialistul commit-uie → mută în `IN_REVIEW`
3. QA verifică gate-ul → mută în `RESOLVED`
4. Dacă apare blocaj: `BLOCKED` + nota motivului în câmpul „Note"
5. Dacă scope-ul exclude din MVP: `ACCEPTED_POST_MVP` + decizie PM logată

**Items noi descoperite în testing → se adaugă cu prefix `NEW-XX` și sunt triate de PM + Architect.**

---

## Sumar status

| Status | Count |
|--------|-------|
| OPEN | 27 |
| IN_PROGRESS | 0 |
| IN_REVIEW | 0 |
| RESOLVED | 1 (DEV-24) |
| BLOCKED | 0 |
| ACCEPTED_POST_MVP | 0 |

---

## Backlog Items

---

### DEV-01 — `GET /ceremonies/:goalId` expune `sprint_score` raw

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.3 |
| **Owner** | Security Engineer |
| **Reviewer** | Backend Senior + Architect |
| **Estimare** | 15 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/achievements.go:67-82` |
| **Impact viitor dacă nu se fixează** | Scurgere de internal scoring → competitori pot face reverse-engineering al algoritmului. Violare directă a CLAUDE.md §7 (security invariant). |
| **Test gate** | TS-12 + opacity scan automatizat trebuie să confirme zero matches pentru `sprint_score` în răspunsuri |
| **Note** | — |

---

### DEV-02 — Tabele DB ne-definite în migrări locale

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.1 |
| **Owner** | DBA |
| **Reviewer** | Architect |
| **Estimare** | 45 min (din total 60 min faza) |
| **Dependențe** | — (faza fundație) |
| **Fișiere** | `backend/migrations/002_runtime_baseline.sql` (nou), `backend/internal/db/db.go` |
| **Tabele lipsă** | `srm_events`, `context_adjustments`, `stagnation_events`, `ceremonies`, `sprint_results`, `go_metrics`, `achievements`/`achievement_badges`, `growth_trajectories`, `evolution_sprints`, `completion_ceremonies` |
| **Impact viitor dacă nu se fixează** | Aplicația crash-uiește pe DB curată; deploy-uri noi imposibile; dezvoltare locală blocată. |
| **Test gate** | Schema check script verde + `db.RunMigrations()` fără erori pe DB curată |
| **Note** | Trebuie făcut export schema reală de pe server pentru reconciliere (pe regula CLAUDE.md §4) |

---

### DEV-03 — Schema `users` din 001_schema.sql nu se potrivește cu cod

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.1 |
| **Owner** | DBA |
| **Reviewer** | Architect |
| **Estimare** | 10 min |
| **Dependențe** | DEV-02 (același fișier de migrare) |
| **Fișiere** | `backend/migrations/001_schema.sql:8-17` (sau nouă migrare ALTER TABLE) |
| **Lipsă în schema** | `email_encrypted`, `salt`, `full_name`, `locale`, `theme`, `avatar_url`, `mfa_secret`, `mfa_enabled`, `is_active` |
| **Impact viitor dacă nu se fixează** | Auth flow rupt complet pe DB nouă; register/login imposibil. |
| **Test gate** | TS-01 (register + login) verde |

---

### DEV-04 — Schema `sessions` vs cod `user_sessions`

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.1 |
| **Owner** | DBA |
| **Reviewer** | Backend Senior |
| **Estimare** | 5 min |
| **Dependențe** | DEV-02 |
| **Fișiere** | `backend/migrations/001_schema.sql:22-29` (rename) sau `backend/internal/db/queries.go:91-119` (rename) |
| **Decizie necesară** | Architect alege numele canonic: `user_sessions` (cod) sau `sessions` (schema) |
| **Impact viitor dacă nu se fixează** | Sessions/refresh token-uri rupte. |
| **Test gate** | TS-01 + refresh token flow verde |

---

### DEV-05 — `growth_trajectories` nu este populat (SA-1)

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 20 min |
| **Dependențe** | F8.1 ✅ (tabelul `growth_trajectories` trebuie să existe) |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:171-235`, `backend/internal/engine/growth.go` |
| **Impact viitor dacă nu se fixează** | TS-03 fail; LineChart la `/goals/[id]` rămâne empty; user nu vede evoluție pe sprint-uri. |
| **Test gate** | După 2 rulări `jobComputeDailyScore`, `growth_trajectories` are 2 rânduri pentru goal-ul activ |

---

### DEV-06 — `GetGoalVisualize` nu are fallback Day 1

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 25 min |
| **Dependențe** | F8.2 ✅ (funcția `GenerateProgressVisualization`) |
| **Fișiere** | `backend/internal/api/handlers/goals.go:447-494`, `backend/internal/engine/growth.go` (sau visualization.go nou) |
| **Impact viitor dacă nu se fixează** | TS-08 fail; user vede chart gol în prima zi → feedback negativ imediat. |
| **Test gate** | TS-08: Day 1 după create → exact 1 entry, `expected_pct > 0` |

---

### DEV-07 — `fn_award_achievement_if_earned()` niciodată apelat (SA-2)

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 25 min |
| **Dependențe** | F8.1 ✅ (tabel `achievement_badges`), F8.2 ✅ |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:292-403, :638-688`, eventual creare migrare cu funcția SQL dacă nu există |
| **Impact viitor dacă nu se fixează** | TS-07 fail; gamification mort; `/achievements` empty pentru toți userii. |
| **Test gate** | După simulare sprint închidere: `achievement_badges` are minim 1 row pentru user |

---

### DEV-08 — `jobCheckSRMTimeouts` log-only (SA-6)

| Câmp | Valoare |
|------|---------|
| **Severitate** | CRITICAL |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 20 min |
| **Dependențe** | F8.2 ✅ (`engine.ApplySRMFallback`) |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:587-633`, `backend/internal/engine/srm.go` (extindere) |
| **Impact viitor dacă nu se fixează** | Goals stuck în L3 unconfirmed la nesfârșit; user blocked. |
| **Test gate** | Test scenariu: L3 unconfirmed > 24h → după rulare job, status DB schimbat conform fallback |

---

### DEV-09 — `jobDetectEvolution` placeholder

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | PM (scope decision) |
| **Estimare** | 30 min |
| **Dependențe** | F8.1 ✅ (tabel `evolution_sprints`), F8.2 ✅ (`MarkEvolutionSprint`, `ApplyEvolveOverride`) |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:693-700`, `backend/internal/engine/evolution.go` (nou) |
| **Decizie scope** | C31 simplified MVP per CLAUDE.md §2 → trebuie implementat minimal (delta threshold 0.05 + INSERT idempotent) |
| **Impact viitor dacă nu se fixează** | PLATINUM ceremony niciodată acordat; gamification incomplet. |
| **Test gate** | Sprint cu delta ≥ 0.05 vs sprint anterior → INSERT în `evolution_sprints` + tier PLATINUM la ceremony |

---

### DEV-10 — `jobComputeWeeklyALI` placeholder

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | PM (scope decision) |
| **Estimare** | 25 min |
| **Dependențe** | F8.1 ✅ (tabel `go_metrics` sau `ali_snapshots`) |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:467-474`, `backend/internal/engine/growth.go` |
| **Decizie scope** | C38 (GORI) este MVP simplified; ALI breakdown e parte din C38. |
| **Impact viitor dacă nu se fixează** | `GetSRMStatus` nu poate returna ali_current/ali_projected; velocity control mort. |
| **Test gate** | După un sprint complet: `go_metrics` (sau echivalent) are ali calculat |

---

### DEV-11 — Engine file structure stale față de docs

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.2 |
| **Owner** | Architect + Backend Senior |
| **Reviewer** | PM |
| **Estimare** | 90 min (faza completă) |
| **Dependențe** | F8.1 ✅ |
| **Fișiere** | `backend/internal/engine/` — adaugă: `visualization.go`, `regulatory.go`, `evolution.go`. Extinde: `growth.go`, `srm.go` |
| **Funcții lipsă de adăugat** | `GenerateProgressVisualization`, `FreezeExpectedTrajectory`, `UnfreezeExpectedTrajectory`, `MarkEvolutionSprint`, `GenerateCompletionCeremony`, `ApplyEvolveOverride`, `CheckAndRecordRegressionEvent`, `ApplySRMFallback`, `ComputeALIBreakdown` |
| **Impact viitor dacă nu se fixează** | Toate fazele F8.4–F8.5 sunt blocate; nu pot apela ce nu există. |
| **Test gate** | Toate funcțiile au unit tests; coverage engine ≥ 80% |
| **Note** | NU se elimină funcțiile vechi care funcționează (preserve, don't rewrite — CLAUDE.md §8) |

---

### DEV-12 — `ConfirmSRML3` nu îngheață trajectory

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 20 min |
| **Dependențe** | F8.2 ✅ (`FreezeExpectedTrajectory`), F8.1 ✅ (coloane `expected_pct_frozen`, `frozen_expected_pct` pe sprints) |
| **Fișiere** | `backend/internal/api/handlers/srm.go:69-90` |
| **Impact viitor dacă nu se fixează** | Drift loop paradox (GAP #20) reapare; user în PAUSE primește scoruri tot mai proaste degeaba. |
| **Test gate** | TS-06: după L3 confirm, response include `frozen_expected: <float>`; daily_scores ulterior nu mai schimbă expected_pct |

---

### DEV-13 — `POST /context/energy` no-op

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 15 min |
| **Dependențe** | F8.1 ✅ (tabel `context_adjustments`) |
| **Fișiere** | `backend/internal/api/handlers/today.go:182-197` |
| **Impact viitor dacă nu se fixează** | Reducere intensitate la energie scăzută inactivă; framework C32 incomplet. |
| **Test gate** | `POST /context/energy {level: low}` → INSERT verificat în DB |

---

### DEV-14 — Frontend SetEnergy folosește label-uri diferite

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.5 (backend accept) + F8.6 (frontend send) |
| **Owner** | Backend Senior + Frontend Senior |
| **Reviewer** | UX |
| **Estimare** | 10 min |
| **Dependențe** | DEV-13 |
| **Fișiere** | `backend/internal/api/handlers/today.go:191-195`, frontend componenta de energie |
| **Decizie** | Standard: `low`/`normal`/`high` (per user-workflow.md). Backend acceptă și legacy `mid`/`hi` pentru compat. |
| **Impact viitor dacă nu se fixează** | Mismatch între ce trimite UI și ce înțelege backend; energie ignorată tăcut. |
| **Test gate** | Frontend trimite `normal`, backend acceptă; legacy `mid` funcționează ca alias |

---

### DEV-15 — `GetSRMStatus` nu calculează ALI breakdown

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 25 min |
| **Dependențe** | F8.2 ✅ (`ComputeALIBreakdown`), DEV-10 ✅ (date ali persistente) |
| **Fișiere** | `backend/internal/api/handlers/srm.go:13-38` |
| **Impact viitor dacă nu se fixează** | Velocity control mort; user nu vede ambition buffer warning. |
| **Test gate** | `GET /srm/status/:id` returnează ali block complet conform user-workflow.md §4.4 |

---

### DEV-16 — `jobCloseExpiredSprints` nu calculează tier cu isEvolution

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | PM (Framework alignment) |
| **Estimare** | 15 min |
| **Dependențe** | F8.2 ✅ (CeremonyTier(score, isEvolution)), DEV-09 |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:339-373`, `backend/internal/engine/growth.go:29-40` |
| **Impact viitor dacă nu se fixează** | PLATINUM și GOLD identice; nu există diferențiere semantică. |
| **Test gate** | Sprint cu score=0.95 și isEvolution=true → tier=PLATINUM; isEvolution=false → tier=GOLD |

---

### DEV-17 — `GET /achievements` referă tabel ne-definit

| Câmp | Valoare |
|------|---------|
| **Severitate** | HIGH |
| **Status** | OPEN |
| **Faza** | F8.1 (decizie nume) + F8.5 (aliniere query) |
| **Owner** | DBA + Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 15 min |
| **Dependențe** | DEV-02 |
| **Decizie** | Nume canonic: `achievement_badges` (per user-workflow.md). Tabelul are coloanele `id`, `user_id`, `badge_type`, `go_id`, `sprint_id`, `awarded_at`. |
| **Fișiere** | `backend/migrations/002_*.sql`, `backend/internal/api/handlers/achievements.go:24-29` |
| **Impact viitor dacă nu se fixează** | Endpoint crash sau empty constant. |
| **Test gate** | După DEV-07 fix, `GET /achievements` returnează badges reale |

---

### DEV-18 — Trigger SRM L1 folosește drift, nu 5-zile-inactive

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | PM (semantic alignment) |
| **Estimare** | 20 min |
| **Dependențe** | DEV-19 |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:240-287` |
| **Decizie** | Adaugă (NU înlocui) detecția 5-zile-inactive; ambele criterii pot trigger L1 (drift critic SAU 5d inactive). |
| **Impact viitor dacă nu se fixează** | TS-04 nu trece curat; user inactiv 5 zile nu primește SRM L1. |
| **Test gate** | TS-04: 5 zile fără MAIN done → `srm_level=L1` |

---

### DEV-19 — `jobCheckStagnation` nu inserează în `srm_events`

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.4 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 15 min |
| **Dependențe** | F8.1 ✅ |
| **Fișiere** | `backend/internal/scheduler/scheduler.go:538-582` |
| **Impact viitor dacă nu se fixează** | Same ca DEV-18 — stagnation events orfane, nu se traduc în SRM. |
| **Test gate** | După rulare job: pentru fiecare row în `stagnation_events` există row în `srm_events` (level=L1, event_type=STAGNATION_5D) |

---

### DEV-20 — Sprint creation hardcodat 30 zile, ignoră goal end_date

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 5 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/goals.go:309-310` |
| **Impact viitor dacă nu se fixează** | Sprint depășește goal end_date; corner case dar genera rezultate inconsistente. |
| **Test gate** | Goal cu durată 20 zile → sprint 1 end = goal end (nu +30) |

---

### DEV-21 — Onboarding hardcodează durata 90 zile

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.6 |
| **Owner** | Frontend Senior |
| **Reviewer** | UX + PM |
| **Estimare** | 30 min |
| **Dependențe** | — |
| **Fișiere** | `frontend/app/app/onboarding/page.tsx:145-149` |
| **Soluție recomandată** | Adaugă step de selecție durată: 30 / 90 / 180 / 365 zile (preset-uri) |
| **Impact viitor dacă nu se fixează** | User nu poate seta GO de 6 luni sau 1 an; flexibilitatea framework C4 (max 365) nu e expusă. |
| **Test gate** | TS-01: user creează GO cu durată custom → DB respectă selecția |

---

### DEV-22 — `SuggestGOCategory` fără context timeout

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 10 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/goals.go:88-89` |
| **Impact viitor dacă nu se fixează** | TS-11 fail dacă Anthropic e lent; onboarding blocked. |
| **Test gate** | TS-11: cu cheie invalidă, răspuns < 2s |

---

### DEV-23 — `AnalyzeGO` timeout 8s

| Câmp | Valoare |
|------|---------|
| **Severitate** | MEDIUM |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 5 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/goals.go:32-33` |
| **Impact viitor dacă nu se fixează** | UX slab dacă AI lent. |
| **Test gate** | Răspuns < 3s în 95th percentile |

---

### DEV-24 — Cron `*/90` rezolvat (SA-7)

| Câmp | Valoare |
|------|---------|
| **Severitate** | LOW |
| **Status** | RESOLVED |
| **Faza** | — (deja rezolvat) |
| **Note** | Confirmat în AUDIT-01: cron este `0 2 * * 0`. Nu mai este nevoie de fix. |

---

### DEV-25 — Streak compute potențial off-by-one

| Câmp | Valoare |
|------|---------|
| **Severitate** | LOW |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | QA |
| **Estimare** | 15 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/today.go:200-230` |
| **Impact viitor dacă nu se fixează** | Streak displayed wrong după ce userul nu bifează azi; UX subtil greșit. |
| **Test gate** | Unit test cu fixture: 7 zile DONE consecutive ieri și înainte → streak=7 indiferent dacă azi a bifat |

---

### DEV-26 — `ListAchievements` returnează array direct

| Câmp | Valoare |
|------|---------|
| **Severitate** | LOW |
| **Status** | OPEN |
| **Faza** | F8.5 + F8.6 |
| **Owner** | Backend Senior + Frontend Senior |
| **Reviewer** | UX |
| **Estimare** | 10 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/achievements.go:44`, `frontend/app/app/achievements/page.tsx` |
| **Impact viitor dacă nu se fixează** | Inconsistență API; frontend trebuie cazuri speciale per endpoint. |
| **Test gate** | Răspuns: `{"achievements": [...]}`; frontend update sincron |

---

### DEV-27 — `GetCeremony` returnează `null` body pe miss

| Câmp | Valoare |
|------|---------|
| **Severitate** | LOW |
| **Status** | OPEN |
| **Faza** | F8.5 + F8.6 |
| **Owner** | Backend Senior + Frontend Senior |
| **Reviewer** | UX |
| **Estimare** | 10 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/achievements.go:73-74`, frontend CeremonyModal |
| **Decizie** | 200 + `{"ceremony": null}` envelope. Frontend verifică `ceremony !== null`. |
| **Test gate** | API contract clar; frontend nu mai parsează body-ul `null` literal |

---

### DEV-28 — Behavior model inconsistency `dominant_behavior_model` vs `behavior_model`

| Câmp | Valoare |
|------|---------|
| **Severitate** | LOW |
| **Status** | OPEN |
| **Faza** | F8.5 |
| **Owner** | Backend Senior |
| **Reviewer** | Architect |
| **Estimare** | 10 min |
| **Dependențe** | — |
| **Fișiere** | `backend/internal/api/handlers/goals.go:224, 274, 302` |
| **Decizie Architect** | API field name: `behavior_model` peste tot (input și output). Drop `dominant_behavior_model` ca nume duplicat. Documentează în `docs/integrations.md`. |
| **Impact viitor dacă nu se fixează** | Confuzie pentru clienți API; mai mult cod de mapping. |
| **Test gate** | Frontend + API folosesc același nume; tests pass |

---

## Items noi descoperite în execuție

> Format: `NEW-XX` cu aceleași câmpuri ca DEV-XX. Triat de PM săptămânal.

(niciun item nou la momentul 2026-05-10)

---

## Decizii blocante deschise

| ID | Întrebare | Decizia | Owner | Termen |
|----|-----------|---------|-------|--------|
| DEC-01 | Numele canonic pentru achievement table: `achievements` sau `achievement_badges`? | RECOMANDAT: `achievement_badges` (per user-workflow.md) | Architect | F8.1 start |
| DEC-02 | C31 evolution detection: MVP-required sau POST-MVP? | RECOMANDAT: MVP simplified (delta 0.05 + INSERT) | PM | F8.4 start |
| DEC-03 | C38 ALI: persistat săptămânal sau computed on-demand? | RECOMANDAT: săptămânal în `go_metrics`, on-demand cache 1h | Architect | F8.4 start |

---

## Impact analysis (ce se rupe dacă nu rezolvăm)

| Fără fix | Scenarii TS care eșuează | Impact business |
|----------|--------------------------|-----------------|
| F8.1 (schema) | TS-01, TS-04..TS-07 | Deploy nou imposibil; onboarding rupt; gamification mort |
| F8.2 (engine) | TS-03, TS-06, TS-07, TS-08 | Visualization, ceremonies, achievements toate rupte |
| F8.3 (security) | TS-12 | Scurgere internal scoring; reverse-engineering posibil |
| F8.4 (scheduler) | TS-03, TS-04, TS-05, TS-07 | Engine "trăiește" doar la cerere; SRM nu auto-detect; achievements nu populate |
| F8.5 (handlers) | TS-04, TS-06, TS-08, TS-11 | Day-1 chart vid; energie ignorată; AI timeouts depășesc |
| F8.6 (frontend) | TS-01 | Onboarding rigid; UX slab |
| F8.7 (tests) | Toate (regresie viitoare) | Fără gate CI = fix-uri se rup în iterații viitoare |
| F8.8 (staging) | — | MVP nu poate fi declarat "gata de pilot" |
