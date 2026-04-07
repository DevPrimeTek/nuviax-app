# FORMULAS_QUICK_REFERENCE.md — NuviaX Engine Formulas

> Versiune: 1.0.0  
> Actualizat: 2026-04-07  
> Implementat în: `backend/internal/engine/`

---

## GO Validation (C2, C3, C4, C14)

| Regulă | Condiție |
|---|---|
| C14 | `name ≠ ""` și `bm ≠ ""` |
| C2  | `bm ∈ {CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE}` |
| C3  | `activeCount < 3` |
| C4  | `endDate - startDate ≤ 365 days` |

---

## C5 — 30-Day Sprint Expected Progress

```
Expected(day) = day / 30.0
```

`day ∈ [1, 30]`

---

## C24 — Progress Computation

```
Progress = Clamp(completedCheckpoints / totalCheckpoints, 0, 1)
```

Returns 0 when `totalCheckpoints = 0`.

---

## C25 — Execution Variance (Drift)

```
Drift = realProgress - expected
```

Not clamped. Positive = ahead, negative = behind.

---

## C20 + C21 — Sprint Target (80% Rule)

```
SprintTarget = (annualTarget - currentProgress) / sprintsRemaining × 0.80
```

Returns 0 when `sprintsRemaining ≤ 0`.

---

## C37 — Sprint Score

```
SprintScore = Clamp(progress×0.50 + consistency×0.30 + deviation×0.20, 0, 1)
```

---

## C11 — Relevance Scoring

```
Relevance = impact×0.35 + urgency×0.25 + alignment×0.25 + feasibility×0.15
```

---

## C7 + C13 — Priority Weight

| Relevance | Weight |
|---|---|
| `< 0.40` | 1 (Low) |
| `≥ 0.40 and < 0.75` | 2 (Medium) |
| `≥ 0.75` | 3 (High) |

---

## C8 — Priority Balance

```
sum(weights) ≤ 7
```

---

## Score → Grade

| Score | Grade |
|---|---|
| `≥ 0.90` | A+ |
| `≥ 0.80` | A  |
| `≥ 0.65` | B  |
| `≥ 0.45` | C  |
| `< 0.45` | D  |

---

## C26 — Drift Critical (SRM trigger)

```
IsDriftCritical = all(last_3_drift_values < -0.15)
```

Returns false if fewer than 3 values.

---

## C28 — Chaos Index

```
ChaosIndex = Clamp(drift×0.30 + stagnation×0.25 + inconsistency×0.25, 0, 1)
```

Velocity component (weight 0.20) omitted in MVP.

| ChaosIndex | Level  |
|---|---|
| `< 0.30`   | GREEN  |
| `< 0.40`   | YELLOW |
| `< 0.60`   | AMBER  |
| `≥ 0.60`   | RED    |

---

## C33 — SRM Fallback (rule-based)

| Hours since last activity | Level |
|---|---|
| `≥ 168` | PAUSE |
| `≥ 72`  | L1    |
| `≥ 24`  | L2    |
| `< 24`  | (none) |

---

## C38 — GORI (Global Objective Realization Index)

```
GORI = Clamp(avg(sprintScores) × (completed / max(total, 1)), 0, 1)
```

---

## C37 — Ceremony Tier

| SprintScore | Tier     |
|---|---|
| `≥ 0.90`    | PLATINUM |
| `≥ 0.80`    | GOLD     |
| `≥ 0.65`    | SILVER   |
| `< 0.65`    | BRONZE   |

---

## Helpers

```
Clamp(x, min, max) = max(min, min(x, max))
```
