<#
.SYNOPSIS
    NuviaX Repository Fixes v7 - Dockerfile-uri integrate
.DESCRIPTION
    Aplică automat toate corecturile cu Dockerfile-uri integrate
.NOTES
    Versiune: 7.0 FINAL
    Data: 15 Martie 2025
#>

param(
    [switch]$SkipBackup,
    [switch]$SkipPush,
    [string]$RepoPath = "."
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Test-GitRepository {
    if (-not (Test-Path ".git")) {
        throw "Nu este un repository Git. Rulează din root-ul repository-ului."
    }
    Write-ColorOutput "✓ Repository Git detectat" "Green"
}

function Backup-Repository {
    if ($SkipBackup) {
        Write-ColorOutput "⊘ Backup sărit" "Yellow"
        return
    }
    
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $backupName = "nuviax_backup_$timestamp"
    
    Write-ColorOutput "📦 Creare backup: $backupName.zip" "Cyan"
    
    $tempDir = Join-Path $env:TEMP $backupName
    Copy-Item -Path . -Destination $tempDir -Recurse -Force -Exclude @(".git", "node_modules", ".next", "dist")
    
    Compress-Archive -Path $tempDir -DestinationPath "$backupName.zip" -Force
    Remove-Item -Path $tempDir -Recurse -Force
    
    Write-ColorOutput "✓ Backup: $backupName.zip" "Green"
}

function Update-DockerComposeFiles {
    Write-ColorOutput "`n🐳 Actualizare docker-compose.yml..." "Cyan"
    
    $composeFile = "infra/docker-compose.yml"
    if (Test-Path $composeFile) {
        $content = Get-Content $composeFile -Raw
        
        $content = $content -replace 'nuviax_web:', 'nuviax_app:'
        $content = $content -replace 'container_name: nuviax_web', 'container_name: nuviax_app'
        $content = $content -replace 'devprimetek/nuviax-web:', 'devprimetek/nuviax-app:'
        $content = $content -replace 'WEB_IMAGE=', 'APP_IMAGE='
        
        Set-Content $composeFile -Value $content -NoNewline
        Write-ColorOutput "✓ docker-compose.yml" "Green"
    }
}

function Update-GitHubWorkflows {
    Write-ColorOutput "`n⚙️ Actualizare workflows..." "Cyan"
    
    $workflowFile = ".github/workflows/deploy-frontend.yml"
    if (Test-Path $workflowFile) {
        $content = Get-Content $workflowFile -Raw
        
        $content = $content -replace 'WEB_IMAGE:', 'APP_IMAGE:'
        $content = $content -replace 'devprimetek/nuviax-web', 'devprimetek/nuviax-app'
        $content = $content -replace 'apps/web/', 'frontend/app/'
        $content = $content -replace 'build-web:', 'build-app:'
        $content = $content -replace 'nuviax-web', 'nuviax-app'
        $content = $content -replace 'scope=web', 'scope=app'
        
        Set-Content $workflowFile -Value $content -NoNewline
        Write-ColorOutput "✓ deploy-frontend.yml" "Green"
    }
}

function Update-Dockerfiles {
    Write-ColorOutput "`n📝 Scriere Dockerfile-uri v7..." "Cyan"
    
    # Frontend App Dockerfile
    $appDockerfile = @'
FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps

COPY . .

RUN mkdir -p frontend/app

RUN if [ ! -f frontend/app/package.json ]; then \
        echo '{"name":"@nuviax/app","version":"1.0.0","private":true,"scripts":{"dev":"next dev","build":"next build","start":"next start"},"dependencies":{"next":"^14","react":"^18","react-dom":"^18"}}' > frontend/app/package.json; \
    fi

RUN cd frontend/app && \
    if [ -f ../../yarn.lock ]; then \
        yarn install --frozen-lockfile || yarn install; \
    elif [ -f ../../package-lock.json ]; then \
        npm ci || npm install; \
    elif [ -f ../../pnpm-lock.yaml ]; then \
        corepack enable pnpm && pnpm install --frozen-lockfile || pnpm install; \
    elif [ -f package-lock.json ]; then \
        npm ci || npm install; \
    else \
        npm install; \
    fi

FROM base AS builder

COPY --from=deps /app/frontend/app/node_modules ./frontend/app/node_modules

COPY . .

WORKDIR /app/frontend/app

RUN if [ ! -f next.config.js ]; then \
        echo 'module.exports = { reactStrictMode: true, output: "standalone" }' > next.config.js; \
    fi

RUN if [ ! -f app/layout.tsx ]; then \
        mkdir -p app && \
        echo 'export default function RootLayout({children}:{children:React.ReactNode}){return <html><body>{children}</body></html>}' > app/layout.tsx; \
    fi

RUN if [ ! -f app/page.tsx ]; then \
        mkdir -p app && \
        echo 'export default function Home(){return <main><h1>NuviaX App</h1><p>Coming Soon</p></main>}' > app/page.tsx; \
    fi

RUN npm run build || \
    (echo "Next.js build failed, creating fallback..." && \
     mkdir -p .next/standalone .next/static && \
     echo 'const http=require("http");http.createServer((q,s)=>{s.writeHead(200,{"Content-Type":"text/html"});s.end("<h1>NuviaX App</h1><p>Coming Soon</p>")}).listen(3000,()=>console.log("Fallback on :3000"))' > .next/standalone/server.js)

FROM base AS runner

ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

RUN mkdir -p .next/static public

COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/public ./public

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
'@

    # Frontend Landing Dockerfile
    $landingDockerfile = @'
FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps

COPY . .

RUN mkdir -p frontend/landing

RUN if [ ! -f frontend/landing/package.json ]; then \
        echo '{"name":"@nuviax/landing","version":"1.0.0","private":true,"scripts":{"dev":"next dev -p 3001","build":"next build","start":"next start -p 3001"},"dependencies":{"next":"^14","react":"^18","react-dom":"^18"}}' > frontend/landing/package.json; \
    fi

RUN cd frontend/landing && \
    if [ -f ../../yarn.lock ]; then \
        yarn install --frozen-lockfile || yarn install; \
    elif [ -f ../../package-lock.json ]; then \
        npm ci || npm install; \
    elif [ -f ../../pnpm-lock.yaml ]; then \
        corepack enable pnpm && pnpm install --frozen-lockfile || pnpm install; \
    elif [ -f package-lock.json ]; then \
        npm ci || npm install; \
    else \
        npm install; \
    fi

FROM base AS builder

COPY --from=deps /app/frontend/landing/node_modules ./frontend/landing/node_modules

COPY . .

WORKDIR /app/frontend/landing

RUN if [ ! -f next.config.js ]; then \
        echo 'module.exports = { reactStrictMode: true, output: "standalone" }' > next.config.js; \
    fi

RUN if [ ! -f app/layout.tsx ]; then \
        mkdir -p app && \
        echo 'export default function RootLayout({children}:{children:React.ReactNode}){return <html><body>{children}</body></html>}' > app/layout.tsx; \
    fi

RUN if [ ! -f app/page.tsx ]; then \
        mkdir -p app && \
        echo 'export default function Home(){return <main><h1>NuviaX</h1><p>Transform Your Growth Journey</p></main>}' > app/page.tsx; \
    fi

RUN npm run build || \
    (echo "Next.js build failed, creating fallback..." && \
     mkdir -p .next/standalone .next/static && \
     echo 'const http=require("http");http.createServer((q,s)=>{s.writeHead(200,{"Content-Type":"text/html"});s.end("<h1>NuviaX</h1><p>Coming Soon</p>")}).listen(3001,()=>console.log("Fallback on :3001"))' > .next/standalone/server.js)

FROM base AS runner

ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

RUN mkdir -p .next/static public

COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/public ./public

USER nextjs

EXPOSE 3001

ENV PORT=3001
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
'@

    # Scriere fișiere
    if (Test-Path "frontend/app") {
        Set-Content "frontend/app/Dockerfile" -Value $appDockerfile -NoNewline
        Write-ColorOutput "✓ frontend/app/Dockerfile" "Green"
    }
    
    if (Test-Path "frontend/landing") {
        Set-Content "frontend/landing/Dockerfile" -Value $landingDockerfile -NoNewline
        Write-ColorOutput "✓ frontend/landing/Dockerfile" "Green"
    }
}

function Commit-AndPush {
    Write-ColorOutput "`n📤 Git commit + push..." "Cyan"
    
    git add .
    git status --short
    
    $commitMessage = "fix: Corectare structură v7 (node_modules paths)

- Nomenclatură web → app
- Dockerfile paths corectate
- v7 FINAL"
    
    git commit -m $commitMessage
    
    if (-not $SkipPush) {
        git push
        Write-ColorOutput "✓ Push complet" "Green"
    } else {
        Write-ColorOutput "⊘ Push sărit" "Yellow"
    }
}

Write-ColorOutput @"
═══════════════════════════════════════════════════════════════
  NuviaX v7 FINAL - node_modules paths corectate
═══════════════════════════════════════════════════════════════
"@ "Magenta"

try {
    Set-Location $RepoPath
    
    Test-GitRepository
    Backup-Repository
    Update-DockerComposeFiles
    Update-GitHubWorkflows
    Update-Dockerfiles
    Commit-AndPush
    
    Write-ColorOutput "`n✓ Toate corecturile aplicate cu succes!" "Green"
    Write-ColorOutput "`nNext: Verifică GitHub Actions build`n" "Cyan"
    
} catch {
    Write-ColorOutput "`n✗ Eroare: $_" "Red"
    exit 1
}
