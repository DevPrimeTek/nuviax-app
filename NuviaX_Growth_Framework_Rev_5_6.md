**NUViaX Growth Framework™**

**Revizia 5.6**

*Ediție Consolidată --- Documentul Complet Integrat*

\---

**Versiune**                5.6 --- Ediție Consolidată

**Bază**                    Rev. 5.3 (text original integrat) +
ajustări 5.4, 5.5, 5.6

**Componente**              40 (C1--C40, inclusiv C39 și C40 introduse
în Rev. 5.5)

**Arhitectură**             Layer 0 + Level 1--5, 8 niveluri

**Patch 5.6**               SEASONAL\_PAUSE integrat în C3, C19, C38

**Limbă**                   Română

\---

**IDENTITATEA SISTEMULUI**

NUViaX Growth Framework™ este un sistem proprietar de organizare
strategică anuală a progresului uman, bazat pe limitarea controlată a
focusului, clasificarea formală a tipurilor de transformare, execuție
ciclică fixă, control matematic al deviației și capacității, și reglare
contextuală adaptivă.

**Scopul sistemului este:**

* realizarea sustenabilă a obiectivelor dominante
* prevenirea haosului și supra-încărcării
* eliminarea stagnării
* menținerea stabilității pe termen lung

**PRINCIPIUL FUNDAMENTAL**

Framework-ul este Rigid Structural -- Flexibil Operațional. Acesta
reprezintă principiul arhitectural central prin care elementele
fundamentale ale sistemului sunt invariabile, iar parametrii
operaționali pot fi ajustați contextual fără a altera structura
metodologică.

\---

**CE ÎNSEAMNĂ RIGID STRUCTURAL**    **CE ÎNSEAMNĂ FLEXIBIL
OPERAȚIONAL**

Rigiditatea structurală reprezintă  Flexibilitatea operațională
ansamblul elementelor arhitecturale reprezintă capacitatea sistemului
care: --- nu pot fi modificate de   de a adapta (fără a altera
utilizator --- nu pot fi ajustate   structura): --- intensitatea
contextual --- nu se schimbă în     acțiunilor --- estimarea efortului
funcție de tipul obiectivului ---   și ritmul de execuție --- contextul
definesc identitatea sistemului     de aplicare --- fragmentarea și
sensibilitatea

\---

\---

**RIGIDUL --- protejează SISTEMUL** **FLEXIBILUL --- protejează
UTILIZATORUL**

\--- protejează sistemul --- previne --- protejează utilizatorul ---
haosul --- menține coerența         previne rigiditatea excesivă ---
menține umanitatea

\---

\---

**ELEMENTE RIGID STRUCTURALE**      **ELEMENTE FLEXIBILE OPERAȚIONALE**

Maximum 3 Global Objectives active  Conținutul concret al GO
simultan

Exact 1 Behavior Model dominant per Valoarea numerică a targetului
GO

Sprint fix de 30 zile               Estimarea Sprint Load

80% Probability Rule aplicată la    Numărul de Milestone (max 5)
Sprint Design

Dynamic Drift calculat prin         Ordinea milestone-urilor
Expected(t)=t/30

Chaos definit ca metrică globală    Daily Stack specific

ALI ca mecanism numeric de limitare Activarea Pause / activarea Crisis
Mode

Clamp universal al tuturor          Intensitatea Velocity Control
metricilor în \[0,1]

Interdicția de a crea Behavior      Parametrii temporari ai Context
Models noi                          Engine

Interdicția de a modifica           ---
arhitectura pentru cazuri speciale

\---

**SUMAR MODIFICĂRI --- REV. 5.4 → 5.5 → 5.6**

Aceasta este ediția consolidată a NUViaX Growth Framework. Toate
ajustările introduse în reviziile 5.4, 5.5 și 5.6 sunt integrate direct
în corpul componentelor --- fără secțiuni separate de versiune. Această
pagină oferă indexul complet de referință față de Rev. 5.3 (textul de
bază).

**Rev. 5.4 --- Stabilitate Operațională**

Rev. 5.4 a adresat incoerențe operaționale identificate în testarea Rev.
5.3. Modificările vizează determinismul numeric, limitele de capacitate
și prioritizarea protocoalelor.

\---

**C#**    **Modificare**          **Rațiune**

**C5**    t = integer 1..30       Elimină nedeterminismul la granițele temporale.
explicit, Drift 1×/zi

**C6**    Context\_disruption =    Metrici de alertă rămân neclamped pentru a
min(1.0, n/3) ---       semnaliza amplitudinea reală.
hibrid

**C8**    Auto-rezoluție când     Mecanism activ, nu pasiv --- GO cu Relevance
Σ(weight) > 7          minimă reduce weight automat.

**C11**   Recalibrare             Fără întrerupere dacă < 50% zile lucrate;
post-Sprint1            trigger automat check\_priority\_balance().
silențioasă

**C20**   Sprint\_Target\_realist = Factorul 80% integrat explicit în calcul, nu
brut × 0.80 explicit    implicit.

**C24**   Regression Event → SRM  Excepție de la regula de 3 zile --- regresul
L1 IMEDIAT              măsurabil este semnal de urgență.

**C27**   Stagnation ESB          Evitare false-positive în perioadele de șoc
threshold: 5 → 10 zile  extern.

**C28**   Drift\_comp =            Cel mai slab GO determină alerta --- principiu
max(|Drift|) nu medie conservativ.

**C32**   Retroactive Pause: max  Limite anti-abuz retrospectiv.
48h retroactiv, max  
3/sprint

**C34**   C34 > C18 dacă GO cu   Stabilization Review are prioritate față de
status SUSPENDED        recalibrarea anuală.

**C35**   Expected(t) înghețat în Previne loop paradoxal: Stabilization → Drift
Core Stabilization      negativ → SRM L1 imediat.

**C37**   Zile ESB și Retroactive Zilele de pauză legitimă nu penalizează scorul
Pause excluse din       de consistență.
Consistency

**C38**   Sprinturi SUSPENDED     Nu 0, nu medie --- excluse, ca și cum nu ar fi
excluse complet din     existat.
GORI

\---

**Rev. 5.5 --- 26 Ajustări Structurale (A1--D3)**

Rev. 5.5 a introdus 2 componente noi (C39, C40) și a rezolvat 17 din 18
gap-uri identificate în Mega Stress Test. Ajustările sunt clasificate
după severitate: P0 Fatal, P1 Critical, P2 Major, P3 Minor.

\---

**Prioritate**   **Cod**   **Comp.**   **Descriere**

**P0 --- Fatal** **A1**    C9, C23     Physical Delta Safety Signal:
domain=Sănătate\_fizică + REDUCE + delta > 25% →
SINGLE\_QUESTION\_FLAG. Mesaj unic la activare +
conector specialist Sprint1 M1.

**P0 --- Fatal** **A2**    C39 NOU     Engagement Signal: 3 proxy-uri motivaționale
monitorizate în background. Dacă 2/3 negative ≥ 7
zile → o singură întrebare. Zero interfață
vizibilă în mod normal.

**P0 --- Fatal** **A3**    C23, C37    Temporal Validity: completări > 48h = Late
Completion → Progress YES, Consistency NO. Bulk
> 5 acțiuni în < 10 min → C39 Proxy 1 flag.

**P0 --- Fatal** **A4**    C33         SRM L3 Timeout Protocol: 24h neconfirmat →
auto-aplicare L2. 72h → re-propune L3. 7 zile →
auto-SUSPEND cu opțiune reactivare.

**P0 --- Fatal** **A5**    C9, C10     GO\_REJECTED\_LOGICAL\_CONTRADICTION: categorie nouă
pentru BM opuse pe aceeași metrică. 3 opțiuni
predefinite. Timeout 7 zile → Vault.

**P1 ---         B1    C8          Priority Balance Universal:
Critical**                             check\_priority\_balance() apelat automat după
ORICE modificare de status GO.

**P1 ---         B2    C33         SRM Ierarhie Formală: L3 > L2 > L1. Un singur
Critical**                             nivel activ simultan. Sprint Target recalculat o
singură dată.

**P1 ---         B3    C38         GORI Continuity Factor: GORI\_final =
Critical**                             GORI\_calculat × (sprinturi\_active /
sprinturi\_totale\_eligibile).

**P1 ---         B4    C29         Focus Rotation: Σ(weights) exclusiv pe GO cu
Critical**                             status=ACTIVE. SUSPENDED exclus complet din
normalizare.

**P1 ---         B5    C36, C33    Reactivation Immunity: SRM L1 dezactivat pe
Critical**                             durata Reactivation. Threshold SRM L2 crescut la
0.60.

**P1 ---         B6    C7          Relevance round(score, 2) înainte de orice
Critical**                             mapping --- elimină nedeterminismul
floating-point la granițe.

**P1 ---         B7    C11, C8     C11 Recalibration declanșează automat
Critical**                             check\_priority\_balance() --- sincronizare
completă.

**P2 --- Major** **C1**    C16         Capacity Validation Gate: C\_daily < 0.5h sau >
14h → avertisment informativ (nu blocare
structurală).

**P2 --- Major** **C2**    C13         Minimum Relevance Floor 0.30: GO cu Relevance <
0.30 → Vault automat, indiferent de alte
condiții.

**P2 --- Major** **C3**    C19         execution\_windows: GO sezoniere cu ferestre
active definite. Expected(t) înghețat în
perioadele inactive.

**P2 --- Major** **C4**    C9, C10     Reformulation Queue: cereri de clarificare strict
secvențiale, una câte una. Timeout 48h → Vault.

**P2 --- Major** **C5**    C38         GORI Variance Penalty: GORI × (1 −
Variance(Sprint\_Scores) × 0.25). Maxim 25%
penalizare.

**P2 --- Major** **C6**    C20         Sprint Target Plafon: compensația ≤ 1.5× target
inițial/sprint. 3+ Regression consecutive →
reducere target anual.

**P2 --- Major** **C7**    C40 NOU     Sprint Reflection Gate: 3 întrebări opționale la
tranziția dintre sprinturi. Zero impact pe niciun
scor.

**P3 --- Minor** **D1**    C5          t = integer. Drift calculat o dată/zi la finalul
zilei t.

**P3 --- Minor** **D2**    C28         Context\_disruption = min(1.0, nr\_eventi/3).

**P3 --- Minor** **D3**    C22         Milestone ordered flag: true/false opțional per
sprint.

\---

**Rev. 5.6 --- Patch Sezonier (Gap rezidual Rev. 5.5)**

Rev. 5.6 rezolvă singurul gap structural identificat în Mega Stress Test
Rev. 5.5: GO-urile cu execution\_windows primeau Continuity Factor scăzut
artificial, deoarece sprinturile inactive sezonier erau incluse în
numitor. Patch-ul introduce statusul SEASONAL\_PAUSE, exclus din
Continuity\_factor identic cu sprinturile SUSPENDED prin SRM L3.

\---

**C#**    **Modificare**        **Descriere**

**C3**    execution\_windows +   Sprinturile din perioadele inactive ale unui GO
status SEASONAL\_PAUSE sezonier primesc statusul distinct
SEASONAL\_PAUSE.

**C19**   Sprint status         La intrarea într-o perioadă inactivă:
SEASONAL\_PAUSE        Expected(t) înghețat, Progress conservat,
introdus formal       Consistency neutru, ALI zero.

**C38**   Formula Continuity    Continuity = sprinturi\_active / (totale −
Factor actualizată    seasonal\_pause − suspended). Ambele tipuri
excluse din numitor.

\---

**LAYER 0 --- AXIOMATIC FOUNDATION**

Layer 0 -- Axiomatic Foundation reprezintă setul de reguli structurale
invariabile care definesc identitatea, limitele și stabilitatea
matematică a NUViaX Growth Framework™. Este fundamentul pe care toate
celelalte nivele funcționează.

**Scopul acestui bloc**

* Prevenirea haosului structural.
* Stabilirea limitelor sistemului.
* Asigurarea stabilității matematice.
* Menținerea identității metodologice.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C1 Structural Supremacy Principle**                                 |
+-----------------------------------------------------------------------+

**Definiție**

Principiul prin care structura sistemului este invariabilă, iar
operaționalul este ajustabil doar în interiorul acestei structuri.

**De unde vine logica**

Orice sistem complex care permite modificarea structurii în timp real
devine instabil. Pentru stabilitate matematică și comportamentală:
Structura ≠ Parametri. Parametrii pot varia doar în interiorul
structurii, nu o pot modifica.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| S = setul de reguli structurale                                       |
|                                                                       |
| P = setul de parametri ajustabili                                     |
|                                                                       |
| S ∩ P = ∅ --- regulile structurale și parametrii nu se suprapun       |
| niciodată                                                             |
|                                                                       |
| P ⊂ Domain(S) --- parametrii pot exista exclusiv în interiorul        |
| regulilor structurale                                                 |
+-----------------------------------------------------------------------+

**Explicație**

Utilizatorul poate modifica valori (timp, target, intensitate), dar nu
poate modifica structura (număr GO, durată sprint, modele
comportamentale). Cu alte cuvinte: ce se face este rigid --- cum se face
este flexibil.

**Rol**

* Protejează framework-ul de extindere haotică.
* Previne „personalizarea excesivă".
* Menține consistența pe termen lung.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C2 Behavior Model System**                                          |
+-----------------------------------------------------------------------+

**Definiție**

Set finit și închis de tipuri universale de transformare acceptate în
sistem. Conține: CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE.

**De unde vine logica**

Orice transformare umană poate fi redusă la una dintre cele 5 direcții:
creare (nu există → există), creștere (există → mai mult), reducere
(există → mai puțin), menținere (stabilizare), evoluție (transformare
progresivă). Set finit → sistem închis → stabilitate matematică.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Fie T = orice transformare                                            |
|                                                                       |
| T ∈ { CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE }                    |
|                                                                       |
| |BM(GO)| = 1 --- unicitate obligatorie per GO                       |
|                                                                       |
| |T| = 5 --- set finit și închis, nu există al 6-lea model permis    |
+-----------------------------------------------------------------------+

**Cele 5 Behavior Models**

\---

**Model**      **Direcție**    **Exemple tipice**

**CREATE**     Construire din  Lansare produs, scriere carte, fondare
zero            companie, prototip nou

**INCREASE**   Creștere        MRR, număr clienți, masă musculară, vocabular,
măsurabilă      audiență

**REDUCE**     Reducere        Greutate corporală, timp de livrare, costuri
măsurabilă      fixe, timp ecran

**MAINTAIN**   Menținere în    Greutate stabilă, relație activă, skill activ,
interval        nivel fitness

**EVOLVE**     Transformare    Schimbare carieră, restructurare identitate,
calitativă      tranziție la leadership

\---

**Explicație**

Orice schimbare pe care utilizatorul o dorește trebuie să se încadreze
într-una dintre cele 5 categorii definite. Dacă un obiectiv nu poate fi
încadrat în niciuna, trebuie reformulat. Nu sunt permise obiective cu 2
direcții simultane pe aceeași metrică --- aceasta generează
GO\_REJECTED\_LOGICAL\_CONTRADICTION (C9).

**Rol**

* Standardizare universală --- permite formule matematice coerente
pentru sprint și progres.
* Elimină ambiguitatea direcțională înainte de activare.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C3 Maximum 3 Active GO Constraint**                                 |
+-----------------------------------------------------------------------+

**Definiție**

Numărul maxim de Global Objectives active simultan este 3. Dacă
utilizatorul încearcă să adauge un al patrulea obiectiv, sistemul îl va
respinge sau îl va trimite în Future Vault.

**De unde vine logica**

Fragmentarea atenției crește non-liniar cu numărul de obiective. Dacă n
= numărul de GO active, costul cognitiv ≈ n² (interferență între
obiective). La n > 3, interferența crește exponențial și niciun
obiectiv nu mai primește atenție suficientă pentru progres real.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| n\_active ≤ 3                                                          |
|                                                                       |
| Cost\_cognitiv ≈ n² (interferență exponențială, nu liniară)            |
|                                                                       |
| Dacă n = 3 → orice tentativă de activare nouă → C17 Future Vault      |
+-----------------------------------------------------------------------+

**GO Sezoniere și SEASONAL\_PAUSE**

Un GO cu execution\_windows definite ocupă un slot activ permanent ---
inclusiv în perioadele în care fereastra este inactivă. Aceasta reflectă
existența sa ca intenție strategică anuală. Pe durata perioadelor
inactive, statusul sprint-ului devine SEASONAL\_PAUSE: Expected(t) este
înghețat, acțiunile nu sunt generate, iar sarcina operațională este nulă
--- dar slotul este ocupat.

+-----------------------------------------------------------------------+
| SEASONAL\_PAUSE este un status distinct: diferit de SUSPENDED (act al  |
| sistemului ca răspuns la instabilitate) și diferit de VAULT           |
| (niciodată activat).                                                  |
|                                                                       |
| Sprinturile SEASONAL\_PAUSE sunt excluse din calculul                  |
| Continuity\_factor în C38 --- identic cu sprinturile SUSPENDED.        |
+-----------------------------------------------------------------------+

**Rol**

* Previne dispersia --- menține focus strategic real.
* Stabilizează ALI și calculele de capacitate.
* Forțează decizia de prioritizare conștientă.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C4 365-Day Maximum GO Duration Constraint**                         |
+-----------------------------------------------------------------------+

**Definiție**

Un GO nu poate depăși 365 de zile calendaristice de la data activării
până la deadline.

**De unde vine logica**

Obiectivele fără limită temporală devin identitare permanente, nu pot fi
evaluate și nu pot intra în cicluri de consolidare. Limitarea la 1 an
permite: 12 sprinturi măsurabile, evaluare anuală prin GORI, recalibrare
structurată. Dacă utilizatorul setează un obiectiv pe 2 ani, sistemul îl
obligă să îl împartă în etape anuale.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| deadline − start\_date ≤ 365 zile                                      |
|                                                                       |
| Deadline\_sugerat = start + Benchmark\_domeniu × Amplitudine            |
|                                                                       |
| Interval acceptat de utilizator: \[1, 365] zile                      |
+-----------------------------------------------------------------------+

**Rol**

* Forțează concretizarea --- obiectivele vagi sau nelimitate temporal
sunt respinse.
* Permite calculul GORI anual complet și comparabil.
* Menține ritmul strategic prin cicluri anuale definite.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C5 30-Day Fixed Sprint Constraint**                                 |
+-----------------------------------------------------------------------+

**Definiție**

Sprintul are durată fixă de exact 30 de zile calendaristice. Variabila
temporală t este un integer în intervalul \[1, 30]. Drift-ul se
calculează o singură dată pe zi, la finalul zilei.

**De unde vine logica**

Dacă durata sprintului este variabilă: progresul devine imposibil de
comparat, Drift devine instabil, GORI devine distorsionat.
Standardizarea la 30 de zile fixe permite măsurare uniformă și
comparabilitate istorică completă.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Sprint\_length = 30 (invariant structural)                             |
|                                                                       |
| t ∈ { 1, 2, 3, ..., 30 } --- integer explicit, nu real               |
|                                                                       |
| Expected(t) = t / 30                                                  |
|                                                                       |
| Exemplu: ziua 15 → 15/30 = 0.50 → 50% progres așteptat                |
|                                                                       |
| Drift calculat: 1 dată/zi, la finalul zilei t                         |
|                                                                       |
| Late completion (> 48h): → Progress\_comp YES / Consistency\_comp NO   |
|                                                                       |
| SEASONAL\_PAUSE: Expected(t) înghețat --- t nu avansează în perioade   |
| inactive                                                              |
+-----------------------------------------------------------------------+

**Rol**

* Permite calcul standardizat și determinist al Drift.
* Permite comparabilitate istorică între sprinturi și între GO-uri.
* Asigură că GORI este construit din date structurate uniform.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C6 Normalization Rule --- Clamp \[0,1]**                           |
+-----------------------------------------------------------------------+

**Definiție**

Toate valorile de progres și performanță sunt limitate în intervalul
\[0, 1] prin funcția clamp. Metricile de alertă și monitorizare nu sunt
clamped --- ele trebuie să poată semnaliza orice magnitudine de deviere.

**De unde vine logica**

Fără clamp: progresul poate deveni negativ sau poate depăși 100%, Drift
poate deveni instabil, ALI poate distorsiona sistemul. Clamp =
stabilitate numerică. Metricile de alertă (Drift, ALI) rămân libere
pentru a detecta amplitudinea reală a problemelor.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| clamp(x) = 0 dacă x < 0                                              |
|                                                                       |
| = x dacă 0 ≤ x ≤ 1                                                    |
|                                                                       |
| = 1 dacă x > 1                                                       |
|                                                                       |
| Aplicate clamp: Real\_Progress, Sprint\_Score, GORI, Focus\_weights      |
|                                                                       |
| Neclamped: Drift, ALI, Chaos\_Index (pot depăși 1.0)                   |
|                                                                       |
| Hibrid (plafonat explicit):                                           |
|                                                                       |
| Context\_disruption = min(1.0, nr\_eventi\_majori / 3)                   |
+-----------------------------------------------------------------------+

**Explicație**

Indiferent ce valoare rezultă din calcule: dacă este negativă → devine
0; dacă este mai mare de 100% → devine 1.0; dacă este între 0 și 1 →
rămâne neschimbată. Această regulă previne erorile matematice și
valorile imposibile care ar destabiliza engine-urile de evaluare.

**Rol**

* Previne instabilitatea matematică --- asigură că metricile de
progres rămân în domeniu valid.
* Asigură comparabilitate universală între utilizatori și GO-uri.
* Normalizează toate engine-urile din Levels 2--5.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C7 Priority Weight System (1--3)**                                  |
+-----------------------------------------------------------------------+

**Definiție**

Fiecărui GO i se atribuie un coeficient de prioritate între 1 și 3,
derivat automat din scorul de Relevance strategică. Utilizatorul nu
poate seta manual un weight arbitrar.

**De unde vine logica**

Nu toate obiectivele au importanță egală. Dacă weight = w, impactul
strategic este proporțional cu w. O scală 1--3 este suficient de
granulară pentru a diferenția, dar suficient de simplă pentru a evita
iluzia de precizie a unei scale 1--10.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Relevance\_adj = round(Relevance\_brut, 2)                              |
|                                                                       |
| Dacă Relevance\_adj < 0.40 → weight = 1                               |
|                                                                       |
| Dacă 0.40 ≤ Relevance\_adj < 0.75 → weight = 2                        |
|                                                                       |
| Dacă Relevance\_adj ≥ 0.75 → weight = 3                                |
|                                                                       |
| Notă implementare: Math.round(score \* 100) / 100                     |
|                                                                       |
| --- nu toFixed(2): produce string, necesită conversie suplimentară    |
+-----------------------------------------------------------------------+

**De ce nu scale 1--5 sau 1--10?**

Scala prea mare creează iluzie de precizie. Scala 1--3 produce claritate
decizională. Un sistem cu 3 GO poate funcționa cu weight 3+2+2=7 (sumă
permisă) sau 3+2+1=6, dar nu 3+3+3=9 (depășire). Diferența între 1 și 3
este semnificativă strategic; diferența între 7 și 8 pe o scală 1--10 nu
ar fi.

**Rol**

* Influențează ALI și capacitatea alocată per GO.
* Influențează Focus Rotation și distribuția atenției zilnice.
* Influențează GORI ponderat și evaluarea strategică anuală.

+-----------------------------------------------------------------------+
| LAYER 0 --- Axiomatic Foundation                                      |
|                                                                       |
| **C8 Priority Balance Constraint**                                    |
+-----------------------------------------------------------------------+

**Definiție**

Limitare structurală care previne supra-încărcarea strategică prin
acumularea simultană a mai multor GO cu prioritate maximă. Suma Priority
Weight-urilor tuturor GO-urilor active simultan nu poate depăși 7.

**De unde vine logica**

Dacă toate GO au weight = 3, sistemul pierde diferențierea strategică.
Dacă toate obiectivele sunt „critice", decizia reală de prioritizare
devine imposibilă, iar ALI și Focus Rotation se distorsionează.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Σ(weight\_i) ≤ 7 pentru toate GO\_i cu status = ACTIVE                  |
|                                                                       |
| Exemplu permis: 3 + 2 + 2 = 7 ✓                                       |
|                                                                       |
| Exemplu interzis: 3 + 3 + 3 = 9 ✗                                     |
|                                                                       |
| Dacă Σ > 7 → auto-rezoluție:                                         |
|                                                                       |
| GO cu Relevance minimă → weight redus automat                         |
|                                                                       |
| La paritate Relevance: GO cu GORI mai mic → reduce primul             |
|                                                                       |
| Trigger universal --- check\_priority\_balance() apelat după ORICE:     |
|                                                                       |
| activare / suspendare / reactivare GO                                 |
|                                                                       |
| modificare Relevance (C11, C18)                                       |
+-----------------------------------------------------------------------+

**Explicație**

Această regulă împiedică utilizatorul să considere toate obiectivele
„critice". Suma maximă de 7 permite o configurație cu un GO dominant
(w=3) și două obiective medii (w=2+2=4). Nu permite trei obiective
simultane de prioritate maximă (3+3+3=9). Verificarea este continuă ---
după orice eveniment care poate modifica suma, nu periodic.

**Rol**

* Menține echilibrul strategic --- forțează decizia reală de
prioritate.
* Previne distorsiunea în ALI, Focus Rotation și GORI.
* Creează ierarhie clară între GO-urile active simultan.

**LEVEL 1 --- STRUCTURAL AUTHORITY**

Structural Authority este nivelul sistemului responsabil pentru
transformarea intenției brute a utilizatorului într-o structură
strategică coerentă și validă.

Acest nivel nu execută obiective și nu monitorizează progresul. Rolul
lui este să clarifice, să filtreze și să limiteze obiectivele înainte ca
acestea să devină parte a arhitecturii executive a sistemului. În
această etapă sistemul transformă intențiile utilizatorului în Global
Objectives valide și stabilește direcția strategică a anului.

**Scopul acestui nivel**

* transformarea intenției brute în obiective clare
* eliminarea ambiguității strategice
* prevenirea supraîncărcării cognitive
* stabilirea unei direcții dominante pentru anul curent
* crearea unei baze stabile pentru execuția ulterioară

**Regula fundamentală**

Un obiectiv nu poate deveni activ dacă nu este clar, limitat și validat
structural. Dacă utilizatorul definește obiective vagi sau
contradictorii, sistemul: clarifică formularea, identifică Behavior
Model dominant, analizează relevanța strategică, limitează numărul de
obiective active. Doar obiectivele validate pot intra în arhitectura de
execuție.

**Relația cu restul sistemului**

Structural Authority urmează după User Intent (Raw Intent Input) și
precede Execution Architecture, Monitoring Authority și Capacity
Regulation. Fără acest nivel, sistemul ar transforma intenții confuze în
execuție instabilă.

**PHASE I --- CLARITY ARCHITECTURE**

Clarity Architecture este mecanismul structural prin care intențiile
nestructurate ale utilizatorului sunt transformate în direcții
strategice formalizate, compatibile cu arhitectura Framework-ului.
Această fază este obligatorie pentru utilizatorii care: nu au claritate
strategică, formulează obiective ambigue, exprimă intenții multiple
simultan sau formulează obiective hibride.

**Scopul acestei faze**

* identificarea direcției dominante
* eliminarea ambiguității
* prevenirea supra-extinderii
* forțarea încadrării într-un Behavior Model

\---

Principiu Structural: Clarity Architecture NU creează obiective
suplimentare. Este o fază de disciplinare, nu de expansiune.

\---

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase I: Clarity Architecture                             |
|                                                                       |
| **C9 Semantic Parsing**                                               |
+-----------------------------------------------------------------------+

**Definiție**

Procesul prin care textul introdus de utilizator este analizat pentru
extragerea elementelor structurale necesare definirii unui Global
Objective valid.

**De unde vine logica**

Intenția umană poate fi exprimată în limbaj vag. Exemplu: „Vreau
libertate financiară." Aceasta nu este măsurabilă și nu poate fi
executată direct. Sistemul trebuie să transforme intenția în variabile
clare înainte de orice evaluare strategică.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| I = text brut introdus                                                |
|                                                                       |
| SP(I) = { domain, direction, metric, timeframe }                      |
|                                                                       |
| Elementele extrase:                                                   |
|                                                                       |
| domain: domeniu (financiar, sănătate, familie, carieră etc.)          |
|                                                                       |
| direction: direcție (creare, creștere, reducere, menținere, evoluție) |
|                                                                       |
| metric: ce anume se schimbă concret și măsurabil                      |
|                                                                       |
| timeframe: orizontul temporal                                         |
|                                                                       |
| Condiții speciale detectate la parsare:                               |
|                                                                       |
| (1) domain=Sănătate\_fizică AND direction=REDUCE AND delta > 25%    |
|                                                                       |
| → SINGLE\_QUESTION\_FLAG = TRUE                                         |
|                                                                       |
| (2) BM\_opus\_1 și BM\_opus\_2 pe aceeași metrică                       |
|                                                                       |
| → GO\_REJECTED\_LOGICAL\_CONTRADICTION                                   |
|                                                                       |
| (3) Parametri insuficienți sau ambigui → Reformulation Queue        |
+-----------------------------------------------------------------------+

**Physical Delta Safety Signal**

Dacă domeniul este Sănătate fizică, direcția este REDUCE și delta
depășește 25% din valoarea inițială, sistemul setează
SINGLE\_QUESTION\_FLAG. La activarea GO-ului utilizatorul primește un
mesaj unic care informează despre importanța consultării unui specialist
--- un singur mesaj, un singur buton de confirmare. Sprint 1, Milestone
1 conține obligatoriu o acțiune de tip conector specialist. Dacă
acțiunea nu este bifată până la finalul Sprint 1, reapare în Sprint 3 o
singură dată.

**GO\_REJECTED\_LOGICAL\_CONTRADICTION**

Detectarea a două BM-uri opuse pe aceeași metrică generează această
categorie specifică. Utilizatorul primește 3 opțiuni predefinite pentru
a rezolva contradicția. Dacă nu răspunde în 48h → PENDING\_CLARIFICATION.
La 7 zile fără răspuns → arhivat automat în Vault. Nu există stare de
limbo.

**Reformulation Queue**

Când Semantic Parsing detectează parametri insuficienți pentru mai multe
GO-uri simultan, cererile de clarificare sunt puse în coadă și
prezentate strict secvențial --- un singur GO la un moment dat, în
ordinea descrescătoare a Relevanței estimate. Dacă utilizatorul nu
răspunde la o cerere în 48h, GO-ul respectiv este arhivat în Vault
automat și coada continuă cu GO-ul următor.

**Rol**

* Elimină ambiguitatea --- creează baza structurată pentru GO valid.
* Permite clasificarea comportamentală în Behavior Model.
* Detectează timpuriu conflictele logice și riscurile speciale.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase I: Clarity Architecture                             |
|                                                                       |
| **C10 Behavior Model Classification**                                 |
+-----------------------------------------------------------------------+

**Definiție**

Atribuirea unui singur Behavior Model dominant fiecărui GO, pe baza
parametrilor extrași de Semantic Parsing.

**De unde vine logica**

Un obiectiv cu două direcții simultane devine instabil. Exemplu: „Vreau
să cresc venit și să reduc programul de lucru." Acesta conține două
modele diferite --- INCREASE și REDUCE. Framework-ul impune alegerea
unui model dominant. Sistemul nu poate genera milestone-uri coerente, nu
poate calcula Sprint Target și nu poate evalua progresul pentru un
obiectiv cu direcție dublă.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| ∀ GO → ∃! Behavior Model (pentru fiecare GO există exact un singur    |
| BM)                                                                   |
|                                                                       |
| BM\_scores = { CREATE: p1, INCREASE: p2, REDUCE: p3, MAINTAIN: p4,     |
| EVOLVE: p5 }                                                          |
|                                                                       |
| BM\_dominant = argmax(BM\_scores)                                       |
|                                                                       |
| confidence ≥ 0.70 → auto-select, fără interacțiune cu utilizatorul    |
|                                                                       |
| confidence ∈ \[0.50,0.70) → Confidence Gate: clarificare utilizator   |
|                                                                       |
| confidence < 0.50 → returnare la C9 Reformulation Queue              |
+-----------------------------------------------------------------------+

**Explicație**

Sistemul întreabă: „Acest obiectiv este în esență despre creare,
creștere, reducere, menținere sau evoluție?" Dacă obiectivul încearcă
să facă mai multe lucruri fundamentale simultan, sistemul îl obligă să
aleagă direcția principală. EVOLVE este selectat automat când analiza
detectează transformare calitativă multi-dimensională.

**Rol**

* Asigură coerență direcțională --- un GO = o direcție.
* Permite formularea corectă și coerentă a milestone-urilor.
* Stabilizează Sprint Target Calculation și calculul efortului.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase I: Clarity Architecture                             |
|                                                                       |
| **C11 Strategic Relevance Scoring**                                   |
+-----------------------------------------------------------------------+

**Definiție**

Evaluarea impactului real al unui GO în contextul strategic al
utilizatorului în următoarele 12 luni. Relevance ∈ \[0, 1], unde 0 =
relevanță foarte scăzută și 1 = relevanță strategică maximă.

**De unde vine logica**

Nu toate dorințele sunt obiective strategice. Unele sunt emoționale sau
temporare. Un obiectiv cu Relevance scăzută ar consuma resurse reale
(timp, energie, capacitate ALI) fără contribuție strategică
semnificativă. Minimum Relevance Floor la 0.30 elimină această categorie
înainte de activare.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Relevance = Impact × 0.35 + Urgency × 0.25 + Alignment × 0.25 +       |
| Feasibility × 0.15                                                    |
|                                                                       |
| Impact: contribuția la obiective de viață de ordin superior           |
|                                                                       |
| Urgency: presiunea temporală și costul amânării                       |
|                                                                       |
| Alignment: coerența cu valorile, identitatea și alte GO active        |
|                                                                       |
| Feasibility: probabilitatea de execuție cu resursele disponibile      |
|                                                                       |
| Relevance\_adj = round(Relevance\_brut, 2) --- înainte de orice mapping |
|                                                                       |
| Minimum Relevance Floor: Relevance\_adj < 0.30 → Vault automat        |
+-----------------------------------------------------------------------+

**Recalibrare Post-Sprint 1**

La finalul Sprint 1, sistemul recalculează Relevance pe baza
comportamentului real de execuție. Dacă utilizatorul a lucrat mai puțin
de 50% din zilele Sprint 1, componentele Urgency și Feasibility sunt
recalibrate descendent. Actualizarea este silențioasă --- nu generează
notificare și nu întrerupe fluxul utilizatorului. Orice modificare a
Relevance declanșează automat check\_priority\_balance() prin C8.

**Interacțiuni cu alte componente**

* Determină Priority Weight (C7, Layer 0).
* Influențează selecția Top-3 (C13).
* Este utilizat în Annual Relevance Recalibration (C18).

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase I: Clarity Architecture                             |
|                                                                       |
| **C12 Resource Conflict Detection**                                   |
+-----------------------------------------------------------------------+

**Definiție**

Identificarea suprapunerii de resurse între GO-uri. Acest mecanism este
un filtru informativ pre-activare și nu înlocuiește ALI Engine. Resource
Conflict detectează suprapuneri calitative între obiective înainte de
activare; ALI Engine efectuează ulterior evaluarea numerică a
capacității reale. Cele două mecanisme sunt complementare.

**De unde vine logica**

Dacă două obiective cer aceleași resurse simultan, conflictul crește și
scade probabilitatea de reușită pentru ambele. Detectarea timpurie
permite ajustarea load-ului înainte ca problema să devină vizibilă în
execuție (Drift, ALI).

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Conflict\_score = Overlap(Timp) × 0.40 + Overlap(Energie) × 0.30 +     |
| Overlap(Capital) × 0.30                                               |
|                                                                       |
| Conflict\_score < 0.40 → activare normală                             |
|                                                                       |
| 0.40 ≤ Conflict\_score < 0.70 → AMBER: load ajustat −15%, fără        |
| blocare                                                               |
|                                                                       |
| Conflict\_score ≥ 0.70 → activare blocată → C17 Future Vault           |
+-----------------------------------------------------------------------+

**Explicație**

Sistemul verifică: dacă două obiective cer același timp, dacă cer
aceeași energie, dacă cer același capital. Dacă suprapunerea este mare,
apare avertizare sau blocare. Ajustarea AMBER (−15% load) este aplicată
silențios în Daily Stack Generator --- utilizatorul vede un plan mai
puțin dens, fără explicație explicită.

**Rol**

* Previne supraîncărcarea --- reduce conflictul de resurse înainte de
activare.
* Ajustează selecția strategică prin influențarea Feasibility în C11.
* Influențează ALI ulterior prin load-ul real generat.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase I: Clarity Architecture                             |
|                                                                       |
| **C13 Top-3 Candidate Selection**                                     |
+-----------------------------------------------------------------------+

**Definiție**

Selectarea a maximum 3 GO cu relevanța cea mai mare, după aplicarea
tuturor filtrelor anterioare.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Pas 1: Filtrare --- elimină toate GO cu Relevance\_adj < 0.30         |
|                                                                       |
| Pas 2: Sortare --- descrescător după Relevance\_adj                    |
|                                                                       |
| Pas 3: Selecție --- primele 3 → ACTIVE, restul → C17 Future Vault     |
+-----------------------------------------------------------------------+

**Explicație**

Obiectivele sunt ordonate după importanța lor strategică calculată.
Primele trei devin active. Restul intră în Future Vault. GO-urile
eliminate la Pasul 1 (Relevance < 0.30) intră în Vault cu mesajul:
„Salvat pentru mai târziu --- scopul acesta nu pare suficient de
important în prezent." GO-urile eliminate la Pasul 3 (Relevance ≥ 0.30
dar al 4-lea sau mai jos) intră în Vault cu mesajul: „Prioritizat pentru
o perioadă viitoare --- ai 3 obiective active acum."

**Rol**

* Aplică constrângerea Maximum 3 Active GO (C3, Layer 0).
* Menține focus strategic --- forțează decizia de prioritizare.

**PHASE II --- STRATEGIC LIMITATION**

Strategic Limitation este mecanismul formal prin care sistemul limitează
utilizatorul la maximum trei Global Objectives active simultan. Aceasta
nu este o sugestie, este o constrângere structurală.

**Scopul acestei faze**

* prevenirea fragmentării atenției
* prevenirea supra-ambiției
* stabilirea unei direcții dominante clare
* protejarea capacității cognitive

**Regula Fundamentală**

Maxim 3 GO active simultan. Dacă utilizatorul definește 5 obiective,
sistemul: solicită prioritizare, recomandă Top 3, mută restul în Future
Vault. Nu este permisă activarea a 4 sau mai multe obiective simultan.

**Validarea GO**

Fiecare GO trebuie: să fie încadrat într-un Behavior Model dominant, să
aibă deadline ≤ 365 zile, să aibă priority\_weight derivat din Relevance
și să fie compatibil cu Strategic Feasibility Analysis (pre-activare).
Dacă nu îndeplinește aceste condiții, GO este respins sau trimis la
reformulare.

**Relația cu restul sistemului**

Strategic Limitation precede: Sprint Design, ALI, Drift și Chaos. Fără
limitare structurală, sistemul devine instabil --- execuția ar porni de
la obiective nevalidate sau contradictorii.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase II: Strategic Limitation                            |
|                                                                       |
| **C14 Global Objective Validation**                                   |
+-----------------------------------------------------------------------+

**Definiție**

Procesul prin care sistemul verifică dacă un GO este structural valid și
eligibil pentru analiza de fezabilitate și activare. Aceasta este ultima
verificare structurală înainte ca un GO să poată deveni activ.

**De unde vine logica**

Dacă un GO este vag, nemăsurabil, nelimitat temporal sau cu model
comportamental ambiguu, atunci orice analiză ulterioară (fezabilitate,
ALI, sprinturi) devine instabilă. Validarea trebuie să fie un filtru
strict înainte de orice activare.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Valid(GO) = TRUE dacă TOATE condițiile sunt satisfăcute:              |
|                                                                       |
| (1) deadline − start\_date ≤ 365 zile                                |
|                                                                       |
| (2) |Behavior Model(GO)| = 1 (unicitate verificată de C10/C15)    |
|                                                                       |
| (3) Metric definit și măsurabil                                     |
|                                                                       |
| (4) Domeniu clar identificat în taxonomy                            |
|                                                                       |
| (5) Valoarea inițială (start) definită                              |
|                                                                       |
| Valid(GO) = FALSE dacă oricare condiție este nesatisfăcută            |
|                                                                       |
| → cerere de completare specifică                                      |
|                                                                       |
| → Timeout 7 zile fără completare → Vault automat                      |
+-----------------------------------------------------------------------+

**Explicație**

Sistemul verifică: Are obiectivul un termen clar? Are o singură direcție
principală? Poate fi măsurat progresul? Este clar în ce domeniu
acționează? Dacă oricare dintre aceste condiții nu este îndeplinită, GO
nu poate trece la următorul pas. Sistemul nu respinge GO-ul definitiv
--- îl suspendă în PENDING\_VALIDATION până la completare.

**Rol**

* Previne activarea obiectivelor abstracte sau incomplet definite.
* Protejează axioma de 365 zile și unicitatea BM.
* Asigură coerența bazei de date necesară pentru Strategic Feasibility
Analysis.

**Interacțiuni cu alte componente**

* Primește input din Phase I (Clarity Architecture).
* Trimite GO validat către Strategic Feasibility Analysis (C16).
* Dacă invalid → cere reformulare specifică.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase II: Strategic Limitation                            |
|                                                                       |
| **C15 Behavior Dominance Enforcement**                                |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care sistemul impune ca fiecare Global Objective să fie
încadrat într-un singur Behavior Model dominant la intrarea în C14
Global Objective Validation.

**De unde vine logica**

Dacă un GO conține mai multe direcții comportamentale simultan: devine
ambiguu, generează milestone-uri contradictorii, produce conflicte de
prioritizare, destabilizează Sprint Target Calculation și afectează
Drift și GORI. Pentru stabilitate matematică și executivă, fiecare GO
trebuie să aibă o singură direcție principală.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| BM(GO) = setul modelelor comportamentale asociate GO                  |
|                                                                       |
| Condiție: |BM(GO)| = 1                                              |
|                                                                       |
| |BM(GO)| = 1 → avansare în C14 Global Objective Validation          |
|                                                                       |
| |BM(GO)| = 0 → returnare la C9 (parsare incompletă)                 |
|                                                                       |
| |BM(GO)| > 1 → GO trebuie reformulat                               |
|                                                                       |
| GO\_REJECTED\_LOGICAL\_CONTRADICTION → blocat, așteptare rezoluție       |
+-----------------------------------------------------------------------+

**Exemplu**

„Vreau să slăbesc și să cresc masă musculară" --- amestec între REDUCE
și INCREASE pe metrici de compoziție corporală. Trebuie stabilit care
este dominant. Dacă ambele sunt la paritate de intensitate pe aceeași
metrică numerică → GO\_REJECTED\_LOGICAL\_CONTRADICTION (C9). Dacă sunt pe
metrici separate → EVOLVE poate fi potrivit.

**Rol**

* Elimină ambiguitatea strategică înainte de activare.
* Permite formularea corectă a milestone-urilor și calcul coerent al
efortului.
* Stabilizează Strategic Feasibility Analysis și comparabilitatea
între GO-uri.

**Interacțiuni cu alte componente**

* Primește date din Behavior Model Classification (C10, Phase I).
* Precede Global Objective Validation (C14).
* Influențează Sprint Target Calculation (C20).

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase II: Strategic Limitation                            |
|                                                                       |
| **C16 Strategic Feasibility Analysis**                                |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care sistemul determină dacă un GO poate fi finalizat
realist în termen de 365 zile, ținând cont de capacitatea utilizatorului
și distribuția resurselor între GO-urile active.

**De unde vine logica**

Axioma 365-Day este rigidă. Pentru a preveni eșecul inevitabil,
carry-over și haos inter-anual, fezabilitatea trebuie evaluată înainte
de activare. Strategic Feasibility Analysis este evaluarea pre-activare
bazată pe estimări. ALI Engine este evaluarea dinamică post-activare
bazată pe execuție reală. Cele două mecanisme sunt complementare și nu
se suprapun.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| E\_total = efort total estimat pentru GO                               |
|                                                                       |
| C\_annual = capacitate anuală estimată a utilizatorului                |
|                                                                       |
| n = numărul de GO active simultan                                     |
|                                                                       |
| C\_per\_GO = C\_annual / n                                               |
|                                                                       |
| Load\_ratio = E\_total / C\_per\_GO                                       |
|                                                                       |
| Load\_ratio ≤ 1.0 → activare normală                                   |
|                                                                       |
| Load\_ratio ∈ (1.0, 1.10] → Ambition Buffer: activare cu avertisment  |
|                                                                       |
| Load\_ratio > 1.10 → capacitate insuficientă → C17 Future Vault       |
|                                                                       |
| Capacity Validation Gate:                                             |
|                                                                       |
| C\_daily < 0.5 h/zi → avertisment de subcapacitate (nu blocare)       |
|                                                                       |
| C\_daily > 14 h/zi → avertisment de supracapacitate (nu blocare)      |
+-----------------------------------------------------------------------+

**Explicație**

Sistemul verifică: „Poți termina acest obiectiv în 1 an cu resursele
tale actuale?" Dacă obiectivul cere mai mult decât poate fi realizat,
nu este activat în forma actuală. Intervalul (1.0, 1.10] este Ambition
Buffer --- o planificare ușor optimistă este acceptabilă. Dacă load-ul
nu scade în Sprint 1, recalibrarea post-Sprint1 din C11 și C20 va ajusta
organic.

**Rol**

* Protejează axioma anuală --- previne activarea obiectivelor
nefinalizabile în 365 zile.
* Reduce eșecul structural și stabilizează GORI.
* Elimină necesitatea carry-over inter-anual.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase II: Strategic Limitation                            |
|                                                                       |
| **C17 Future Vault System**                                           |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul structural prin care sistemul stochează GO care nu pot deveni
active în momentul curent, fără a le elimina din sistem. Un GO poate
intra în Vault din mai multe motive: depășește limita de 3 GO active, nu
trece Strategic Feasibility Analysis, Relevance < 0.30, sau timeout-uri
de clarificare/validare depășite.

**De unde vine logica**

Sistemul trebuie să păstreze ideile utilizatorului, dar să mențină
stabilitatea structurală. Eliminarea definitivă ar genera frustrare.
Activarea directă cu toate obiectivele dorite ar genera haos. Vault este
mecanismul de echilibru: „Nu acum. Poate mai târziu."

**Reprezentare formală**

+-----------------------------------------------------------------------+
| GO.status = VAULT dacă oricare din condițiile de intrare:             |
|                                                                       |
| (1) |Active\_GO| = 3 și tentativă de activare nouă (C3)            |
|                                                                       |
| (2) Load\_ratio > 1.10 după Feasibility Analysis (C16)              |
|                                                                       |
| (3) Relevance\_adj < 0.30 după scoring (C11)                        |
|                                                                       |
| (4) GO\_REJECTED\_LOGICAL\_CONTRADICTION fără răspuns 7 zile (C9)      |
|                                                                       |
| (5) Reformulation Queue abandonată > 48h (C9)                      |
|                                                                       |
| (6) PENDING\_VALIDATION > 7 zile fără completare (C14)              |
+-----------------------------------------------------------------------+

**Semantica VAULT vs alte statusuri**

\---

**Status**           **Semnificație**      **Reactivare**

**VAULT**            Niciodată activ ---   Când slot disponibil sau Load\_ratio ≤
așteptare strategică  1.10

**SUSPENDED**        A fost activ ---      Prin C36 Reactivation Protocol, cu
suspendat de sistem   Relevance ≥ 0.60
prin SRM L3

**SEASONAL\_PAUSE**   Activ sezonier ---    Automat la intrarea în
inactiv în fereastra  execution\_window activă
curentă

**ARCHIVED**         Închis definitiv de   Nu se reactivează
utilizator

\---

**Condiții formale GO în Vault**

* GO în Vault nu este inclus în ALI.
* GO în Vault nu are Sprint activ.
* GO în Vault poate fi activat doar dacă există slot liber (n < 3).
* GO în Vault este reevaluat la recalibrarea periodică (C18).

**Rol**

* Protejează limita de 3 GO --- menține stabilitate anuală.
* Previne activarea obiectivelor imposibil de finalizat în 365 zile.
* Permite restructurare strategică: obiectivele stocate nu sunt
pierdute.

**Interacțiuni cu alte componente**

* Primește GO din Top-3 Selection (C13) când limita este depășită.
* Primește GO din Strategic Feasibility Analysis (C16) când nerealist.
* Nu influențează ALI sau Drift --- GO în Vault este exclus complet
din calcule operative.

+-----------------------------------------------------------------------+
| LEVEL 1 --- Phase II: Strategic Limitation                            |
|                                                                       |
| **C18 Annual Relevance Recalibration**                                |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care sistemul reevaluează periodic relevanța strategică
a fiecărui Global Objective, indiferent dacă este activ sau stocat în
Future Vault. Scopul este să determine dacă obiectivul rămâne strategic
valid, trebuie reformulat, trebuie închis sau trebuie eliminat.

**De unde vine logica**

Obiectivele pot deveni irelevante din cauza schimbării contextului
personal, a resurselor, a priorităților sau a finalizării parțiale care
modifică direcția. Fără recalibrare periodică: Vault devine depozit
inert, GO active pot continua mecanic fără relevanță, sistemul devine
rigid operațional.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| R\_current = scorul de relevanță actual                                |
|                                                                       |
| R\_initial = scorul de relevanță la activare                           |
|                                                                       |
| Relevance\_ratio = R\_current / R\_initial                               |
|                                                                       |
| Relevance\_ratio ≥ 0.70 → GO rămâne activ fără intervenție             |
|                                                                       |
| 0.40 ≤ Relevance\_ratio < 0.70 → REVIEW\_REQUIRED: o întrebare         |
|                                                                       |
| Relevance\_ratio < 0.40 → recomandare formală de închidere sau Vault  |
|                                                                       |
| Prioritate: dacă GO.status = SUSPENDED → C34 Stabilization Review     |
|                                                                       |
| are prioritate față de C18                                            |
+-----------------------------------------------------------------------+

**Explicație**

Sistemul verifică: „Este acest obiectiv la fel de important ca atunci
când a fost creat?" Dacă relevanța scade semnificativ (sub aproximativ
70% din importanța inițială), sistemul recomandă reformulare, închidere
sau înlocuire. Recalibrarea este advisory --- nu modifică automat
statusul fără confirmare utilizator, cu excepția cazurilor severe
(Relevance\_ratio < 0.40, care generează recomandare formală).

**Rol**

* Previne menținerea obiectivelor depășite sau devenite irelevante.
* Curăță Future Vault --- forțează decizii strategice conștiente
periodic.
* Menține coerența anuală a sistemului pe termen lung.

**LEVEL 2 --- EXECUTION AUTHORITY**

Execution Authority este nivelul sistemului responsabil pentru
transformarea obiectivelor strategice validate în execuție operațională
structurată.

Acest nivel nu definește obiective și nu reglează capacitatea
utilizatorului. Rolul lui este să creeze arhitectura executivă prin care
obiectivele sunt realizate în mod sistematic. Execution Authority
stabilește ritmul execuției și structura progresului prin Sprint-uri,
Milestone-uri și execuție zilnică.

**Scopul acestui nivel**

* transformarea obiectivelor strategice în acțiune concretă
* stabilirea unui ritm executiv constant
* fragmentarea controlată a obiectivelor
* crearea unei structuri clare de progres
* generarea datelor necesare pentru monitorizare

**Regula fundamentală**

Execuția este structurată și ciclică. Fiecare Global Objective activ
este împărțit în Sprint-uri fixe de 30 zile, fiecare Sprint are un
target măsurabil, fiecare Sprint este structurat în Milestone-uri,
execuția zilnică derivă din structura sprintului. Execuția liberă, fără
structură, nu este permisă.

**Relația cu restul sistemului**

Execution Authority urmează după Structural Authority și precede
Monitoring Authority și Capacity Regulation. Datele generate de execuție
sunt utilizate pentru Drift, Chaos, ALI și Reinforcement Model.

**PHASE III --- EXECUTION ARCHITECTURE**

Execution Architecture este mecanismul formal prin care sistemul
transformă Global Objectives active în execuție ciclică structurată pe
Sprint-uri fixe de 30 zile. Aceasta nu este o listă de task-uri, ci o
arhitectură operațională standardizată.

**Scopul acestei faze**

* fragmentarea controlată a obiectivelor anuale
* standardizarea progresului
* permiterea calculului deviației (Drift)
* prevenirea stagnării
* crearea bazei pentru evaluarea strategică (Sprint Score și GORI)

**Regula Fundamentală**

Execuția este ciclică și standardizată. Fiecare GO activ este împărțit
în Sprint-uri fixe de 30 zile, fiecare Sprint are un target măsurabil,
fiecare Sprint este structurat în maximum 5 Milestone-uri, execuția
zilnică derivă din Milestone-ul activ. Nu sunt permise Sprint-uri cu
durată variabilă. Nu sunt permise Milestone-uri nelimitate numeric. Nu
sunt permise Daily Stack independent de structură.

**Relația cu restul sistemului**

Execution Architecture urmează după Strategic Limitation și Strategic
Feasibility Analysis. Precede Drift, Chaos, ALI (calcul real) și
Reinforcement Model. Fără această fază, obiectivul rămâne strategic dar
neexecutabil.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C19 Sprint Structuring --- 30-Day Cycle**                           |
+-----------------------------------------------------------------------+

**Definiție**

Structura ciclică fixă prin care fiecare GO este împărțit în unități de
execuție de 30 zile.

**De unde vine logica**

Dacă durata sprintului este variabilă: progresul devine imposibil de
comparat, Drift devine instabil, GORI devine distorsionat.
Standardizarea la 30 de zile permite măsurare uniformă.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Sprint\_length = 30 (invariant)                                        |
|                                                                       |
| t ∈ { 1, 2, 3, ..., 30 } --- integer                                 |
|                                                                       |
| Expected(t) = t / 30                                                  |
|                                                                       |
| Statusuri posibile sprint:                                            |
|                                                                       |
| ACTIVE --- sprint în curs de execuție                                 |
|                                                                       |
| COMPLETED --- finalizat normal la t=30                                |
|                                                                       |
| SUSPENDED --- GO suspendat prin SRM L3 în mijlocul sprintului         |
|                                                                       |
| SEASONAL\_PAUSE --- GO sezonier intrat în fereastră inactivă (patch    |
| Rev. 5.6)                                                             |
+-----------------------------------------------------------------------+

**Execution Windows și SEASONAL\_PAUSE**

Un GO cu execution\_windows definite (exemplu: activ noiembrie--martie
pentru un obiectiv sezonier) intră automat în SEASONAL\_PAUSE la ieșirea
din fereastra activă. La intrarea în SEASONAL\_PAUSE: Expected(t) este
înghețat la valoarea curentă, progresul acumulat este conservat, Drift
nu crește negativ, Stagnation Detection este dezactivat, ALI nu
înregistrează ore pentru acel GO. La re-intrarea în fereastra activă:
Expected(t) reia avansarea, Sprint Planning generează sprint nou cu
target recalibrat.

+-----------------------------------------------------------------------+
| SEASONAL\_PAUSE nu este un SUSPENDED: GO-ul nu a eșuat, nu a necesitat |
| intervenție sistemică. Este o pauză structurală planificată, cu       |
| semantică proprie.                                                    |
|                                                                       |
| Sprinturile cu status SEASONAL\_PAUSE sunt excluse din                 |
| Continuity\_factor în GORI (C38) --- identic cu SUSPENDED.             |
+-----------------------------------------------------------------------+

**Rol**

* Permite progres liniar și calcul standardizat al Drift.
* Permite comparabilitate istorică între sprinturi.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C20 Sprint Target Calculation**                                     |
+-----------------------------------------------------------------------+

**Definiție**

Procesul prin care este stabilit rezultatul măsurabil al fiecărui
Sprint, derivat din obiectivul anual și din progresul rămas de realizat.

**De unde vine logica**

Un Sprint fără target clar produce execuție difuză, creează
milestone-uri incoerente și face progresul imposibil de evaluat.
Targetul trebuie să fie ambițios dar realizabil --- de aici factorul
0.80 și plafonul de compensație.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Sprint\_Target\_brut = (Target\_anual − Progress\_actual) /               |
| Sprinturi\_rămase                                                      |
|                                                                       |
| Sprint\_Target\_realist = Sprint\_Target\_brut × 0.80                     |
|                                                                       |
| Compensație după Regression Event:                                    |
|                                                                       |
| Sprint\_Target\_compensat ≤ Sprint\_Target\_inițial × 1.50 (plafon)       |
|                                                                       |
| Dacă 3+ Regression Events consecutive:                                |
|                                                                       |
| → propunere formală de reducere Target\_anual                          |
+-----------------------------------------------------------------------+

**Plafonul de Compensație 1.5×**

Fără plafon, un Regression Event major ar genera un Sprint Target de
recuperare de 200--300% din normal --- un plan imposibil care duce fie
la abandon, fie la bifat mecanic. Plafonul la 1.5× produce un target
ambițios dar realist, iar diferența de recuperare se distribuie pe
sprinturile rămase. La 3+ Regression Events consecutive, sistemul nu
penalizează utilizatorul --- propune o reducere a Target-ului anual ca
semnal că targetul inițial a fost supraestimat.

**Rol**

* Definește direcția și ambiția sprintului.
* Creează baza pentru structurarea Milestone-urilor.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C21 80% Probability Rule**                                          |
+-----------------------------------------------------------------------+

**Definiție**

Regulă prin care targetul sprintului este definit la un nivel realist de
realizare --- aproximativ 80% probabilitate de succes --- nu la maximul
teoretic.

**De unde vine logica**

Targeturi setate la nivel ideal (100% ambiție) cresc rata de eșec, cresc
Drift negativ și cresc riscul de abandon. Planificarea realistă
stabilizează sistemul și protejează Momentum-ul pe termen lung.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Sprint\_Target\_aplicat = Sprint\_Target\_brut × 0.80                     |
|                                                                       |
| Pentru metrici continue (greutate, MRR, ore):                         |
|                                                                       |
| factorul 0.80 se aplică direct pe delta rămasă                        |
|                                                                       |
| Pentru metrici discrete (nr. clienți, nr. certificări):               |
|                                                                       |
| Sprint 1--2: 0--1 unitate                                             |
|                                                                       |
| Sprinturi medii: 1--2 unități/sprint                                  |
|                                                                       |
| Sprint final: unitățile rămase pentru target total                    |
+-----------------------------------------------------------------------+

**Rol**

* Reduce eșecul repetitiv --- stabilizează execuția pe termen lung.
* Protejează Momentum-ul și motivația utilizatorului.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C22 Milestone Structuring (max 5)**                                 |
+-----------------------------------------------------------------------+

**Definiție**

Fragmentarea Sprint Target în maximum 5 praguri intermediare de progres.
Milestone-urile sunt subobiective intermediare verificabile --- puncte
de control care fac progresul vizibil în cadrul sprint-ului.

**De unde vine logica**

Prea multe milestone-uri produc fragmentare excesivă și transformă
sistemul într-un task tracker. Prea puține lasă lipsa de claritate
progresivă. Limitarea la maximum 5 păstrează echilibrul și oferă
feedback de completare la momente semnificative.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| 1 ≤ |Milestone\_set| ≤ 5 per sprint                                  |
|                                                                       |
| Ordonare milestone-uri:                                               |
|                                                                       |
| ordered = false (default):                                            |
|                                                                       |
| Milestone-urile pot fi bifate în orice ordine                         |
|                                                                       |
| ordered = true:                                                       |
|                                                                       |
| Milestone-urile au dependențe logice --- M2 nu poate fi bifat         |
|                                                                       |
| înainte de M1. Bifarea out-of-order generează avertisment,            |
|                                                                       |
| nu blocare. Utilizatorul poate forța ordinea dacă dorește.            |
|                                                                       |
| Flag-ul ordered este opțional și se setează per sprint, nu per GO     |
+-----------------------------------------------------------------------+

**Rol**

* Structurează progresul --- permite vizualizarea etapelor
intermediare.
* Permite Sprint Score să reflecte granularitatea execuției.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C23 Daily Stack Generator**                                         |
+-----------------------------------------------------------------------+

**Definiție**

Daily Stack este setul de acțiuni zilnice recomandate utilizatorului,
compus din două componente distincte: Core Stack (structural,
obligatoriu) și Optional Stack (operațional, suplimentar).

**Core Stack și Optional Stack**

Core Stack reprezintă acțiunile generate automat din Milestone-ul activ
și constituie execuția oficială a progresului. Optional Stack reprezintă
acțiuni adăugate manual de utilizator pentru a susține sau accelera
progresul, fără a modifica structura oficială a sprintului.

**De unde vine logica**

Daily Stack nu este o listă liberă de sarcini. Este derivat din
structura sprintului. Dacă este independent de Sprint și Milestone,
sistemul devine task tracker. Separarea dintre Core Stack și Optional
Stack permite menținerea disciplinei structurale și a flexibilității
operaționale simultan.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Daily\_Stack = Core\_Stack ∪ Optional\_Stack                             |
|                                                                       |
| Core\_Stack = funcție(Milestone\_activ, Context)                        |
|                                                                       |
| Optional\_Stack = acțiuni\_adăugate\_manual (limitate numeric)           |
|                                                                       |
| Timestamping obligatoriu pentru fiecare acțiune:                      |
|                                                                       |
| Late completion (> 48h după ziua programată):                        |
|                                                                       |
| → Progress\_comp: YES (acțiunea contribuie la progres)                 |
|                                                                       |
| → Consistency\_comp: NO (nu demonstrează execuție zilnică)             |
|                                                                       |
| Bulk detection:                                                       |
|                                                                       |
| > 5 acțiuni completate în < 10 minute → C39 Proxy 1 flag            |
|                                                                       |
| Physical Delta Safety --- Sprint 1, Milestone 1:                      |
|                                                                       |
| Dacă SINGLE\_QUESTION\_FLAG=TRUE → acțiune obligatorie "Consultă       |
| specialist"                                                          |
+-----------------------------------------------------------------------+

**Rol**

* Traduce strategia în execuție zilnică concretă.
* Menține coerența direcțională --- fiecare acțiune derivă din
Milestone activ.
* Permite flexibilitate operațională controlată prin Optional Stack.
* Previne transformarea sistemului într-un task tracker.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C24 Progress Computation Engine**                                   |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care este calculat progresul real al Sprintului, ca
rație dintre milestone-urile finalizate și milestone-urile totale.

**De unde vine logica**

Progresul nu poate fi intuitiv. Trebuie să fie numeric și comparabil cu
progresul așteptat pentru a putea alimenta Drift și a permite evaluarea
obiectivă a execuției.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Real\_Progress = milestone\_finalizate / milestone\_totale               |
|                                                                       |
| Real\_Progress = clamp(Real\_Progress, 0, 1)                            |
|                                                                       |
| Acțiunile din Optional Stack contribuie la progres DOAR dacă          |
|                                                                       |
| sunt marcate explicit ca suport direct pentru Milestone-ul activ.     |
|                                                                       |
| Regression Event:                                                     |
|                                                                       |
| Metrica GO scade sub valoarea de la startul sprintului curent         |
|                                                                       |
| → Real\_Progress = 0                                                   |
|                                                                       |
| → regression\_flag = TRUE                                              |
|                                                                       |
| → SRM L1 activat IMEDIAT (fără așteptarea a 3 zile consecutive)       |
|                                                                       |
| → Sprint Target compensat ≤ 1.5× (C20)                                |
+-----------------------------------------------------------------------+

**Notă Late Completion**

Acțiunile bifate ca Late Completion (> 48h) contribuie la Real\_Progress
--- sunt incluse în calculul milestone\_finalizate. Nu contribuie la
Consistency\_comp (C37). Distincția reflectă realitatea: dacă o acțiune a
fost executată, valoarea sa strategică este reală chiar dacă a venit cu
întârziere.

**Rol**

* Permite compararea cu Expected(t) --- baza calculului Drift.
* Alimentează Stability Control Layer și toate calculele de
performanță.

+-----------------------------------------------------------------------+
| LEVEL 2 --- Execution Authority                                       |
|                                                                       |
| **C25 Execution Variance Tracker**                                    |
+-----------------------------------------------------------------------+

**Definiție**

Mecanism backend care calculează diferența dintre progresul real și
progresul așteptat și o transmite Stability Control Layer. Este
componenta de legătură între execuția zilnică și mecanismele de
monitorizare.

**De unde vine logica**

Fără măsurarea deviației: stagnarea nu poate fi detectată, intervenția
devine întârziată și progresul aparent poate masca probleme reale. C25
nu ia decizii --- calculează și transmite.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Drift\_raw = Real\_Progress − Expected(t)                               |
|                                                                       |
| unde Expected(t) = t / 30, t ∈ { 1..30 }                              |
|                                                                       |
| Dacă progresul real < progresul planificat → Drift\_raw negativ       |
|                                                                       |
| Dacă progresul real > progresul planificat → Drift\_raw pozitiv       |
|                                                                       |
| Acțiunile din Optional Stack nu influențează calculul Drift           |
|                                                                       |
| decât dacă sunt validate ca suport direct pentru Milestone-ul activ.  |
+-----------------------------------------------------------------------+

**Rol**

* Detectează deviația executivă --- baza pentru intervenție timpurie.
* Alimentează Stability Control Layer (C26, C27, C28).
* Poate declanșa ajustări strategice prin SRM.

**LEVEL 3 --- MONITORING AUTHORITY**

Monitoring Authority este mecanismul formal prin care sistemul
monitorizează stabilitatea executivă și capacitatea strategică a
utilizatorului în timpul execuției.

Acest nivel nu creează obiective și nu execută task-uri. Rolul lui este
să observe, să măsoare și să intervină atunci când apare instabilitate.

**Scopul acestui nivel**

* detectarea deviației față de plan
* identificarea stagnării
* măsurarea haosului executiv
* reglarea capacității reale
* prevenirea supraîncărcării

**Regula fundamentală**

Execuția este permisă doar în condiții de stabilitate controlată. Dacă
sistemul detectează: deviație excesivă, stagnare prelungită, haos global
sau depășirea capacității --- intervine prin mecanismele de control.
Monitoring Authority nu modifică structura anuală --- ajustează
comportamentul executiv.

**Relația cu restul sistemului**

Monitoring Authority urmează după Execution Architecture și influențează
Adaptive Context Engine, Strategic Reset Mode, Velocity Control și
Reinforcement Model. Fără monitorizare, execuția devine instabilă în
timp.

**STABILITY CONTROL LAYER**

Acest sub-bloc monitorizează stabilitatea execuției. Componentele
C26--C29 detectează deviația, stagnarea, haosul global și gestionează
distribuția atenției.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Stability Control Layer                                   |
|                                                                       |
| **C26 Dynamic Drift Engine**                                          |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul care calculează deviația dintre progresul așteptat și
progresul real într-un Sprint și monitorizează acumularea acestei
deviații în timp.

**De unde vine logica**

Fără măsurarea deviației: stagnarea este invizibilă, intervenția este
întârziată, progresul devine iluzoriu. Drift reprezintă diferența dintre
ce ar fi trebuit realizat și ce s-a realizat efectiv --- semnal direct
de sănătate a execuției.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Drift(GO\_i, t) = Real\_Progress(GO\_i) − Expected(t)                    |
|                                                                       |
| Expected(t) = t / 30, t ∈ { 1..30 }                                   |
|                                                                       |
| Drift pozitiv: progres real > planificat (performanță peste          |
| așteptări)                                                            |
|                                                                       |
| Drift negativ: progres real < planificat (întârziere acumulată)      |
|                                                                       |
| Trigger SRM L1:                                                       |
|                                                                       |
| Drift(GO\_i) < −0.15 pentru 3 zile CONSECUTIVE → SRM L1               |
|                                                                       |
| SAU regression\_flag = TRUE din C24 → SRM L1 IMEDIAT                   |
|                                                                       |
| Expected(t) ÎNGHEȚAT în:                                              |
|                                                                       |
| Core Stabilization (C35), Planned Pause (C32)                         |
|                                                                       |
| External Shock Buffer (C32), Reactivation Protocol (C36)              |
|                                                                       |
| SEASONAL\_PAUSE (C19)                                                  |
+-----------------------------------------------------------------------+

**Explicație**

Dacă progresul real este mai mic decât cel planificat, Drift devine
negativ și indică o întârziere. Dacă progresul real depășește planul,
Drift devine pozitiv și indică performanță peste așteptări. Pragul de
−0.15 corespunde unui decalaj de aproximativ 4--5 zile de execuție față
de traiectoria liniară --- semnal că GO-ul pierde ritm sistematic, nu că
a avut o zi slabă.

Acțiunile din Optional Stack nu influențează calculul Drift decât dacă
sunt validate ca direct legate de Milestone-ul activ. Această regulă
previne distorsionarea măsurării progresului prin activități secundare.

**Înghețarea Expected(t)**

Înghețarea Expected(t) previne acumularea Drift negativ în perioadele
când sistemul știe că execuția nu poate sau nu trebuie să aibă loc. Fără
înghețare, Drift-ul ar escalada mecanic și ar declanșa SRM în perioadele
de protecție --- un loop paradoxal.

**Rol**

* Identifică întârzierile --- permite intervenție timpurie.
* Alimentează SRM și Chaos Index Engine.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Stability Control Layer                                   |
|                                                                       |
| **C27 Stagnation Detection Engine**                                   |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul care detectează absența progresului pe o perioadă relevantă.
Spre deosebire de Drift (care măsoară decalajul față de traiectoria
așteptată), Stagnation detectează platoul complet --- utilizatorul a
încetat să genereze orice progres.

**De unde vine logica**

Progres zero pe termen lung indică blocaj psihologic, supraîncărcare sau
obiectiv nerealist. Aceasta indică lipsă de avans, chiar dacă nu există
regres activ.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Stagnant(GO\_i) = TRUE dacă:                                           |
|                                                                       |
| Real\_Progress(t2) − Real\_Progress(t1) = 0 pe un interval relevant     |
|                                                                       |
| Threshold standard: 5 zile consecutive de zero progres                |
|                                                                       |
| Threshold ESB: 10 zile consecutive (External Shock Buffer activ)      |
|                                                                       |
| DEZACTIVAT în:                                                        |
|                                                                       |
| Planned Pause, Crisis Protocol, ESB, Core Stabilization,              |
| SEASONAL\_PAUSE                                                        |
+-----------------------------------------------------------------------+

**Rol**

* Previne stagnarea prelungită --- poate declanșa Focus Rotation.
* Poate activa SRM dacă stagnarea persistă după threshold.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Stability Control Layer                                   |
|                                                                       |
| **C28 Chaos Index Engine**                                            |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul care măsoară instabilitatea globală a execuției. Haosul nu
este lipsa progresului, ci dezorganizarea globală. Poate exista progres
cu Chaos Index ridicat.

**De unde vine logica**

Un singur indicator (Drift sau Stagnation) nu poate surprinde
complexitatea instabilității. Chaos Index agregă semnalele din toate
sursele --- Drift, Stagnare, Inconsistență și Context extern --- într-un
singur indice care determină nivelul de intervenție necesar.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Chaos\_Index = Drift\_comp × 0.30 + Stagnation\_comp × 0.25              |
|                                                                       |
| + Inconsistency\_comp × 0.25 + Context\_disruption × 0.20              |
|                                                                       |
| Drift\_comp = max(|Drift(GO\_i)|) pentru toate GO\_i ACTIVE            |
|                                                                       |
| (cel mai slab GO determină --- nu media)                              |
|                                                                       |
| Stagnation\_comp = { 0 dacă nicio stagnare | 0.5 dacă 1 GO | 1.0     |
| dacă 2+ GO }                                                          |
|                                                                       |
| Inconsistency\_comp = Variance(Completion\_rate\_zilnic, ultimele 14     |
| zile)                                                                 |
|                                                                       |
| Context\_disruption = min(1.0, nr\_eventi\_majori / 3)                   |
|                                                                       |
| Praguri intervenție:                                                  |
|                                                                       |
| Chaos\_Index < 0.30 → Verde: sistem stabil                            |
|                                                                       |
| 0.30--0.40 → Galben: monitorizare crescută                            |
|                                                                       |
| 0.40--0.60 → Amber: SRM L2 recomandat                                 |
|                                                                       |
| Chaos\_Index ≥ 0.60 → Roșu: SRM L3 recomandat                          |
+-----------------------------------------------------------------------+

**Principiul Conservativ --- max() în loc de medie**

Drift\_comp folosește maximul valorilor absolute ale Drift-ului per GO,
nu media. Un sistem cu un GO la Drift −0.05 și un al doilea la Drift
−0.80 are Drift\_comp = 0.80, nu 0.425. Media ar masca problemele grave.
Cel mai slab element determină sănătatea sistemului.

**Context\_disruption**

Plafonul la min(1.0, nr\_eventi/3) asigură că Chaos\_Index rămâne
calculabil chiar în perioadele cu număr mare de evenimente
perturbatoare. Chaos\_Index în sine rămâne neclamped --- poate depăși 1.0
teoretic pentru a semnaliza magnitudinea reală a instabilității.

**Rol**

* Detectează dezorganizarea sistemică --- poate activa Core
Stabilization Mode.
* Influențează Velocity Control și nivelul de intervenție SRM.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Stability Control Layer                                   |
|                                                                       |
| **C29 Focus Rotation Logic**                                          |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care atenția strategică este redistribuită între GO
active, balansând prioritatea strategică cu starea curentă a execuției.

**De unde vine logica**

Dacă toate GO primesc atenție egală permanent, apare diluare, apare
supraîncărcare și scade eficiența. Obiectivele cu prioritate mai mare
trebuie să primească mai multă atenție; obiectivele cu deviații mari
trebuie să primească atenție suplimentară pentru recuperare.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Focus\_weight\_brut(GO\_i) = w(GO\_i) / (1 + |Drift(GO\_i)|)             |
|                                                                       |
| Normalizare exclusiv pe GO cu status = ACTIVE:                        |
|                                                                       |
| Focus\_norm(GO\_i) = Focus\_weight\_brut(GO\_i) / Σ(                       |
| Focus\_weight\_brut(GO\_j) )                                             |
|                                                                       |
| unde j parcurge EXCLUSIV GO\_j.status = ACTIVE                         |
|                                                                       |
| GO SUSPENDED, VAULT, SEASONAL\_PAUSE: excluse complet din Σ            |
|                                                                       |
| Garanție minimă: min 1 acțiune/zi/GO activ, indiferent de Focus\_norm  |
+-----------------------------------------------------------------------+

**Excluderea GO SUSPENDED din Normalizare**

Includerea unui GO SUSPENDED în normalizare ar produce o distribuție
distorsionată: GO-ul ar consuma o fracție din Σ fără a genera acțiuni
reale, iar celelalte GO active ar primi mai puțin decât justifică
greutățile lor relative. Excluderea completă asigură că Focus Rotation
reflectă realitatea operațională.

**Rol**

* Redistribuie efortul --- protejează GO dominante strategice.
* Stabilizează execuția --- ajustează atenția în funcție de starea
reală a fiecărui GO.

**CAPACITY REGULATION LAYER**

Acest sub-bloc reglează încărcarea anuală și ritmul de execuție.
Componentele C30--C31 monitorizează sarcina reală față de capacitate și
ajustează viteza de execuție proporțional.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Capacity Regulation Layer                                 |
|                                                                       |
| **C30 ALI Engine --- Annual Load Index**                              |
+-----------------------------------------------------------------------+

**Definiție**

Mecanism numeric care evaluează încărcarea anuală reală comparativ cu
capacitatea utilizatorului. ALI detectează supraîncărcarea reală și
proiectată, nu doar estimată inițial.

**De unde vine logica**

Estimarea inițială poate fi greșită. Execuția reală poate consuma mai
mult efort decât planificat. ALI detectează supraîncărcarea reală în
timp real, nu retrospectiv.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| ALI\_curent = Ore\_acumulate\_total / C\_annual                           |
|                                                                       |
| ALI\_proiectat = ALI\_curent × (365 / Zile\_scurse)                      |
|                                                                       |
| Final\_ALI = ALI\_proiectat × Execution\_Reliability                     |
|                                                                       |
| Execution\_Reliability: factor de corecție bazat pe rata istorică de   |
| finalizare                                                            |
|                                                                       |
| (floor 0.6 --- nu scade sub 0.6 pentru a evita suprapenalizarea)      |
|                                                                       |
| Praguri intervenție:                                                  |
|                                                                       |
| Final\_ALI ≤ 1.0 → capacitate normală                                  |
|                                                                       |
| Final\_ALI ∈ (1.0, 1.10] → Ambition Buffer: avertisment, activare     |
| permisă                                                               |
|                                                                       |
| Final\_ALI > 1.10 → Core Stabilization recomandat                     |
|                                                                       |
| Trigger-urile ALI dezactivate în primele 14 zile de utilizare         |
|                                                                       |
| GO SUSPENDED sau SEASONAL\_PAUSE: excluse din calculul orelor          |
| consumate                                                             |
+-----------------------------------------------------------------------+

**ALI\_curent vs ALI\_proiectat**

ALI\_curent este o metrică retrospectivă --- câtă capacitate a fost
consumată până acum. ALI\_proiectat este metrica de decizie --- dacă
utilizatorul menține ritmul actual tot anul, va depăși capacitatea?
Deciziile sistemice se bazează pe ALI\_proiectat. Un utilizator cu
ALI\_curent = 0.6 la jumătatea anului poate părea ok, dar dacă ritmul
curent este intensificat, ALI\_proiectat poate fi 1.2 --- sistem în risc.

**Rol**

* Previne supraîncărcarea anuală --- poate declanșa Velocity Control.
* Poate activa Core Stabilization Mode când Final\_ALI > 1.10.

+-----------------------------------------------------------------------+
| LEVEL 3 --- Capacity Regulation Layer                                 |
|                                                                       |
| **C31 Velocity Control Mechanism**                                    |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul care ajustează ritmul de execuție în funcție de capacitatea
detectată de ALI Engine. Ritmul constant în condiții de supraîncărcare
produce burnout.

**De unde vine logica**

Dacă utilizatorul este supraîncărcat și ritmul nu scade, execuția devine
nesustenabilă. Reducerea proporțională a ritmului protejează
sustenabilitatea pe termen lung și stabilizează ALI.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Capacity\_factor = 1.0 − max(0, Final\_ALI − 1.0)                       |
|                                                                       |
| Velocity\_ajustat = Velocity\_base × Capacity\_factor                    |
|                                                                       |
| Exemplu: Final\_ALI = 1.08                                             |
|                                                                       |
| Capacity\_factor = 1.0 − 0.08 = 0.92                                   |
|                                                                       |
| Velocity\_ajustat = Velocity\_base × 0.92 (−8% față de normal)          |
+-----------------------------------------------------------------------+

Reducerea este progresivă și proporțională. La Final\_ALI = 1.10 (pragul
maxim al Ambition Buffer), reducerea este de 10%. Ajustarea este
aplicată silențios în Daily Stack Generator --- utilizatorul vede un
plan mai puțin dens.

**Rol**

* Reduce intensitatea temporar --- protejează sustenabilitatea
execuției.
* Stabilizează ALI --- previne escaladarea spre Core Stabilization
Mode.

**LEVEL 4 --- REGULATORY AUTHORITY**

Regulatory Authority este mecanismul formal prin care sistemul adaptează
execuția la contextul real al utilizatorului fără a modifica structura
fundamentală a Framework-ului.

Acest nivel nu creează obiective și nu redefinește strategia anuală.
Rolul lui este să regleze execuția atunci când apar schimbări în
contextul utilizatorului.

**Contextul poate include**

* variații de energie
* evenimente externe
* perioade planificate de pauză
* crize personale sau profesionale
* fluctuații de motivație

**Scopul acestui nivel**

* protejarea stabilității pe termen lung
* prevenirea burnout-ului
* menținerea continuității execuției
* adaptarea temporară la context
* prevenirea abandonului sistemului

**Regula Fundamentală**

Contextul poate modifica ritmul execuției, dar nu poate modifica
structura sistemului. Sistemul permite: ajustarea intensității,
suspendarea temporară controlată, reducerea temporară a obiectivelor
operaționale. Nu este permisă modificarea structurii GO, a duratei
sprintului sau creșterea numărului de GO active.

**Relația cu restul sistemului**

Regulatory Authority urmează după Monitoring Authority și influențează
Daily Stack, Velocity Control, Strategic Reset Mode și Reinforcement
Model. Fără acest nivel, sistemul ar deveni rigid și ar produce abandon
în situații reale de viață.

+-----------------------------------------------------------------------+
| LEVEL 4 --- Regulatory Authority                                      |
|                                                                       |
| **C32 Adaptive Context Engine**                                       |
+-----------------------------------------------------------------------+

**Definiție**

Adaptive Context Engine este mecanismul care analizează contextul
utilizatorului și adaptează execuția fără a modifica structura
strategică. Acest engine funcționează continuu în fundal și poate ajusta
ritmul execuției.

**Energy Modulation**

Mecanismul care ajustează intensitatea execuției în funcție de nivelul
de energie al utilizatorului. Energia personală nu este constantă ---
execuția constantă într-un context de energie scăzută produce epuizare,
stagnare și abandon.

+-----------------------------------------------------------------------+
| Execution\_Intensity = Base\_Intensity × Energy\_Factor                  |
|                                                                       |
| Energy\_Factor ∈ \[0.6, 1.2]                                          |
|                                                                       |
| EF < 1.0 → execuție redusă (energie scăzută)                         |
|                                                                       |
| EF = 1.0 → execuție normală                                           |
|                                                                       |
| EF > 1.0 → execuție accelerată moderat (energie ridicată)            |
+-----------------------------------------------------------------------+

**Planned Pause Protocol**

Mecanismul prin care utilizatorul poate introduce perioade planificate
de pauză fără a destabiliza sistemul. Viața reală include perioade în
care execuția trebuie suspendată temporar: vacanțe, evenimente
familiale, recuperare fizică.

+-----------------------------------------------------------------------+
| Pause\_interval ≤ 30 zile/an (total)                                   |
|                                                                       |
| Expected(t) înghețat pe durata pauzei                                 |
|                                                                       |
| Stagnation Detection dezactivat                                       |
|                                                                       |
| Zile de pauză excluse din Consistency\_comp (C37)                      |
|                                                                       |
| Retroactive Pause:                                                    |
|                                                                       |
| Marcaj retroactiv --- maxim 48h, maxim 3 ori/sprint                   |
|                                                                       |
| Limite anti-abuz retrospectiv                                         |
+-----------------------------------------------------------------------+

**Crisis Protocol**

Mecanismul activat în situații de criză majoră care afectează
capacitatea utilizatorului de a executa obiectivele. Situațiile extreme
pot include: accidente, pierdere financiară, probleme medicale.

+-----------------------------------------------------------------------+
| Execution\_Intensity → minim stabil (0.1--0.2)                         |
|                                                                       |
| SRM L1 și SRM L2 dezactivate                                          |
|                                                                       |
| Expected(t) înghețat                                                  |
|                                                                       |
| Minim 1 acțiune/zi/GO activ (menținere contact)                       |
+-----------------------------------------------------------------------+

**External Shock Buffer (ESB)**

Mecanism care absoarbe impactul evenimentelor externe neprevăzute.
Evenimentele externe pot afecta execuția chiar dacă utilizatorul este
motivat: schimbări profesionale, probleme logistice, evenimente
familiale.

+-----------------------------------------------------------------------+
| Stagnation threshold: 5 → 10 zile (extins)                            |
|                                                                       |
| SRM L1 threshold extins: 3 → 5 zile consecutive                       |
|                                                                       |
| Previne activarea prematură a SRM în context volatil                  |
+-----------------------------------------------------------------------+

**Momentum Monitor**

Mecanism care urmărește continuitatea execuției și detectează pierderea
momentum-ului. Pierderea momentum-ului este un risc major de abandon pe
termen lung --- mai important de detectat decât Drift-ul zilnic.

**Burnout Prevention Mechanism**

Mecanism care detectează semnele de supraîncărcare și reduce temporar
intensitatea execuției. Supraîncărcarea prelungită produce scăderea
performanței, abandon al sistemului și deteriorare motivațională.

**Rol general Adaptive Context Engine**

* Adaptează intensitatea execuției la realitatea contextuală.
* Protejează sustenabilitatea --- previne abandonul sistemului în
perioade dificile.
* Menține contactul utilizatorului cu sistemul chiar în perioadele de
criză.

+-----------------------------------------------------------------------+
| LEVEL 4 --- Regulatory Authority                                      |
|                                                                       |
| **C33 Strategic Reset Mode --- SRM (3 Levels)**                       |
+-----------------------------------------------------------------------+

**Definiție**

Strategic Reset Mode este mecanismul prin care sistemul restructurează
execuția atunci când instabilitatea devine critică. Cele 3 niveluri au
grade de intervenție crescând și grade de automatism descrescând. Un
singur nivel SRM poate fi activ simultan per GO --- nivelul mai înalt
suprascrie și anulează complet nivelul inferior.

**Ierarhia SRM: L3 > L2 > L1**

+-----------------------------------------------------------------------+
| SRM L1 --- Adjustment (automat, silențios):                           |
|                                                                       |
| Trigger: Drift < −0.15 / 3 zile consecutive, SAU regression\_flag     |
|                                                                       |
| Acțiune: Sprint Target −20%. Dashboard nemodificat vizual.            |
|                                                                       |
| Revocare: automat dacă Drift revine > −0.10 timp de 3 zile.          |
|                                                                       |
| DEZACTIVAT în: Reactivation Protocol, Crisis Protocol.                |
|                                                                       |
| SRM L2 --- Structural Adjustment (automat + notificare push):         |
|                                                                       |
| Trigger: Chaos\_Index ∈ \[0.40, 0.60) SAU Final\_ALI ∈ (1.0, 1.10]     |
|                                                                       |
| Acțiune: recalcul Sprint Target, redistribuție resurse.               |
|                                                                       |
| Notificare: "Am ajustat temporar ritmul tău."                       |
|                                                                       |
| Threshold crescut la 0.60 dacă Reactivation Protocol activ.           |
|                                                                       |
| SRM L3 --- Strategic Reset (confirmare obligatorie):                  |
|                                                                       |
| Trigger: Chaos\_Index ≥ 0.60 SAU Final\_ALI > 1.10                     |
|                                                                       |
| Acțiune: C34 Suspension + C35 Core Stabilization                      |
|                                                                       |
| Necesită double confirmation din partea utilizatorului.               |
+-----------------------------------------------------------------------+

**Timeout Protocol SRM L3**

Dacă utilizatorul nu confirmă SRM L3 în intervalele definite, sistemul
aplică protecție graduală pentru a evita starea de limbo:

* 24h fără confirmare → SRM L2 aplicat automat cu notificare: „Am
ajustat temporar ritmul tău."
* 72h fără confirmare → SRM L3 re-propus cu context actualizat.
* 7 zile fără confirmare → GO cu Priority\_weight minim (și GORI mai
mic la paritate) suspendat automat. Opțiune de reactivare
disponibilă imediat.

**Imunitate în Reactivation Protocol**

Pe durata Reactivation Protocol (C36), SRM L1 este dezactivat complet și
threshold-ul SRM L2 este crescut de la 0.40 la 0.60. La ieșirea din Core
Stabilization, Chaos\_Index este natural crescut și ar declanșa fals SRM
L2 imediat fără această protecție. Imunitatea permite tranziție
graduală.

+-----------------------------------------------------------------------+
| LEVEL 4 --- Regulatory Authority                                      |
|                                                                       |
| **C34 Weighted GO Suspension Logic**                                  |
+-----------------------------------------------------------------------+

**Definiție**

Mecanism care permite suspendarea temporară a unui GO în funcție de
prioritatea strategică. Nu toate GO au aceeași importanță --- GO cu
prioritate mai mică pot fi suspendate temporar pentru a proteja
obiectivele critice.

**De unde vine logica**

Când SRM L3 este confirmat, sistemul trebuie să reducă sarcina
operațională prin suspendarea unui sau mai multor GO-uri. Selecția
trebuie să fie obiectivă și bazată pe criterii strategice clare, nu
arbitrară.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Selecție GO de suspendat:                                             |
|                                                                       |
| Probabilitate suspendare ∝ 1 / Priority\_weight                        |
|                                                                       |
| La paritate weight: GO cu GORI mai mic → suspendat primul             |
|                                                                       |
| Consecințe SUSPENDED:                                                 |
|                                                                       |
| → GO exclus din C29 Focus Rotation (Σ normalizare)                    |
|                                                                       |
| → Sprint curent marcat SUSPENDED --- exclus din GORI și Continuity    |
|                                                                       |
| → check\_priority\_balance() apelat automat (C8)                        |
|                                                                       |
| C34 > C18: dacă recalibrare periodică coincide cu suspendarea,       |
|                                                                       |
| Stabilization Review are prioritate față de recalibrarea anuală       |
+-----------------------------------------------------------------------+

**Rol**

* Protejează obiectivele dominante --- reduce supraîncărcarea prin
suspendare selectivă.
* Menține coerența sistemului --- GO suspendat nu mai consumă resurse
operaționale.

+-----------------------------------------------------------------------+
| LEVEL 4 --- Regulatory Authority                                      |
|                                                                       |
| **C35 Core Stabilization Mode**                                       |
+-----------------------------------------------------------------------+

**Definiție**

Mod de execuție minimal activat în condiții de instabilitate majoră. În
perioade extreme, menținerea unui progres minimal este mai importantă
decât performanța maximă.

**De unde vine logica**

Revenirea bruscă la execuție normală după o perioadă de criză poate
produce o recădere. Core Stabilization Mode menține contactul
utilizatorului cu sistemul la intensitate minimă --- baza din care poate
reveni gradual.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Execution\_mode = MINIMAL\_SUSTAIN                                      |
|                                                                       |
| Intensity = 0.1 → 0.3 (ajustabil manual în interval)                  |
|                                                                       |
| Minimum: 1 acțiune/zi/GO activ                                        |
|                                                                       |
| Comportament componente:                                              |
|                                                                       |
| Expected(t) = ÎNGHEȚAT (nu crește, nu acumulează Drift)               |
|                                                                       |
| SRM L1 = DEZACTIVAT                                                   |
|                                                                       |
| SRM L2 = DEZACTIVAT                                                   |
|                                                                       |
| Stagnation Detection = DEZACTIVAT                                     |
|                                                                       |
| Durată: nelimitată --- până la decizia utilizatorului sau reducerea   |
| ALI                                                                   |
|                                                                       |
| Ieșire: → C36 Reactivation Protocol (rampă de revenire gradată)       |
+-----------------------------------------------------------------------+

**Prevenirea Loop-ului Paradoxal**

Înghețarea Expected(t) este critică. Fără ea, apare un loop paradoxal:
Stabilization → Drift crește zilnic (Expected avansează, Progress
stagnează) → SRM L1 declanșat → dar Stabilization dezactivează deja L1 →
incoerență de stare. Înghețarea elimină complet această posibilitate.

**Rol**

* Menține continuitatea --- previne abandonul complet al sistemului.
* Creează baza pentru revenire graduală prin Reactivation Protocol.

+-----------------------------------------------------------------------+
| LEVEL 4 --- Regulatory Authority                                      |
|                                                                       |
| **C36 Reactivation Protocol**                                         |
+-----------------------------------------------------------------------+

**Definiție**

Mecanism care permite revenirea treptată la execuția normală după
perioade de instabilitate sau Core Stabilization. Revenirea bruscă la
intensitate maximă poate produce recădere.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Rampă de revenire:                                                    |
|                                                                       |
| Intensity = 0.2 → +0.1/zi → 1.0 (8 zile)                              |
|                                                                       |
| Pe durata Reactivation:                                               |
|                                                                       |
| SRM L1 = DEZACTIVAT                                                   |
|                                                                       |
| Threshold SRM L2 = 0.60 (față de 0.40 standard)                       |
|                                                                       |
| La Intensity = 1.0 → threshold-urile revin la valorile standard       |
|                                                                       |
| Reactivare GO SUSPENDED în cadrul Reactivation Protocol:              |
|                                                                       |
| Relevance GO ≥ 0.60 necesar (față de 0.40 standard)                   |
|                                                                       |
| check\_priority\_balance() apelat automat la reactivare                 |
+-----------------------------------------------------------------------+

**Explicație**

Revenirea la execuția normală se face treptat pentru a evita recăderile.
Rampa de 8 zile permite recalibrarea graduală a ritmului și a
Chaos\_Index fără a declanșa fals SRM L2.

**Threshold Crescut pentru GO SUSPENDED**

Pragul crescut de Relevance la 0.60 pentru reactivarea unui GO SUSPENDED
asigură că GO-ul care revine la activitate are suficientă valoare
strategică pentru a justifica resursele suplimentare necesare în
perioada de recuperare. Un GO cu Relevance 0.42 (deasupra floor-ului de
0.30 dar sub 0.60) rămâne SUSPENDED până când contextul strategic se
schimbă.

**Rol**

* Restabilește progresul --- reintegrarea graduală a GO-urilor active.
* Protejează stabilitatea pe termen lung --- previne recăderea după
criză.

**LEVEL 5 --- STRATEGIC CONSOLIDATION**

Strategic Consolidation este mecanismul formal prin care sistemul
consolidează progresul obținut și întărește comportamentele productive
pentru a menține continuitatea dezvoltării pe termen lung.

Acest nivel nu definește obiective noi și nu execută activități. Rolul
lui este să transforme rezultatele obținute în stabilitate motivațională
și coerență strategică.

**Scopul acestui nivel**

* evaluarea performanței executive
* consolidarea progresului obținut
* stabilizarea motivației utilizatorului
* prevenirea regresului după finalizarea sprinturilor
* susținerea continuității pe termen lung

**Regula fundamentală**

Progresul realizat trebuie consolidat pentru a deveni stabil și
repetabil. Fără consolidare: progresul devine temporar, motivația scade,
apare regres comportamental și sistemul devine instabil pe termen lung.

**Relația cu restul sistemului**

Strategic Consolidation urmează după Monitoring Authority și Regulatory
Authority. Influențează stabilitatea motivațională, continuitatea
execuției, calitatea planificării viitoare și evaluarea anuală a
progresului. Fără consolidare, progresul nu devine sustenabil.

**REINFORCEMENT MODEL**

Reinforcement Model este mecanismul prin care sistemul evaluează
performanța și consolidează comportamentele productive. Conține Sprint
Score Calculation (C37), GORI (C38), Engagement Signal (C39) și Sprint
Reflection Gate (C40).

+-----------------------------------------------------------------------+
| LEVEL 5 --- Strategic Consolidation                                   |
|                                                                       |
| **C37 Sprint Score Calculation**                                      |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul prin care este evaluată performanța unui sprint pe 3
dimensiuni: progresul față de target, consistența execuției zilnice și
abaterea față de traiectoria planificată.

**De unde vine logica**

Execuția fără evaluare: nu permite învățare, nu permite optimizare, nu
permite comparabilitate. Evaluarea periodică stabilizează progresul și
oferă baza pentru recalibrare inteligentă.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Sprint\_Score = Progress\_comp × 0.50 + Consistency\_comp × 0.30 +       |
| Deviation\_comp × 0.20                                                 |
|                                                                       |
| Progress\_comp:                                                        |
|                                                                       |
| Real\_Progress / Sprint\_Target (clamped \[0,1])                       |
|                                                                       |
| Consistency\_comp:                                                     |
|                                                                       |
| Zile\_cu\_minim\_50%\_Core\_Stack\_completat\_în\_ziua\_respectivă            |
|                                                                       |
| ÷ Zile\_eligibile                                                      |
|                                                                       |
| Zile\_eligibile = Zile\_sprint − Zile\_ESB − Zile\_Retroactive\_Pause      |
|                                                                       |
| Late completions (> 48h): EXCLUSE din Consistency, incluse în        |
| Progress                                                              |
|                                                                       |
| Deviation\_comp:                                                       |
|                                                                       |
| 1 − |Drift\_final\_sprint| / 0.30 (clamped \[0,1])                   |
|                                                                       |
| Sprint cu status SUSPENDED: EXCLUS complet --- nu 0, nu medie         |
+-----------------------------------------------------------------------+

**Grile Sprint Score**

\---

**Interval**     **Calificativ**   **Interpretare**

**0.85--1.00**   **S ---           Sprint excepțional: target depășit, execuție
Excelent**        consistentă.

**0.70--0.84**   **A --- Foarte    Performanță solidă pe toate dimensiunile.
Bun**

**0.55--0.69**   **B --- Bun**     Progres real, consistență cu variabilitate
acceptabilă.

**0.40--0.54**   **C ---           Progres parțial; plan de sprint prea
Satisfăcător**    ambițios sau execuție neregulată.

**0.25--0.39**   **D ---           Probleme structurale; sprint-ul următor
Recalibrare       necesită ajustare semnificativă.
Timpurie**

**0.00--0.24**   **F ---           Execuție minimă sau absentă; SRM L2/L3
Intervenție**     probabil deja activ.

\---

**Notă**

Acțiunile provenite din Optional Stack pot influența evaluarea
calitativă a execuției, dar nu modifică scorul structural al sprintului
decât dacă sunt validate ca suport direct pentru milestone-uri.

**Rol**

* Evaluează performanța executivă pe fiecare sprint.
* Susține consolidarea progresului și contribuie la evaluarea anuală
GORI.

+-----------------------------------------------------------------------+
| LEVEL 5 --- Strategic Consolidation                                   |
|                                                                       |
| **C38 GORI --- Global Objective Return Index**                        |
+-----------------------------------------------------------------------+

**Definiție**

Mecanismul care consolidează performanța strategică pe termen lung prin
agregarea ponderată a rezultatelor sprinturilor. Spre deosebire de
Sprint Score (care evaluează un sprint izolat), GORI reflectă tendința
pe termen lung și consecvența strategică anuală.

**De unde vine logica**

Performanța strategică nu poate fi evaluată pe termen scurt. Este
necesară o măsurare agregată, ponderată temporal și ajustată pentru
consistență, pentru a reflecta stabilitatea reală a progresului. Un
utilizator care alternează sprinturi excelente cu sprinturi slabe nu
demonstrează aceeași soliditate cu unul care menține performanță medie
constantă.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| GORI\_calculat = Σ(Sprint\_Score\_i × w\_i) / Σ(w\_i)                      |
|                                                                       |
| Ponderi temporale (recency bias):                                     |
|                                                                       |
| Sprint curent (cel mai recent): w = 1.0                               |
|                                                                       |
| Sprint n−1: w = 0.8                                                   |
|                                                                       |
| Sprint n−2: w = 0.6                                                   |
|                                                                       |
| Sprint n−3 și mai vechi: w = 0.4                                      |
|                                                                       |
| GORI\_final = GORI\_calculat × Continuity\_factor × (1 −                 |
| Variance\_penalty)                                                     |
|                                                                       |
| Continuity\_factor:                                                    |
|                                                                       |
| = sprinturi\_active / (sprinturi\_totale                                |
|                                                                       |
| − sprinturi\_suspended − sprinturi\_seasonal\_pause)                     |
|                                                                       |
| Variance\_penalty:                                                     |
|                                                                       |
| = Variance(Sprint\_Scores\_active) × 0.25                               |
|                                                                       |
| Penalizare maximă: 25%                                                |
|                                                                       |
| Sprinturi SUSPENDED: EXCLUSE complet din calcul (nu 0, nu medie)      |
+-----------------------------------------------------------------------+

**Interpretarea Continuity\_factor**

Continuity\_factor reflectă ce proporție din durata strategică
planificată a GO-ului a fost executată efectiv. Un GO sezonier cu 5 luni
active și 7 luni SEASONAL\_PAUSE are Continuity = 5/5 = 1.0 --- a
executat perfect în fereastra sa activă. Același calcul fără excluderea
SEASONAL\_PAUSE ar produce 5/12 = 0.42, o penalizare artificială pentru
arhitectura temporală a obiectivului, nu pentru performanța reală.

**Grile GORI**

\---

**Interval**     **Clasificare**    **Acțiune sugerată**

**0.80--1.00**   **Excellent**      Consolidare strategică puternică.
Continuă cu aceeași abordare.

**0.60--0.79**   **Good**           Progres consistent. Micro-ajustări
opționale la recalibrare.

**0.45--0.59**   **Advisory         Invitație pentru reflecție. Sistemul
Review**           propune C40 și C18 Recalibration.

**0.00--0.44**   **Early            Intervenție necesară. C18 accelerat,
Recalibration**    posibil Vault sau reformulare GO.

\---

**Rol**

* Consolidează progresul strategic --- susține evaluarea anuală
completă.
* Contribuie la deciziile de recalibrare și la selecția GO-urilor de
suspendat (C34).

+-----------------------------------------------------------------------+
| LEVEL 5 --- Strategic Consolidation                                   |
|                                                                       |
| **C39 Engagement Signal**                                             |
+-----------------------------------------------------------------------+

**Definiție**

Engagement Signal monitorizează în background trei proxy-uri de
motivație pentru fiecare GO activ, pentru a detecta obiectivele care au
pierdut relevanța subiectivă a utilizatorului --- chiar dacă scorul
formal (Sprint Score, GORI) rămâne acceptabil.

**De unde vine logica**

Un utilizator poate continua să bifeze acțiuni mecanic, fără angajament
real față de obiectiv. Sprint Score și GORI nu pot detecta această
situație --- evaluează execuția formală. Engagement Signal detectează
dezangajarea prin proxy-uri comportamentale observabile.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| 3 Proxy-uri de Engagement monitorizate în background:                 |
|                                                                       |
| Proxy 1 --- Timp de completare:                                       |
|                                                                       |
| Flag dacă > 5 acțiuni completate în < 10 minute                     |
|                                                                       |
| (semnal de bifat mecanic, fără angajament real)                       |
|                                                                       |
| Proxy 2 --- Click depth:                                              |
|                                                                       |
| Flag dacă 0 acțiuni de explorare a detaliilor GO ≥ 7 zile             |
|                                                                       |
| (utilizatorul nu mai accesează informații despre obiectiv)            |
|                                                                       |
| Proxy 3 --- Optional Stack rata:                                      |
|                                                                       |
| Flag dacă rata de completare Optional Stack → 0 pe ≥ 7 zile           |
|                                                                       |
| (utilizatorul face strictul minim, fără nicio inițiativă)             |
|                                                                       |
| Engagement\_Signal = WEAK dacă 2/3 proxy-uri negative ≥ 7 zile         |
| consecutive                                                           |
|                                                                       |
| La WEAK:                                                              |
|                                                                       |
| O singură întrebare: "Scopul acesta mai pare important pentru        |
| tine?"                                                               |
|                                                                       |
| Da → semnal resetat, monitorizare continuă                            |
|                                                                       |
| Nu → C18 Recalibration accelerat pentru GO respectiv                  |
+-----------------------------------------------------------------------+

**Toleranța de design**

Sistemul de proxy-uri nu este un mecanism anti-fraudă --- este o
detecție de dezangajare. Un utilizator care manipulează deliberat toate
cele 3 proxy-uri cheltuie mai mult efort manipulând sistemul decât
executând GO-ul --- problema se auto-rezolvă. Pragul de 2/3 proxy-uri
negative reduce semnificativ false positive-urile datorate
variabilității normale.

**Integrare cu C40**

Dacă la Sprint Reflection Gate (C40) utilizatorul răspunde cu un scor de
vitalitate Q3 ≤ 5, Engagement Signal este activat preventiv ---
monitorizare crescută în sprint-ul următor.

+-----------------------------------------------------------------------+
| LEVEL 5 --- Strategic Consolidation                                   |
|                                                                       |
| **C40 Sprint Reflection Gate**                                        |
+-----------------------------------------------------------------------+

**Definiție**

Sprint Reflection Gate este o oprire opțională la tranziția dintre
sprinturi --- după calculul Sprint Score (C37) și înainte de Sprint
Planning pentru sprint-ul următor. Scopul este reflexiv, nu evaluativ.

**De unde vine logica**

Execuția continuă fără momente de reflecție duce la pierderea sensului
și la dezangajare progresivă. O oprire structurată, dar complet
voluntară, la finalul fiecărui sprint oferă utilizatorului spațiul să
proceseze experiența fără presiune. Zero impact pe scor înseamnă că
utilizatorul poate fi sincer.

**Reprezentare formală**

+-----------------------------------------------------------------------+
| Momentul apariției: după C37 Sprint Score, înainte de Sprint Planning |
|                                                                       |
| Cele 3 întrebări (toate opționale, skippable în orice moment):        |
|                                                                       |
| Q1: "Ce a funcționat cel mai bine în acest sprint?"                 |
|                                                                       |
| (răspuns liber --- text sau selecție din liste pre-generate)          |
|                                                                       |
| Q2: "Unde ai fost blocat sau ai întâmpinat rezistență?"             |
|                                                                       |
| (răspuns liber)                                                       |
|                                                                       |
| Q3: "Cât de important simți că este acest obiectiv acum?"           |
|                                                                       |
| (scală 1--10)                                                         |
|                                                                       |
| Consecințe directe:                                                   |
|                                                                       |
| Răspunsurile Q1/Q2 → context opțional în Sprint Planning următor      |
|                                                                       |
| Q3 ≤ 5 → C39 Engagement Signal activat preventiv                      |
|                                                                       |
| Sprint Skipped → Sprint Planning continuă normal, fără penalizare     |
+-----------------------------------------------------------------------+

**Zero Impact pe Scor**

Sprint Reflection Gate nu modifică Sprint Score calculat, nu alimentează
GORI și nu declanșează SRM. Este singurul mecanism din NUViaX care are
acces direct la starea subiectivă a utilizatorului. Designul
intenționat: dacă reflecția ar modifica scoruri, utilizatorul ar fi
stimulat să răspundă strategic (pozitiv) în loc de sincer.

**Integrarea cu Sprint Planning**

Dacă utilizatorul completează Q1 sau Q2, răspunsurile sunt transmise
Sprint Planning (C20) ca context opțional. Sprint Planning poate sugera
ajustări de milestone sau de Daily Stack format pe baza feedback-ului
din Q2. Integrarea este sugestivă, nu automată --- utilizatorul poate
ignora sugestiile.

**Rol**

* Susține consolidarea motivațională --- oferă spațiu de procesare
fără presiune.
* Detectează timpuriu dezangajarea prin Q3 și integrarea cu C39.
* Îmbunătățește calitatea Sprint Planning prin context subiectiv
opțional.

**SUMAR GENERAL AL FRAMEWORK-ULUI**

NUViaX Growth Framework™ este un sistem care transformă dorințele în
direcție clară și direcția în progres real.

Framework-ul oferă structură acolo unde apare haosul, limită acolo unde
apare supraîncărcarea și ritm acolo unde apare stagnarea. Utilizatorul
nu mai navighează prin obiective confuze, ci urmează un traseu
organizat, construit pentru a susține rezultate vizibile și sustenabile.

Obiectivele sunt clarificate, limitate și transformate în pași
realizabili, iar progresul este monitorizat constant pentru a menține
echilibrul dintre ambiție și capacitate.

Sistemul se adaptează realității vieții utilizatorului, protejându-l de
epuizare și ajutându-l să continue chiar și în perioade dificile.

Prin această arhitectură, NUViaX Growth Framework™ devine un ghid stabil
pentru evoluție personală --- oferind claritate, direcție și încredere
în fiecare etapă a progresului.

