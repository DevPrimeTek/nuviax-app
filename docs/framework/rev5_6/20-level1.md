# **LEVEL 1 - STRUCTURAL AUTHORITY**

Structural Authority este nivelul sistemului responsabil pentru transformarea intenției brute a utilizatorului într-o structură strategică coerentă și validă.

Acest nivel nu execută obiective și nu monitorizează progresul. Rolul lui este să clarifice, să filtreze și să limiteze obiectivele înainte ca acestea să devină parte a arhitecturii executive a sistemului. În această etapă sistemul transformă intențiile utilizatorului în Global Objectives valide și stabilește direcția strategică a anului.

### **Scopul acestui nivel**

- transformarea intenției brute în obiective clare
- eliminarea ambiguității strategice
- prevenirea supraîncărcării cognitive
- stabilirea unei direcții dominante pentru anul curent
- crearea unei baze stabile pentru execuția ulterioară

### **Regula fundamentală**

Un obiectiv nu poate deveni activ dacă nu este clar, limitat și validat structural. Dacă utilizatorul definește obiective vagi sau contradictorii, sistemul: clarifică formularea, identifică Behavior Model dominant, analizează relevanța strategică, limitează numărul de obiective active. Doar obiectivele validate pot intra în arhitectura de execuție.

### **Relația cu restul sistemului**

Structural Authority urmează după User Intent (Raw Intent Input) și precede Execution Architecture, Monitoring Authority și Capacity Regulation. Fără acest nivel, sistemul ar transforma intenții confuze în execuție instabilă.

## **PHASE I - CLARITY ARCHITECTURE**

Clarity Architecture este mecanismul structural prin care intențiile nestructurate ale utilizatorului sunt transformate în direcții strategice formalizate, compatibile cu arhitectura Framework-ului. Această fază este obligatorie pentru utilizatorii care: nu au claritate strategică, formulează obiective ambigue, exprimă intenții multiple simultan sau formulează obiective hibride.

### **Scopul acestei faze**

- identificarea direcției dominante
- eliminarea ambiguității
- prevenirea supra-extinderii
- forțarea încadrării într-un Behavior Model

Principiu Structural: Clarity Architecture NU creează obiective suplimentare. Este o fază de disciplinare, nu de expansiune.

LEVEL 1 - Phase I: Clarity Architecture

**C9 Semantic Parsing**

#### **Definiție**

Procesul prin care textul introdus de utilizator este analizat pentru extragerea elementelor structurale necesare definirii unui Global Objective valid.

#### **De unde vine logica**

Intenția umană poate fi exprimată în limbaj vag. Exemplu: „Vreau libertate financiară." Aceasta nu este măsurabilă și nu poate fi executată direct. Sistemul trebuie să transforme intenția în variabile clare înainte de orice evaluare strategică.

#### **Reprezentare formală**

I = text brut introdus

SP(I) = { domain, direction, metric, timeframe }

Elementele extrase:

domain: domeniu (financiar, sănătate, familie, carieră etc.)

direction: direcție (creare, creștere, reducere, menținere, evoluție)

metric: ce anume se schimbă concret și măsurabil

timeframe: orizontul temporal

Condiții speciale detectate la parsare:

(1) domain=Sănătate_fizică AND direction=REDUCE AND delta > 25%

→ SINGLE_QUESTION_FLAG = TRUE

(2) BM_opus_1 și BM_opus_2 pe aceeași metrică

→ GO_REJECTED_LOGICAL_CONTRADICTION

(3) Parametri insuficienți sau ambigui → Reformulation Queue

#### **Physical Delta Safety Signal**

Dacă domeniul este Sănătate fizică, direcția este REDUCE și delta depășește 25% din valoarea inițială, sistemul setează SINGLE_QUESTION_FLAG. La activarea GO-ului utilizatorul primește un mesaj unic care informează despre importanța consultării unui specialist - un singur mesaj, un singur buton de confirmare. Sprint 1, Milestone 1 conține obligatoriu o acțiune de tip conector specialist. Dacă acțiunea nu este bifată până la finalul Sprint 1, reapare în Sprint 3 o singură dată.

#### **GO_REJECTED_LOGICAL_CONTRADICTION**

Detectarea a două BM-uri opuse pe aceeași metrică generează această categorie specifică. Utilizatorul primește 3 opțiuni predefinite pentru a rezolva contradicția. Dacă nu răspunde în 48h → PENDING_CLARIFICATION. La 7 zile fără răspuns → arhivat automat în Vault. Nu există stare de limbo.

#### **Reformulation Queue**

Când Semantic Parsing detectează parametri insuficienți pentru mai multe GO-uri simultan, cererile de clarificare sunt puse în coadă și prezentate strict secvențial - un singur GO la un moment dat, în ordinea descrescătoare a Relevanței estimate. Dacă utilizatorul nu răspunde la o cerere în 48h, GO-ul respectiv este arhivat în Vault automat și coada continuă cu GO-ul următor.

#### **Rol**

- Elimină ambiguitatea - creează baza structurată pentru GO valid.
- Permite clasificarea comportamentală în Behavior Model.
- Detectează timpuriu conflictele logice și riscurile speciale.

LEVEL 1 - Phase I: Clarity Architecture

**C10 Behavior Model Classification**

#### **Definiție**

Atribuirea unui singur Behavior Model dominant fiecărui GO, pe baza parametrilor extrași de Semantic Parsing.

#### **De unde vine logica**

Un obiectiv cu două direcții simultane devine instabil. Exemplu: „Vreau să cresc venit și să reduc programul de lucru." Acesta conține două modele diferite - INCREASE și REDUCE. Framework-ul impune alegerea unui model dominant. Sistemul nu poate genera milestone-uri coerente, nu poate calcula Sprint Target și nu poate evalua progresul pentru un obiectiv cu direcție dublă.

#### **Reprezentare formală**

∀ GO → ∃! Behavior Model (pentru fiecare GO există exact un singur BM)

BM_scores = { CREATE: p1, INCREASE: p2, REDUCE: p3, MAINTAIN: p4, EVOLVE: p5 }

BM_dominant = argmax(BM_scores)

confidence ≥ 0.70 → auto-select, fără interacțiune cu utilizatorul

confidence ∈ \[0.50,0.70) → Confidence Gate: clarificare utilizator

confidence < 0.50 → returnare la C9 Reformulation Queue

#### **Explicație**

Sistemul întreabă: „Acest obiectiv este în esență despre creare, creștere, reducere, menținere sau evoluție?" Dacă obiectivul încearcă să facă mai multe lucruri fundamentale simultan, sistemul îl obligă să aleagă direcția principală. EVOLVE este selectat automat când analiza detectează transformare calitativă multi-dimensională.

#### **Rol**

- Asigură coerență direcțională - un GO = o direcție.
- Permite formularea corectă și coerentă a milestone-urilor.
- Stabilizează Sprint Target Calculation și calculul efortului.

LEVEL 1 - Phase I: Clarity Architecture

**C11 Strategic Relevance Scoring**

#### **Definiție**

Evaluarea impactului real al unui GO în contextul strategic al utilizatorului în următoarele 12 luni. Relevance ∈ \[0, 1\], unde 0 = relevanță foarte scăzută și 1 = relevanță strategică maximă.

#### **De unde vine logica**

Nu toate dorințele sunt obiective strategice. Unele sunt emoționale sau temporare. Un obiectiv cu Relevance scăzută ar consuma resurse reale (timp, energie, capacitate ALI) fără contribuție strategică semnificativă. Minimum Relevance Floor la 0.30 elimină această categorie înainte de activare.

#### **Reprezentare formală**

Relevance = Impact × 0.35 + Urgency × 0.25 + Alignment × 0.25 + Feasibility × 0.15

Impact: contribuția la obiective de viață de ordin superior

Urgency: presiunea temporală și costul amânării

Alignment: coerența cu valorile, identitatea și alte GO active

Feasibility: probabilitatea de execuție cu resursele disponibile

Relevance_adj = round(Relevance_brut, 2) - înainte de orice mapping

Minimum Relevance Floor: Relevance_adj < 0.30 → Vault automat

#### **Recalibrare Post-Sprint 1**

La finalul Sprint 1, sistemul recalculează Relevance pe baza comportamentului real de execuție. Dacă utilizatorul a lucrat mai puțin de 50% din zilele Sprint 1, componentele Urgency și Feasibility sunt recalibrate descendent. Actualizarea este silențioasă - nu generează notificare și nu întrerupe fluxul utilizatorului. Orice modificare a Relevance declanșează automat check_priority_balance() prin C8.

#### **Interacțiuni cu alte componente**

- Determină Priority Weight (C7, Layer 0).
- Influențează selecția Top-3 (C13).
- Este utilizat în Annual Relevance Recalibration (C18).

LEVEL 1 - Phase I: Clarity Architecture

**C12 Resource Conflict Detection**

#### **Definiție**

Identificarea suprapunerii de resurse între GO-uri. Acest mecanism este un filtru informativ pre-activare și nu înlocuiește ALI Engine. Resource Conflict detectează suprapuneri calitative între obiective înainte de activare; ALI Engine efectuează ulterior evaluarea numerică a capacității reale. Cele două mecanisme sunt complementare.

#### **De unde vine logica**

Dacă două obiective cer aceleași resurse simultan, conflictul crește și scade probabilitatea de reușită pentru ambele. Detectarea timpurie permite ajustarea load-ului înainte ca problema să devină vizibilă în execuție (Drift, ALI).

#### **Reprezentare formală**

Conflict_score = Overlap(Timp) × 0.40 + Overlap(Energie) × 0.30 + Overlap(Capital) × 0.30

Conflict_score < 0.40 → activare normală

0.40 ≤ Conflict_score < 0.70 → AMBER: load ajustat −15%, fără blocare

Conflict_score ≥ 0.70 → activare blocată → C17 Future Vault

#### **Explicație**

Sistemul verifică: dacă două obiective cer același timp, dacă cer aceeași energie, dacă cer același capital. Dacă suprapunerea este mare, apare avertizare sau blocare. Ajustarea AMBER (−15% load) este aplicată silențios în Daily Stack Generator - utilizatorul vede un plan mai puțin dens, fără explicație explicită.

#### **Rol**

- Previne supraîncărcarea - reduce conflictul de resurse înainte de activare.
- Ajustează selecția strategică prin influențarea Feasibility în C11.
- Influențează ALI ulterior prin load-ul real generat.

LEVEL 1 - Phase I: Clarity Architecture

**C13 Top-3 Candidate Selection**

#### **Definiție**

Selectarea a maximum 3 GO cu relevanța cea mai mare, după aplicarea tuturor filtrelor anterioare.

#### **Reprezentare formală**

Pas 1: Filtrare - elimină toate GO cu Relevance_adj < 0.30

Pas 2: Sortare - descrescător după Relevance_adj

Pas 3: Selecție - primele 3 → ACTIVE, restul → C17 Future Vault

#### **Explicație**

Obiectivele sunt ordonate după importanța lor strategică calculată. Primele trei devin active. Restul intră în Future Vault. GO-urile eliminate la Pasul 1 (Relevance < 0.30) intră în Vault cu mesajul: „Salvat pentru mai târziu - scopul acesta nu pare suficient de important în prezent." GO-urile eliminate la Pasul 3 (Relevance ≥ 0.30 dar al 4-lea sau mai jos) intră în Vault cu mesajul: „Prioritizat pentru o perioadă viitoare - ai 3 obiective active acum."

#### **Rol**

- Aplică constrângerea Maximum 3 Active GO (C3, Layer 0).
- Menține focus strategic - forțează decizia de prioritizare.

## **PHASE II - STRATEGIC LIMITATION**

Strategic Limitation este mecanismul formal prin care sistemul limitează utilizatorul la maximum trei Global Objectives active simultan. Aceasta nu este o sugestie, este o constrângere structurală.

### **Scopul acestei faze**

- prevenirea fragmentării atenției
- prevenirea supra-ambiției
- stabilirea unei direcții dominante clare
- protejarea capacității cognitive

### **Regula Fundamentală**

Maxim 3 GO active simultan. Dacă utilizatorul definește 5 obiective, sistemul: solicită prioritizare, recomandă Top 3, mută restul în Future Vault. Nu este permisă activarea a 4 sau mai multe obiective simultan.

### **Validarea GO**

Fiecare GO trebuie: să fie încadrat într-un Behavior Model dominant, să aibă deadline ≤ 365 zile, să aibă priority_weight derivat din Relevance și să fie compatibil cu Strategic Feasibility Analysis (pre-activare). Dacă nu îndeplinește aceste condiții, GO este respins sau trimis la reformulare.

### **Relația cu restul sistemului**

Strategic Limitation precede: Sprint Design, ALI, Drift și Chaos. Fără limitare structurală, sistemul devine instabil - execuția ar porni de la obiective nevalidate sau contradictorii.

LEVEL 1 - Phase II: Strategic Limitation

**C14 Global Objective Validation**

#### **Definiție**

Procesul prin care sistemul verifică dacă un GO este structural valid și eligibil pentru analiza de fezabilitate și activare. Aceasta este ultima verificare structurală înainte ca un GO să poată deveni activ.

#### **De unde vine logica**

Dacă un GO este vag, nemăsurabil, nelimitat temporal sau cu model comportamental ambiguu, atunci orice analiză ulterioară (fezabilitate, ALI, sprinturi) devine instabilă. Validarea trebuie să fie un filtru strict înainte de orice activare.

#### **Reprezentare formală**

Valid(GO) = TRUE dacă TOATE condițiile sunt satisfăcute:

(1) deadline − start_date ≤ 365 zile

(2) |Behavior Model(GO)| = 1 (unicitate verificată de C10/C15)

(3) Metric definit și măsurabil

(4) Domeniu clar identificat în taxonomy

(5) Valoarea inițială (start) definită

Valid(GO) = FALSE dacă oricare condiție este nesatisfăcută

→ cerere de completare specifică

→ Timeout 7 zile fără completare → Vault automat

#### **Explicație**

Sistemul verifică: Are obiectivul un termen clar? Are o singură direcție principală? Poate fi măsurat progresul? Este clar în ce domeniu acționează? Dacă oricare dintre aceste condiții nu este îndeplinită, GO nu poate trece la următorul pas. Sistemul nu respinge GO-ul definitiv - îl suspendă în PENDING_VALIDATION până la completare.

#### **Rol**

- Previne activarea obiectivelor abstracte sau incomplet definite.
- Protejează axioma de 365 zile și unicitatea BM.
- Asigură coerența bazei de date necesară pentru Strategic Feasibility Analysis.

#### **Interacțiuni cu alte componente**

- Primește input din Phase I (Clarity Architecture).
- Trimite GO validat către Strategic Feasibility Analysis (C16).
- Dacă invalid → cere reformulare specifică.

LEVEL 1 - Phase II: Strategic Limitation

**C15 Behavior Dominance Enforcement**

#### **Definiție**

Mecanismul prin care sistemul impune ca fiecare Global Objective să fie încadrat într-un singur Behavior Model dominant la intrarea în C14 Global Objective Validation.

#### **De unde vine logica**

Dacă un GO conține mai multe direcții comportamentale simultan: devine ambiguu, generează milestone-uri contradictorii, produce conflicte de prioritizare, destabilizează Sprint Target Calculation și afectează Drift și GORI. Pentru stabilitate matematică și executivă, fiecare GO trebuie să aibă o singură direcție principală.

#### **Reprezentare formală**

BM(GO) = setul modelelor comportamentale asociate GO

Condiție: |BM(GO)| = 1

|BM(GO)| = 1 → avansare în C14 Global Objective Validation

|BM(GO)| = 0 → returnare la C9 (parsare incompletă)

|BM(GO)| > 1 → GO trebuie reformulat

GO_REJECTED_LOGICAL_CONTRADICTION → blocat, așteptare rezoluție

#### **Exemplu**

„Vreau să slăbesc și să cresc masă musculară" - amestec între REDUCE și INCREASE pe metrici de compoziție corporală. Trebuie stabilit care este dominant. Dacă ambele sunt la paritate de intensitate pe aceeași metrică numerică → GO_REJECTED_LOGICAL_CONTRADICTION (C9). Dacă sunt pe metrici separate → EVOLVE poate fi potrivit.

#### **Rol**

- Elimină ambiguitatea strategică înainte de activare.
- Permite formularea corectă a milestone-urilor și calcul coerent al efortului.
- Stabilizează Strategic Feasibility Analysis și comparabilitatea între GO-uri.

#### **Interacțiuni cu alte componente**

- Primește date din Behavior Model Classification (C10, Phase I).
- Precede Global Objective Validation (C14).
- Influențează Sprint Target Calculation (C20).

LEVEL 1 - Phase II: Strategic Limitation

**C16 Strategic Feasibility Analysis**

#### **Definiție**

Mecanismul prin care sistemul determină dacă un GO poate fi finalizat realist în termen de 365 zile, ținând cont de capacitatea utilizatorului și distribuția resurselor între GO-urile active.

#### **De unde vine logica**

Axioma 365-Day este rigidă. Pentru a preveni eșecul inevitabil, carry-over și haos inter-anual, fezabilitatea trebuie evaluată înainte de activare. Strategic Feasibility Analysis este evaluarea pre-activare bazată pe estimări. ALI Engine este evaluarea dinamică post-activare bazată pe execuție reală. Cele două mecanisme sunt complementare și nu se suprapun.

#### **Reprezentare formală**

E_total = efort total estimat pentru GO

C_annual = capacitate anuală estimată a utilizatorului

n = numărul de GO active simultan

C_per_GO = C_annual / n

Load_ratio = E_total / C_per_GO

Load_ratio ≤ 1.0 → activare normală

Load_ratio ∈ (1.0, 1.10\] → Ambition Buffer: activare cu avertisment

Load_ratio > 1.10 → capacitate insuficientă → C17 Future Vault

Capacity Validation Gate:

C_daily < 0.5 h/zi → avertisment de subcapacitate (nu blocare)

C_daily > 14 h/zi → avertisment de supracapacitate (nu blocare)

#### **Explicație**

Sistemul verifică: „Poți termina acest obiectiv în 1 an cu resursele tale actuale?" Dacă obiectivul cere mai mult decât poate fi realizat, nu este activat în forma actuală. Intervalul (1.0, 1.10\] este Ambition Buffer - o planificare ușor optimistă este acceptabilă. Dacă load-ul nu scade în Sprint 1, recalibrarea post-Sprint1 din C11 și C20 va ajusta organic.

#### **Rol**

- Protejează axioma anuală - previne activarea obiectivelor nefinalizabile în 365 zile.
- Reduce eșecul structural și stabilizează GORI.
- Elimină necesitatea carry-over inter-anual.

LEVEL 1 - Phase II: Strategic Limitation

**C17 Future Vault System**

#### **Definiție**

Mecanismul structural prin care sistemul stochează GO care nu pot deveni active în momentul curent, fără a le elimina din sistem. Un GO poate intra în Vault din mai multe motive: depășește limita de 3 GO active, nu trece Strategic Feasibility Analysis, Relevance < 0.30, sau timeout-uri de clarificare/validare depășite.

#### **De unde vine logica**

Sistemul trebuie să păstreze ideile utilizatorului, dar să mențină stabilitatea structurală. Eliminarea definitivă ar genera frustrare. Activarea directă cu toate obiectivele dorite ar genera haos. Vault este mecanismul de echilibru: „Nu acum. Poate mai târziu."

#### **Reprezentare formală**

GO.status = VAULT dacă oricare din condițiile de intrare:

(1) |Active_GO| = 3 și tentativă de activare nouă (C3)

(2) Load_ratio > 1.10 după Feasibility Analysis (C16)

(3) Relevance_adj < 0.30 după scoring (C11)

(4) GO_REJECTED_LOGICAL_CONTRADICTION fără răspuns 7 zile (C9)

(5) Reformulation Queue abandonată > 48h (C9)

(6) PENDING_VALIDATION > 7 zile fără completare (C14)

#### **Semantica VAULT vs alte statusuri**

| **Status**         | **Semnificație**                               | **Reactivare**                                      |
| ------------------ | ---------------------------------------------- | --------------------------------------------------- |
| **VAULT**          | Niciodată activ - așteptare strategică         | Când slot disponibil sau Load_ratio ≤ 1.10          |
| **SUSPENDED**      | A fost activ - suspendat de sistem prin SRM L3 | Prin C36 Reactivation Protocol, cu Relevance ≥ 0.60 |
| **SEASONAL_PAUSE** | Activ sezonier - inactiv în fereastra curentă  | Automat la intrarea în execution_window activă      |
| **ARCHIVED**       | Închis definitiv de utilizator                 | Nu se reactivează                                   |

#### **Condiții formale GO în Vault**

- GO în Vault nu este inclus în ALI.
- GO în Vault nu are Sprint activ.
- GO în Vault poate fi activat doar dacă există slot liber (n < 3).
- GO în Vault este reevaluat la recalibrarea periodică (C18).

#### **Rol**

- Protejează limita de 3 GO - menține stabilitate anuală.
- Previne activarea obiectivelor imposibil de finalizat în 365 zile.
- Permite restructurare strategică: obiectivele stocate nu sunt pierdute.

#### **Interacțiuni cu alte componente**

- Primește GO din Top-3 Selection (C13) când limita este depășită.
- Primește GO din Strategic Feasibility Analysis (C16) când nerealist.
- Nu influențează ALI sau Drift - GO în Vault este exclus complet din calcule operative.

LEVEL 1 - Phase II: Strategic Limitation

**C18 Annual Relevance Recalibration**

#### **Definiție**

Mecanismul prin care sistemul reevaluează periodic relevanța strategică a fiecărui Global Objective, indiferent dacă este activ sau stocat în Future Vault. Scopul este să determine dacă obiectivul rămâne strategic valid, trebuie reformulat, trebuie închis sau trebuie eliminat.

#### **De unde vine logica**

Obiectivele pot deveni irelevante din cauza schimbării contextului personal, a resurselor, a priorităților sau a finalizării parțiale care modifică direcția. Fără recalibrare periodică: Vault devine depozit inert, GO active pot continua mecanic fără relevanță, sistemul devine rigid operațional.

#### **Reprezentare formală**

R_current = scorul de relevanță actual

R_initial = scorul de relevanță la activare

Relevance_ratio = R_current / R_initial

Relevance_ratio ≥ 0.70 → GO rămâne activ fără intervenție

0.40 ≤ Relevance_ratio < 0.70 → REVIEW_REQUIRED: o întrebare

Relevance_ratio < 0.40 → recomandare formală de închidere sau Vault

Prioritate: dacă GO.status = SUSPENDED → C34 Stabilization Review

are prioritate față de C18

#### **Explicație**

Sistemul verifică: „Este acest obiectiv la fel de important ca atunci când a fost creat?" Dacă relevanța scade semnificativ (sub aproximativ 70% din importanța inițială), sistemul recomandă reformulare, închidere sau înlocuire. Recalibrarea este advisory - nu modifică automat statusul fără confirmare utilizator, cu excepția cazurilor severe (Relevance_ratio < 0.40, care generează recomandare formală).

#### **Rol**

- Previne menținerea obiectivelor depășite sau devenite irelevante.
- Curăță Future Vault - forțează decizii strategice conștiente periodic.
- Menține coerența anuală a sistemului pe termen lung.

