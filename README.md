# deploy/ — Scripturi NuviaX

> Această mapă **NU apare pe GitHub** (adaugă `deploy/` în `.gitignore` rădăcină).
> Rulează scripturile din rădăcina proiectului.

---

## Sync-NuviaX.ps1

Script PowerShell 7.4.5+ pentru sync local → GitHub.

**Ce face:**
1. Verifică status git și branch curent
2. Fetch + merge cu remote (actualizează structura și fișierele)
3. Crează backup local al fișierelor modificate (în `deploy/backups/`)
4. Verifică că README.md e actualizat (versiunile coincid)
5. Verifică că nu sunt fișiere sensibile staged
6. Commit + push

### Utilizare

```powershell
# Sync simplu pe branch-ul curent
.\deploy\Sync-NuviaX.ps1 -CommitMessage "feat: G-11 behavior model dominance"

# Pe un branch specific
.\deploy\Sync-NuviaX.ps1 -CommitMessage "feat: i18n framework" -Branch "claude/i18n-sprint3"

# Fără verificare README
.\deploy\Sync-NuviaX.ps1 -CommitMessage "chore: cleanup" -SkipReadmeCheck

# Test fără să faci modificări reale
.\deploy\Sync-NuviaX.ps1 -CommitMessage "test" -DryRun
```

### Parametri

| Parametru | Obligatoriu | Descriere |
|-----------|------------|-----------|
| `-CommitMessage` | ✅ Da | Mesajul commit-ului |
| `-Branch` | Nu | Branch target (default: branch curent) |
| `-SkipReadmeCheck` | Nu | Sare verificarea versiunii README |
| `-DryRun` | Nu | Simulează fără modificări reale |

### Structura backup-uri

```
deploy/
├── backups/
│   ├── 2026-03-29_14-30-00/    # timestamp sesiune
│   │   ├── backend/
│   │   │   └── internal/engine/level5_growth.go
│   │   └── README.md
│   └── ...  (se șterg automat după 30 zile)
├── .gitignore                   # exclude backups/ din git
├── README.md                    # acest fișier
└── Sync-NuviaX.ps1             # scriptul principal
```

---

## Adaugă deploy/ în .gitignore rădăcină

```
# În .gitignore din rădăcina proiectului, adaugă:
deploy/
```

Astfel mapa `deploy/` nu apare niciodată pe GitHub.

---

*PowerShell 7.4.5+ necesar. Testează cu `-DryRun` la prima rulare.*
