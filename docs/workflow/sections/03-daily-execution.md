## 3. Daily Execution Flow

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

### 3.1 Input (User Action)

- User opens `/today`
- User sets energy level (low / normal / high)
- User completes tasks
- User optionally adds personal tasks (max 2/day)

### 3.2 Backend Processing (`handlers.go` → `GetTodayTasks`, `CompleteTask`, `SetEnergy`)

**Loading today's tasks (`GET /today`):**
1. Redis cache checked first (`cache.GetTodayTasks`) — returns immediately on hit
2. On miss: `db.GetTodayTasks()` for today's date
3. If DB returns empty: `engine.GenerateDailyTasks()` called on-demand (fallback to scheduler)
4. Tasks split into `MAIN` (sprint-generated) and `PERSONAL` (user-added)
5. Streak days, current checkpoint (status `IN_PROGRESS`), day-in-sprint number all fetched
6. Response cached in Redis

**Setting energy (`POST /context/energy`):**
1. Frontend label normalized: `mid → normal`, `hi → high`
2. `normal` energy → no DB action, returns `200` immediately
3. `low` / `high` → `db.CreateContextAdjustment()` with `adjType = AdjEnergyLow / AdjEnergyHigh`, active from today to tomorrow
4. Today's Redis task cache invalidated → next load regenerates with adjusted intensity

**Completing a task (`POST /today/complete/:id`):**
1. `db.CompleteTask()` — sets `completed = TRUE`, records timestamp
2. Both today-tasks and dashboard Redis caches invalidated
3. No immediate score recalculation — score computed on-demand via `engine.ComputeGoalScore()`

**Adding a personal task (`POST /today/personal`):**
1. Max 2 personal tasks/day enforced via `db.CountPersonalTasksToday()`
2. Active sprint resolved from first ACTIVE goal
3. `db.CreateTask()` inserts with `task_type = PERSONAL`
4. Today-tasks cache invalidated

### 3.3 DB Changes

| Table | Operation |
|---|---|
| `daily_tasks` | UPDATE `completed = TRUE` on task completion |
| `context_adjustments` | INSERT energy adjustment (low/high only) |
| `daily_tasks` | INSERT personal task |

### 3.4 API Response

**`GET /today`:**
```json
{
  "date": "2026-03-30T00:00:00Z",
  "goal_name": "...",
  "day_number": 14,
  "main_tasks": [ { "id": "...", "text": "...", "completed": false } ],
  "personal_tasks": [ ... ],
  "done_count": 2,
  "total_count": 5,
  "streak_days": 7,
  "checkpoint": { "id": "...", "name": "Progres: ...", "status": "IN_PROGRESS" }
}
```

**`POST /today/complete/:id`:** `200` + `{ "message": "Activitate bifată." }`

**`POST /context/energy` (low/high):** `200` + `{ "message": "...Activitățile de mâine vor fi adaptate." }`

### 3.5 Frontend Behavior

1. `/today` renders main tasks + personal tasks in separate lists
2. Checkbox tap → optimistic UI update → `POST /today/complete/:id`
3. Energy selector: 3 options (low / normal / high); selection calls `POST /context/energy`; no page reload
4. "Add personal task" button disabled after 2 tasks/day (client enforced + server enforced)
5. Streak counter and checkpoint banner update on each page load (no real-time push)

---

