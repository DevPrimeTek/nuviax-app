# **LAYER 0 - AXIOMATIC FOUNDATION**

Layer 0 - Axiomatic Foundation reprezintă setul de reguli structurale invariabile care definesc identitatea, limitele și stabilitatea matematică a NUViaX Growth Framework™. Este fundamentul pe care toate celelalte nivele funcționează.

### **Scopul acestui bloc**

- Prevenirea haosului structural.
- Stabilirea limitelor sistemului.
- Asigurarea stabilității matematice.
- Menținerea identității metodologice.

LAYER 0 - Axiomatic Foundation

**C1 Structural Supremacy Principle**

#### **Definiție**

Principiul prin care structura sistemului este invariabilă, iar operaționalul este ajustabil doar în interiorul acestei structuri.

#### **De unde vine logica**

Orice sistem complex care permite modificarea structurii în timp real devine instabil. Pentru stabilitate matematică și comportamentală: Structura ≠ Parametri. Parametrii pot varia doar în interiorul structurii, nu o pot modifica.

#### **Reprezentare formală**

S = setul de reguli structurale

P = setul de parametri ajustabili

S ∩ P = ∅ - regulile structurale și parametrii nu se suprapun niciodată

P ⊂ Domain(S) - parametrii pot exista exclusiv în interiorul regulilor structurale

#### **Explicație**

Utilizatorul poate modifica valori (timp, target, intensitate), dar nu poate modifica structura (număr GO, durată sprint, modele comportamentale). Cu alte cuvinte: ce se face este rigid - cum se face este flexibil.

#### **Rol**

- Protejează framework-ul de extindere haotică.
- Previne „personalizarea excesivă".
- Menține consistența pe termen lung.

LAYER 0 - Axiomatic Foundation

**C2 Behavior Model System**

#### **Definiție**

Set finit și închis de tipuri universale de transformare acceptate în sistem. Conține: CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE.

#### **De unde vine logica**

Orice transformare umană poate fi redusă la una dintre cele 5 direcții: creare (nu există → există), creștere (există → mai mult), reducere (există → mai puțin), menținere (stabilizare), evoluție (transformare progresivă). Set finit → sistem închis → stabilitate matematică.

#### **Reprezentare formală**

Fie T = orice transformare

T ∈ { CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE }

|BM(GO)| = 1 - unicitate obligatorie per GO

|T| = 5 - set finit și închis, nu există al 6-lea model permis

#### **Cele 5 Behavior Models**

| **Model**    | **Direcție**            | **Exemple tipice**                                                   |
| ------------ | ----------------------- | -------------------------------------------------------------------- |
| **CREATE**   | Construire din zero     | Lansare produs, scriere carte, fondare companie, prototip nou        |
| **INCREASE** | Creștere măsurabilă     | MRR, număr clienți, masă musculară, vocabular, audiență              |
| **REDUCE**   | Reducere măsurabilă     | Greutate corporală, timp de livrare, costuri fixe, timp ecran        |
| **MAINTAIN** | Menținere în interval   | Greutate stabilă, relație activă, skill activ, nivel fitness         |
| **EVOLVE**   | Transformare calitativă | Schimbare carieră, restructurare identitate, tranziție la leadership |

#### **Explicație**

Orice schimbare pe care utilizatorul o dorește trebuie să se încadreze într-una dintre cele 5 categorii definite. Dacă un obiectiv nu poate fi încadrat în niciuna, trebuie reformulat. Nu sunt permise obiective cu 2 direcții simultane pe aceeași metrică - aceasta generează GO_REJECTED_LOGICAL_CONTRADICTION (C9).

#### **Rol**

- Standardizare universală - permite formule matematice coerente pentru sprint și progres.
- Elimină ambiguitatea direcțională înainte de activare.

LAYER 0 - Axiomatic Foundation

**C3 Maximum 3 Active GO Constraint**

#### **Definiție**

Numărul maxim de Global Objectives active simultan este 3. Dacă utilizatorul încearcă să adauge un al patrulea obiectiv, sistemul îl va respinge sau îl va trimite în Future Vault.

#### **De unde vine logica**

Fragmentarea atenției crește non-liniar cu numărul de obiective. Dacă n = numărul de GO active, costul cognitiv ≈ n² (interferență între obiective). La n > 3, interferența crește exponențial și niciun obiectiv nu mai primește atenție suficientă pentru progres real.

#### **Reprezentare formală**

n_active ≤ 3

Cost_cognitiv ≈ n² (interferență exponențială, nu liniară)

Dacă n = 3 → orice tentativă de activare nouă → C17 Future Vault

#### **GO Sezoniere și SEASONAL_PAUSE**

Un GO cu execution_windows definite ocupă un slot activ permanent - inclusiv în perioadele în care fereastra este inactivă. Aceasta reflectă existența sa ca intenție strategică anuală. Pe durata perioadelor inactive, statusul sprint-ului devine SEASONAL_PAUSE: Expected(t) este înghețat, acțiunile nu sunt generate, iar sarcina operațională este nulă - dar slotul este ocupat.

SEASONAL_PAUSE este un status distinct: diferit de SUSPENDED (act al sistemului ca răspuns la instabilitate) și diferit de VAULT (niciodată activat).

Sprinturile SEASONAL_PAUSE sunt excluse din calculul Continuity_factor în C38 - identic cu sprinturile SUSPENDED.

#### **Rol**

- Previne dispersia - menține focus strategic real.
- Stabilizează ALI și calculele de capacitate.
- Forțează decizia de prioritizare conștientă.

LAYER 0 - Axiomatic Foundation

**C4 365-Day Maximum GO Duration Constraint**

#### **Definiție**

Un GO nu poate depăși 365 de zile calendaristice de la data activării până la deadline.

#### **De unde vine logica**

Obiectivele fără limită temporală devin identitare permanente, nu pot fi evaluate și nu pot intra în cicluri de consolidare. Limitarea la 1 an permite: 12 sprinturi măsurabile, evaluare anuală prin GORI, recalibrare structurată. Dacă utilizatorul setează un obiectiv pe 2 ani, sistemul îl obligă să îl împartă în etape anuale.

#### **Reprezentare formală**

deadline − start_date ≤ 365 zile

Deadline_sugerat = start + Benchmark_domeniu × Amplitudine

Interval acceptat de utilizator: \[1, 365\] zile

#### **Rol**

- Forțează concretizarea - obiectivele vagi sau nelimitate temporal sunt respinse.
- Permite calculul GORI anual complet și comparabil.
- Menține ritmul strategic prin cicluri anuale definite.

LAYER 0 - Axiomatic Foundation

**C5 30-Day Fixed Sprint Constraint**

#### **Definiție**

Sprintul are durată fixă de exact 30 de zile calendaristice. Variabila temporală t este un integer în intervalul \[1, 30\]. Drift-ul se calculează o singură dată pe zi, la finalul zilei.

#### **De unde vine logica**

Dacă durata sprintului este variabilă: progresul devine imposibil de comparat, Drift devine instabil, GORI devine distorsionat. Standardizarea la 30 de zile fixe permite măsurare uniformă și comparabilitate istorică completă.

#### **Reprezentare formală**

Sprint_length = 30 (invariant structural)

t ∈ { 1, 2, 3, ..., 30 } - integer explicit, nu real

Expected(t) = t / 30

Exemplu: ziua 15 → 15/30 = 0.50 → 50% progres așteptat

Drift calculat: 1 dată/zi, la finalul zilei t

Late completion (> 48h): → Progress_comp YES / Consistency_comp NO

SEASONAL_PAUSE: Expected(t) înghețat - t nu avansează în perioade inactive

#### **Rol**

- Permite calcul standardizat și determinist al Drift.
- Permite comparabilitate istorică între sprinturi și între GO-uri.
- Asigură că GORI este construit din date structurate uniform.

LAYER 0 - Axiomatic Foundation

**C6 Normalization Rule - Clamp \[0,1\]**

#### **Definiție**

Toate valorile de progres și performanță sunt limitate în intervalul \[0, 1\] prin funcția clamp. Metricile de alertă și monitorizare nu sunt clamped - ele trebuie să poată semnaliza orice magnitudine de deviere.

#### **De unde vine logica**

Fără clamp: progresul poate deveni negativ sau poate depăși 100%, Drift poate deveni instabil, ALI poate distorsiona sistemul. Clamp = stabilitate numerică. Metricile de alertă (Drift, ALI) rămân libere pentru a detecta amplitudinea reală a problemelor.

#### **Reprezentare formală**

clamp(x) = 0 dacă x < 0

\= x dacă 0 ≤ x ≤ 1

\= 1 dacă x > 1

Aplicate clamp: Real_Progress, Sprint_Score, GORI, Focus_weights

Neclamped: Drift, ALI, Chaos_Index (pot depăși 1.0)

Hibrid (plafonat explicit):

Context_disruption = min(1.0, nr_eventi_majori / 3)

#### **Explicație**

Indiferent ce valoare rezultă din calcule: dacă este negativă → devine 0; dacă este mai mare de 100% → devine 1.0; dacă este între 0 și 1 → rămâne neschimbată. Această regulă previne erorile matematice și valorile imposibile care ar destabiliza engine-urile de evaluare.

#### **Rol**

- Previne instabilitatea matematică - asigură că metricile de progres rămân în domeniu valid.
- Asigură comparabilitate universală între utilizatori și GO-uri.
- Normalizează toate engine-urile din Levels 2-5.

LAYER 0 - Axiomatic Foundation

**C7 Priority Weight System (1-3)**

#### **Definiție**

Fiecărui GO i se atribuie un coeficient de prioritate între 1 și 3, derivat automat din scorul de Relevance strategică. Utilizatorul nu poate seta manual un weight arbitrar.

#### **De unde vine logica**

Nu toate obiectivele au importanță egală. Dacă weight = w, impactul strategic este proporțional cu w. O scală 1-3 este suficient de granulară pentru a diferenția, dar suficient de simplă pentru a evita iluzia de precizie a unei scale 1-10.

#### **Reprezentare formală**

Relevance_adj = round(Relevance_brut, 2)

Dacă Relevance_adj < 0.40 → weight = 1

Dacă 0.40 ≤ Relevance_adj < 0.75 → weight = 2

Dacă Relevance_adj ≥ 0.75 → weight = 3

Notă implementare: Math.round(score \* 100) / 100

- nu toFixed(2): produce string, necesită conversie suplimentară

#### **De ce nu scale 1-5 sau 1-10?**

Scala prea mare creează iluzie de precizie. Scala 1-3 produce claritate decizională. Un sistem cu 3 GO poate funcționa cu weight 3+2+2=7 (sumă permisă) sau 3+2+1=6, dar nu 3+3+3=9 (depășire). Diferența între 1 și 3 este semnificativă strategic; diferența între 7 și 8 pe o scală 1-10 nu ar fi.

#### **Rol**

- Influențează ALI și capacitatea alocată per GO.
- Influențează Focus Rotation și distribuția atenției zilnice.
- Influențează GORI ponderat și evaluarea strategică anuală.

LAYER 0 - Axiomatic Foundation

**C8 Priority Balance Constraint**

#### **Definiție**

Limitare structurală care previne supra-încărcarea strategică prin acumularea simultană a mai multor GO cu prioritate maximă. Suma Priority Weight-urilor tuturor GO-urilor active simultan nu poate depăși 7.

#### **De unde vine logica**

Dacă toate GO au weight = 3, sistemul pierde diferențierea strategică. Dacă toate obiectivele sunt „critice", decizia reală de prioritizare devine imposibilă, iar ALI și Focus Rotation se distorsionează.

#### **Reprezentare formală**

Σ(weight_i) ≤ 7 pentru toate GO_i cu status = ACTIVE

Exemplu permis: 3 + 2 + 2 = 7 ✓

Exemplu interzis: 3 + 3 + 3 = 9 ✗

Dacă Σ > 7 → auto-rezoluție:

GO cu Relevance minimă → weight redus automat

La paritate Relevance: GO cu GORI mai mic → reduce primul

Trigger universal - check_priority_balance() apelat după ORICE:

activare / suspendare / reactivare GO

modificare Relevance (C11, C18)

#### **Explicație**

Această regulă împiedică utilizatorul să considere toate obiectivele „critice". Suma maximă de 7 permite o configurație cu un GO dominant (w=3) și două obiective medii (w=2+2=4). Nu permite trei obiective simultane de prioritate maximă (3+3+3=9). Verificarea este continuă - după orice eveniment care poate modifica suma, nu periodic.

#### **Rol**

- Menține echilibrul strategic - forțează decizia reală de prioritate.
- Previne distorsiunea în ALI, Focus Rotation și GORI.
- Creează ierarhie clară între GO-urile active simultan.

