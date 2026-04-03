# **LEVEL 3 - MONITORING AUTHORITY**

Monitoring Authority este mecanismul formal prin care sistemul monitorizează stabilitatea executivă și capacitatea strategică a utilizatorului în timpul execuției.

Acest nivel nu creează obiective și nu execută task-uri. Rolul lui este să observe, să măsoare și să intervină atunci când apare instabilitate.

### **Scopul acestui nivel**

- detectarea deviației față de plan
- identificarea stagnării
- măsurarea haosului executiv
- reglarea capacității reale
- prevenirea supraîncărcării

### **Regula fundamentală**

Execuția este permisă doar în condiții de stabilitate controlată. Dacă sistemul detectează: deviație excesivă, stagnare prelungită, haos global sau depășirea capacității - intervine prin mecanismele de control. Monitoring Authority nu modifică structura anuală - ajustează comportamentul executiv.

### **Relația cu restul sistemului**

Monitoring Authority urmează după Execution Architecture și influențează Adaptive Context Engine, Strategic Reset Mode, Velocity Control și Reinforcement Model. Fără monitorizare, execuția devine instabilă în timp.

## **STABILITY CONTROL LAYER**

Acest sub-bloc monitorizează stabilitatea execuției. Componentele C26-C29 detectează deviația, stagnarea, haosul global și gestionează distribuția atenției.

LEVEL 3 - Stability Control Layer

**C26 Dynamic Drift Engine**

#### **Definiție**

Mecanismul care calculează deviația dintre progresul așteptat și progresul real într-un Sprint și monitorizează acumularea acestei deviații în timp.

#### **De unde vine logica**

Fără măsurarea deviației: stagnarea este invizibilă, intervenția este întârziată, progresul devine iluzoriu. Drift reprezintă diferența dintre ce ar fi trebuit realizat și ce s-a realizat efectiv - semnal direct de sănătate a execuției.

#### **Reprezentare formală**

Drift(GO_i, t) = Real_Progress(GO_i) − Expected(t)

Expected(t) = t / 30, t ∈ { 1..30 }

Drift pozitiv: progres real > planificat (performanță peste așteptări)

Drift negativ: progres real < planificat (întârziere acumulată)

Trigger SRM L1:

Drift(GO_i) < −0.15 pentru 3 zile CONSECUTIVE → SRM L1

SAU regression_flag = TRUE din C24 → SRM L1 IMEDIAT

Expected(t) ÎNGHEȚAT în:

Core Stabilization (C35), Planned Pause (C32)

External Shock Buffer (C32), Reactivation Protocol (C36)

SEASONAL_PAUSE (C19)

#### **Explicație**

Dacă progresul real este mai mic decât cel planificat, Drift devine negativ și indică o întârziere. Dacă progresul real depășește planul, Drift devine pozitiv și indică performanță peste așteptări. Pragul de −0.15 corespunde unui decalaj de aproximativ 4-5 zile de execuție față de traiectoria liniară - semnal că GO-ul pierde ritm sistematic, nu că a avut o zi slabă.

Acțiunile din Optional Stack nu influențează calculul Drift decât dacă sunt validate ca direct legate de Milestone-ul activ. Această regulă previne distorsionarea măsurării progresului prin activități secundare.

#### **Înghețarea Expected(t)**

Înghețarea Expected(t) previne acumularea Drift negativ în perioadele când sistemul știe că execuția nu poate sau nu trebuie să aibă loc. Fără înghețare, Drift-ul ar escalada mecanic și ar declanșa SRM în perioadele de protecție - un loop paradoxal.

#### **Rol**

- Identifică întârzierile - permite intervenție timpurie.
- Alimentează SRM și Chaos Index Engine.

LEVEL 3 - Stability Control Layer

**C27 Stagnation Detection Engine**

#### **Definiție**

Mecanismul care detectează absența progresului pe o perioadă relevantă. Spre deosebire de Drift (care măsoară decalajul față de traiectoria așteptată), Stagnation detectează platoul complet - utilizatorul a încetat să genereze orice progres.

#### **De unde vine logica**

Progres zero pe termen lung indică blocaj psihologic, supraîncărcare sau obiectiv nerealist. Aceasta indică lipsă de avans, chiar dacă nu există regres activ.

#### **Reprezentare formală**

Stagnant(GO_i) = TRUE dacă:

Real_Progress(t2) − Real_Progress(t1) = 0 pe un interval relevant

Threshold standard: 5 zile consecutive de zero progres

Threshold ESB: 10 zile consecutive (External Shock Buffer activ)

DEZACTIVAT în:

Planned Pause, Crisis Protocol, ESB, Core Stabilization, SEASONAL_PAUSE

#### **Rol**

- Previne stagnarea prelungită - poate declanșa Focus Rotation.
- Poate activa SRM dacă stagnarea persistă după threshold.

LEVEL 3 - Stability Control Layer

**C28 Chaos Index Engine**

#### **Definiție**

Mecanismul care măsoară instabilitatea globală a execuției. Haosul nu este lipsa progresului, ci dezorganizarea globală. Poate exista progres cu Chaos Index ridicat.

#### **De unde vine logica**

Un singur indicator (Drift sau Stagnation) nu poate surprinde complexitatea instabilității. Chaos Index agregă semnalele din toate sursele - Drift, Stagnare, Inconsistență și Context extern - într-un singur indice care determină nivelul de intervenție necesar.

#### **Reprezentare formală**

Chaos_Index = Drift_comp × 0.30 + Stagnation_comp × 0.25

\+ Inconsistency_comp × 0.25 + Context_disruption × 0.20

Drift_comp = max(|Drift(GO_i)|) pentru toate GO_i ACTIVE

(cel mai slab GO determină - nu media)

Stagnation_comp = { 0 dacă nicio stagnare | 0.5 dacă 1 GO | 1.0 dacă 2+ GO }

Inconsistency_comp = Variance(Completion_rate_zilnic, ultimele 14 zile)

Context_disruption = min(1.0, nr_eventi_majori / 3)

Praguri intervenție:

Chaos_Index < 0.30 → Verde: sistem stabil

0.30-0.40 → Galben: monitorizare crescută

0.40-0.60 → Amber: SRM L2 recomandat

Chaos_Index ≥ 0.60 → Roșu: SRM L3 recomandat

#### **Principiul Conservativ - max() în loc de medie**

Drift_comp folosește maximul valorilor absolute ale Drift-ului per GO, nu media. Un sistem cu un GO la Drift −0.05 și un al doilea la Drift −0.80 are Drift_comp = 0.80, nu 0.425. Media ar masca problemele grave. Cel mai slab element determină sănătatea sistemului.

#### **Context_disruption**

Plafonul la min(1.0, nr_eventi/3) asigură că Chaos_Index rămâne calculabil chiar în perioadele cu număr mare de evenimente perturbatoare. Chaos_Index în sine rămâne neclamped - poate depăși 1.0 teoretic pentru a semnaliza magnitudinea reală a instabilității.

#### **Rol**

- Detectează dezorganizarea sistemică - poate activa Core Stabilization Mode.
- Influențează Velocity Control și nivelul de intervenție SRM.

LEVEL 3 - Stability Control Layer

**C29 Focus Rotation Logic**

#### **Definiție**

Mecanismul prin care atenția strategică este redistribuită între GO active, balansând prioritatea strategică cu starea curentă a execuției.

#### **De unde vine logica**

Dacă toate GO primesc atenție egală permanent, apare diluare, apare supraîncărcare și scade eficiența. Obiectivele cu prioritate mai mare trebuie să primească mai multă atenție; obiectivele cu deviații mari trebuie să primească atenție suplimentară pentru recuperare.

#### **Reprezentare formală**

Focus_weight_brut(GO_i) = w(GO_i) / (1 + |Drift(GO_i)|)

Normalizare exclusiv pe GO cu status = ACTIVE:

Focus_norm(GO_i) = Focus_weight_brut(GO_i) / Σ( Focus_weight_brut(GO_j) )

unde j parcurge EXCLUSIV GO_j.status = ACTIVE

GO SUSPENDED, VAULT, SEASONAL_PAUSE: excluse complet din Σ

Garanție minimă: min 1 acțiune/zi/GO activ, indiferent de Focus_norm

#### **Excluderea GO SUSPENDED din Normalizare**

Includerea unui GO SUSPENDED în normalizare ar produce o distribuție distorsionată: GO-ul ar consuma o fracție din Σ fără a genera acțiuni reale, iar celelalte GO active ar primi mai puțin decât justifică greutățile lor relative. Excluderea completă asigură că Focus Rotation reflectă realitatea operațională.

#### **Rol**

- Redistribuie efortul - protejează GO dominante strategice.
- Stabilizează execuția - ajustează atenția în funcție de starea reală a fiecărui GO.

## **CAPACITY REGULATION LAYER**

Acest sub-bloc reglează încărcarea anuală și ritmul de execuție. Componentele C30-C31 monitorizează sarcina reală față de capacitate și ajustează viteza de execuție proporțional.

LEVEL 3 - Capacity Regulation Layer

**C30 ALI Engine - Annual Load Index**

#### **Definiție**

Mecanism numeric care evaluează încărcarea anuală reală comparativ cu capacitatea utilizatorului. ALI detectează supraîncărcarea reală și proiectată, nu doar estimată inițial.

#### **De unde vine logica**

Estimarea inițială poate fi greșită. Execuția reală poate consuma mai mult efort decât planificat. ALI detectează supraîncărcarea reală în timp real, nu retrospectiv.

#### **Reprezentare formală**

ALI_curent = Ore_acumulate_total / C_annual

ALI_proiectat = ALI_curent × (365 / Zile_scurse)

Final_ALI = ALI_proiectat × Execution_Reliability

Execution_Reliability: factor de corecție bazat pe rata istorică de finalizare

(floor 0.6 - nu scade sub 0.6 pentru a evita suprapenalizarea)

Praguri intervenție:

Final_ALI ≤ 1.0 → capacitate normală

Final_ALI ∈ (1.0, 1.10\] → Ambition Buffer: avertisment, activare permisă

Final_ALI > 1.10 → Core Stabilization recomandat

Trigger-urile ALI dezactivate în primele 14 zile de utilizare

GO SUSPENDED sau SEASONAL_PAUSE: excluse din calculul orelor consumate

#### **ALI_curent vs ALI_proiectat**

ALI_curent este o metrică retrospectivă - câtă capacitate a fost consumată până acum. ALI_proiectat este metrica de decizie - dacă utilizatorul menține ritmul actual tot anul, va depăși capacitatea? Deciziile sistemice se bazează pe ALI_proiectat. Un utilizator cu ALI_curent = 0.6 la jumătatea anului poate părea ok, dar dacă ritmul curent este intensificat, ALI_proiectat poate fi 1.2 - sistem în risc.

#### **Rol**

- Previne supraîncărcarea anuală - poate declanșa Velocity Control.
- Poate activa Core Stabilization Mode când Final_ALI > 1.10.

LEVEL 3 - Capacity Regulation Layer

**C31 Velocity Control Mechanism**

#### **Definiție**

Mecanismul care ajustează ritmul de execuție în funcție de capacitatea detectată de ALI Engine. Ritmul constant în condiții de supraîncărcare produce burnout.

#### **De unde vine logica**

Dacă utilizatorul este supraîncărcat și ritmul nu scade, execuția devine nesustenabilă. Reducerea proporțională a ritmului protejează sustenabilitatea pe termen lung și stabilizează ALI.

#### **Reprezentare formală**

Capacity_factor = 1.0 − max(0, Final_ALI − 1.0)

Velocity_ajustat = Velocity_base × Capacity_factor

Exemplu: Final_ALI = 1.08

Capacity_factor = 1.0 − 0.08 = 0.92

Velocity_ajustat = Velocity_base × 0.92 (−8% față de normal)

Reducerea este progresivă și proporțională. La Final_ALI = 1.10 (pragul maxim al Ambition Buffer), reducerea este de 10%. Ajustarea este aplicată silențios în Daily Stack Generator - utilizatorul vede un plan mai puțin dens.

#### **Rol**

- Reduce intensitatea temporar - protejează sustenabilitatea execuției.
- Stabilizează ALI - previne escaladarea spre Core Stabilization Mode.

