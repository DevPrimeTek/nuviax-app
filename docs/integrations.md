# Integrări Externe — NuviaX

> Detalii implementare AI (Claude Haiku) și Email (Resend).

---

## AI — Claude Haiku 4.5

**Status:** ✅ Implementat în v10.2

| Parametru | Valoare |
|-----------|---------|
| Model | `claude-haiku-4-5-20251001` |
| Provider | Anthropic API |
| Fișier | `backend/internal/ai/ai.go` |
| Client | HTTP direct (stdlib `net/http`, fără SDK) |
| Timeout | 12 secunde |
| Cost estimat | $4–5/lună la 1.000 utilizatori activi |

**Environment variable:**
```env
ANTHROPIC_API_KEY=sk-ant-...
```

**Graceful degradation:** dacă `ANTHROPIC_API_KEY` lipsește → `IsAvailable()` returnează false → fallback automat pe rule-based, fără erori vizibile.

### Unde e folosit

| Locație | Funcție | Fallback |
|---------|---------|---------|
| `level1_structural.go → generateTaskTexts` | Generare sarcini zilnice | Template-uri statice |
| `handlers.go → AnalyzeGO` | Analiză și clasificare GO | Rule-based |

### Structura clientului

```go
// New() returnează nil dacă ANTHROPIC_API_KEY lipsește
client, err := ai.New()
if err != nil {
    // folosește fallback
}

// Verificare rapidă
if !ai.IsAvailable() {
    // skip AI, go to fallback
}
```

---

## Email — Resend.com

**Status:** ✅ Implementat în v10.3

| Parametru | Valoare |
|-----------|---------|
| Provider | Resend.com |
| Fișier | `backend/internal/email/email.go` |
| Client | HTTP direct (fără SDK) |
| Domeniu trimitere | `noreply@nuviax.app` |
| Fallback | Fire-and-forget în goroutine (erori nu blochează requestul) |

**Environment variables:**
```env
RESEND_API_KEY=re_...
EMAIL_FROM=noreply@nuviax.app
```

### DNS records necesare pe name.com

```
TXT  @                    "v=spf1 include:spf.resend.com ~all"
TXT  resend._domainkey    [DKIM key din Resend dashboard]
CNAME send                [tracking domain din Resend]
```

### Email-uri tranzacționale implementate

| Email | Declanșator | Handler |
|-------|------------|---------|
| Welcome | `POST /auth/register` | `handlers.go → Register` (goroutine) |
| Sprint completat | Scheduler `jobCloseExpiredSprints` | `scheduler.go` |
| Reset parolă | `POST /auth/forgot-password` | `handlers.go → ForgotPassword` (timing-safe) |

### Flow forgot-password

```
POST /auth/forgot-password
  → generează token (1h TTL, single-use)
  → INSERT password_reset_tokens (migration 009)
  → trimite email cu link reset
  → returnează MEREU 200 (previne user enumeration)

POST /auth/reset-password
  → validează token (hash, expirat, folosit)
  → actualizează parola (bcrypt cost 14)
  → marchează token ca used_at
```

---

*Actualizat: v10.4.1 — 2026-03-26*
