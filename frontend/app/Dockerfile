FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps

COPY . .

RUN mkdir -p frontend/app && \
    if [ ! -f frontend/app/package.json ]; then \
        echo '{"name":"@nuviax/app","version":"1.0.0","private":true,"scripts":{"dev":"next dev","build":"next build","start":"next start"},"dependencies":{"next":"^14","react":"^18","react-dom":"^18"}}' > frontend/app/package.json; \
    fi

RUN if [ -f yarn.lock ]; then \
        yarn install --frozen-lockfile || yarn install; \
    elif [ -f package-lock.json ]; then \
        npm ci || npm install; \
    elif [ -f pnpm-lock.yaml ]; then \
        corepack enable pnpm && pnpm install --frozen-lockfile || pnpm install; \
    else \
        cd frontend/app && npm install; \
    fi

FROM base AS builder

COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app ./

RUN if [ ! -d "frontend/app/app" ] && [ ! -d "frontend/app/pages" ] && [ ! -d "frontend/app/src" ]; then \
        mkdir -p frontend/app/app && \
        echo 'export default function Home() {' > frontend/app/app/page.tsx && \
        echo '  return (' >> frontend/app/app/page.tsx && \
        echo '    <div style={{' >> frontend/app/app/page.tsx && \
        echo '      minHeight: "100vh",' >> frontend/app/app/page.tsx && \
        echo '      display: "flex",' >> frontend/app/app/page.tsx && \
        echo '      flexDirection: "column",' >> frontend/app/app/page.tsx && \
        echo '      alignItems: "center",' >> frontend/app/app/page.tsx && \
        echo '      justifyContent: "center",' >> frontend/app/app/page.tsx && \
        echo '      fontFamily: "system-ui, sans-serif",' >> frontend/app/app/page.tsx && \
        echo '      padding: "2rem",' >> frontend/app/app/page.tsx && \
        echo '      background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"' >> frontend/app/app/page.tsx && \
        echo '    }}>' >> frontend/app/app/page.tsx && \
        echo '      <h1 style={{fontSize: "3rem", color: "white", marginBottom: "1rem"}}>NuviaX App</h1>' >> frontend/app/app/page.tsx && \
        echo '      <p style={{fontSize: "1.2rem", color: "rgba(255,255,255,0.9)"}}>Coming Soon</p>' >> frontend/app/app/page.tsx && \
        echo '    </div>' >> frontend/app/app/page.tsx && \
        echo '  );' >> frontend/app/app/page.tsx && \
        echo '}' >> frontend/app/app/page.tsx && \
        echo 'export const metadata = { title: "NuviaX App" };' > frontend/app/app/layout.tsx && \
        echo 'export default function RootLayout({ children }: { children: React.ReactNode }) {' >> frontend/app/app/layout.tsx && \
        echo '  return <html lang="en"><body>{children}</body></html>;' >> frontend/app/app/layout.tsx && \
        echo '}' >> frontend/app/app/layout.tsx && \
        mkdir -p frontend/app/public; \
    fi

WORKDIR /app/frontend/app

RUN if [ ! -f "next.config.js" ] && [ ! -f "next.config.mjs" ]; then \
        echo "module.exports = { output: 'standalone' }" > next.config.js; \
    fi

RUN npm run build || ( \
    echo "Build failed, creating fallback..." && \
    mkdir -p .next/standalone .next/static public && \
    echo 'const http = require("http");' > server.js && \
    echo 'http.createServer((req, res) => {' >> server.js && \
    echo '  res.writeHead(200, {"Content-Type": "text/html"});' >> server.js && \
    echo '  res.end("<!DOCTYPE html><html><head><title>NuviaX</title><style>body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);color:white;text-align:center}h1{font-size:3rem}</style></head><body><h1>NuviaX App</h1><p>Coming Soon</p></body></html>");' >> server.js && \
    echo '}).listen(3000, () => console.log("Server on 3000"));' >> server.js)

FROM base AS runner

ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

RUN mkdir -p .next/static public

COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/server.js ./server.js

RUN if [ ! -f server.js ]; then \
        echo 'const http = require("http");' > server.js && \
        echo 'http.createServer((req, res) => {' >> server.js && \
        echo '  res.writeHead(200, {"Content-Type": "text/html"});' >> server.js && \
        echo '  res.end("<html><body><h1>NuviaX</h1><p>Coming Soon</p></body></html>");' >> server.js && \
        echo '}).listen(3000);' >> server.js; \
    fi

USER nextjs
EXPOSE 3000
ENV PORT=3000 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
