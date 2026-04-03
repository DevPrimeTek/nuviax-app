## 5. Achievement Flow

### CURRENT SYSTEM (REALITY)

**Scheduler trigger chain:**
- ✅ `jobCloseExpiredSprints` (00:01 UTC) — sets `sprints.status = 'COMPLETED'` correctly
- ✅ `jobDetectEvolutionSprints` (01:00 UTC) — evaluates delta, inserts into `evolution_sprints` (idempotent)
- ✅ G-11 override (`ApplyEvolveOverride()`) — ANALYTIC/TACTICAL/STRATEGIC/REACTIVE thresholds applied
- ✅ `jobGenerateCeremonies` (01:05 UTC) — generates ceremony per closed sprint (idempotent)

**Achievement award:**
- ❌ `fn_award_achievement_if_earned()` exists in migration 006 but NEVER called from Go (SA-2)
- ❌ `achievement_badges` table always empty for real users
- ❌ `GET /achievements` always returns `[]` — badges only appear via direct DB insert

**Ceremony:**
- ✅ `engine.GenerateCompletionCeremony()` assigns tier (BRONZE/SILVER/GOLD/PLATINUM) correctly
- ✅ Stored in `completion_ceremonies`; generated exactly once per sprint (`ON CONFLICT DO NOTHING`)
- ✅ `GET /ceremonies/:goalId` returns ceremony with `viewed` flag
- ✅ `POST /ceremonies/:id/view` marks as seen — modal not re-shown

**Frontend:**
- ✅ `CeremonyModal.tsx` renders on login when `viewed = false`; dismissal works
- ✅ `/achievements` page renders empty grid without error when `achievements: []`
- ✅ `GET /achievements/progress` returns progress bars from `get_achievement_progress()`

---

### TARGET SYSTEM (FRAMEWORK)

**Scheduler trigger chain:**
- ✅ `jobCloseExpiredSprints` → calls `fn_award_achievement_if_earned(user_id, sprint_id)` per closed sprint
- ✅ `jobDetectEvolutionSprints` → calls `fn_award_achievement_if_earned()` when evolution detected
- ✅ `jobGenerateCeremonies` → ceremony + badge award run in sequence after sprint close

**Achievement award:**
- ✅ `fn_award_achievement_if_earned()` wired to scheduler — `achievement_badges` populated automatically
- ✅ `GET /achievements` returns non-empty badge array after sprint close
- ✅ Badge types earned based on sprint performance, streaks, evolution detection

**Ceremony:**
- ✅ Same tier logic (BRONZE/SILVER/GOLD/PLATINUM) — score opaque, never exposed
- ✅ PLATINUM only on `score >= 0.90` AND `isEvolution = true`
- ✅ `CeremonyModal.tsx` displays on next login; one-time view enforced

**Frontend:**
- ✅ `/achievements` badge grid populated after each sprint close
- ✅ Progress bars reflect real advancement toward each badge type
- ✅ `/profile` links to `/achievements`; heatmap and badge system remain separate

---

### 5.1 Achievement Trigger Conditions

Achievements and ceremonies are evaluated **after sprint close only** — not on individual task completion.

**Trigger chain (scheduler, nightly UTC):**

1. `jobCloseExpiredSprints` (00:01 UTC) — sets `sprints.status = 'COMPLETED'` for sprints past their `end_date`
2. `jobDetectEvolutionSprints` (01:00 UTC) — queries sprints completed yesterday; calls `engine.MarkEvolutionSprint()` per sprint
   - Evolution condition: `current_sprint_score - prev_sprint_score >= 0.05` (delta threshold)
   - G-11 override: if `dominant_behavior_model` set, `ApplyEvolveOverride()` applies model-specific thresholds:
     - `ANALYTIC`: requires consistency >= 0.75 in addition to delta
     - `TACTICAL`: delta threshold lowered to 0.02 (more responsive to quick wins)
     - `STRATEGIC`: standard delta threshold (0.05)
     - `REACTIVE`: adaptive threshold based on recent performance volatility
   - Evolution detected → INSERT into `evolution_sprints` (idempotent via `ON CONFLICT sprint_id DO NOTHING`)
3. `jobGenerateCeremonies` (01:05 UTC) — queries sprints completed yesterday with no existing ceremony; calls `engine.GenerateCompletionCeremony()`

**⚠️ NOT IMPLEMENTED — Achievement auto-award (see SA-2, Sprint 3.1):** `fn_award_achievement_if_earned()` exists in migration 006 but is never called from Go. `achievement_badges` is not populated by the scheduler. `GET /achievements` always returns `[]` for real users. Badges only appear if inserted directly via DB.

### 5.2 Ceremony Tiers

Tier assignment in `engine.GenerateCompletionCeremony()` (`level5_growth.go:185`):

| Tier | Condition |
|---|---|
| `BRONZE` | `score < 0.75` (any sprint) |
| `SILVER` | `score >= 0.75` |
| `GOLD` | `score >= 0.90` AND not an evolution sprint |
| `PLATINUM` | `score >= 0.90` AND `isEvolution = true` |

- Score = `engine.ComputeSprintScore()` — opaque value; never exposed raw
- `isEvolution` flag passed from `jobGenerateCeremonies` query via `evolution_sprints` join
- Ceremony stored in `completion_ceremonies`: `sprint_id`, `go_id`, `ceremony_tier`, `viewed = false`
- On conflict (`ON CONFLICT sprint_id DO NOTHING`) — ceremony generated exactly once per sprint

### 5.3 Badge Storage and Award

**Tables involved:**
- `achievement_badges` — awarded badges: `id`, `user_id`, `badge_type`, `go_id`, `sprint_id`, `awarded_at`
- `completion_ceremonies` — sprint ceremonies: `id`, `sprint_id`, `go_id`, `ceremony_tier`, `viewed`, `generated_at`
- `evolution_sprints` — evolution markers: `sprint_id`, `evolution_score`, `delta_performance`, `consistency_weight`

**Read path (`GET /achievements`):**
1. Handler calls `engine.GetUserAchievements(ctx, userID)`
2. Query: `SELECT ... FROM achievement_badges WHERE user_id = $1 ORDER BY awarded_at DESC`
3. Returns `[]models.AchievementBadge` — never null, empty array `[]` on no badges (nil guard in handler)

**Progress path (`GET /achievements/progress`):**
1. Handler calls DB function directly: `SELECT * FROM get_achievement_progress($1)`
2. Returns progress toward each badge type (from migration 006)

**`fn_award_achievement_if_earned()` — when it should be called:**
- After each `jobCloseExpiredSprints` for the closed sprint
- After `jobDetectEvolutionSprints` when evolution is detected
- Currently not called (SA-2 gap) — must be wired to scheduler in Sprint 3.1

### 5.4 Frontend Display

**`/achievements` page (`achievements/page.tsx`):**
1. Fetches `GET /achievements` → renders badge grid from `achievements` array
2. Empty state: `achievements: []` → renders empty grid (no error shown)
3. Fetches `GET /achievements/progress` → renders progress bars per badge type

**Ceremony modal (`CeremonyModal.tsx`):**
1. Dashboard checks `GET /ceremonies/:goalId` on each login per active goal
2. If `viewed = false` → `CeremonyModal` rendered with tier (BRONZE/SILVER/GOLD/PLATINUM) and message
3. User dismisses → `POST /ceremonies/:id/view` → `viewed = true` in DB → modal not shown again
4. Colors/icons vary by tier — defined in `CeremonyModal.tsx` component

**`/profile` page:**
- Does not show achievement history directly; links to `/achievements`
- Shows activity heatmap (`ActivityHeatmap.tsx`) and stats — separate from badge system

---

