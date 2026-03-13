# NUViaX App

Monorepo pentru aplicația NUViaX Growth Framework.

## Structura

```
nuviax-app/
├── backend/          # Go API (Fiber framework)
├── frontend/         # Next.js 14 (Faza 2 - în lucru)
├── landing/          # Next.js static (Faza 2 - în lucru)
├── infra/            # Docker Compose, DB schema, setup
└── .github/          # CI/CD workflows
```

## Setup server

```bash
bash infra/setup-server.sh
```

## Documentație API

Vezi `backend/API.md`
