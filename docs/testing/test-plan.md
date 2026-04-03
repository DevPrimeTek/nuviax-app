# docs/testing/test-plan.md — NuviaX Master Test Plan (Unit + Integration)

> Version: 11.1.0  
> Last updated: 2026-04-03

---

## 1) Obiectiv

**Prerequisite pentru orice task de testare:** pornește din `CLAUDE.md` pentru a selecta exact fișierele relevante și a evita context inutil.

Acest document definește planul oficial de testare pentru validarea:
1. comportamentului curent (as-built),
2. aliniamentului cu Framework Rev 5.6,
3. criteriilor de release.

---

## 2) Tipuri de teste obligatorii

## A. Unit tests (backend)
Scop: validare logică izolată pentru engine/scheduler/helpers.

Comandă principală:
```bash
cd backend && go test ./internal/engine/... -v
```

Comandă validare pachete compilabile + formatare:
```bash
bash backend/scripts/test_all.sh
```

## B. Integration tests (API + DB + scheduler)
Scop: validare fluxuri end-to-end pe endpoint-uri și persistență.

Smoke API:
```bash
TOKEN=<jwt> bash backend/scripts/test_api.sh http://localhost:8080/api/v1
```

Scenarii detaliate:
- `docs/testing/scenarios/critical.md`
- `docs/testing/scenarios/regression.md`

Admin sanity check (după login ca admin):
```bash
curl -i -H "Authorization: Bearer <token_admin>" http://localhost:8080/api/v1/admin/stats
```

## C. Framework stress tests
Scop: validarea regulilor avansate Rev 5.6 (anti-abuz, sezonalitate, SRM ierarhic).

Set minim obligatoriu:
- T1 Canonical BM integrity
- T2 Seasonal pause continuity
- T3 SRM hierarchy conflict
- T4 Regression immediate signal
- T5 Temporal validity abuse
- T6 Opaque API security

Referință: `docs/framework_100_percent_implementation_playbook.md`.

---

## 3) Test matrix (ce blochează release-ul)

| Arie | Test minim | Blocker release dacă pică? |
|---|---|---|
| Engine scoring | Unit tests engine | DA |
| SRM flow | TS-04, TS-05, TS-06 | DA |
| Goals + daily loop | TS-01, TS-02 | DA |
| Visualization | TS-03, TS-08 | DA |
| Achievements/ceremonies | TS-07 | DA |
| API opacity/security | TS-12 + security checks | DA |
| Framework stress T1..T6 | Stress suite completă | DA (pentru milestone sign-off) |

---

## 4) Mapping la programul M1–M4

| Milestone | Coverage minimă de test |
|---|---|
| M1 (BM + SRM single-active) | Unit BM validation + SRM integration + T1 + T3 |
| M2 (Sezonalitate) | Seasonal integration tests + T2 |
| M3 (Regression + A3) | Regression integration + temporal validity tests + T4 + T5 |
| M4 (Verification) | Full regression suite + T6 + document checks |

---

## 5) Reguli de execuție pentru echipă

1. Nu închizi task arhitectural fără test mapat explicit.
2. Nu marchezi release ready dacă lipsește cel puțin un test blocker din matrix.
3. Orice modificare la scoring/SRM necesită:
   - minim 1 unit test nou/actualizat,
   - minim 1 integration scenario executat.
4. Orice contradicție docs-vs-code descoperită în testare se tratează ca defect de prioritate înaltă.

---

## 6) Evidență rezultate test

La finalul fiecărui PR, include obligatoriu:
- comenzi rulate,
- rezultat (pass/fail/warn),
- scenarii TS/Tx acoperite,
- ce a rămas neacoperit și de ce.

Format recomandat:

```md
✅ go test ./internal/engine/... -v
✅ bash backend/scripts/test_all.sh
⚠️ TOKEN=<jwt> bash backend/scripts/test_api.sh ... (TOKEN indisponibil în mediu CI local)
Covered: TS-04, TS-05, T3
Not covered: T2 (depinde de execution_windows, încă neimplementat)
```
