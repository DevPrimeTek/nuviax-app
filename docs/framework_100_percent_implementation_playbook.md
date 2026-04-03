# NuviaX Playbook — Cum ajungem la implementare 100% a Framework-ului (explicat simplu)

> Audiență: Product Owner / Founder / PM (fără background de Software Architect)  
> Data: 2026-04-03

---

## 1) Pe scurt: ce nu e încă 100% aliniat

Dacă simplificăm maxim, sunt 5 zone critice:

1. **Limbajul de Behavior Model nu e același cu Framework-ul**  
   Framework-ul cere: `CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE`, dar aplicația validează `ANALYTIC/STRATEGIC/TACTICAL/REACTIVE`.
2. **Nu există sezonalitate completă în modelul de date**  
   Framework-ul Rev 5.6 cere `SEASONAL_PAUSE` + `execution_windows`, dar DB/statusurile actuale nu le au explicit.
3. **SRM nu are „single source of truth” pe nivel activ**  
   Trebuie un singur nivel activ (L1/L2/L3) în orice moment; acum pot exista în istoric mai multe neînchise.
4. **Regression pipeline există, dar nu e legat cap-coadă în fluxul zilnic**  
   Avem funcție de detectare, dar nu este conectată în job-urile zilnice care calculează scoruri.
5. **Documentația workflow e parțial depășită**  
   Unele secțiuni spun „NOT IMPLEMENTED”, deși codul deja implementează acele părți.

---

## 2) Ce modifici concret și unde (hartă exactă)

## GAP A — Unificare Behavior Model (obligatoriu)

### Unde modifici
- `backend/migrations/011_behavior_model.sql`
- `backend/internal/api/handlers/handlers.go` (request validation create goal)
- `backend/internal/engine/level5_growth.go` (logică de override/evolution)
- `frontend/app/app/onboarding/page.tsx` și `frontend/app/app/goals/page.tsx` (opțiuni UI)
- `docs/user-workflow.md` (contract API actualizat)

### Ce schimbi
- Introduci **BM canonic**: `CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE` ca câmp principal.
- Muți `ANALYTIC/STRATEGIC/TACTICAL/REACTIVE` într-un câmp secundar (ex: `execution_profile`) sau îl elimini dacă nu este necesar.

### Criteriu DONE
- API acceptă și returnează BM canonic.
- UI permite selectarea doar BM canonical.
- Motorul de scor operează explicit pe BM canonical.

---

## GAP B — Sezonalitate Rev 5.6 (`SEASONAL_PAUSE`)

### Unde modifici
- `backend/migrations/001_base_schema.sql` (sau migrare nouă 013): extindere `sprint_status`
- migrare nouă: `execution_windows` (JSONB sau tabel dedicat)
- `backend/internal/scheduler/scheduler.go` (job-uri care generează task-uri / închid sprint-uri)
- `backend/internal/engine/engine.go` + `level5_growth.go` (Continuity/GORI)
- `frontend/app/components/GoalTabs.tsx` și `/goals/[id]/page.tsx` (afișare fereastră activă/inactivă)

### Ce schimbi
- Adaugi status nou sprint: `SEASONAL_PAUSE`.
- Când un GO intră în fereastră inactivă: fără task-uri, expected înghețat, consistency neutru.
- În formulele de continuitate excluzi explicit sprinturile `SEASONAL_PAUSE` și `PAUSED/SUSPENDED` din denominator.

### Criteriu DONE
- GO sezonier nu este penalizat în lunile inactive.
- Reintrarea în fereastră activă reia progresul fără „salturi false” în scor.

---

## GAP C — SRM ierarhic strict (L3 > L2 > L1)

### Unde modifici
- `backend/internal/db/queries.go` (`InsertSRMEvent`, query active level)
- `backend/internal/api/handlers/srm.go` (confirm L2/L3)
- `backend/internal/scheduler/scheduler.go` (trigger L1/L2)
- posibil migrare nouă: index/constraint pentru „un singur eveniment SRM activ/GO”
- `frontend/app/components/SRMWarning.tsx` (afișare strict pe nivelul curent)

### Ce schimbi
- La activare L2/L3: revoci automat nivelurile inferioare active.
- Definești clar: `active_event_id`, `current_level`, `trigger_reason`, `expires_at/timeout`.

### Criteriu DONE
- Pentru orice GO, un singur SRM activ în DB.
- UI și API arată același nivel, fără ambiguități.

---

## GAP D — Regression event integrat complet în runtime

### Unde modifici
- `backend/internal/scheduler/scheduler.go` (jobComputeDailyScore)
- `backend/internal/engine/level2_execution.go` (funcția deja există)
- `backend/internal/api/handlers/visualization.go` sau endpoint dedicat pentru alertă regression
- `frontend/app/components/DashboardClientLayer.tsx` (badge/alert)

### Ce schimbi
- După fiecare calcul de scor zilnic, rulezi `CheckAndRecordRegressionEvent(...)` pe GO eligibile.
- La regression confirmat: trigger SRM L1 imediat + alertă în dashboard.

### Criteriu DONE
- Regression real apare în `regression_events` în aceeași zi.
- Utilizatorul primește semnal clar și acțiune recomandată.

---

## GAP E — Documentație „as-built” reală

### Unde modifici
- `docs/user-workflow.md`
- `docs/testing/test-plan.md`
- `docs/testing/scenarios/*.md`

### Ce schimbi
- Separi clar:
  1) ce e implementat azi,
  2) ce e gap față de Rev 5.6,
  3) ce e roadmap.

### Criteriu DONE
- Echipa QA/Product nu mai urmărește bug-uri deja rezolvate.
- Toate testele sunt mapate la cerințe C1…C40.

---

## 3) Ordinea recomandată (ca să nu blochezi echipa)

1. **Sprint 1 (P0):** Behavior Model + SRM single-level.
2. **Sprint 2 (P0):** Sezonalitate (`SEASONAL_PAUSE` + execution windows).
3. **Sprint 3 (P1):** Regression runtime integration + alerte UI.
4. **Sprint 4 (P2):** Curățare documentație + test matrix final C1…C40.

---

## 4) „Teste sofisticate” care trebuie să treacă (gata de execuție)

## T1 — Canonical BM Integrity Test
- Creezi 5 GO, câte unul per BM canonic.
- Verifici că fiecare trece validarea API și generează scor fără fallback logic.
- **Pass:** niciun GO nu e respins din cauza taxonomiei.

## T2 — Seasonal Pause Continuity Test
- GO cu 2 ferestre active + 1 inactivă.
- Rulezi 3 sprint-uri simulate.
- **Pass:** lunile inactive nu scad artificial continuitatea/GORI.

## T3 — SRM Hierarchy Conflict Test
- Forțezi L1, apoi L2, apoi L3 pe același GO.
- **Pass:** mereu un singur eveniment activ în DB.

## T4 — Regression Immediate Signal Test
- Simulezi regres metric semnificativ într-o zi.
- **Pass:** `regression_event` + SRM L1 în aceeași rulare de job.

## T5 — Temporal Validity Abuse Test
- Simulezi completare la 72h + bulk completions într-un interval foarte scurt.
- **Pass:** Progress și Consistency sunt tratate diferit conform regulii A3.

## T6 — Opaque API Security Test
- Verifici payload-ul tuturor endpoint-urilor user-facing.
- **Pass:** niciun weight/drift intern sensibil nu este expus.

---

## 5) Cum știi că ai ajuns la „100% lucrativ”

Ai ajuns la 100% când toate cele 4 condiții sunt adevărate simultan:

1. **Conformitate structurală:** C1…C40 marcate `Implemented` sau `Out-of-scope` justificat.
2. **Conformitate operațională:** testele T1…T6 trec green în CI.
3. **Conformitate UX:** utilizatorul vede acțiuni clare la SRM/regression fără ambiguități.
4. **Conformitate documentară:** `docs/user-workflow.md` descrie exact ce face sistemul azi.

---

## 6) Mesaj simplu de business

Dacă implementezi exact pașii de mai sus, rezultatul este:
- utilizatorii primesc evaluare corectă (fără penalizare sezonieră falsă),
- SRM devine previzibil și coerent,
- scorurile devin mai greu de „păcălit”,
- QA poate valida obiectiv conformitatea cu Framework Rev 5.6.

Cu alte cuvinte: **mai puține surprize în producție, mai multă încredere în sistem și în rezultate**.

