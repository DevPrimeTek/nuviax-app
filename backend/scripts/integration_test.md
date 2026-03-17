# Integration Test Flow — NUViaX Framework REV 5.6

## Prerequisites

| Service | Command | Port |
|---------|---------|------|
| PostgreSQL | Running with `nuviax_dev` DB | 5432 |
| Backend | `go run cmd/server/main.go` | 8080 |
| Frontend | `cd frontend/app && npm run dev` | 3000 |

Ensure all migrations are applied:
```bash
psql -U postgres -d nuviax_dev -f backend/scripts/verify_db.sql
```

---

## Test Sequence

### 1. Authentication
- Navigate to `http://localhost:3000/auth/login`
- Register a new user (if needed) at `/auth/register`
- Verify redirect to `/app/dashboard`
- Verify JWT cookie `nv_access` is set

### 2. Create a Goal
- Click **Obiectiv nou** on dashboard
- Fill: name, description, start date, end date (≥ 14 days)
- Submit and verify goal appears in dashboard active list
- Verify `SRMWarning` renders without error (should be invisible at `srm_level=NONE`)

### 3. Generate Daily Tasks
- Navigate to `/app/today`
- If no tasks appear, trigger scheduler manually:
  ```sql
  -- psql nuviax_dev
  SELECT generate_daily_tasks_for_all();
  ```
- Verify 1–3 tasks are shown

### 4. Complete Tasks & Consistency
- Mark tasks as completed
- Verify progress bar on dashboard updates
- Repeat over multiple days (or seed data) to build streak

### 5. Complete a Sprint → Ceremony
- Complete enough tasks to finish sprint day count
- Trigger ceremony job manually if not waiting for 01:05 UTC:
  ```sql
  UPDATE sprints SET end_date = NOW() - INTERVAL '1 day'
  WHERE goal_id = '<your-goal-id>' AND status = 'ACTIVE';
  ```
  Then re-run job or call engine directly.
- Verify ceremony is created in `completion_ceremonies`
- Reload dashboard → `DashboardClientLayer` polls `/ceremonies/unviewed`
- Verify **CeremonyModal** appears with tier (BRONZE → PLATINUM)
- Click Close → verify `mark_ceremony_viewed` is called (row `viewed_at` set)

### 6. Achievements Page
- Navigate to `/app/achievements`
- Verify badge grid renders (locked badges show as dimmed)
- Complete milestones to unlock badges
- Verify unlocked badges show with color and description

### 7. Progress Visualization
- Navigate to `/app/goals/<id>`
- Click **Progres** tab
- Verify recharts LineChart and BarChart render
- Verify trajectory table shows data points
- If no trajectory data exists yet, verify fallback snapshot is used

### 8. SRM Warning Flow
- Force an SRM event (or seed directly):
  ```sql
  INSERT INTO srm_events (id, goal_id, triggered_at, srm_level, reason, status)
  VALUES (gen_random_uuid(), '<goal-id>', NOW(), 'L1', 'test', 'ACTIVE');
  ```
- Reload goal detail or dashboard
- Verify **SRMWarning** renders yellow (L1), orange (L2), or red (L3)
- For L3: verify **Confirm Resetare** button appears and POST works

### 9. Evolution Detection
- Verify two consecutive sprints exist with scores
- Trigger evolution detection or wait for 01:00 UTC job
- Check `evolution_sprints` table for inserted row
- Verify GOLD/PLATINUM tier ceremony if evolution detected

### 10. Scheduler Jobs Verification
```bash
# Check cron job registration in server logs at startup:
# Expected: "All jobs scheduled" total_jobs=14
grep "total_jobs" /var/log/nuviax-server.log
```

---

## Expected Results

| Test | Expected |
|------|----------|
| Auth flow | JWT set, redirect works |
| Goal creation | Goal in DB + dashboard |
| Daily tasks | 1–3 tasks generated |
| Task completion | Progress score updates |
| Sprint ceremony | Modal appears, tier correct |
| Achievements | Grid renders, unlock works |
| Progress charts | Charts render with data |
| SRM warning | Color + level correct |
| L3 confirm | POST 200, SRM resolved |
| Evolution sprint | Row in evolution_sprints |
| Scheduler jobs | 14 jobs at startup |

All tests should complete with no console errors in browser or server logs.
