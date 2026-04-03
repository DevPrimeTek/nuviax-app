# ROADMAP.md — NuviaX Delivery Roadmap (Framework Rev 5.6)

> Versiune: 11.0.0  
> Ultima actualizare: 2026-04-03

---

## Snapshot

| Arie | Status curent | Acțiune necesară |
|---|---|---|
| Core platform (auth, goals, today, dashboard) | ✅ Stabil | Întreținere + monitorizare |
| Framework compliance C1–C40 | ⚠️ Parțial | Program de aliniere pe 4 milestones |
| Docs principale | ⚠️ În curs de sincronizare | Un singur "source of truth" |
| Testare Unit/Integration | ⚠️ Bună, dar incomplet mapată pe Rev 5.6 | Extindere suită + gating strict |

---

## Milestone Roadmap

## M1 — Structural Alignment (P0)
**Perioadă:** Sprint 1

- Behavior Model canonic (`CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE`)
- SRM single-active-level enforcement
- Actualizare contracte API/UI pentru cele două schimbări

**Rezultat așteptat:** dispar conflictele semantice și stările SRM ambigue.

---

## M2 — Seasonal Engine Alignment (P0)
**Perioadă:** Sprint 2

- Introducere `execution_windows`
- Introducere `SEASONAL_PAUSE`
- Ajustare continuity/GORI conform Rev 5.6

**Rezultat așteptat:** GO sezoniere evaluate corect, fără penalizări false.

---

## M3 — Runtime Integrity (P1)
**Perioadă:** Sprint 3

- Regression pipeline integrat end-to-end
- Temporal validity A3 implementat formal
- Semnalizare UX clară la regression/SRM

**Rezultat așteptat:** scoruri robuste, rezistente la abuz și mai predictibile.

---

## M4 — Verification & Governance (P2)
**Perioadă:** Sprint 4

- Actualizare completă documentație "as-built"
- Compliance matrix C1..C40
- Finalizare test suites Unit + Integration + Stress

**Rezultat așteptat:** release gate obiectiv pentru "100% lucrativ".

---

## Backlog după M4

- Monetizare (Stripe)
- Export PDF recap
- PWA + notificări push
- Advanced analytics

> Aceste inițiative intră în execuție doar după închiderea completă a M1–M4.

---

## KPIs de roadmap

1. % componente C1..C40 marcate `Implemented`.
2. Pass rate Unit tests pe module engine/scheduler/API.
3. Pass rate Integration tests pe fluxuri user critice.
4. Număr deviații docs-vs-code = 0.
5. Timp mediu de închidere incident SRM/regression.

