# docs/testing/test-plan.md — NuviaX Test Plan

> Version: 10.5.0 | Last updated: 2026-03-30

---

## Overview

This directory contains the complete NuviaX testing documentation, split into modular sections.

A feature that is not covered by a passing test scenario does not exist.

### Structure

```
docs/testing/
├── test-plan.md          ← this file (overview + how to run)
├── flows/
│   ├── goal-flow.md      ← User journey, goal creation, daily execution
│   ├── srm-flow.md       ← SRM L1–L3 trigger, confirm, escalation
│   ├── achievements.md   ← Achievement award, ceremony, badge storage
│   └── visualization.md  ← Trajectory, charts, heatmap, progress bar
└── scenarios/
    ├── critical.md       ← TS-01–TS-12 + critical checkpoints (8.1–8.6)
    └── regression.md     ← SA-1–SA-7 fix mapping + post-fix checklist
```

---

## How to Run Tests

### Before implementing any feature

1. Open the relevant flow file under `flows/`
2. Identify the related test scenario in `scenarios/critical.md` (TS-xx)
3. Confirm the implementation satisfies the inputs, DB changes, API response, and frontend behavior defined in that flow

### After implementing any feature

State explicitly which test scenarios should now pass:

```
# Example commit message or session close note:
# "SA-4 fix complete — TS-05 should pass: L2 confirm creates ENERGY_LOW
#  context adjustment; next-day task count reduced."
```

Do not close a session without naming the test scenarios the implementation covers.

### Quick Lookup

| Task type | File to read |
|---|---|
| New feature | `flows/<relevant-flow>.md` → nearest TS-xx in `scenarios/critical.md` |
| Bug fix | `scenarios/critical.md` (TS-xx that reproduces the bug) |
| Scheduler job | `flows/achievements.md` or `flows/visualization.md` |
| SRM change | `flows/srm-flow.md` → TS-04, TS-05, TS-06 |
| API response change | `scenarios/critical.md` (8.1–8.2 opaque API rules → TS-12) |
| Auth/security change | `scenarios/critical.md` (8.3–8.6) |
| Sprint 3.1 system alignment fixes | `scenarios/regression.md` |

### Hard Rules

```
NEVER implement a feature without first reading its flow file
NEVER mark a task complete if its TS-xx scenario would still fail
NEVER assume a feature works — identify the exact DB change and API response
NEVER ship a fix for SA-1 through SA-7 without running its mapped test scenario
```

---

## Current Known Gaps (Sprint 3.1)

| Gap | Status | Verified By |
|---|---|---|
| SA-1: `growth_trajectories` not populated | ❌ NOT IMPLEMENTED | TS-03, TS-08 |
| SA-2: `fn_award_achievement_if_earned()` never called | ❌ NOT IMPLEMENTED | TS-07 |
| SA-3: SRM L1 auto-trigger not wired | ❌ NOT IMPLEMENTED | TS-04 |
| SA-4: SRM L2 confirm does not reduce task intensity | ❌ NOT IMPLEMENTED | TS-05 |
| SA-5: `SRMWarning.tsx` missing L2 confirm button | ❌ NOT IMPLEMENTED | TS-05 (frontend) |
| SA-6: `jobCheckSRMTimeouts` fallback not applied | ❌ NOT IMPLEMENTED | TS-06 |
| SA-7: `jobRecalibrateRelevance` invalid cron (`*/90`) | ❌ NOT IMPLEMENTED | TS-04 (indirect) |

Full fix details and scenario mappings: `scenarios/regression.md`
