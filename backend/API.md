# NUViaX API — Documentație Endpoints
## Base URL: `https://api.nuviax.app/api/v1`

---

## AUTH

### POST /auth/register
```json
// Request
{ "email": "user@email.com", "password": "min8chars", "full_name": "Ion Popescu", "locale": "ro" }

// Response 201
{ "access_token": "eyJ...", "refresh_token": "hex64...", "expires_in": 900 }
```

### POST /auth/login
```json
// Request
{ "email": "user@email.com", "password": "parola123" }

// Response 200 (fără MFA)
{ "access_token": "eyJ...", "refresh_token": "hex64...", "expires_in": 900 }

// Response 200 (cu MFA)
{ "mfa_required": true, "mfa_token": "hex32..." }
```

### POST /auth/mfa/verify
```json
// Request
{ "mfa_token": "hex32...", "code": "123456" }
// Response 200 → same as login
```

### POST /auth/refresh
```json
// Request
{ "refresh_token": "hex64..." }
// Response 200 → new token pair
```

### POST /auth/logout  `[JWT required]`
```json
// Body: { "refresh_token": "hex64..." }
// Response: { "message": "Deconectat cu succes." }
```

---

## DASHBOARD

### GET /dashboard  `[JWT required]`
```json
// Response 200
{
  "user": { "id": "uuid", "full_name": "Ion", "locale": "ro" },
  "active_goals": [
    {
      "id": "uuid", "name": "MRR Freelancing 2k→5k RON",
      "status": "ACTIVE",
      "progress_score": 0.84,   // opac 0-1
      "grade": "A",
      "days_left": 194,
      "sprint_number": 3,
      "total_sprints": 9
    }
  ],
  "waiting_goals": [...],
  "today_tasks_count": 2
}
```

---

## GOALS

### GET /goals  `[JWT required]`
Lista tuturor obiectivelor (active + în așteptare)

### POST /goals  `[JWT required]`
```json
// Request
{
  "name": "MRR Freelancing 2.000 → 5.000 RON",
  "description": "Creștere venit lunar prin clienți noi",
  "start_date": "2026-01-01",
  "end_date": "2026-09-30",
  "waiting_list": false   // true → merge în lista de așteptare
}
// Response 201 → Goal object
// Sprint 1 creat automat dacă status=ACTIVE
```

### GET /goals/:id  `[JWT required]`
```json
// Response 200
{
  "goal": { ... },
  "score": 0.84,           // opac
  "grade": "A",
  "grade_label": "Excelent",
  "progress_pct": 56,
  "days_left": 194,
  "sprint_history": [{ "sprint_id": "...", "score": 0.78, "grade": "B" }],
  "current_sprint": { ... },
  "checkpoints": [
    { "name": "Identifică 5 clienți", "status": "COMPLETED", "progress_pct": 100 },
    { "name": "Creează pachetul fix", "status": "IN_PROGRESS", "progress_pct": 45 },
    { "name": "Semnează contract nou", "status": "UPCOMING", "progress_pct": 0 }
  ]
}
```

### PATCH /goals/:id  `[JWT required]`
```json
{ "name": "Nume nou", "description": "Descriere actualizată" }
```

### DELETE /goals/:id  `[JWT required]`
Arhivează obiectivul (soft delete)

### GET /goals/:id/progress  `[JWT required]`
```json
{ "score": 0.84, "grade": "A", "progress_pct": 56 }
```

### POST /goals/:id/activate  `[JWT required]`
Activează un obiectiv din lista de așteptare
```json
// Response 200
{ "message": "Obiectivul a fost activat.", "warning": "" }
// Response 422 dacă sunt deja 3 active
{ "error": "Poți lucra la maxim 3 obiective în același timp." }
```

---

## TODAY

### GET /today  `[JWT required]`
```json
// Response 200
{
  "date": "2026-03-12T00:00:00Z",
  "goal_name": "MRR Freelancing",
  "day_number": 14,
  "main_tasks": [
    {
      "id": "uuid", "text": "Trimite oferta la ClientX",
      "type": "MAIN", "completed": false, "sort_order": 0
    }
  ],
  "personal_tasks": [
    { "id": "uuid", "text": "Citesc articol", "type": "PERSONAL", "completed": false }
  ],
  "done_count": 1,
  "total_count": 3,
  "streak_days": 7,
  "checkpoint": { "name": "Creează pachetul fix", "progress_pct": 45 }
}
```

### POST /today/complete/:id  `[JWT required]`
```json
// Response 200
{ "message": "Activitate bifată." }
```

### POST /today/personal  `[JWT required]`
```json
// Request
{ "text": "Citesc un capitol din Pricing Design", "duration_minutes": 30 }
// Response 201 → DailyTask object
// Response 422 dacă sunt deja 2 activități personale azi
```

---

## SPRINTS

### GET /sprints/current/:goalId  `[JWT required]`
```json
{
  "sprint": { "id": "uuid", "sprint_number": 3, "start_date": "...", "end_date": "..." },
  "checkpoints": [...]
}
```

### GET /sprints/:id/score  `[JWT required]`
```json
{ "score": 0.84, "grade": "A" }
```

### POST /sprints/:id/reflection  `[JWT required]`
```json
// Request (toate câmpurile opționale)
{ "q1": "A mers bine contactul cu clienții", "q2": "Lipsa timpului", "energy_level": 8 }
// Response 200 → { "message": "Reflecție salvată." }
```

### POST /sprints/:id/close  `[JWT required]`
Finalizează etapa curentă + creează etapa următoare automat
```json
// Response 200
{ "message": "Etapă finalizată.", "score": 0.84, "grade": "A" }
```

---

## CONTEXT (Ritm)

### POST /context/pause  `[JWT required]`
```json
// Request
{ "goal_id": "uuid", "days": 3, "note": "Plecare în vacanță" }
// Response 201
{ "message": "Pauză activată.", "start_date": "...", "end_date": "..." }
// Max 30 zile pauză pe etapă (validat în engine)
```

### POST /context/energy  `[JWT required]`
```json
// Request
{ "goal_id": "uuid", "level": "low" }   // "low" | "normal" | "high"
// Response 200
{ "message": "Nivel de energie actualizat. Activitățile de mâine vor fi adaptate." }
```

### GET /context/current/:goalId  `[JWT required]`
```json
{ "adjustments": [...] }
```

---

## SETTINGS

### GET /settings  `[JWT required]`
```json
{
  "locale": "ro",
  "notifications_on": true,
  "reminder_hour": 8,
  "sprint_reflection": true,
  "show_progress_chart": true
}
```

### PATCH /settings  `[JWT required]`
```json
{ "locale": "en" }
```

### GET /settings/sessions  `[JWT required]`
```json
{ "sessions": [...], "count": 2 }
```

### DELETE /settings/sessions/:id  `[JWT required]`
Deconectează un dispozitiv specific

### GET /settings/export  `[JWT required]`
Exportă toate datele personale (GDPR)

---

## CODURI DE EROARE

| Cod | Semnificație |
|-----|-------------|
| 400 | Date invalide / format greșit |
| 401 | Neautentificat / token expirat |
| 404 | Resursa nu există |
| 409 | Conflict (ex: email deja folosit) |
| 422 | Regulă business violată (ex: max 3 obiective) |
| 429 | Rate limit depășit |
| 503 | Serviciu degradat (DB / Redis down) |

---

## NOTE DE SECURITATE

- Toate endpoint-urile `/auth/*` au rate limit 10 req/min per IP
- Token-ul de acces expiră în **15 minute** (RS256)
- Refresh token-ul expiră în **7 zile** și se rotează la fiecare refresh
- Logout invalidează imediat access token-ul (Redis blacklist)
- Email-ul este stocat **criptat AES-256-GCM** în baza de date
- API-ul nu returnează NICIODATĂ: formule, ponderi, parametri interni
