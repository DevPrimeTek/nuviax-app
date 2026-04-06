# MVP_SCOPE.md — Matrice prioritizare C1–C40 (completată)

> **Versiune:** 1.0.0  
> **Completat:** 2026-04-06  
> **Criteriu de selecție:** Fluxul minim MVP funcțional:  
> Register → Create GO (AI) → Sprint generat → Task-uri zilnice → Complete → Scor → Dashboard

**Legenda:**
- **FULL** — implementare completă conform framework-ului
- **SIMPLIFIED** — logică de bază funcțională, fără edge cases avansate
- **POST_MVP** — nu se implementează acum

---

## Layer 0 — Axiomatic Foundation

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C1 | Structural Supremacy | FULL | Principiu — aplicat prin design, nu prin cod explicit |
| C2 | Behavior Model System | FULL | 5 BM-uri necesare pentru creare GO; validare la input |
| C3 | Max 3 Active GO | FULL | Validare simplă, critică. SEASONAL_PAUSE → SIMPLIFIED (status există în DB, logica detaliată post-MVP) |
| C4 | 365-Day Max Duration | FULL | O singură verificare la creare GO |
| C5 | 30-Day Fixed Sprint | FULL | Fundament — Expected(t) = t/30, Drift derivat din el |
| C6 | Normalization [0,1] | FULL | Funcție clamp, 5 linii de cod |
| C7 | Priority Weight (1-3) | SIMPLIFIED | Weight derivat din Relevance cu mapping fix. Fără auto-rezoluție complexă la depășire — avertisment simplu |
| C8 | Priority Balance ≤7 | SIMPLIFIED | Check sumă ≤7 la activare GO. Auto-rezoluție simplă: reject al 4-lea dacă depășește, fără recalcul weight automat |

**Layer 0 total:** 6 FULL + 2 SIMPLIFIED = 8/8

---

## Level 1 — Structural Authority

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C9 | Semantic Parsing | SIMPLIFIED | AI extrage domain/direction/metric/timeframe. Fără Reformulation Queue, fără GO_REJECTED_LOGICAL_CONTRADICTION, fără timeout 48h→Vault automat |
| C10 | BM Classification | SIMPLIFIED | AI atribuie BM cu confidence. Fără Confidence Gate interactiv (confidence<0.50 → fallback pe selecție manuală, nu reformulation queue) |
| C11 | Strategic Relevance | SIMPLIFIED | Scor fix la creare (Impact×0.35 + Urgency×0.25 + Alignment×0.25 + Feasibility×0.15). Fără recalibrare post-Sprint 1 |
| C12 | Future Vault | FULL | Status WAITING pe GO — deja există în DB, logica e simplă |
| C13 | Relevance Thresholds | SIMPLIFIED | Floor 0.30 → reject. Mapping la weight (1/2/3) conform C7 |
| C14 | GO Validation | FULL | Verificare: deadline ≤365d, BM definit, metric prezent. Critic pentru integritate date |
| C15 | Strategic Feasibility | POST_MVP | Analiză pre-activare complexă — MVP-ul nu are suficiente date istorice |
| C16 | Capacity Calibration | POST_MVP | Avertisment C_daily — nice to have, nu blochează fluxul |
| C17 | Deep Work Estimation | POST_MVP | Optimizare avansată de planificare — post-MVP |
| C18 | Annual Recalibration | POST_MVP | Relevantă doar după luni de utilizare — imposibil de testat în MVP |

**Level 1 total:** 2 FULL + 4 SIMPLIFIED + 4 POST_MVP = 10/10

---

## Level 2 — Execution Authority

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C19 | Sprint Structuring | FULL | Sprint 30 zile + statusuri ACTIVE/COMPLETED. SEASONAL_PAUSE există ca status dar logica de execution_windows e POST_MVP |
| C20 | Sprint Target Calc | FULL | Formula directă: (Target - Progress) / Sprints_rămase × 0.80 |
| C21 | 80% Probability Rule | FULL | Factor 0.80 — o multiplicare, parte din C20 |
| C22 | Milestone Structuring | SIMPLIFIED | Creare 3 checkpoints per sprint automat. Fără ordered flag, fără milestone dependencies |
| C23 | Daily Stack Generator | SIMPLIFIED | 1-3 task-uri/zi bazat pe sprint + milestone activ. Fără Physical Delta Safety Signal (A1), fără SINGLE_QUESTION_FLAG |
| C24 | Progress Computation | FULL | Formula: Σ(completed×weight)/Σ(total_weight). Critică |
| C25 | Execution Variance | FULL | Drift_raw = Real_Progress - Expected(t). Transmis la C26 |

**Level 2 total:** 5 FULL + 2 SIMPLIFIED = 7/7

---

## Level 3 — Monitoring Authority

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C26 | Dynamic Drift Engine | SIMPLIFIED | Calcul Drift, trigger SRM L1 la <-0.15/3 zile. Fără Expected(t) freeze scenarios complexe (depind de C35 care e POST_MVP) |
| C27 | Stagnation Detection | SIMPLIFIED | ≥5 zile consecutive fără task completat → flag. Fără ESB threshold extension (10 zile) |
| C28 | Chaos Index Engine | SIMPLIFIED | Formula de bază cu 4 componente, praguri Verde/Galben/Amber/Roșu. Drift_comp cu max() nu medie. Fără context_disruption complex |
| C29 | Focus Rotation | POST_MVP | Redistribuire atenție — optimizare, nu blochează fluxul MVP |
| C30 | Consistency Tracking | SIMPLIFIED | active_days/eligible_days basic. Fără weighted recency (7d×0.60 + 30d×0.40) — media simplă |
| C31 | Behavioral Patterns | POST_MVP | Detectare PROCRASTINATION/STEADY etc. — ML-ready, complexitate mare, zero impact pe fluxul MVP |

**Level 3 total:** 0 FULL + 4 SIMPLIFIED + 2 POST_MVP = 6/6

---

## Level 4 — Regulatory Authority

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C32 | Adaptive Context Engine | SIMPLIFIED | Doar Planned Pause (adj_type PAUSE există în DB). Fără: Energy Modulation, Crisis Protocol, ESB, Momentum Monitor, Burnout Prevention |
| C33 | SRM (3 Levels) | SIMPLIFIED | L1 auto (Sprint Target -20%), L2 notificare, L3 confirmare + suspendare. Ierarhie strictă single-active. Fără: Timeout Protocol complet (24h/72h/7d), fără Imunitate Reactivation |
| C34 | Weighted GO Suspension | POST_MVP | Suspendare selectivă la SRM L3 — utilizatorii MVP nu vor ajunge aici rapid |
| C35 | Core Stabilization | POST_MVP | Mod criză — necesită luni de utilizare pentru a fi relevant |
| C36 | Reactivation Protocol | POST_MVP | Rampă 8 zile — dependent de C35 |

**Level 4 total:** 0 FULL + 2 SIMPLIFIED + 3 POST_MVP = 5/5

---

## Level 5 — Strategic Consolidation

| C# | Componentă | Scope | Justificare |
|----|-----------|-------|-------------|
| C37 | Sprint Score | FULL | Formula: Progress×0.50 + Consistency×0.30 + Deviation×0.20. Fundament evaluare |
| C38 | GORI | SIMPLIFIED | Media ponderată sprint scores × Continuity factor. Fără Variance Penalty (×0.25), fără excludere SEASONAL_PAUSE (nu e implementat complet) |
| C39 | Engagement Signal | POST_MVP | 3 proxy-uri comportamentale — necesită date istorice, imposibil de testat în MVP |
| C40 | Sprint Reflection Gate | POST_MVP | 3 întrebări opționale, zero impact scor — polish, nu core |

**Level 5 total:** 1 FULL + 1 SIMPLIFIED + 2 POST_MVP = 4/4

---

## Rezumat

| Scope | Componente | Procent |
|-------|-----------|---------|
| **FULL** | C1, C2, C3, C4, C5, C6, C12, C14, C19, C20, C21, C24, C25, C37 | **14 (35%)** |
| **SIMPLIFIED** | C7, C8, C9, C10, C11, C13, C22, C23, C26, C27, C28, C30, C32, C33, C38 | **15 (37.5%)** |
| **POST_MVP** | C15, C16, C17, C18, C29, C31, C34, C35, C36, C39, C40 | **11 (27.5%)** |
| **Total** | | **40 (100%)** |

**MVP implementează: 29 componente (14 FULL + 15 SIMPLIFIED)**  
**POST_MVP: 11 componente (cele mai complexe sau cele care necesită date istorice)**

---

## Verificare flux MVP cu această matrice

| Pas flux | Componente folosite | Acoperit? |
|----------|-------------------|-----------|
| Register + Login | Auth (nu e componentă framework) | ✅ Deja funcțional |
| Create GO cu AI | C9, C10, C14, C2, C3, C4, C11, C12, C13 | ✅ Toate minim SIMPLIFIED |
| Sprint generat automat | C19, C20, C21, C22 | ✅ Toate minim SIMPLIFIED |
| Task-uri zilnice | C23 | ✅ SIMPLIFIED |
| User completează task | C24, C25 | ✅ FULL |
| Scor calculat zilnic | C26, C27, C28, C30, C37 | ✅ Toate minim SIMPLIFIED |
| Dashboard cu progress | C6, C7, C24, C37 | ✅ Toate minim SIMPLIFIED |
| SRM la probleme | C33, C32 | ✅ SIMPLIFIED |
| Sprint close → GORI | C38 | ✅ SIMPLIFIED |
| Ceremonie sprint | Ceremonies (tabel există) | ✅ Implementabil cu C37 |

**Flux MVP complet acoperit. Zero componente lipsă.**

---

*v1.0.0 | 2026-04-06*  
*Aprobat de: [Ștefan — marchează aici după review]*
