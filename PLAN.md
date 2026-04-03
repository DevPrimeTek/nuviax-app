# PLAN.md — NuviaX Framework Rev 5.6 Full Alignment Plan

> Versiune: 11.0.0  
> Actualizat: 2026-04-03  
> Scop: implementare 100% a logicii Framework Rev 5.6 + pregătire pentru execuție Unit/Integration tests

---

## 1) Obiectiv executiv

Acest plan înlocuiește planul vechi bazat pe "session prompts" și devine documentul principal de implementare arhitecturală.

**Ținta finală:**
1. Framework C1–C40 mapat explicit în cod și DB.
2. Workflow documentat "as-built" (fără contradicții față de cod).
3. Test plan executabil pentru Unit + Integration + Stress tests.

---

## 2) Scope (ce livrăm)

### Livrabil 1 — Plan de implementare (acest fișier)
- backlog tehnic pe workstreams
- ordonare P0/P1/P2
- Definition of Done pe fiecare workstream

### Livrabil 2 — Roadmap actualizat
- repere calendaristice reale
- dependențe între echipe
- criterii de intrare/ieșire pe faze

### Livrabil 3 — CLAUDE.md actualizat
- context unic pentru echipa tehnică
- reguli de execuție și validare

### Livrabil 4 — README.md actualizat
- starea reală a produsului
- comenzi de rulare și testare

### Livrabil 5 — Test plan actualizat
- mapare clară între framework rules și teste Unit/Integration
- suită obligatorie pentru validare release

---

## 3) Workstreams arhitecturale (cu impact direct în cod)

## WS-A (P0) — Behavior Model Canonic
**Problemă:** taxonomie diferită între framework și implementare.

**Implementare:**
- migrare DB pentru BM canonic (`CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE`)
- API validation + payload update
- UI onboarding/goals update

**Done când:** toate endpoint-urile și UI folosesc BM canonic.

---

## WS-B (P0) — Sezonalitate Rev 5.6
**Problemă:** lipsă `SEASONAL_PAUSE` + `execution_windows` în fluxul runtime.

**Implementare:**
- extindere schema sprint status
- modelare ferestre sezoniere
- actualizare scheduler + engine pentru freeze/neutral consistency
- update formule continuitate/GORI

**Done când:** GO sezoniere nu mai sunt penalizate artificial în perioade inactive.

---

## WS-C (P0) — SRM ierarhic strict (single-active-level)
**Problemă:** lipsă enforcement explicit pentru un singur SRM activ.

**Implementare:**
- revocare automată niveluri inferioare la escaladare
- contract API clar (`active_event_id`, `current_level`)
- frontend afișează strict nivelul activ

**Done când:** pentru fiecare GO există exact un SRM activ în orice moment.

---

## WS-D (P1) — Regression pipeline end-to-end
**Problemă:** detecția există, dar nu este orchestrat complet în job-ul zilnic.

**Implementare:**
- integrare în fluxul de calcul scor
- trigger SRM imediat la regres confirmat
- semnal vizual utilizator + audit DB complet

**Done când:** regression event + semnal SRM apar în aceeași zi.

---

## WS-E (P1) — Temporal validity (A3)
**Problemă:** reguli >48h / bulk completions trebuie formalizate cap-coadă.

**Implementare:**
- câmpuri + reguli score explicite
- separare `Progress` vs `Consistency`
- teste anti-abuz

**Done când:** comportamentul A3 este testabil și determinist.

---

## WS-F (P2) — Documentație "as-built"
**Problemă:** documente istorice au afirmații depășite.

**Implementare:**
- sincronizare `docs/user-workflow.md`, `ROADMAP.md`, `README.md`, `CLAUDE.md`, `docs/testing/*`
- compliance matrix C1..C40

**Done când:** toate documentele descriu exact comportamentul curent.

---

## 4) Milestones

| Milestone | Workstreams | Target | Exit criteria |
|---|---|---|---|
| M1 | WS-A + WS-C | 1 sprint | BM canonic + SRM single-active-level funcționale |
| M2 | WS-B | +1 sprint | Sezonalitate Rev 5.6 activă în DB/engine/scheduler |
| M3 | WS-D + WS-E | +1 sprint | Regression + A3 complet testabile |
| M4 | WS-F | +1 sprint | Toate docurile + test plan sincronizate |

---

## 5) Gating pentru release

Un release se aprobă doar dacă toate condițiile sunt adevărate:
1. Unit tests green pentru componentele modificate.
2. Integration tests green pentru fluxurile critice.
3. Stress tests T1..T6 / ST-* green.
4. Documentația principală actualizată în același PR.

---

## 6) Referințe operative

- `docs/framework_100_percent_implementation_playbook.md`
- `docs/framework_workflow_deviations_stress_test.md`
- `docs/testing/test-plan.md`
- `docs/user-workflow.md`

