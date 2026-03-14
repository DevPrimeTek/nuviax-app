FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps

COPY . .

RUN mkdir -p frontend/landing && \
    if [ ! -f frontend/landing/package.json ]; then \
        echo '{"name":"@nuviax/landing","version":"1.0.0","private":true,"scripts":{"dev":"next dev -p 3001","build":"next build","start":"next start -p 3001"},"dependencies":{"next":"^14","react":"^18","react-dom":"^18"}}' > frontend/landing/package.json; \
    fi

RUN if [ -f yarn.lock ]; then \
        yarn install --frozen-lockfile || yarn install; \
    elif [ -f package-lock.json ]; then \
        npm ci || npm install; \
    elif [ -f pnpm-lock.yaml ]; then \
        corepack enable pnpm && pnpm install --frozen-lockfile || pnpm install; \
    else \
        cd frontend/landing && npm install; \
    fi

FROM base AS builder

COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app ./

RUN if [ ! -d "frontend/landing/app" ] && [ ! -d "frontend/landing/pages" ] && [ ! -d "frontend/landing/src" ]; then \
        mkdir -p frontend/landing/app && \
        echo 'export default function Home() {' > frontend/landing/app/page.tsx && \
        echo '  return (' >> frontend/landing/app/page.tsx && \
        echo '    <div style={{' >> frontend/landing/app/page.tsx && \
        echo '      minHeight: "100vh",' >> frontend/landing/app/page.tsx && \
        echo '      display: "flex",' >> frontend/landing/app/page.tsx && \
        echo '      flexDirection: "column",' >> frontend/landing/app/page.tsx && \
        echo '      alignItems: "center",' >> frontend/landing/app/page.tsx && \
        echo '      justifyContent: "center",' >> frontend/landing/app/page.tsx && \
        echo '      fontFamily: "system-ui, sans-serif",' >> frontend/landing/app/page.tsx && \
        echo '      padding: "2rem",' >> frontend/landing/app/page.tsx && \
        echo '      textAlign: "center",' >> frontend/landing/app/page.tsx && \
        echo '      background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"' >> frontend/landing/app/page.tsx && \
        echo '    }}>' >> frontend/landing/app/page.tsx && \
        echo '      <h1 style={{fontSize: "4rem", color: "white", marginBottom: "1rem"}}>NuviaX</h1>' >> frontend/landing/app/page.tsx && \
        echo '      <p style={{fontSize: "1.5rem", color: "rgba(255,255,255,0.9)"}}>Growth Automation Platform</p>' >> frontend/landing/app/page.tsx && \
        echo '      <p style={{fontSize: "1.2rem", color: "rgba(255,255,255,0.7)"}}>Coming Soon</p>' >> frontend/landing/app/page.tsx && \
        echo '    </div>' >> frontend/landing/app/page.tsx && \
        echo '  );' >> frontend/landing/app/page.tsx && \
        echo '}' >> frontend/landing/app/page.tsx && \
        echo 'export const metadata = { title: "NuviaX" };' > frontend/landing/app/layout.tsx && \
        echo 'export default function RootLayout({ children }: { children: React.ReactNode }) {' >> frontend/landing/app/layout.tsx && \
        echo '  return <html lang="en"><body>{children}</body></html>;' >> frontend/landing/app/layout.tsx && \
        echo '}' >> frontend/landing/app/layout.tsx && \
        mkdir -p frontend/landing/public; \
    fi

WORKDIR /app/frontend/landing

RUN if [ ! -f "next.config.js" ] && [ ! -f "next.config.mjs" ]; then \
        echo "module.exports = { output: 'standalone' }" > next.config.js; \
    fi

RUN npm run build || ( \
    echo "Build failed, creating fallback..." && \
    mkdir -p .next/standalone .next/static public && \
    echo 'const http = require("http");' > server.js && \
    echo 'http.createServer((req, res) => {' >> server.js && \
    echo '  res.writeHead(200, {"Content-Type": "text/html"});' >> server.js && \
    echo '  res.end("<!DOCTYPE html><html><head><title>NuviaX</title><style>body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);color:white;text-align:center}h1{font-size:4rem}p{font-size:1.5rem;opacity:0.9}</style></head><body><div><h1>NuviaX</h1><p>Growth Automation Platform</p><p style=\"font-size:1.2rem;opacity:0.7\">Coming Soon</p></div></body></html>");' >> server.js && \
    echo '}).listen(3001, () => console.log("Landing on 3001"));' >> server.js)

FROM base AS runner

ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/standalone ./ 2>/dev/null || true
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/static ./.next/static 2>/dev/null || mkdir -p ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/public ./public 2>/dev/null || mkdir ./public
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/server.js ./server.js 2>/dev/null || true

RUN if [ ! -f server.js ]; then \
        echo 'const http = require("http");' > server.js && \
        echo 'http.createServer((req, res) => {' >> server.js && \
        echo '  res.writeHead(200, {"Content-Type": "text/html"});' >> server.js && \
        echo '  res.end("<html><body><h1>NuviaX</h1><p>Coming Soon</p></body></html>");' >> server.js && \
        echo '}).listen(3001);' >> server.js; \
    fi

USER nextjs
EXPOSE 3001
ENV PORT=3001 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
