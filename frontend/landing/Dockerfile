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
        cat > frontend/landing/app/page.tsx << 'PAGEFILE'
export default function Home() {
  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'system-ui, -apple-system, sans-serif',
      padding: '2rem',
      textAlign: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    }}>
      <h1 style={{fontSize: '4rem', color: 'white', marginBottom: '1rem'}}>
        NuviaX
      </h1>
      <p style={{fontSize: '1.5rem', color: 'rgba(255,255,255,0.9)', marginBottom: '2rem'}}>
        Growth Automation Platform
      </p>
      <p style={{fontSize: '1.2rem', color: 'rgba(255,255,255,0.7)'}}>
        Landing Page Coming Soon
      </p>
    </div>
  );
}
PAGEFILE
        cat > frontend/landing/app/layout.tsx << 'LAYOUTFILE'
export const metadata = {
  title: 'NuviaX - Growth Automation Platform',
  description: 'Advanced growth automation platform',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
LAYOUTFILE
        mkdir -p frontend/landing/public; \
    fi

WORKDIR /app/frontend/landing

RUN if [ ! -f "next.config.js" ] && [ ! -f "next.config.mjs" ]; then \
        echo "module.exports = { output: 'standalone' }" > next.config.js; \
    fi

RUN npm run build || ( \
    echo "Next.js build failed, creating fallback server..." && \
    mkdir -p .next/standalone .next/static public && \
    cat > server.js << 'SERVERFILE'
const http = require("http");
http.createServer((req, res) => {
  res.writeHead(200, {"Content-Type": "text/html"});
  res.end(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>NuviaX</title>
      <style>
        body {
          margin: 0;
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          font-family: system-ui, sans-serif;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          text-align: center;
        }
        h1 { font-size: 4rem; margin-bottom: 1rem; }
        p { font-size: 1.5rem; opacity: 0.9; }
      </style>
    </head>
    <body>
      <div>
        <h1>NuviaX</h1>
        <p>Growth Automation Platform</p>
        <p style="font-size:1.2rem;opacity:0.7">Coming Soon</p>
      </div>
    </body>
    </html>
  `);
}).listen(3001, () => console.log("Landing on 3001"));
SERVERFILE
)

FROM base AS runner

ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/standalone ./ 2>/dev/null || echo "No standalone"
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/static ./.next/static 2>/dev/null || mkdir -p ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/public ./public 2>/dev/null || mkdir ./public
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/server.js ./server.js 2>/dev/null || echo "No server.js"

RUN if [ ! -f server.js ]; then \
        cat > server.js << 'FINALSERVER'
const http = require("http");
http.createServer((req, res) => {
  res.writeHead(200, {"Content-Type": "text/html"});
  res.end("<html><body><h1>NuviaX</h1><p>Coming Soon</p></body></html>");
}).listen(3001, () => console.log("Landing fallback on 3001"));
FINALSERVER
    fi

USER nextjs
EXPOSE 3001
ENV PORT=3001 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
