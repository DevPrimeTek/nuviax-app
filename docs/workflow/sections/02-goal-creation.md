## 2. Goal Creation Flow

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

### 2.1 Input (User Action)

- User submits: `name`, `start_date`, `end_date` (YYYY-MM-DD), optional `description`, optional `dominant_behavior_model`, optional `waiting_list: true`
- Source: `/onboarding` wizard or `/goals` create form

### 2.2 Backend Processing (`handlers.go` → `CreateGoal`)

1. **Date validation:** `end_date > start_date`; max duration 365 days — returns `400` on failure
2. **G-10 capacity check:** `engine.ValidateGoalActivation()` — max 3 active goals
   - If at capacity and `waiting_list: false` → goal auto-routed to `WAITING` (`vaulted: true` in response)
   - If `waiting_list: true` → status set to `WAITING` directly
3. **DB insert:** `db.CreateGoal()` → row in `global_objectives`
4. **G-11 behavior model:** if `dominant_behavior_model` provided → `db.SetGoalBehaviorModel()` updates `global_objectives.dominant_behavior_model`
5. **Sprint creation (ACTIVE goals only):**
   - Sprint 1 end = `start_date + 30 days` (capped at `end_date`)
   - `db.CreateSprint()` → row in `sprints`
   - 3 checkpoints created: `"Fundament: <name>"`, `"Progres: <name>"`, `"Consolidare: <name>"`
   - `engine.GenerateDailyTasks()` called immediately (no waiting for midnight scheduler)
6. **Cache:** `cache.InvalidateDashboard()` clears Redis dashboard cache for user

### 2.3 DB Changes

| Table | Operation |
|---|---|
| `global_objectives` | INSERT — new goal row with status ACTIVE / WAITING |
| `global_objectives` | UPDATE `dominant_behavior_model` (if G-11 provided) |
| `sprints` | INSERT Sprint 1 (ACTIVE goals only) |
| `checkpoints` | INSERT × 3 (ACTIVE goals only) |
| `daily_tasks` | INSERT tasks for today (ACTIVE goals only, via engine) |

### 2.4 API Response

**ACTIVE goal (standard):** `201` + goal object

**Auto-vaulted to WAITING (G-10):**
```json
{
  "goal": { ... },
  "message": "Ai deja 3 obiective active. Obiectivul a fost adăugat în Vault-ul viitor...",
  "status": "WAITING",
  "vaulted": true
}
```

**Validation error:** `400` / `422` + `{ "error": "..." }`

### 2.5 Frontend Behavior

1. On success (ACTIVE): redirect → `/today`; dashboard cache busted → fresh load
2. On `vaulted: true`: show vault notice banner; stay on `/goals`
3. On `422`: display inline error; no redirect
4. AI category suggestion (`POST /goals/suggest-category`) runs debounced on title input before form submit — result pre-fills category pill; no block if timeout (2s fallback)

---

