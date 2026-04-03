## 4. SRM Flow (L1–L3)

### CURRENT SYSTEM (REALITY)

**Trigger detection:**
- ❌ L1 auto-trigger NOT IMPLEMENTED (SA-3) — `jobDetectStagnation` writes to `stagnation_events` only; `srm_events` never populated; `GET /srm/status` always returns `NONE` after 5+ inactive days
- ❌ `jobRecalibrateRelevance` cron expression `*/90` is invalid — job never fires; chaos_index threshold for L2 never evaluated (SA-7)

**L1 — Automatic intensity reduction:**
- ❌ NOT IMPLEMENTED — no `srm_events` row created; no task intensity reduction triggered

**L2 — Structural recalibration:**
- ✅ `POST /srm/confirm-l2` stamps `confirmed_at` on `srm_events`
- ❌ `CreateContextAdjustment(AdjEnergyLow)` NOT called — next-day task count unchanged (SA-4)
- ❌ `SRMWarning.tsx` L2 confirm button absent in UI (SA-5)

**L3 — Strategic reset:**
- ✅ `POST /srm/confirm-l3` pauses goal (`status = 'PAUSED'`)
- ✅ `FreezeExpectedTrajectory()` runs — drift loop paradox prevented (GAP #20)
- ✅ `frozen_expected` computed from elapsed time / total duration

**Status endpoint:**
- ✅ `GET /srm/status/:goalId` returns `srm_level`, `ali_current`, `ali_projected`
- ✅ `velocity_control_on: true` when `ALI_projected > 1.15`

---

### TARGET SYSTEM (FRAMEWORK)

**Trigger detection (C32–C34):**
- ✅ L1: after 5 consecutive days with 0 MAIN task completions → `jobCheckDailyProgress` inserts `srm_events (srm_level='L1')` → task intensity auto-reduced
- ✅ L2: `chaos_index ≥ 0.40` evaluated by `jobRecalibrateRelevance` (weekly) → `srm_events (srm_level='L2')` inserted
- ✅ L3: `chaos_index ≥ 0.60` OR unresolved L2 after N days → `srm_events (srm_level='L3')` inserted

**L1 — Automatic intensity reduction (C32):**
- ✅ No user action required
- ✅ Task count reduced automatically via `CreateContextAdjustment(AdjEnergyLow)`
- ✅ Banner: informational only, no confirm button

**L2 — Structural recalibration (C33–C34):**
- ✅ User confirms via `SRMWarning.tsx` confirm button
- ✅ `ConfirmSRML2()` stamps `confirmed_at` AND calls `CreateContextAdjustment(AdjEnergyLow)` → next-day task count reduced
- ✅ Goal remains `ACTIVE`

**L3 — Strategic reset (C35–C36):**
- ✅ Goal paused; trajectory frozen; stabilization mode active
- ✅ Reactivation proposed by scheduler after 7 days
- ✅ Unconfirmed L3 after timeout → `jobCheckSRMTimeouts` applies fallback state (SA-6)

**ALI Velocity Control:**
- ✅ `ALI_projected > 1.15` → `velocity_control_on: true` in status response
- ✅ Ambition buffer zone `1.0–1.15` → warning only, no SRM escalation

---

### 4.1 Input (User Action)

- SRM is **engine-triggered**, not user-initiated
- User confirms L2 or L3 when banner is shown
- Entry points: dashboard SRM warning banner (`SRMWarning.tsx`), goal detail page
- Endpoints: `GET /srm/status/:goalId`, `POST /srm/confirm-l2/:goalId`, `POST /srm/confirm-l3/:goalId`

### 4.2 Backend Processing (`srm.go` + `level4_regulatory.go`)

**SRM Status (`GET /srm/status/:goalId`):**
1. Queries `srm_events` for most recent non-revoked event
2. Returns `srm_level: NONE / L1 / L2 / L3`
3. Computes ALI breakdown (current + projected): `computeALIBreakdown()`
   - `ALI_current` = tasks completed / expected by now
   - `ALI_projected` = if current pace continues to sprint end
   - Ambition buffer zone: `ALI_projected` between 1.0–1.15 → Velocity Control warning
   - `velocity_control_on: true` if `ALI_projected > 1.15`

**L1 — Automatic adjustment (no user action required):** ⚠️ NOT IMPLEMENTED (see SA-3, Sprint 3.1)
- Intended: triggered by `level4_regulatory.go` during scheduler run; task intensity reduced automatically; `srm_events` row inserted with `srm_level = 'L1'`
- Actual: `jobDetectStagnation` inserts into `stagnation_events` only — no `srm_events` row is created; `GET /srm/status` returns `NONE` even after 5+ inactive days

**L2 — Structural recalibration (`POST /srm/confirm-l2/:goalId`):**
1. Verifies access: `db.GetGoalByID()` — returns `404` if not owner
2. `UPDATE srm_events SET confirmed_at = NOW(), confirmed_by = $2` on most recent unconfirmed L2
3. If no active unconfirmed L2 event → `404`
4. ⚠️ Task intensity adjustment NOT IMPLEMENTED (see SA-4, Sprint 3.1) — `ConfirmSRML2` stamps `confirmed_at` only; no `CreateContextAdjustment` call; next-day task count is unchanged
5. Goal status remains `ACTIVE`

**L3 — Strategic reset (`POST /srm/confirm-l3/:goalId`):**
1. Verifies goal ownership
2. `UPDATE global_objectives SET status = 'PAUSED'`
3. `INSERT INTO srm_events` with `srm_level = 'L3'`, trigger_reason = `'user_confirmed_stabilization'`
4. `engine.FreezeExpectedTrajectory(sprint.ID)` — freezes projected trajectory to prevent drift loop paradox (GAP #20)
5. `frozen_expected` percentage computed from elapsed time / total goal duration

### 4.3 DB Changes

| Table | Operation | When |
|---|---|---|
| `srm_events` | INSERT new event | L1/L2/L3 trigger |
| `srm_events` | UPDATE `confirmed_at` | L2 user confirmation |
| `srm_events` | INSERT with `trigger_reason` | L3 user confirmation |
| `global_objectives` | UPDATE `status = 'PAUSED'` | L3 only |
| `sprint_trajectories` | UPDATE frozen flag | L3 only (GAP #20) |

### 4.4 API Response

**`GET /srm/status/:goalId`:**
```json
{
  "goal_id": "...",
  "srm_level": "L2",
  "message": "Ajustare structurală în curs. Am recalibrat obiectivele.",
  "ali": {
    "ali_current": 0.72,
    "ali_projected": 0.68,
    "in_ambition_buffer": false,
    "velocity_control_on": false,
    "goal_breakdown": [ ... ],
    "note": "ali_current = progres realizat până acum. ali_projected = proiecție la finalul sprintului."
  }
}
```

**`POST /srm/confirm-l2`:** `200` + message + `next_step` ⚠️ intensity reduction NOT IMPLEMENTED (SA-4)

**`POST /srm/confirm-l3`:** `200` + `new_status: PAUSED` + `frozen_expected` percentage + `next_step`

### 4.5 Frontend Behavior

1. `SRMWarning.tsx` banner displayed on dashboard when `srm_level ≠ NONE`
2. L1: informational banner only — no action button
3. L2: banner shows "Confirm recalibration" button → calls `POST /srm/confirm-l2`; on success banner dismissed
4. L3: banner shows "Activate stabilization mode" button → calls `POST /srm/confirm-l3`; on success goal card shows `PAUSED` badge; reactivation proposed after 7 days (scheduler)

---

