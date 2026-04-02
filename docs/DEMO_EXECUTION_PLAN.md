# DEMO_EXECUTION_PLAN.md — NuviaX: Plan Execuție Demo

> **Obiectiv:** Flux complet funcțional register → goal → sprint → today → progress
> **Termen:** Înainte de demo (sesiuni consecutive, fiecare max 45 minute)
> **Versiune curentă:** 10.5.0 | **Sprint activ:** Sprint 3.1 (System Alignment)
> **Blocker principal:** SA-1–SA-7 neimplementate — fără acestea fluxul de progres e broken

---

## De ce este demo-ul blocat acum

Situația actuală (v10.5.0) are codul scris și deploy-ul funcțional, dar 7 gap-uri de sistem
neimplementate fac ca fluxul vizibil pentru un utilizator să fie incomplet:

| Simptom vizibil în demo | Cauza reală | Fix |
|---|---|---|
| Graficul de progres e gol / null | `growth_trajectories` nu se populează | SA-1 + CE-1 |
| SRM nu se activează după zile ratate | `srm_events` nu primește L1 | SA-3 |
| Butonul "Confirmare L2" lipsește din UI | `SRMWarning.tsx` incomplet | SA-5 |
| Confirmarea L2 nu reduce task-urile | `CreateContextAdjustment` nelipsit | SA-4 |
| Achievements returneaza array gol | `fn_award_achievement_if_earned` nechelat | SA-2 |
| Goal cu L3 neconfirmat rămâne blocat | `ApplySRMFallback` are TODO neimplementat | SA-6 |
| L2 chaos_index nu se evaluează corect | Cron `*/90` invalid, job nu rulează | SA-7 |

---

## Regulă de sesiune (OBLIGATORIU)

**Fiecare sesiune Claude Code respectă:**

1. **Durată maximă: 45 minute** — după 45 minute, indiferent de starea task-ului:
   - Commite ce e gata cu sufix `[WIP]` dacă e incomplet
   - Deschide PR cu descrierea stării curente
   - Închide sesiunea — continuă în sesiunea următoare

2. **La STARTUL fiecărei sesiuni noi:**
   ```
   Read CLAUDE.md. Current version: X.X.X. Active branch: Y.
   Continuing from previous session: [ce s-a terminat].
   Task: [exact ce urmează].
   ```

3. **`/compact` după fiecare subtask major** — nu la finalul sesiunii

4. **Fișiere citite: MAXIM 3** — nu explorare globală

5. **Un singur task per sesiune** — commit → sesiune nouă

---

## Sesiunea 1 — SA-7 + CE-1

**Model:** Sonnet | **Durată estimată:** 20 minute | **Risc:** Minim (2 linii de cod)

**De ce primul:**
SA-7 repară cronul invalid care face ca `jobRecalibrateRelevance` să nu ruleze niciodată.
Fără acest job, `chaos_index` nu se evaluează și L2 nu se poate triggera automat.
CE-1 repară bug-ul de tabel care face ca trajectories să returneze `null` în ziua 1.
Ambele sunt modificări de 1 linie — cel mai mic risc, cel mai mare impact de deblocare.

**Ce citești (max 3 fișiere):**
1. `backend/internal/scheduler/scheduler.go` — caută `jobRecalibrateRelevance`
2. `backend/internal/engine/level5_growth.go` liniile 80–95 — caută `FROM goals`

**Ce schimbi:**

Fix 1 — SA-7 în `scheduler.go`:
```
Găsește: "0 2 */90 * *"
Înlocuiește cu: "0 2 * * 0"
```

Fix 2 — CE-1 în `level5_growth.go` (aprox. linia 85):
```
Găsește: FROM goals   (în query-ul de fallback pentru live snapshot)
Înlocuiește cu: FROM global_objectives
```

**Verificare (grep obligatoriu înainte de commit):**
```bash
grep "*/90" backend/internal/scheduler/scheduler.go        # → GOL (nimic)
grep "FROM goals" backend/internal/engine/level5_growth.go  # → GOL (nimic)
```

**Commit:**
```
fix: SA-7 cron expression weekly + CE-1 trajectory table name (level5_growth.go)
```

**Actualizări după commit:**
- `CLAUDE.md` secțiunea 4: SA-7 ✅, CE-1 ✅
- `docs/testing/test-plan.md`: SA-7 → ✅, CE-1 → ✅

**Scenarii deblocate:** TS-04 (indirect), TS-08

**→ ÎNCHIDE SESIUNEA. Deschide sesiunea 2.**

---

## Sesiunea 2 — SA-1 (growth_trajectories populate)

**Model:** Sonnet | **Durată estimată:** 30 minute | **Risc:** Mic

**De ce al doilea:**
Fără SA-1, `GET /goals/:id/visualize` returnează `trajectory: null` sau array gol —
graficul de progres e gol în demo. Acesta e cel mai vizibil bug pentru un demo.
Fix-ul adaugă un singur apel de funcție în scheduler + o funcție DB de 5 linii.

**Contextul tehnic:**
`growth_trajectories` table există (migration aplicată).
`fn_compute_growth_trajectory(goal_id, date)` este funcție PostgreSQL — există deja.
Problema: nu e apelată niciodată din Go.
`jobComputeDailyScore` rulează la 23:50 UTC, apelează `db.UpsertGoalScore()` dar nu
apelează `fn_compute_growth_trajectory` după.

**Ce citești (max 3 fișiere):**
1. `backend/internal/scheduler/scheduler.go` — caută `jobComputeDailyScore`
2. `backend/internal/db/queries.go` — verifică dacă `ComputeGrowthTrajectory` există

**Ce adaugi:**

În `scheduler.go`, în `jobComputeDailyScore`, după apelul `db.UpsertGoalScore(...)`:
```go
if err := db.ComputeGrowthTrajectory(ctx, goalID, time.Now()); err != nil {
    log.Printf("[scheduler] growth trajectory failed for %s: %v", goalID, err)
    // non-fatal: continuă loop-ul
}
```

În `db/queries.go`, adaugă funcția nouă:
```go
func (db *DB) ComputeGrowthTrajectory(ctx context.Context, goalID uuid.UUID, date time.Time) error {
    _, err := db.Pool.Exec(ctx,
        "SELECT fn_compute_growth_trajectory($1, $2)",
        goalID, date.UTC().Truncate(24*time.Hour),
    )
    return err
}
```

**Verificare:**
```bash
grep "ComputeGrowthTrajectory" backend/internal/scheduler/scheduler.go  # → 1 match
grep "fn_compute_growth_trajectory" backend/internal/db/queries.go       # → 1 match
```

**Commit:**
```
feat: SA-1 wire fn_compute_growth_trajectory in jobComputeDailyScore
```

**Actualizări după commit:**
- `CLAUDE.md` secțiunea 4: SA-1 ✅
- `docs/testing/test-plan.md`: SA-1 → ✅

**Scenarii deblocate:** TS-03 (după 2 rulări scheduler), TS-08 (ziua 1 snapshot)

**→ ÎNCHIDE SESIUNEA. Deschide sesiunea 3.**

---

## Sesiunea 3 — SA-3 (SRM L1 auto-trigger)

**Model:** Sonnet | **Durată estimată:** 25 minute | **Risc:** Mediu

**De ce al treilea:**
Fără SA-3, monitorizarea progresului nu funcționează. Utilizatorul poate rata 5 zile
consecutive și sistemul nu reacționează — SRM rămâne "NONE". Aceasta rupe logica
de bază a Framework-ului de monitorizare a obiectivelor.

**Contextul tehnic:**
`jobDetectStagnation` (23:58 UTC) populează `stagnation_events` corect când
`inactive_days >= 5`. Problema: NU scrie în `srm_events` cu `srm_level = 'L1'`.
Deci `GET /srm/status` returnează mereu "NONE" chiar și după 5 zile ratate.

**Ce citești (max 3 fișiere):**
1. `backend/internal/scheduler/scheduler.go` — caută `jobDetectStagnation`
2. `backend/internal/engine/srm.go` — caută structura `SRMEvent` și `InsertSRMEvent`
3. `backend/internal/db/queries.go` — verifică `GetActiveSRMLevel` sau similar

**Ce adaugi în `scheduler.go`, în `jobDetectStagnation`, după detectarea `inactive_days >= 5`:**

```go
// Verifică dacă există deja un event activ L1/L2/L3 pentru acest goal
existingLevel, _ := db.GetActiveSRMLevel(ctx, goalID)
if existingLevel == "" {
    srmEvent := engine.SRMEvent{
        GoID:          goalID,
        SRMLevel:      "L1",
        TriggerReason: "stagnation_5days",
        TriggeredAt:   time.Now(),
    }
    if err := db.InsertSRMEvent(ctx, srmEvent); err != nil {
        log.Printf("[scheduler] SRM L1 insert failed for %s: %v", goalID, err)
    }
}
```

Dacă `GetActiveSRMLevel` nu există în `db/queries.go`, adaugă:
```go
func (db *DB) GetActiveSRMLevel(ctx context.Context, goalID uuid.UUID) (string, error) {
    var level string
    err := db.Pool.QueryRow(ctx,
        `SELECT srm_level FROM srm_events
         WHERE go_id = $1 AND confirmed_at IS NULL
         ORDER BY triggered_at DESC LIMIT 1`,
        goalID,
    ).Scan(&level)
    if errors.Is(err, pgx.ErrNoRows) { return "", nil }
    return level, err
}
```

**Verificare:**
```bash
grep "stagnation_5days" backend/internal/scheduler/scheduler.go  # → 1 match
grep "GetActiveSRMLevel" backend/internal/scheduler/scheduler.go  # → 1 match
```

**Commit:**
```
feat: SA-3 SRM L1 auto-trigger in jobDetectStagnation after 5 inactive days
```

**Actualizări după commit:**
- `CLAUDE.md` secțiunea 4: SA-3 ✅
- `docs/testing/test-plan.md`: SA-3 → ✅

**Scenarii deblocate:** TS-04

**→ ÎNCHIDE SESIUNEA. Deschide sesiunea 4.**

---

## Sesiunea 4 — SA-4 + SA-5 (SRM L2 reduce intensitate + buton frontend)

**Model:** Sonnet | **Durată estimată:** 30 minute | **Risc:** Mediu

**De ce al patrulea:**
SA-4 + SA-5 sunt strâns legate — backend-ul crează ajustarea, frontend-ul expune butonul.
Fără SA-4, confirmarea L2 nu are efect real (task count rămâne același a doua zi).
Fără SA-5, utilizatorul nu are cum să confirme L2 din UI — butonul pur și simplu lipsește.
Ambele sunt necesare pentru ca TS-05 să treacă.

**Contextul tehnic:**
SA-4: `ConfirmSRML2()` în `srm.go` stampilează `confirmed_at` corect dar NU apelează
`CreateContextAdjustment()`. Deci task count a doua zi e neschimbat.
SA-5: `SRMWarning.tsx` afișează banner-ul L2 dar nu are buton de confirmare.
Ruta `POST /srm/confirm-l2/:goalId` există — nu e legată în frontend.

**Ce citești (max 3 fișiere):**
1. `backend/internal/engine/srm.go` — caută `ConfirmSRML2` (aprox. 30 linii)
2. `backend/internal/db/queries.go` — caută `CreateContextAdjustment` și signatura
3. `frontend/app/components/SRMWarning.tsx` — caută secțiunea condițională L2

**SA-4 — în `srm.go`, în `ConfirmSRML2()`, după stampilarea `confirmed_at`:**
```go
tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
adj := db.ContextAdjustment{
    GoID:      goalID,
    AdjType:   db.AdjEnergyLow,
    StartDate: tomorrow,
    EndDate:   tomorrow.AddDate(0, 0, 7),
    Source:    "srm_l2_confirm",
}
if err := dbConn.CreateContextAdjustment(ctx, adj); err != nil {
    return fmt.Errorf("ConfirmSRML2 context adjustment: %w", err)
}
```

**SA-5 — în `SRMWarning.tsx`, în blocul condițional L2, adaugă după textul mesajului:**
```tsx
{status.srm_level === 'L2' && !status.confirmed && (
  <button
    onClick={handleConfirmL2}
    className="mt-3 px-4 py-2 rounded-lg bg-amber-500 text-white text-sm font-medium hover:bg-amber-600 transition-colors"
  >
    Confirmare — Reduc intensitatea
  </button>
)}
```

Adaugă funcția `handleConfirmL2` înainte de `return`:
```tsx
const handleConfirmL2 = async () => {
  await fetch(`/api/v1/srm/confirm-l2/${goalId}`, {
    method: 'POST',
    headers: authHeaders(),
  });
  refreshSRMStatus();
};
```

**Verificare:**
```bash
grep "AdjEnergyLow" backend/internal/engine/srm.go          # → 1 match în ConfirmSRML2
grep "confirm-l2" frontend/app/components/SRMWarning.tsx     # → 1 match
```

**Commit:**
```
feat: SA-4 ConfirmSRML2 creates context adjustment + SA-5 SRMWarning L2 confirm button
```

**Actualizări după commit:**
- `CLAUDE.md` secțiunea 4: SA-4 ✅, SA-5 ✅
- `docs/testing/test-plan.md`: SA-4 → ✅, SA-5 → ✅

**Scenarii deblocate:** TS-05 (backend + frontend)

**→ ÎNCHIDE SESIUNEA. Deschide sesiunea 5.**

---

## Sesiunea 5 — SA-2 + SA-6 (achievements + SRM fallback)

**Model:** Sonnet | **Durată estimată:** 35 minute | **Risc:** Mediu

**De ce al cincilea:**
SA-2 și SA-6 nu blochează demo-ul de bază (register → goal → today → progress funcționează
după sesiunile 1–4), dar fără ele demo-ul e incomplet:
- SA-2: `GET /achievements` returnează `[]` → achievements nu sunt niciodată acordate
- SA-6: goaluri cu L3 neconfirmat rămân blocate indefinit (TODO neimplementat)

**Contextul tehnic:**
SA-2: `jobGenerateCeremonies` apelează `GenerateCompletionCeremony()` dar niciodată
`fn_award_achievement_if_earned(user_id, sprint_id)`.
SA-6: `jobCheckSRMTimeouts` are comentariu `// TODO: engine.ApplySRMFallback(...)` —
fallback-ul nu e implementat, goalurile cu L3 rămân blocate.

**Ce citești (max 3 fișiere):**
1. `backend/internal/scheduler/scheduler.go` — caută `jobGenerateCeremonies` și `jobCheckSRMTimeouts`
2. `backend/internal/db/queries.go` — caută sau verifică `AwardAchievementIfEarned`
3. `backend/internal/engine/srm.go` — caută sau verifică `ComputeSRMFallback`

**SA-2 — în `scheduler.go`, în `jobGenerateCeremonies`, după `GenerateCompletionCeremony()` reușit:**
```go
if err := db.AwardAchievementIfEarned(ctx, userID, sprintID); err != nil {
    log.Printf("[scheduler] achievement award failed for sprint %s: %v", sprintID, err)
    // non-fatal
}
```

Adaugă în `db/queries.go`:
```go
func (db *DB) AwardAchievementIfEarned(ctx context.Context, userID, sprintID uuid.UUID) error {
    _, err := db.Pool.Exec(ctx,
        "SELECT fn_award_achievement_if_earned($1, $2)",
        userID, sprintID,
    )
    return err
}
```

**SA-6 — în `scheduler.go`, în `jobCheckSRMTimeouts`, înlocuiește TODO cu:**
```go
fallbackLevel := engine.ComputeSRMFallback(currentLevel, hoursUnconfirmed)
srmEvent := engine.SRMEvent{
    GoID:          goalID,
    SRMLevel:      fallbackLevel,
    TriggerReason: "srm_timeout_fallback",
    TriggeredAt:   time.Now(),
}
if err := db.InsertSRMEvent(ctx, srmEvent); err != nil {
    log.Printf("[scheduler] SRM fallback insert failed: %v", err)
}
```

Dacă `ComputeSRMFallback` nu există, adaugă în `srm.go`:
```go
func ComputeSRMFallback(current string, hoursUnconfirmed float64) string {
    if current == "L3" && hoursUnconfirmed > 72 {
        return "L1"
    }
    return current
}
```

**Verificare:**
```bash
grep "AwardAchievementIfEarned" backend/internal/scheduler/scheduler.go  # → 1 match
grep "TODO.*ApplySRMFallback" backend/internal/scheduler/scheduler.go     # → GOLS (nimic)
```

**Commit:**
```
feat: SA-2 fn_award_achievement_if_earned + SA-6 ApplySRMFallback implementation
```

**Actualizări după commit:**
- `CLAUDE.md` secțiunea 4: SA-2 ✅, SA-6 ✅
- `docs/testing/test-plan.md`: SA-2 → ✅, SA-6 → ✅
- Rulează checklist complet din `docs/testing/scenarios/regression.md`

**Scenarii deblocate:** TS-06, TS-07

**→ DEMO READY după această sesiune. Sesiunea 6 (CI/CD) nu blochează demo-ul.**

---

## Sesiunea 6 — CI/CD Tests (GitHub Actions)

**Model:** Sonnet | **Durată estimată:** 45 minute | **Risc:** Mic (nu atinge codul de producție)

**Notă:** Această sesiune nu este necesară pentru demo. O faci după ce demo-ul e confirmat.

**Ce implementezi:**

### Layer 1: Unit tests (`.github/workflows/test-unit.yml`)
Trigger: push orice branch, PR to main.
```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: backend/go.mod
- run: go test ./internal/engine/... -v -count=1
- run: go test ./internal/db/... -v -count=1 -tags=!integration
- run: go vet ./...
```

### Layer 2: Integration tests (`.github/workflows/test-integration.yml`)
Trigger: push main, PR to main.
Services block: `postgres:16-alpine`, `redis:7-alpine`.
```yaml
- run: go test ./... -tags=integration -v -count=1
  env:
    TEST_DB_URL: postgres://postgres:postgres@localhost:5432/nuviax_test
    TEST_REDIS_URL: redis://localhost:6379
```

### Layer 3: E2E Playwright — SEPARAT (Sprint 4)
Nu implementa în această sesiune. Adaugă job placeholder cu `if: false`.

### Teste Go de creat dacă lipsesc:

`backend/internal/engine/engine_test.go`:
```go
func TestEngineNeverExposesDrift(t *testing.T) {
    // Apelează engine cu date mock
    // Asertează că răspunsul nu conține: drift, chaos_index, weights
}
```

`backend/internal/engine/srm_test.go`:
```go
func TestComputeSRMFallbackL3After72h(t *testing.T) {
    result := ComputeSRMFallback("L3", 73)
    if result != "L1" {
        t.Errorf("expected L1, got %s", result)
    }
}
```

**Commit:**
```
feat: CI/CD GitHub Actions unit + integration tests (E2E Playwright planned Sprint 4)
```

**Actualizări după commit:**
- `README.md`: adaugă badge-uri CI pentru ambele workflow-uri

**→ ÎNCHIDE SESIUNEA.**

---

## Checklist final înainte de demo

Rulează după sesiunile 1–5 sunt commituite și deployate:

```bash
# Health check VPS
curl https://api.nuviax.app/health
# → {"status":"ok","db":true,"redis":true}

# Verificare trajectories (după cel puțin o rulare scheduler sau manual)
curl -H "Authorization: Bearer TOKEN" https://api.nuviax.app/api/v1/goals/GOAL_ID/visualize
# → trajectory NU trebuie să fie null

# Verificare SRM status
curl -H "Authorization: Bearer TOKEN" https://api.nuviax.app/api/v1/srm/status/GOAL_ID
# → srm_level prezent (NONE / L1 / L2 / L3)

# Verificare opacitate API (TS-12)
# Răspunsurile NU trebuie să conțină: drift, chaos_index, weights, thresholds
```

**Post-fix validation checklist completă:** `docs/testing/scenarios/regression.md`

---

## Sumar timp total estimat

| Sesiune | Task | Timp estimat |
|---------|------|-------------|
| 1 | SA-7 + CE-1 | 20 min |
| 2 | SA-1 | 30 min |
| 3 | SA-3 | 25 min |
| 4 | SA-4 + SA-5 | 30 min |
| 5 | SA-2 + SA-6 | 35 min |
| 6 | CI/CD | 45 min (după demo) |
| **Total demo** | **Sesiunile 1–5** | **~2h 20min** |

---

*Creat: 2026-04-02 — v10.5.0*
*Adaugă în repo la: `docs/DEMO_EXECUTION_PLAN.md`*
*Referință prompturi complete: `PROMPTS.md` (sesiunile 1–6)*
