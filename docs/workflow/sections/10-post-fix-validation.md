## 10. Post-Fix Validation Checklist

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

Run after all SA-1 through SA-7 fixes are deployed. All items must pass before Sprint 3.1 is closed.

- [ ] **TS-03** — `GET /goals/:id/visualize` returns ≥2 trajectory entries after 2 scheduler runs
- [ ] **TS-04** — `GET /srm/status/:goalId` returns `srm_level: "L1"` after 5 consecutive missed days
- [ ] **TS-05** — SRM L2 banner has confirm button; after confirm, next-day task count is reduced
- [ ] **TS-06** — L3 unconfirmed >N hours → fallback applied; goal not stuck indefinitely
- [ ] **TS-07** — Sprint close → `GET /achievements` returns ≥1 badge; `GET /ceremonies/latest` returns ceremony with correct tier
- [ ] **TS-08** — Day 1 visualization returns exactly 1 entry; `trajectory` never null or empty
- [ ] **TS-12** — Zero internal fields (`drift`, `chaos_index`, `weights`, thresholds) in any API response
- [ ] **8.3** — All protected routes return `401` without Authorization header
- [ ] **8.4** — Admin routes return `404` for non-admin users
- [ ] **8.6** — `POST /auth/forgot-password` returns `200` for both known and unknown emails
- [ ] **Cron fix (SA-7)** — `jobRecalibrateRelevance` runs without error; verify via scheduler logs
