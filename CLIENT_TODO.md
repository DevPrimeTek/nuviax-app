# CLIENT_TODO.md — Acțiuni necesare din partea ta

> Codul e scris, dar aceste task-uri necesită acțiuni manuale din partea ta.
> Bifează fiecare punct după ce îl finalizezi.

---

## ⚠️ ACȚIUNE URGENTĂ — Rotire chei API

**Problemă:** Cheile API reale au fost commituite accidental în `.env.example` (commits `76166c2`, `f678ae9`, `ddd5e9e`, `958a4ce`). Istoricul git este public.

**Pași imediat:**
1. Mergi pe [console.anthropic.com](https://console.anthropic.com) → **API Keys** → revocă cheia curentă → creează una nouă
2. Mergi pe [resend.com/api-keys](https://resend.com/api-keys) → șterge cheia curentă → creează una nouă
3. Actualizează **GitHub Secrets** cu noile chei (aceeași procedură ca mai jos)
4. Actualizează **VPS `.env`** cu noile chei
5. Actualizează `.env.example` — pune înapoi placeholder `CHANGE_ME` (nu cheia reală)

---

## 1. 🤖 Anthropic API Key (Claude Haiku)

**Status cod:** ✅ Implementat în `backend/internal/ai/ai.go`
**Status tău:** ✅ Cheia a fost adăugată — ⚠️ rotire necesară (vezi mai sus)

### Ce trebuie să faci

**Pasul 1 — Obține API Key:**
1. Mergi pe [console.anthropic.com](https://console.anthropic.com)
2. Creează cont sau loghează-te
3. Navighează la **API Keys** → **Create Key**
4. Copiază cheia (începe cu `sk-ant-api03-...`)
5. **Salvează-o imediat** — nu o mai poți vedea după ce închizi pagina

**Pasul 2 — Adaugă în GitHub Secrets:**
1. Mergi la: `https://github.com/DevPrimeTek/nuviax-app/settings/secrets/actions`
2. Click **New repository secret**
3. Name: `ANTHROPIC_API_KEY`
4. Value: `sk-ant-api03-...` (cheia copiată)
5. Click **Add secret**

**Pasul 3 — Adaugă în `.env` local (pentru development):**
```env
# În fișierul infra/.env
ANTHROPIC_API_KEY=sk-ant-api03-...
```

**Pasul 4 — Adaugă în `.env` pe server:**
```bash
# SSH pe server
ssh sbarbu@83.143.69.103
cd /var/www/wxr-nuviax
nano infra/.env
# Adaugă linia: ANTHROPIC_API_KEY=sk-ant-api03-...
# Ctrl+X → Y → Enter
docker compose -f infra/docker-compose.yml up -d --no-build nuviax_api
```

**Pasul 5 — Verificare:**
```bash
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}

# Testează AI task generation: creează un obiectiv și verifică
# că sarcinile zilnice sunt generate cu text real (nu template)
```

**Cost estimat:** ~$4–5/lună la 1.000 utilizatori activi (Claude Haiku e cel mai ieftin model)

**Fără cheie:** Aplicația funcționează (graceful degradation), dar sarcinile zilnice vor fi template-uri statice generice, nu generate de AI.

---

## 2. 📧 Resend Email

**Status cod:** ✅ Implementat în `backend/internal/email/email.go`
**Status tău:** ✅ Cheia a fost adăugată — ⚠️ rotire necesară (vezi mai sus)

### Ce trebuie să faci

**Pasul 1 — Creează cont Resend:**
1. Mergi pe [resend.com](https://resend.com)
2. **Sign Up** cu emailul tău de business
3. Contul gratuit include 3.000 emailuri/lună — suficient pentru start

**Pasul 2 — Adaugă domeniul `nuviax.app`:**
1. În Resend Dashboard → **Domains** → **Add Domain**
2. Introdu: `nuviax.app`
3. Resend îți va da 3 DNS records de adăugat

**Pasul 3 — Adaugă DNS records pe name.com:**
1. Loghează-te pe [name.com](https://name.com)
2. Mergi la **My Domains** → `nuviax.app` → **Manage DNS**
3. Adaugă exact aceste records (valorile exacte le iei din Resend Dashboard):

```
Tip     Nume                    Valoare
TXT     @                       "v=spf1 include:spf.resend.com ~all"
TXT     resend._domainkey       [valoare DKIM din Resend — e lungă]
CNAME   send                    [tracking domain din Resend]
```

4. Aștepți 10–30 minute pentru propagare DNS
5. În Resend Dashboard → **Verify** — trebuie să apară ✅ lângă toate 3

**Pasul 4 — Obține API Key Resend:**
1. Resend Dashboard → **API Keys** → **Create API Key**
2. Nume: `nuviax-production`
3. Permisiuni: **Full access**
4. Copiază cheia (începe cu `re_...`)

**Pasul 5 — Adaugă în GitHub Secrets:**
1. `https://github.com/DevPrimeTek/nuviax-app/settings/secrets/actions`
2. **New repository secret** → Name: `RESEND_API_KEY` → Value: `re_...`

**Pasul 6 — Adaugă în `.env` pe server:**
```bash
ssh sbarbu@83.143.69.103
cd /var/www/wxr-nuviax
nano infra/.env
# Adaugă:
# RESEND_API_KEY=re_...
# EMAIL_FROM=noreply@nuviax.app
docker compose -f infra/docker-compose.yml up -d --no-build nuviax_api
```

**Pasul 7 — Testare:**
```bash
# Înregistrează un cont nou pe nuviax.app
# Verifică că primești emailul de Welcome
# Testează și forgot-password cu emailul tău
```

**Fără cheie:** Aplicația funcționează, dar niciun email nu se trimite (welcome, reset parolă, sprint complet).

---

## 3. 👤 Admin Panel — Finalizare

**Status cod:** ✅ Codul admin e scris (frontend + backend)
**Status tău:** ❌ Nu există niciun cont admin creat. Panoul nu poate fi accesat.

### Ce trebuie să faci

**Pasul 1 — Asigură-te că aplicația rulează pe server:**
```bash
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}
```

**Pasul 2 — Creează contul tău de admin:**

Opțiunea A — Folosind scriptul automat (recomandat):
```bash
ssh sbarbu@83.143.69.103
cd /var/www/wxr-nuviax
bash scripts/setup_admin.sh
# Scriptul te va întreba email + parolă și setează is_admin=TRUE automat
```

Opțiunea B — Manual (dacă scriptul nu funcționează):
```bash
# 1. Înregistrează-te pe nuviax.app cu emailul tău
# 2. Conectează-te la DB și setează manual:
docker exec -it nuviax_db psql -U nuviax -d nuviax
```
```sql
-- Găsește user-ul tău (emailul e criptat, caută după data înregistrării)
SELECT id, created_at, is_admin FROM users ORDER BY created_at DESC LIMIT 5;

-- Setează admin (înlocuiește UUID-ul)
UPDATE users SET is_admin = TRUE WHERE id = 'uuid-din-query-de-sus';

-- Verificare
SELECT id, is_admin FROM users WHERE id = 'uuid-din-query-de-sus';
-- Expected: is_admin = true
\q
```

**Pasul 3 — Testare panel admin:**
1. Loghează-te pe [nuviax.app](https://nuviax.app) cu contul tău
2. Mergi la: `https://nuviax.app/admin`
3. Trebuie să vezi panoul cu 4 tab-uri: Statistici, Utilizatori, Audit, Sistem
4. Verifică că link-ul "Admin" apare în navigare (doar pentru contul tău)

**Dacă panoul arată eroare "Acces restricționat"** → is_admin nu e setat corect → reparcurge Pasul 2

**Pasul 4 — Ce mai e de făcut în admin (cunoscut):**
- [ ] Verificat că toate statisticile se afișează corect
- [ ] Verificat că lista utilizatori funcționează (search, activate/deactivate)
- [ ] Verificat că Audit log are intrări
- [ ] Verificat că Health tab arată DB + Redis status
- [ ] Testat Dev Reset (ATENȚIE: șterge toate datele — folosește doar în development)

---

## Sumar rapid

| Task | Status | Prioritate |
|------|--------|------------|
| ⚠️ Rotire ANTHROPIC_API_KEY | 🔴 Urgent — cheie expusă în git | Imediat |
| ⚠️ Rotire RESEND_API_KEY | 🔴 Urgent — cheie expusă în git | Imediat |
| Admin cont creat pe VPS | ❌ Nu e creat încă | Ridicată |
| Verificat emailuri (welcome, reset) | ❓ Netestat | Medie |
| Verificat DNS Resend (SPF, DKIM) | ❓ Neverificat | Medie |

**Ordinea recomandată:** Rotire chei (urgent) → Admin cont → Verificare email deliverability

---

*Ultima actualizare: 2026-03-29 — cheile API au fost adăugate dar necesită rotire urgentă*
