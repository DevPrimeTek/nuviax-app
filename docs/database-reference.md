# Database Reference — NuviaX

> Schema completă PostgreSQL, migrări, funcții, triggers.
> Actualizează la fiecare migration nouă.

---

## Statistici curente

**28 tabele + 26 views + 1 materialized view + 10 funcții + 12 triggers**
**Migrări aplicate: 010**

---

## Migrări

| Migrare | Conținut | Tabele noi |
|---------|---------|-----------|
| `001_base_schema.sql` | Core: users, sessions, goals, sprints, tasks, audit_log | 12 |
| `002_layer0_level1.sql` | goal_categories, sprint_configs, goal_metadata, views Layer 0 | 3 |
| `003_level2_execution.sql` | task_executions, daily_metrics, sprint_metrics | 3 |
| `004_level3_adaptive.sql` | behavior_patterns, consistency_snapshots, adaptive_weights | 3 |
| `005_level4_regulatory.sql` | regulatory_events, goal_activation_log, resource_slots | 3 |
| `006_level5_growth.sql` | growth_milestones, achievement_badges, completion_ceremonies, evolution_sprints | 4 |
| `007_admin_fixes.sql` | regression_events, ali_snapshots + cols pe sprints + views admin | — |
| `008_avatar.sql` | `users.avatar_url VARCHAR(500)` | — |
| `009_password_reset.sql` | password_reset_tokens (1h TTL, single-use) | 1 |
| `010_p1_gaps.sql` | srm_events, reactivation_protocols, stagnation_events | 3 |

### Aplicare migrări

```bash
# Toate (idempotent — safe să rulezi de mai multe ori)
docker exec -i nuviax_db psql -U nuviax -d nuviax < backend/migrations/apply_all.sql

# Individual
docker exec -i nuviax_db psql -U nuviax -d nuviax < backend/migrations/010_p1_gaps.sql

# Verificare integritate
docker exec -i nuviax_db psql -U nuviax -d nuviax -f backend/scripts/verify_db.sql
```

**Cerință PostgreSQL:** minimum 16+ pentru `gen_random_uuid()`, JSONB operators, Materialized Views CONCURRENTLY.

---

## Tabele Core (migration 001)

| Tabel | Descriere | Note |
|-------|-----------|------|
| `users` | Utilizatori | `is_admin BOOLEAN` adăugat în migration 007 |
| `user_sessions` | Sesiuni JWT | `token_hash`, `device_fp`, `expires_at` |
| `global_objectives` | Obiective (GO) | `status: ACTIVE/PAUSED/COMPLETED/ARCHIVED/WAITING` |
| `sprints` | Sprint-uri per GO | `expected_pct_frozen`, `frozen_expected_pct` adăugate migration 007 |
| `daily_tasks` | Sarcini zilnice | `MAIN` sau `PERSONAL` |
| `checkpoints` | Milestone-uri per sprint | — |
| `context_adjustments` | Ajustări context | `retroactive BOOL` adăugat migration 007 |
| `audit_log` | Toate evenimentele de securitate | — |

## Tabele importante (migrations 007-010)

| Tabel | Migration | Scop |
|-------|-----------|------|
| `regression_events` | 007 | Detecție valori sub sprint start |
| `ali_snapshots` | 007 | ALI current vs projected per zi |
| `password_reset_tokens` | 009 | Token forgot-password (1h TTL) |
| `srm_events` | 010 | Audit trail complet SRM L1/L2/L3 per obiectiv |
| `reactivation_protocols` | 010 | Tracking 7-day stability per PAUSED GO |
| `stagnation_events` | 010 | Log zile consecutive inactive per GO |

## Tabele Level 5 (migration 006)

| Tabel | Scop |
|-------|------|
| `evolution_sprints` | Evolution Sprint tracking (C37) |
| `completion_ceremonies` | Ceremony data BRONZE/SILVER/GOLD/PLATINUM (C38) |
| `user_achievements` | Achievements permanente (C39) |
| `progress_snapshots` | Cache vizualizări (C40) |

---

## Views importante

```sql
-- Admin
v_admin_platform_stats    -- statistici platformă globale
v_admin_user_list         -- lista utilizatori cu detalii

-- Level 5
evolution_sprints_summary -- summary evolution per goal
latest_ceremonies         -- ultimele ceremonii per goal
unviewed_ceremonies       -- ceremonii nevăzute (notificări)
user_achievement_stats    -- statistici achievements per user

-- Materialized (refresh periodic)
user_progress_overview    -- overview complet per user
```

---

## Funcții PostgreSQL (10)

```sql
check_pause_limit(user_id)             -- verifică limita de pauze
compute_sprint_target_with_compensation() -- 80% rule + compensație
refresh_progress_overview()            -- refresh materialized view
fn_dev_reset_data(admin_id)            -- dev only: reset date utilizator
```

---

## Triggere automate (12)

| Trigger | Tabel | Acțiune |
|---------|-------|---------|
| `auto_update_weight_from_relevance` | global_objectives | Actualizează weight la schimbare relevance |
| `log_relevance_change` | global_objectives | Loghează schimbări relevance |
| `auto_vault_low_relevance` | global_objectives | Auto-vault la relevance < 0.30 |
| `checkpoint_completed_activate_next` | checkpoints | La COMPLETED → activează next UPCOMING |
| `checkpoint_update_detect_regression` | checkpoints | Detectează drift < -0.15 → creează regression_event |
| `task_completion_validate_timestamp` | daily_tasks | Validează timestamp la completare |

---

## Statut verificare (expected)

```
Tables:           28
Regular Views:   ~26
Materialized:     1
Functions:        10
Triggers:         12
```

---

## Environment Variables DB

```env
POSTGRES_HOST=nuviax_db
POSTGRES_PORT=5432
POSTGRES_USER=nuviax
POSTGRES_PASSWORD=<openssl rand -base64 32>
POSTGRES_DB=nuviax

REDIS_HOST=nuviax_redis
REDIS_PORT=6379
REDIS_PASSWORD=<openssl rand -base64 32>
```

---

*Actualizat: v10.4.1 — migration 010 aplicată*
