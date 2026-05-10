# AUDIT-01 — Deviation Report NuviaX MVP vs Framework

**Data audit:** 2026-04-28
**Versiune codebase:** v1.4.0
**Auditor:** Claude Code (Sonnet)
**Branch:** claude/nuviax-system-audit-BLxBt
**Scope:** Read-only audit — zero modificări de cod

---

## Rezumat executiv

| Severitate | Count |
|-----------|-------|
| CRITICAL  | 8 |
| HIGH      | 9 |
| MEDIUM    | 7 |
| LOW       | 4 |
| **TOTAL devieri** | **28** |

**Concluzie principală:** Aproximativ 60–70% din funcționalitățile descrise în Framework Rev 5.6 și `docs/user-workflow.md` (TS-01–TS-12) nu sunt implementate sau nu funcționează corect în runtime. Sursele majore sunt:

1. **Schema DB stale**: `001_schema.sql` definește 8 tabele, dar codul referențiază 17+ tabele (srm_events, ceremonies, context_adjustments etc.) ne-definite local. Aplicația depinde de tabele prezente doar pe server.
2. **Engine fragmentat**: structura de fișiere `level1_structural.go` / `level3_adaptive.go` / `level4_regulatory.go` / `level5_growth.go` descrisă în docs nu există — engine-ul real are `engine.go`, `growth.go`, `helpers.go`, `srm.go`, fără funcții cheie precum `GenerateProgressVisualization`, `FreezeExpectedTrajectory`, `MarkEvolutionSprint`.
3. **Status SA-1 → SA-7**: 3 sunt rezolvate (SA-4, SA-5, SA-7), 1 parțial (SA-3), 3 confirmate ne-rezolvate (SA-1, SA-2, SA-6).
4. **Securitate API**: o scurgere CRITICAL — `GET /ceremonies/:goalId` expune `sprint_score` raw (TS-12 fail).

---

## Devieri identificate

### DEV-01 — `GET /ceremonies/:goalId` expune `sprint_score` raw

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-12 |
| Componentă Framework | C37 (opacity invariant) |
| Fișier | `backend/internal/api/handlers/achievements.go:67-82` |
| Comportament actual | SELECT `sprint_score` din `ceremonies` și-l include în răspunsul JSON ca `"sprint_score": <float>`. |
| Comportament așteptat | Răspunsul trebuie să conțină DOAR `tier` (BRONZE/SILVER/GOLD/PLATINUM) + `viewed_at`. Sprint score este intern (Section 7 CLAUDE.md, Section 8.1/8.2 user-workflow.md). |
| Fix necesar | Elimină `sprint_score` din SELECT și din răspunsul JSON. Rămâne `id`, `tier`, `viewed_at`. |

---

### DEV-02 — Schema 001_schema.sql nu conține tabelele referite de cod (srm_events, ceremonies, context_adjustments etc.)

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-04, TS-05, TS-06, TS-07 |
| Componentă Framework | C32-C38 |
| Fișier | `backend/migrations/001_schema.sql` (singura migrare locală), `backend/internal/db/db.go` |
| Comportament actual | `001_schema.sql` definește 8 tabele (users, sessions, password_reset_tokens, audit_log, global_objectives, sprints, daily_tasks, daily_scores, go_ai_analysis). `db.go::ensureGoalsTables` adaugă încă câteva, dar **nu** include: `srm_events`, `context_adjustments`, `stagnation_events`, `ceremonies`, `sprint_results`, `go_metrics`, `achievements`, `growth_trajectories`. CLAUDE.md menționează 32 tabele și migrări 001-013 — doar 001 există local. Cod-ul referențiază aceste tabele și se bazează pe ele existând pe server. |
| Comportament așteptat | Toate cele 32 de tabele să fie definite în migrări versionate local; `RunMigrations` să le creeze idempotent. |
| Fix necesar | Reconstituie migrările 002-013 sau extinde `ensureGoalsTables` cu definițiile lipsă. **NU** dezactiva codul care depinde de aceste tabele — ele există probabil pe server (vezi regula CLAUDE.md §4 despre env vars). |

---

### DEV-03 — Schema users din 001_schema.sql nu se potrivește cu modelul așteptat de cod

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-01, TS-10 |
| Componentă Framework | Auth |
| Fișier | `backend/migrations/001_schema.sql:8-17` vs `backend/internal/db/queries.go:21-72` |
| Comportament actual | Schema locală definește `users` cu columns: `email, email_hash, password_hash, name, is_admin`. Codul `CreateUser`/`GetUserByID` selectează `email_encrypted, email_hash, password_hash, salt, full_name, locale, theme, avatar_url, mfa_secret, mfa_enabled, is_active, is_admin`. Pe DB creată cu 001_schema.sql, queries vor da erori SQL. |
| Comportament așteptat | Schema să corespundă cu modelele Go (email_encrypted + salt + full_name + locale + theme + mfa_* + is_active). |
| Fix necesar | Aliniază 001_schema.sql cu schema reală a DB-ului de pe server. Adaugă columns lipsă cu `ALTER TABLE … ADD COLUMN IF NOT EXISTS`. |

---

### DEV-04 — Schema sessions vs cod user_sessions

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-01 |
| Componentă Framework | Auth |
| Fișier | `backend/migrations/001_schema.sql:22-29` vs `backend/internal/db/queries.go:91-119` |
| Comportament actual | Schema locală creează tabel `sessions`. Codul query `INSERT INTO user_sessions` și `SELECT … FROM user_sessions`. Numele de tabel diferă. |
| Comportament așteptat | Numele de tabel coerent între schema și cod. |
| Fix necesar | Redenumește tabelul în schema sau în queries — alege un singur nume. |

---

### DEV-05 — `growth_trajectories` nu este populat (SA-1)

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-03, TS-08 |
| Componentă Framework | C25, C37 visualization |
| Fișier | `backend/internal/scheduler/scheduler.go:171-235` (jobComputeDailyScore) + `backend/internal/api/handlers/goals.go:447-494` |
| Comportament actual | `jobComputeDailyScore` execută `UPSERT daily_scores` dar nu apelează niciodată `fn_compute_growth_trajectory()`. Tabelul `growth_trajectories` nu există în migrări locale și nu este populat. `GetGoalVisualize` citește direct din `daily_scores`, fără `growth_trajectories`. |
| Comportament așteptat | După `UpsertGoalScore` să se apeleze `fn_compute_growth_trajectory(go_id, today)`; tabelul `growth_trajectories` să primească 1 rând/zi/GO activ. |
| Fix necesar | Adaugă apel la `fn_compute_growth_trajectory()` la final de `jobComputeDailyScore` (după `db.Exec(ctx, INSERT INTO daily_scores …)`). |

---

### DEV-06 — `GetGoalVisualize` nu are fallback Day 1 (SA-1 frontend impact)

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-08 |
| Componentă Framework | C37 visualization fallback |
| Fișier | `backend/internal/api/handlers/goals.go:462-493` |
| Comportament actual | Dacă `daily_scores` nu are rânduri pentru GO, `points` rămâne `[]`; răspuns: `{"trajectory": []}`. Frontend renderează LineChart gol. |
| Comportament așteptat | Când nu există date, calculează un snapshot live: `expected_pct = elapsed/total`, `actual_pct = 0`, `delta = -expected_pct`, `trend = "ON_TRACK"`. Returnează exact 1 entry — niciodată array gol. |
| Fix necesar | Dacă `len(points) == 0`, fă SELECT din `global_objectives` pentru `start_date`/`end_date`, calculează `expected_pct` linear, push 1 dataPoint sintetic. |

---

### DEV-07 — `fn_award_achievement_if_earned()` niciodată apelat din Go (SA-2)

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-07 |
| Componentă Framework | C37, C38 |
| Fișier | `backend/internal/scheduler/scheduler.go:292-403` (jobCloseExpiredSprints) + `:638-688` (jobGenerateCeremonies) |
| Comportament actual | Niciun job nu apelează `fn_award_achievement_if_earned()`. `jobCloseExpiredSprints` inserează în `ceremonies` și `sprint_results` dar nu acordă badge-uri. `jobGenerateCeremonies` la fel — generează doar ceremony, nu badge-uri. |
| Comportament așteptat | După inserarea ceremony, apelează `SELECT fn_award_achievement_if_earned($1, $2)` cu user_id + sprint_id. Tabelul `achievements`/`achievement_badges` să se populeze automat. |
| Fix necesar | Adaugă apel la `fn_award_achievement_if_earned()` în loop-ul din `jobGenerateCeremonies` (după INSERT ceremony) și/sau în `jobCloseExpiredSprints`. |

---

### DEV-08 — `jobCheckSRMTimeouts` log-only, fără fallback aplicat (SA-6)

| Câmp | Valoare |
|------|---------|
| Severitate | CRITICAL |
| Scenariu afectat | TS-06 |
| Componentă Framework | C33, C35 |
| Fișier | `backend/internal/scheduler/scheduler.go:587-633` |
| Comportament actual | `jobCheckSRMTimeouts` apelează `engine.ComputeSRMFallback(hoursSince)` și înregistrează rezultatul cu `logger.Info(...)`, dar nu execută nicio mutație: nu inserează `srm_events`, nu schimbă `global_objectives.status`, nu reduce intensitatea. |
| Comportament așteptat | Pe baza `fallbackAction` (PAUSE/L1/L2): execută acțiunea — INSERT `srm_events` cu nivelul calculat, sau UPDATE goal status='PAUSED' pentru PAUSE. Goal cu L3 ne-confirmat la timeout să nu rămână blocat. |
| Fix necesar | Implementează `engine.ApplySRMFallback(ctx, db, goalID, fallbackAction)` și apelează din loop. |

---

### DEV-09 — `jobDetectEvolution` placeholder

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-07 (evolution-based ceremony tier) |
| Componentă Framework | C31 |
| Fișier | `backend/internal/scheduler/scheduler.go:693-700` |
| Comportament actual | Funcția conține doar `// TODO(F8): implement C31 behavioral pattern detection (post-MVP)` și un log. Niciun delta între sprint-uri nu este calculat, `evolution_sprints` nu este populat. PLATINUM ceremony (care necesită `isEvolution=true`) nu se va emite vreodată. |
| Comportament așteptat | Per spec C31 (POST-MVP) — dar Framework Rev 5.6 marchează C31 ca SIMPLIFIED MVP în CLAUDE.md §2. Calculează delta `current_sprint_score - prev_sprint_score >= 0.05`, INSERT în `evolution_sprints`. |
| Fix necesar | Conform `MVP_SCOPE.md`: confirmă dacă C31 este MVP-required sau POST-MVP. Dacă MVP, implementează minimal — delta threshold + INSERT idempotent. |

---

### DEV-10 — `jobComputeWeeklyALI` placeholder

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-04, TS-05 (ALI velocity control) |
| Componentă Framework | C38 ALI |
| Fișier | `backend/internal/scheduler/scheduler.go:467-474` |
| Comportament actual | Placeholder cu `// TODO(F8): implement C38 ALI computation`. ALI breakdown nu este pre-calculat săptămânal — `GetSRMStatus` nu poate returna `ali_current`/`ali_projected` (vezi DEV-15). |
| Comportament așteptat | Calculează ALI weekly snapshot per goal, persist (eventual în `go_metrics` sau `ali_snapshots`); `GetSRMStatus` să-l citească. |
| Fix necesar | Implementează în baza formulei: `ALI_current = tasks_done_until_now / expected_until_now`, `ALI_projected` = proiecție liniară. |

---

### DEV-11 — Engine file structure stale față de docs

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-04, TS-05, TS-06, TS-07, TS-08 |
| Componentă Framework | C19-C38 |
| Fișier | `backend/internal/engine/` directory |
| Comportament actual | Engine real conține: `engine.go`, `growth.go`, `helpers.go`, `srm.go`, `engine_test.go` (5 files). Documentele referă constant `level1_structural.go`, `level3_adaptive.go`, `level4_regulatory.go`, `level5_growth.go` și funcții ce nu există: `GenerateDailyTasks`, `GenerateProgressVisualization`, `FreezeExpectedTrajectory`, `UnfreezeExpectedTrajectory`, `MarkEvolutionSprint`, `GenerateCompletionCeremony`, `ApplyEvolveOverride`, `CheckAndRecordRegressionEvent`, `ComputeProgressPct`, `ComputeGoalScore`, `ApplySRMFallback`. |
| Comportament așteptat | Funcții documentate să existe în engine, cu logica descrisă. |
| Fix necesar | Decide: (a) actualizează docs să reflecte structura reală a engine-ului, sau (b) implementează funcțiile lipsă pentru paritate cu Framework. Funcțiile critice de implementat: `FreezeExpectedTrajectory` (DEV-12), `GenerateProgressVisualization` (DEV-06), `ApplyEvolveOverride` (DEV-09). |

---

### DEV-12 — `ConfirmSRML3` nu îngheață trajectory (drift loop paradox)

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-06 |
| Componentă Framework | C35, GAP #20 |
| Fișier | `backend/internal/api/handlers/srm.go:69-90` |
| Comportament actual | UPDATE `global_objectives` SET status='PAUSED' + INSERT `context_adjustments` type='PAUSE'. Nu se apelează `engine.FreezeExpectedTrajectory(sprintID)` (funcția nu există). `expected_pct` continuă să avanseze, drift-ul se înrăutățește în PAUSE → drift loop paradox. Răspunsul JSON nu include `frozen_expected`. |
| Comportament așteptat | Setează `sprints.expected_pct_frozen = TRUE` + `frozen_expected_pct = current_elapsed_ratio`. Răspunde cu `{"new_status": "PAUSED", "frozen_expected": <float 0-1>}`. |
| Fix necesar | Implementează `engine.FreezeExpectedTrajectory(ctx, db, sprintID)` (UPDATE coloanele de freeze pe sprint), apoi apel-o din `ConfirmSRML3`. Adaugă `frozen_expected` în response. |

---

### DEV-13 — `POST /context/energy` no-op (nu scrie în DB)

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-04 (ENERGY_LOW preventiv), Daily Loop §1.4 |
| Componentă Framework | C32 Pause/Adaptive Context |
| Fișier | `backend/internal/api/handlers/today.go:182-197` |
| Comportament actual | Handler validează level-ul, returnează 200 OK fără INSERT. Comentariu in-line: `// context_adjustments table is pending migration — accepted and acknowledged without DB write`. Dar `ConfirmSRML2` și `ConfirmSRML3` scriu în acel tabel — deci tabelul există. Reducerea de intensitate la energie scăzută nu se aplică. |
| Comportament așteptat | Pentru `low`/`high` (sau echivalente): INSERT `context_adjustments` cu `adjType=AdjEnergyLow`/`AdjEnergyHigh`, valid azi+mâine. `normal` → no-op (200). Cache today-tasks invalidat. |
| Fix necesar | Înlocuiește no-op-ul cu INSERT condiționat pe level. Aliniază numele level-urilor cu frontend-ul (`low`/`normal`/`high`, NU `low`/`mid`/`hi`). |

---

### DEV-14 — Frontend SetEnergy folosește label-uri diferite

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-04 |
| Componentă Framework | C32 |
| Fișier | `backend/internal/api/handlers/today.go:191-195` |
| Comportament actual | Backend acceptă `low`/`mid`/`hi`. Frontend (per docs §3.2) trimite `low`/`normal`/`high`. Doc `mid → normal`, `hi → high` apare ca normalizare; nu există cod care normalizează. |
| Comportament așteptat | Backend să accepte ambele seturi sau frontend să trimită strict `low`/`mid`/`hi` (existând test fail dacă mismatched). |
| Fix necesar | Standardizează pe `low`/`normal`/`high` (alinierea cu user-workflow.md §3.2 și `SetEnergy`-ului din docs). Acceptă și legacy values pentru backward compat. |

---

### DEV-15 — `GetSRMStatus` nu calculează ALI breakdown

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-05, user-workflow.md §4.2 |
| Componentă Framework | C38 GORI/ALI |
| Fișier | `backend/internal/api/handlers/srm.go:13-38` |
| Comportament actual | Răspunde `{"srm_level": "...", "triggered_at": ...}`. Lipsesc: `ali_current`, `ali_projected`, `in_ambition_buffer`, `velocity_control_on`, `goal_breakdown`, `note`, `message`. |
| Comportament așteptat | Conform docs §4.4: include `ali` block + `message` localizat per level. |
| Fix necesar | Adaugă `computeALIBreakdown()` (din `daily_scores`/`go_metrics`); compune răspunsul cu structura completă. |

---

### DEV-16 — `jobCloseExpiredSprints` nu apelează detect-evolution înainte de tier

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-07 (PLATINUM tier) |
| Componentă Framework | C37 |
| Fișier | `backend/internal/scheduler/scheduler.go:339-373` |
| Comportament actual | `tier := engine.CeremonyTier(sprintScore)` calculează tier strict pe baza scorului, fără flag `isEvolution`. PLATINUM și GOLD au pragul `≥0.90` la fel — fără diferențiere. |
| Comportament așteptat | Per FORMULAS_QUICK_REFERENCE.md §C37 Ceremony Tier: PLATINUM `≥0.90` apare doar dacă `isEvolution = true`; altfel GOLD. Tier să se calculeze cu signature-ul `CeremonyTier(score, isEvolution)`. |
| Fix necesar | Schimbă signature `CeremonyTier(score float64, isEvolution bool) string`; query `evolution_sprints` JOIN sprint_id; pasează flag-ul. |

---

### DEV-17 — `GET /achievements` query referă tabel `achievements` ne-definit

| Câmp | Valoare |
|------|---------|
| Severitate | HIGH |
| Scenariu afectat | TS-07 |
| Componentă Framework | C37/C38 badges |
| Fișier | `backend/internal/api/handlers/achievements.go:24-29` |
| Comportament actual | Query `SELECT id, type, title, description, earned_at FROM achievements`. Tabelul `achievements` nu apare în 001_schema.sql sau `ensureGoalsTables`. user-workflow.md menționează `achievement_badges` ca nume canonic (cu coloanele `badge_type`, `awarded_at`). Nume diferite — doc vs cod. |
| Comportament așteptat | Standardizează pe `achievement_badges` cu coloanele documentate. |
| Fix necesar | Migrare nouă pentru `achievement_badges`, sau ALTER renaming `achievements` → `achievement_badges`. Aliniază query-ul. |

---

### DEV-18 — Trigger SRM L1 folosește drift-criteriu, nu 5-zile-inactive (SA-3 parțial)

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | TS-04 |
| Componentă Framework | C26 vs C32 spec |
| Fișier | `backend/internal/scheduler/scheduler.go:240-287` (jobCheckDailyProgress) |
| Comportament actual | Inserează `srm_events` (level='L1', event_type='DRIFT_CRITICAL') doar când ultimele 3 valori drift sunt `< -0.15` (`engine.IsDriftCritical`). Logica este **C26 Drift Engine**. Dar TS-04 specifică trigger pe **5 zile consecutive cu 0 task-uri MAIN completate** (C32). Mecanismul există, dar criteriul diferă. |
| Comportament așteptat | Adaugă (sau înlocuiește cu) detecție pe `daily_tasks` — 5 zile consecutive cu zero MAIN done → `srm_events (level='L1', event_type='STAGNATION')`. |
| Fix necesar | În `jobCheckDailyProgress` (sau extinde `jobCheckStagnation`): JOIN `daily_tasks WHERE task_type='MAIN' AND status='DONE'` pe ultimele 5 zile; dacă count = 0 → INSERT `srm_events L1`. |

---

### DEV-19 — `jobCheckStagnation` populează doar `stagnation_events`, nu și `srm_events`

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | TS-04 |
| Componentă Framework | C27, C32 |
| Fișier | `backend/internal/scheduler/scheduler.go:538-582` |
| Comportament actual | Job inserează în `stagnation_events` cu `days_inactive=5`. Nu inserează corespondent în `srm_events`. `GetSRMStatus` returnează `NONE` chiar și după detectare stagnation. |
| Comportament așteptat | După `INSERT stagnation_events`, INSERT și în `srm_events (level='L1', event_type='STAGNATION_5D')`. Idempotent (`ON CONFLICT DO NOTHING` pe (go_id, day)). |
| Fix necesar | Adaugă INSERT `srm_events` în loop. |

---

### DEV-20 — Sprint creation hardcodat 30 zile, ignoră goal end_date

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | Goal Creation §2.2 |
| Componentă Framework | C5 |
| Fișier | `backend/internal/api/handlers/goals.go:309-310` |
| Comportament actual | `sprintEnd := sprintStart.AddDate(0, 0, 30)` — sprint 1 este forțat la 30 zile, nu se face cap la `goal.end_date`. Pentru un goal cu durată < 30 zile, sprint-ul depășește deadline-ul. |
| Comportament așteptat | Per user-workflow.md §2.2 punct 5: `Sprint 1 end = start_date + 30 days (capped at end_date)`. |
| Fix necesar | `if sprintEnd.After(endDate) { sprintEnd = endDate }`. |

---

### DEV-21 — Onboarding hardcodează durata 90 de zile

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | TS-01 |
| Componentă Framework | C4 |
| Fișier | `frontend/app/app/onboarding/page.tsx:145-149` |
| Comportament actual | `end.setDate(end.getDate() + 90)` — toate GO-urile create prin onboarding au `end_date = today + 90 days`, fără posibilitate de selecție. |
| Comportament așteptat | Per docs §1.3: user creează GO cu `start_date`, `end_date`, max 365 zile (C4). Onboarding să permită selecție durată sau cel puțin 3 preset-uri (30/90/180/365 zile). |
| Fix necesar | Adaugă step de selecție durată în onboarding (sau extinde wizard). |

---

### DEV-22 — `SuggestGOCategory` fără context timeout (TS-11 risc)

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | TS-11 |
| Componentă Framework | Graceful degradation §8.5 |
| Fișier | `backend/internal/api/handlers/goals.go:88-89` |
| Comportament actual | `result := h.ai.SuggestGOCategory(req.Title, req.Description)` — apel sincron, fără `context.WithTimeout`. Dacă AI client are timeout intern de 12s (per ai.go), poate dura 12s; doc cere 2s hard timeout. |
| Comportament așteptat | TS-11: răspuns < 2s chiar dacă upstream Anthropic e nedisponibil. |
| Fix necesar | Wrap apelul cu `context.WithTimeout(c.Context(), 2*time.Second)`; pe timeout → fallback la `fallbackCategory()`. |

---

### DEV-23 — `AnalyzeGO` timeout 8s (peste budget 2s pentru endpoints AI)

| Câmp | Valoare |
|------|---------|
| Severitate | MEDIUM |
| Scenariu afectat | TS-11 (indirect) |
| Componentă Framework | Graceful degradation |
| Fișier | `backend/internal/api/handlers/goals.go:32-33` |
| Comportament actual | `context.WithTimeout(c.Context(), 8*time.Second)`. Pentru un endpoint pe path-ul critic onboarding, 8s blochează UX. |
| Comportament așteptat | Endpoint-uri AI să aibă 2-3s budget pentru a permite fallback rapid. |
| Fix necesar | Reducere la 3s + fallback rule-based imediat. |

---

### DEV-24 — Cron `*/90` rezolvat (SA-7 RESOLVED — confirm)

| Câmp | Valoare |
|------|---------|
| Severitate | LOW (confirmare) |
| Scenariu afectat | TS-04 |
| Componentă Framework | C28 |
| Fișier | `backend/internal/scheduler/scheduler.go:66` |
| Comportament actual | `s.cron.AddFunc("0 2 * * 0", s.jobRecalibrateRelevance)` — cron valid (Sunday 02:00 UTC). |
| Comportament așteptat | Cron valid, săptămânal. ✅ |
| Fix necesar | Niciun fix — confirmat rezolvat. |

---

### DEV-25 — Streak compute potențial off-by-one

| Câmp | Valoare |
|------|---------|
| Severitate | LOW |
| Scenariu afectat | TS-01 punct 9 (streak) |
| Componentă Framework | C30 |
| Fișier | `backend/internal/api/handlers/today.go:200-230` |
| Comportament actual | `checkDate := tomorrow`. Dacă astăzi nu există DONE-uri, primul rând returnat va fi de ieri sau mai vechi → bucla nu va incrementa streak chiar dacă au existat 7 zile consecutive de DONE până ieri. Documentația nu clarifică dacă streak include doar zilele care se sfârșesc înaintea zilei curente. |
| Comportament așteptat | Streak să tolereze "azi nu încă bifat" (start checkDate la today, nu tomorrow). |
| Fix necesar | Schimbă `checkDate` la `today` și ajustează loop-ul pentru a permite primul match să fie azi sau ieri. Adaugă unit test. |

---

### DEV-26 — `ListAchievements` returnează array gol direct (vs envelope)

| Câmp | Valoare |
|------|---------|
| Severitate | LOW |
| Scenariu afectat | TS-07 frontend |
| Componentă Framework | API consistency |
| Fișier | `backend/internal/api/handlers/achievements.go:44` |
| Comportament actual | Returnează `[]` direct (Fiber JSON encode al slice-ului). Alte endpoints (ex. `ListGoals`) returnează `{"goals": [...]}` envelope. Inconsistent. |
| Comportament așteptat | `{"achievements": [...]}` envelope, conform user-workflow.md §5.3. |
| Fix necesar | `return c.JSON(fiber.Map{"achievements": items})`. Verifică frontend. |

---

### DEV-27 — `GetCeremony` returnează `null` pe goals fără ceremony (vs 404)

| Câmp | Valoare |
|------|---------|
| Severitate | LOW |
| Scenariu afectat | TS-07 |
| Componentă Framework | API contract |
| Fișier | `backend/internal/api/handlers/achievements.go:73-74` |
| Comportament actual | Pe `pgx.ErrNoRows` returnează `c.JSON(nil)` → body literal `null` cu status 200. user-workflow.md §1.7 punct 4 spune că pe miss ceremony, body ar trebui să fie un obiect cu `viewed=null` sau 404. |
| Comportament așteptat | Decis: 200 + `{"ceremony": null}` envelope, sau 404 dacă lipsește. |
| Fix necesar | Aliniere cu contractul ales. |

---

### DEV-28 — Behavior model API field vs DB column inconsistency

| Câmp | Valoare |
|------|---------|
| Severitate | LOW |
| Scenariu afectat | Goal Creation |
| Componentă Framework | C2 |
| Fișier | `backend/internal/api/handlers/goals.go:224, 274, 302` |
| Comportament actual | API expune `dominant_behavior_model` (input) și `behavior_model` (în Goal detail response). DB column este `behavior_model`. Nume duplu pentru același câmp confunduiește clienți. |
| Comportament așteptat | Un singur nume — `behavior_model` peste tot (sau `dominant_behavior_model` peste tot). |
| Fix necesar | Standardizează pe un nume; documentează în `docs/integrations.md`. |

---

## Devieri documentate anterior (confirmare)

| SA | Descriere | Status confirmat | Detalii |
|----|-----------|------------------|---------|
| SA-1 | `growth_trajectories` nepopulate | ✅ Confirmat (DEV-05, DEV-06) | Nici scheduler-ul nu apelează `fn_compute_growth_trajectory()`, nici endpoint-ul nu are fallback. |
| SA-2 | `fn_award_achievement_if_earned()` lipsă | ✅ Confirmat (DEV-07) | Niciun job nu apelează funcția. `GET /achievements` rămâne empty. Adițional: tabelul referit (`achievements`) nu match-uiește numele canonic `achievement_badges` (DEV-17). |
| SA-3 | `srm_events` nepopulate la L1 | ⚠️ Parțial (DEV-18, DEV-19) | `srm_events` SE populează din `jobCheckDailyProgress` dar pe criteriu DRIFT_CRITICAL, nu pe 5-zile-inactive. `jobCheckStagnation` nu inserează în `srm_events`. TS-04 nu va trece curat. |
| SA-4 | `ConfirmSRML2` fără `CreateContextAdjustment` | ✅ REZOLVAT | `srm.go:60-63` inserează `context_adjustments` cu `type='ENERGY_LOW'`. Verifică pe runtime că tabelul este creat și INSERT-ul reușește. |
| SA-5 | `SRMWarning.tsx` fără buton L2 confirm | ✅ REZOLVAT | `SRMWarning.tsx:66-73` are buton "Confirmare — Reduc intensitatea" care apelează `POST /srm/confirm-l2/:goalId` și pe success ascunde banner-ul (`setSrm(null)`). |
| SA-6 | `ApplySRMFallback` e `// TODO` | ✅ Confirmat (DEV-08) | `jobCheckSRMTimeouts` doar loghează; nu mutează state. |
| SA-7 | Cron `*/90` invalid | ✅ REZOLVAT (DEV-24) | `0 2 * * 0` — Sunday 02:00 UTC, cron expression validă. |

---

## Plan de fix recomandat

Sesiunile de fix sunt grupate după dependențe (DB schema → engine → scheduler → handlers → frontend) și prioritate. Fiecare sesiune este dimensionată pentru 45-60 min, conform CLAUDE.md §4.

| Sesiune | Devieri acoperite | Fișiere afectate | Durată est. |
|---------|------------------|-----------------|------------|
| **FIX-01 — Schema reconciliation** | DEV-02, DEV-03, DEV-04, DEV-17 | `backend/migrations/002_*.sql` (nou), `backend/internal/db/db.go`, `001_schema.sql` | 60 min |
| **FIX-02 — API opacity hardening** | DEV-01 | `backend/internal/api/handlers/achievements.go` | 15 min (rapid, single field) |
| **FIX-03 — Visualization fallback (SA-1 + Day 1)** | DEV-05, DEV-06 | `backend/internal/scheduler/scheduler.go`, `backend/internal/api/handlers/goals.go`, `backend/internal/engine/growth.go` (nou `GenerateProgressVisualization`) | 60 min |
| **FIX-04 — Achievement awards (SA-2)** | DEV-07, DEV-17 | `backend/internal/scheduler/scheduler.go`, `backend/internal/api/handlers/achievements.go` | 45 min |
| **FIX-05 — SRM L1 stagnation trigger (SA-3)** | DEV-18, DEV-19 | `backend/internal/scheduler/scheduler.go` | 45 min |
| **FIX-06 — SRM L3 trajectory freeze + ALI** | DEV-12, DEV-15, DEV-16 | `backend/internal/api/handlers/srm.go`, `backend/internal/engine/srm.go`, `backend/internal/engine/growth.go` | 60 min |
| **FIX-07 — SRM timeouts apply state (SA-6)** | DEV-08 | `backend/internal/scheduler/scheduler.go`, `backend/internal/engine/srm.go` | 45 min |
| **FIX-08 — Context energy DB write** | DEV-13, DEV-14 | `backend/internal/api/handlers/today.go` | 30 min |
| **FIX-09 — Sprint duration cap + onboarding** | DEV-20, DEV-21 | `backend/internal/api/handlers/goals.go`, `frontend/app/app/onboarding/page.tsx` | 45 min |
| **FIX-10 — AI timeouts hardening (TS-11)** | DEV-22, DEV-23 | `backend/internal/api/handlers/goals.go` | 30 min |
| **FIX-11 — Engine restructure & evolution detection** | DEV-09, DEV-11, DEV-16 | `backend/internal/engine/*.go` (potențial split) + `scheduler.go` | 60 min |
| **FIX-12 — Weekly ALI computation** | DEV-10 | `backend/internal/scheduler/scheduler.go`, `backend/internal/engine/growth.go` | 60 min |
| **FIX-13 — Cleanup minor** | DEV-25, DEV-26, DEV-27, DEV-28 | Multiple handlers | 30 min |

**Total estimat:** ~10-12 ore de muncă (echivalent 12 sesiuni de 45-60 min). Recomandare ordine:

1. **Phase A (CRITICAL)**: FIX-01, FIX-02, FIX-03, FIX-04, FIX-05, FIX-07 — restabilește contractul de bază + API opacity.
2. **Phase B (HIGH)**: FIX-06, FIX-08, FIX-09, FIX-11 — restabilește C32-C38 funcționalitate.
3. **Phase C (MEDIUM/LOW)**: FIX-10, FIX-12, FIX-13 — hardening și cleanup.

---

## Note finale

**Fișiere ne-citite (în afara scope):** `backend/internal/api/handlers/admin.go`, `dashboard.go`, `profile.go`, `handlers.go`, `backend/internal/ai/ai.go`, `backend/internal/email/email.go` — nu au fost auditate. Auditul s-a concentrat pe path-urile menționate explicit în Task 4.x al sesiunii AUDIT-01.

**Tabele lipsă verificate explicit pe filesystem:**
- `srm_events`, `context_adjustments`, `stagnation_events`, `ceremonies`, `sprint_results`, `go_metrics`, `achievements`, `growth_trajectories`, `evolution_sprints`, `achievement_badges`, `completion_ceremonies`.

Niciuna nu apare în `001_schema.sql` sau în `db.go::ensureGoalsTables`. Conform regulii CLAUDE.md §4 (variabile/resurse server), aceste tabele există probabil pe DB-ul de producție. Recomandare: re-export schemă DB de pe server în repo ca `migrations/002_runtime_baseline.sql` înainte de orice fix structural.

**Bug-ul `level5_growth.go:85` (FROM goals vs FROM global_objectives):** fișierul `level5_growth.go` **nu există** în codebase. Bug-ul descris în task-ul AUDIT-01 §4.2 era probabil deja eliminat în refactor-ul F0.1 (eliminare engine v10.x). DEV-06 acoperă echivalentul în structura curentă.
