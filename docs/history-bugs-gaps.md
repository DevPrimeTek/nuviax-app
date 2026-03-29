# Istoric Bug-uri și Gap-uri — NuviaX

> Arhivă completă: bug-uri rezolvate și gap-uri din stress test.
> NU necesită citire la fiecare sesiune — consultă doar dacă ai nevoie de context istoric.

---

## Bug-uri Rezolvate (v10.2.0 — toate rezolvate)

| # | Locație | Problema | Status |
|---|---------|---------|--------|
| B-2 | `handlers.go AnalyzeGO` | Analiză GO fără AI | ✅ Rezolvat — Claude Haiku cu fallback |
| B-3 | `handlers.go` | `days_left` folosea `goal.EndDate` în loc de `sprint.EndDate` | ✅ Rezolvat |
| B-4 | `level1_structural.go` | Task generation static fără AI | ✅ Rezolvat — Claude Haiku cu fallback |
| B-5 | `today/page.tsx` | Energy nu se salva — endpoint greșit | ✅ Rezolvat |
| B-6 | `today/page.tsx` | Fără formular sarcini personale | ✅ Rezolvat |
| B-7 | `handlers.go` + `goals/page.tsx` | Goals returna array plat; acum returnează `{goals:[], waiting:[]}` | ✅ Rezolvat |
| B-8 | `server.go` + `handlers.go` | `GET /recap/current` lipsea — 404 | ✅ Rezolvat |
| B-9 | `settings/page.tsx` | Parolă/export neconectate la API | ✅ Rezolvat |
| B-10 | `profile/page.tsx` | Fără upload foto profil | ✅ Rezolvat |
| B-11 | `styles/globals.css` | `--ul`, `--ug`, `--l2g`, `--ff-h` lipseau; Light theme lipsă | ✅ Rezolvat |

---

## Gap-uri Stress Test

### ✅ Rezolvate (P0 — Critical, rezolvate în v10.1)

| Gap # | Problema | Fișier modificat |
|-------|---------|-----------------|
| #8 | ALI total vs per-GO ambiguitate | `srm.go` |
| #13 | ALI current vs projected ambiguitate | `srm.go` |
| #14 | Pauza retroactivă (max 48h) | `handlers.go`, `db/queries.go`, migration 007 |
| #15 | Regression event (valoare sub sprint start) | `level2_execution.go`, migration 007 |
| #20 | Drift loop paradox în Stabilization Mode | `level5_growth.go`, migration 007 |

### ✅ Rezolvate (P1 — Medium, 10/12 în v10.4)

| Gap # | ID | Locație | Status |
|-------|----|---------|--------|
| G-1 | Tie-breaking BM | `engine.go` | ✅ |
| G-2 | Domain Benchmark Library | `level1_structural.go` | ✅ |
| G-3 | Clamp [0,1] clarificare Drift exclus | `engine.go` | ✅ |
| G-4 | Priority Balance auto-rezoluție | `engine.go` | ✅ |
| G-5 | Stagnation detection 5 zile → `stagnation_events` | `level2_execution.go`, `scheduler.go` | ✅ |
| G-6 | Velocity Control ALI_projected > 1.15 | `level2_execution.go`, `engine.go` | ✅ |
| G-7 | Reactivation Protocol 7 zile stabilitate | `level4_regulatory.go`, `scheduler.go` | ✅ |
| G-8 | Sprint score formula 40/25/25/10 completă | `level2_execution.go` | ✅ |
| G-9 | Annual Relevance Recalibration + Chaos storage | `scheduler.go` | ✅ |
| G-10 | Future Vault max 3 active | `level4_regulatory.go`, `handlers.go` | ✅ |
| G-12 | SRM flow complet L2/L3 confirmare | `srm.go`, `server.go` | ✅ |

### ⏳ Nerezolvate

| Gap # | ID | Problema | Sprint |
|-------|----|---------|--------|
| G-11 | Behavior Model dominance | EVOLVE override GO hibride; necesită câmp `dominant_behavior_model VARCHAR(20)` pe `global_objectives` + migration 011 | Sprint 3 |

---

## Valori de referință stress test

Folosite dacă nu sunt specificate altfel în cod:

| Parametru | Valoare |
|-----------|---------|
| Retroactive Pause window | 48h |
| Reactivation stability | 7 zile |
| Stagnation threshold | 5 zile consecutive fără progres |
| ESB Stagnation threshold | 10 zile (extins) |
| Chaos Index L2 threshold | 0.40 |
| ALI Ambition Buffer | 1.0 – 1.15 |
| Evolution delta threshold | ≥ 5% |

---

## Changelog Rapid

| Versiune | Data | Descriere |
|---------|------|-----------|
| v10.4.1 | 2026-03-26 | Admin page standalone; setup_admin.sh; docs restructurate |
| v10.4.0 | 2026-03-26 | P1 Gaps G-1—G-10, G-12 (10/12); migration 010 |
| v10.3.1 | 2026-03-26 | Admin fix: is_admin în nav; cleanup fișiere duplicate |
| v10.3.0 | 2026-03-25 | Email Resend: welcome + sprint complet + forgot/reset parolă |
| v10.2.0 | 2026-03-24 | Fix toate bug-urile B-2—B-11; AI integration Claude Haiku; upload avatar |
| v10.1.0 | 2026-03-20 | P0 Gaps: regresie, pauză retroactivă, drift paradox, ALI |
| v10.0.0 | 2026-03-16 | Restructurare completă, deploy automat CI/CD |

> Changelog complet: `CHANGES.md`

---

*Actualizat: v10.4.1 — 2026-03-26*
