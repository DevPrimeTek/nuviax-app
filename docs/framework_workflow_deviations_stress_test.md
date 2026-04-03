# NuviaX — Audit de conformitate Framework vs Workflow + Stress Test

> Data audit: 2026-04-03  
> Scop: verificare deviații între **NuviaX_Growth_Framework_Rev_5_6.md**, documentația de workflow existentă și implementarea curentă (backend/frontend/docs).

---

## 1) Analiza Framework-ului Rev 5.6 (principii obligatorii)

Am extras criteriile structurale cu impact direct asupra produsului:

1. **Structură rigidă**: max 3 GO active, sprint fix 30 zile, 1 Behavior Model dominant/GO, formule și limite standardizate.
2. **Set BM închis**: `{CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE}` (fără extinderi ad-hoc).
3. **Governanță SRM formală**: ierarhie L3 > L2 > L1, un singur nivel activ, timeout-uri și fallback-uri clare.
4. **Control sezonal**: `execution_windows` + status `SEASONAL_PAUSE`, cu excludere explicită din denominatorii de continuitate.
5. **Integritate temporală**: reguli explicite pentru completări întârziate / bulk completion / consistency fairness.
6. **Metrici de stabilitate**: clamp [0,1], drift determinist, reguli explicite pentru regression/stagnation/chaos.

Concluzie: Framework-ul este foarte bine definit la nivel **principial**, însă necesită mapare strictă în contractele de date/statusuri și în job-uri operaționale pentru a evita derapajele de implementare.

---

## 2) Verificarea `docs/user-workflow.md` vs implementare reală

### 2.1 Deviații de documentație (workflow doc nealiniat cu codul curent)

1. **SA-3 marcat "NOT IMPLEMENTED" în doc, dar implementat în cod**  
   - `jobDetectStagnation` inserează SRM L1 când există 5+ zile inactive (`InsertSRMEvent`).
2. **SA-4 marcat "NOT IMPLEMENTED" în doc, dar implementat în cod**  
   - `ConfirmSRML2` creează `context_adjustments` cu `AdjEnergyLow` pe 7 zile.
3. **SA-1 marcat "NOT IMPLEMENTED" în doc, dar implementat în cod**  
   - `jobComputeDailyScore` cheamă `ComputeGrowthTrajectory` zilnic.
4. **Bug vizualizare "FROM goals" marcat în doc, dar reparat în cod**  
   - fallback query citește acum `global_objectives`, nu `goals`.
5. **SA-2 marcat "NOT IMPLEMENTED" în doc, dar implementat în cod**  
   - scheduler-ul cheamă `AwardAchievementIfEarned` la generarea ceremoniilor.

**Impact:** documentația de workflow produce "false negatives" și poate induce prioritizare greșită în roadmap/testare.

---

## 3) Deviații reale față de Framework Rev 5.6 (produs)

> Mai jos sunt deviații **reale**, nu doar neconcordanțe de documentație.

### D-01 — Behavior Model mismatch structural
- Framework: BM obligatoriu din setul `{CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE}`.
- Implementare: câmpul `dominant_behavior_model` validează `{ANALYTIC, STRATEGIC, TACTICAL, REACTIVE}`.
- Risc: sistemul operează cu o taxonomie diferită față de Layer 0 (incompatibilitate semantică și metrică).

### D-02 — Lipsă `SEASONAL_PAUSE` și `execution_windows`
- Framework Rev 5.6 cere explicit sezonalitate formală.
- Implementare DB/status: `sprint_status` = `ACTIVE|COMPLETED|SKIPPED`, fără `SEASONAL_PAUSE`; nu există câmpuri pentru ferestre sezoniere în schema principală.
- Risc: continuitate/consistență incorectă pentru GO sezoniere; scoruri distorsionate.

### D-03 — Ierarhie SRM parțială (unicitate nivel activ neenforțată)
- Framework: un singur nivel SRM activ simultan + ierarhie strictă L3>L2>L1.
- Implementare: evenimentele se adaugă în `srm_events`, însă nu există mecanism global de revocare/închidere a nivelurilor inferioare la escaladare.
- Risc: stare SRM ambiguă în istoric/analitice, comportament inconsistent în UI/rapoarte.

### D-04 — Regression event pipeline incomplet integrat
- Framework cere reacție imediată la regression măsurabil.
- Implementare: funcția de detectare `CheckAndRecordRegressionEvent` există, dar nu este conectată în fluxul zilnic de calcul/execuție.
- Risc: lipsesc trigger-ele proactive SRM și semnalele de siguranță pe scădere reală.

### D-05 — Lipsă contract formal pentru regulile de validitate temporală (A3)
- Framework 5.5 (A3) impune reguli clare pentru completări >48h, bulk completion, impact separat Progress/Consistency.
- Implementare: există piese de infrastructură, dar fără un contract API/DB explicit end-to-end pentru aplicarea regulii A3 în scorare.
- Risc: evaluarea consistenței poate fi "gameable" și neuniformă între utilizatori.

---

## 4) Ajustări necesare pentru workflow "100% lucrativ"

### Prioritate P0 (blocante de conformitate)
1. **Unificare taxonomie Behavior Model**
   - Introdu model canonic Framework (`CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE`) în DB/API.
   - Păstrează `ANALYTIC/STRATEGIC/TACTICAL/REACTIVE` doar ca strat secundar (ex: `execution_profile`).
2. **Introduce sezonalitate formală în schema și engine**
   - Add: `execution_windows` la goal-level + `SEASONAL_PAUSE` în `sprint_status`.
   - Update calcule continuitate/GORI conform Rev 5.6 (excludere suspended + seasonal_pause din numitor).
3. **Enforce SRM single-active-level**
   - La inserare nivel superior: revocare automată niveluri inferioare active.
   - Expune în API atât `current_level`, cât și `active_event_id`.

### Prioritate P1 (stabilitate operațională)
4. **Integrează efectiv Regression pipeline în job zilnic**
   - Rulează `CheckAndRecordRegressionEvent` după `ComputeGoalScore`.
   - Trigger direct SRM L1 când se detectează regression confirmat.
5. **Formalizează A3 (late/bulk completion) în engine**
   - Introdu câmpuri/evente explicite pentru: `completed_at`, `is_late`, `bulk_flag`.
   - Score contract: `Progress=YES`, `Consistency=NO` pentru completări >48h.

### Prioritate P2 (guvernanță documentară)
6. **Actualizează `docs/user-workflow.md` pe baza codului actual**
   - Elimină marcajele "NOT IMPLEMENTED" care nu mai sunt valide.
   - Adaugă secțiune de "as-built vs framework gap" separată de bug backlog.
7. **Adaugă Compliance Matrix permanentă**
   - Tabel C1…C40 cu status: `Implemented / Partial / Missing / Out-of-scope`.

---

## 5) Document de Stress Test al Framework-ului (pe baza pașilor 2+3)

## 5.1 Obiectiv
Să măsoare cât de bine este definit Framework-ul în practică (nu doar teoretic), prin testarea clarității, testabilității și implementabilității fiecărei reguli majore.

## 5.2 Metodă de scoring
Pentru fiecare regulă auditată (ex: BM, SRM, sezonalitate, temporal validity):
- **Claritate** (0-5): regula este lipsită de ambiguități?
- **Mapare tehnică** (0-5): există model DB/API evident?
- **Testabilitate** (0-5): poate fi verificată automat prin test/e2e/job?
- **Rezistență la abuz** (0-5): poate fi "păcălită" ușor?

Scor total per regulă = /20.

## 5.3 Rezultat stres test (curent)

| Domeniu testat | Claritate | Mapare tehnică | Testabilitate | Anti-abuz | Scor | Verdict |
|---|---:|---:|---:|---:|---:|---|
| Set BM canonic | 5 | 1 | 2 | 3 | 11/20 | **Slab** (mismatch semantic) |
| SRM ierarhic | 4 | 3 | 3 | 2 | 12/20 | **Mediu-** |
| Sezonalitate (Rev 5.6) | 5 | 1 | 1 | 3 | 10/20 | **Slab** |
| Temporal validity (A3) | 4 | 2 | 2 | 2 | 10/20 | **Slab** |
| Regression immediate signal | 4 | 2 | 2 | 3 | 11/20 | **Slab-Mediu** |
| Opaque scoring/API boundary | 5 | 4 | 4 | 4 | 17/20 | **Bun** |

### Scor global estimat
**71 / 120 (59.2%)** — Framework-ul este bine definit conceptual, dar transpunea tehnică este parțială pe zonele cele mai sensibile (taxonomie BM, sezonalitate, validitate temporală, SRM orchestration complet).

## 5.4 Set minim de teste de stres recomandate (obligatoriu înainte de "100% lucrativ")

1. **ST-BM-01**: creare GO pe toate cele 5 BM canonice + verificare score path distinct.
2. **ST-SEAS-01**: GO sezonier cu 2 ferestre active + 1 inactivă → verificare `SEASONAL_PAUSE` și continuity denominator.
3. **ST-SRM-01**: escaladare L1→L2→L3 în 10 zile cu verificare unicitate nivel activ.
4. **ST-TIME-01**: completare task la 72h + bulk 7 task-uri/8 min → verificare impact Progress/Consistency și flag-uri A3.
5. **ST-REG-01**: regres metric > prag → inserare `regression_event` + trigger SRM imediat.
6. **ST-OPAQUE-01**: audit payload API pentru a confirma că metricile interne rămân neexpuse.

---

## 6) Verdict executiv

- **Ce este bun:** direcția arhitecturală și separarea server-side scoring sunt solide.
- **Ce blochează conformitatea full Rev 5.6:** taxonomia BM, sezonalitatea formală, SRM single-level enforcement și contractul temporal validity.
- **Recomandare:** după implementarea P0+P1, rulați pachetul de stres test și marcați explicit conformitatea C1…C40.

