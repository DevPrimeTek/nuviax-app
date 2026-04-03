# **LEVEL 2 - EXECUTION AUTHORITY**

Execution Authority este nivelul sistemului responsabil pentru transformarea obiectivelor strategice validate în execuție operațională structurată.

Acest nivel nu definește obiective și nu reglează capacitatea utilizatorului. Rolul lui este să creeze arhitectura executivă prin care obiectivele sunt realizate în mod sistematic. Execution Authority stabilește ritmul execuției și structura progresului prin Sprint-uri, Milestone-uri și execuție zilnică.

### **Scopul acestui nivel**

- transformarea obiectivelor strategice în acțiune concretă
- stabilirea unui ritm executiv constant
- fragmentarea controlată a obiectivelor
- crearea unei structuri clare de progres
- generarea datelor necesare pentru monitorizare

### **Regula fundamentală**

Execuția este structurată și ciclică. Fiecare Global Objective activ este împărțit în Sprint-uri fixe de 30 zile, fiecare Sprint are un target măsurabil, fiecare Sprint este structurat în Milestone-uri, execuția zilnică derivă din structura sprintului. Execuția liberă, fără structură, nu este permisă.

### **Relația cu restul sistemului**

Execution Authority urmează după Structural Authority și precede Monitoring Authority și Capacity Regulation. Datele generate de execuție sunt utilizate pentru Drift, Chaos, ALI și Reinforcement Model.

## **PHASE III - EXECUTION ARCHITECTURE**

Execution Architecture este mecanismul formal prin care sistemul transformă Global Objectives active în execuție ciclică structurată pe Sprint-uri fixe de 30 zile. Aceasta nu este o listă de task-uri, ci o arhitectură operațională standardizată.

### **Scopul acestei faze**

- fragmentarea controlată a obiectivelor anuale
- standardizarea progresului
- permiterea calculului deviației (Drift)
- prevenirea stagnării
- crearea bazei pentru evaluarea strategică (Sprint Score și GORI)

### **Regula Fundamentală**

Execuția este ciclică și standardizată. Fiecare GO activ este împărțit în Sprint-uri fixe de 30 zile, fiecare Sprint are un target măsurabil, fiecare Sprint este structurat în maximum 5 Milestone-uri, execuția zilnică derivă din Milestone-ul activ. Nu sunt permise Sprint-uri cu durată variabilă. Nu sunt permise Milestone-uri nelimitate numeric. Nu sunt permise Daily Stack independent de structură.

### **Relația cu restul sistemului**

Execution Architecture urmează după Strategic Limitation și Strategic Feasibility Analysis. Precede Drift, Chaos, ALI (calcul real) și Reinforcement Model. Fără această fază, obiectivul rămâne strategic dar neexecutabil.

LEVEL 2 - Execution Authority

**C19 Sprint Structuring - 30-Day Cycle**

#### **Definiție**

Structura ciclică fixă prin care fiecare GO este împărțit în unități de execuție de 30 zile.

#### **De unde vine logica**

Dacă durata sprintului este variabilă: progresul devine imposibil de comparat, Drift devine instabil, GORI devine distorsionat. Standardizarea la 30 de zile permite măsurare uniformă.

#### **Reprezentare formală**

Sprint_length = 30 (invariant)

t ∈ { 1, 2, 3, ..., 30 } - integer

Expected(t) = t / 30

Statusuri posibile sprint:

ACTIVE - sprint în curs de execuție

COMPLETED - finalizat normal la t=30

SUSPENDED - GO suspendat prin SRM L3 în mijlocul sprintului

SEASONAL_PAUSE - GO sezonier intrat în fereastră inactivă (patch Rev. 5.6)

#### **Execution Windows și SEASONAL_PAUSE**

Un GO cu execution_windows definite (exemplu: activ noiembrie-martie pentru un obiectiv sezonier) intră automat în SEASONAL_PAUSE la ieșirea din fereastra activă. La intrarea în SEASONAL_PAUSE: Expected(t) este înghețat la valoarea curentă, progresul acumulat este conservat, Drift nu crește negativ, Stagnation Detection este dezactivat, ALI nu înregistrează ore pentru acel GO. La re-intrarea în fereastra activă: Expected(t) reia avansarea, Sprint Planning generează sprint nou cu target recalibrat.

SEASONAL_PAUSE nu este un SUSPENDED: GO-ul nu a eșuat, nu a necesitat intervenție sistemică. Este o pauză structurală planificată, cu semantică proprie.

Sprinturile cu status SEASONAL_PAUSE sunt excluse din Continuity_factor în GORI (C38) - identic cu SUSPENDED.

#### **Rol**

- Permite progres liniar și calcul standardizat al Drift.
- Permite comparabilitate istorică între sprinturi.

LEVEL 2 - Execution Authority

**C20 Sprint Target Calculation**

#### **Definiție**

Procesul prin care este stabilit rezultatul măsurabil al fiecărui Sprint, derivat din obiectivul anual și din progresul rămas de realizat.

#### **De unde vine logica**

Un Sprint fără target clar produce execuție difuză, creează milestone-uri incoerente și face progresul imposibil de evaluat. Targetul trebuie să fie ambițios dar realizabil - de aici factorul 0.80 și plafonul de compensație.

#### **Reprezentare formală**

Sprint_Target_brut = (Target_anual − Progress_actual) / Sprinturi_rămase

Sprint_Target_realist = Sprint_Target_brut × 0.80

Compensație după Regression Event:

Sprint_Target_compensat ≤ Sprint_Target_inițial × 1.50 (plafon)

Dacă 3+ Regression Events consecutive:

→ propunere formală de reducere Target_anual

#### **Plafonul de Compensație 1.5×**

Fără plafon, un Regression Event major ar genera un Sprint Target de recuperare de 200-300% din normal - un plan imposibil care duce fie la abandon, fie la bifat mecanic. Plafonul la 1.5× produce un target ambițios dar realist, iar diferența de recuperare se distribuie pe sprinturile rămase. La 3+ Regression Events consecutive, sistemul nu penalizează utilizatorul - propune o reducere a Target-ului anual ca semnal că targetul inițial a fost supraestimat.

#### **Rol**

- Definește direcția și ambiția sprintului.
- Creează baza pentru structurarea Milestone-urilor.

LEVEL 2 - Execution Authority

**C21 80% Probability Rule**

#### **Definiție**

Regulă prin care targetul sprintului este definit la un nivel realist de realizare - aproximativ 80% probabilitate de succes - nu la maximul teoretic.

#### **De unde vine logica**

Targeturi setate la nivel ideal (100% ambiție) cresc rata de eșec, cresc Drift negativ și cresc riscul de abandon. Planificarea realistă stabilizează sistemul și protejează Momentum-ul pe termen lung.

#### **Reprezentare formală**

Sprint_Target_aplicat = Sprint_Target_brut × 0.80

Pentru metrici continue (greutate, MRR, ore):

factorul 0.80 se aplică direct pe delta rămasă

Pentru metrici discrete (nr. clienți, nr. certificări):

Sprint 1-2: 0-1 unitate

Sprinturi medii: 1-2 unități/sprint

Sprint final: unitățile rămase pentru target total

#### **Rol**

- Reduce eșecul repetitiv - stabilizează execuția pe termen lung.
- Protejează Momentum-ul și motivația utilizatorului.

LEVEL 2 - Execution Authority

**C22 Milestone Structuring (max 5)**

#### **Definiție**

Fragmentarea Sprint Target în maximum 5 praguri intermediare de progres. Milestone-urile sunt subobiective intermediare verificabile - puncte de control care fac progresul vizibil în cadrul sprint-ului.

#### **De unde vine logica**

Prea multe milestone-uri produc fragmentare excesivă și transformă sistemul într-un task tracker. Prea puține lasă lipsa de claritate progresivă. Limitarea la maximum 5 păstrează echilibrul și oferă feedback de completare la momente semnificative.

#### **Reprezentare formală**

1 ≤ |Milestone_set| ≤ 5 per sprint

Ordonare milestone-uri:

ordered = false (default):

Milestone-urile pot fi bifate în orice ordine

ordered = true:

Milestone-urile au dependențe logice - M2 nu poate fi bifat

înainte de M1. Bifarea out-of-order generează avertisment,

nu blocare. Utilizatorul poate forța ordinea dacă dorește.

Flag-ul ordered este opțional și se setează per sprint, nu per GO

#### **Rol**

- Structurează progresul - permite vizualizarea etapelor intermediare.
- Permite Sprint Score să reflecte granularitatea execuției.

LEVEL 2 - Execution Authority

**C23 Daily Stack Generator**

#### **Definiție**

Daily Stack este setul de acțiuni zilnice recomandate utilizatorului, compus din două componente distincte: Core Stack (structural, obligatoriu) și Optional Stack (operațional, suplimentar).

#### **Core Stack și Optional Stack**

Core Stack reprezintă acțiunile generate automat din Milestone-ul activ și constituie execuția oficială a progresului. Optional Stack reprezintă acțiuni adăugate manual de utilizator pentru a susține sau accelera progresul, fără a modifica structura oficială a sprintului.

#### **De unde vine logica**

Daily Stack nu este o listă liberă de sarcini. Este derivat din structura sprintului. Dacă este independent de Sprint și Milestone, sistemul devine task tracker. Separarea dintre Core Stack și Optional Stack permite menținerea disciplinei structurale și a flexibilității operaționale simultan.

#### **Reprezentare formală**

Daily_Stack = Core_Stack ∪ Optional_Stack

Core_Stack = funcție(Milestone_activ, Context)

Optional_Stack = acțiuni_adăugate_manual (limitate numeric)

Timestamping obligatoriu pentru fiecare acțiune:

Late completion (> 48h după ziua programată):

→ Progress_comp: YES (acțiunea contribuie la progres)

→ Consistency_comp: NO (nu demonstrează execuție zilnică)

Bulk detection:

\> 5 acțiuni completate în < 10 minute → C39 Proxy 1 flag

Physical Delta Safety - Sprint 1, Milestone 1:

Dacă SINGLE_QUESTION_FLAG=TRUE → acțiune obligatorie "Consultă specialist"

#### **Rol**

- Traduce strategia în execuție zilnică concretă.
- Menține coerența direcțională - fiecare acțiune derivă din Milestone activ.
- Permite flexibilitate operațională controlată prin Optional Stack.
- Previne transformarea sistemului într-un task tracker.

LEVEL 2 - Execution Authority

**C24 Progress Computation Engine**

#### **Definiție**

Mecanismul prin care este calculat progresul real al Sprintului, ca rație dintre milestone-urile finalizate și milestone-urile totale.

#### **De unde vine logica**

Progresul nu poate fi intuitiv. Trebuie să fie numeric și comparabil cu progresul așteptat pentru a putea alimenta Drift și a permite evaluarea obiectivă a execuției.

#### **Reprezentare formală**

Real_Progress = milestone_finalizate / milestone_totale

Real_Progress = clamp(Real_Progress, 0, 1)

Acțiunile din Optional Stack contribuie la progres DOAR dacă

sunt marcate explicit ca suport direct pentru Milestone-ul activ.

Regression Event:

Metrica GO scade sub valoarea de la startul sprintului curent

→ Real_Progress = 0

→ regression_flag = TRUE

→ SRM L1 activat IMEDIAT (fără așteptarea a 3 zile consecutive)

→ Sprint Target compensat ≤ 1.5× (C20)

#### **Notă Late Completion**

Acțiunile bifate ca Late Completion (> 48h) contribuie la Real_Progress - sunt incluse în calculul milestone_finalizate. Nu contribuie la Consistency_comp (C37). Distincția reflectă realitatea: dacă o acțiune a fost executată, valoarea sa strategică este reală chiar dacă a venit cu întârziere.

#### **Rol**

- Permite compararea cu Expected(t) - baza calculului Drift.
- Alimentează Stability Control Layer și toate calculele de performanță.

LEVEL 2 - Execution Authority

**C25 Execution Variance Tracker**

#### **Definiție**

Mecanism backend care calculează diferența dintre progresul real și progresul așteptat și o transmite Stability Control Layer. Este componenta de legătură între execuția zilnică și mecanismele de monitorizare.

#### **De unde vine logica**

Fără măsurarea deviației: stagnarea nu poate fi detectată, intervenția devine întârziată și progresul aparent poate masca probleme reale. C25 nu ia decizii - calculează și transmite.

#### **Reprezentare formală**

Drift_raw = Real_Progress − Expected(t)

unde Expected(t) = t / 30, t ∈ { 1..30 }

Dacă progresul real < progresul planificat → Drift_raw negativ

Dacă progresul real > progresul planificat → Drift_raw pozitiv

Acțiunile din Optional Stack nu influențează calculul Drift

decât dacă sunt validate ca suport direct pentru Milestone-ul activ.

#### **Rol**

- Detectează deviația executivă - baza pentru intervenție timpurie.
- Alimentează Stability Control Layer (C26, C27, C28).
- Poate declanșa ajustări strategice prin SRM.

