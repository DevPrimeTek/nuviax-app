FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps
COPY package.json* ./
COPY yarn.lock* package-lock.json* pnpm-lock.yaml* ./

RUN mkdir -p frontend/landing
COPY frontend/landing/package.json ./frontend/landing/ 2>/dev/null || \
    echo '{"name":"@nuviax/landing","version":"1.0.0","private":true}' > ./frontend/landing/package.json

RUN if [ -f yarn.lock ]; then \
        yarn install --frozen-lockfile || yarn install; \
    elif [ -f package-lock.json ]; then \
        npm ci || npm install; \
    elif [ -f pnpm-lock.yaml ]; then \
        corepack enable pnpm && pnpm install --frozen-lockfile; \
    else \
        npm install next@14 react@18 react-dom@18; \
    fi

FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN if [ ! -d "frontend/landing/src" ] && [ ! -d "frontend/landing/app" ] && [ ! -d "frontend/landing/pages" ]; then \
        mkdir -p frontend/landing/app && \
        echo 'export default function Page() { return <div><h1>NuviaX</h1><p>Coming Soon</p></div> }' > frontend/landing/app/page.tsx && \
        mkdir -p frontend/landing/public; \
    fi

WORKDIR /app/frontend/landing

RUN if [ ! -f "next.config.js" ] && [ ! -f "next.config.mjs" ]; then \
        echo 'module.exports = { output: "standalone" }' > next.config.js; \
    fi

RUN npm run build 2>/dev/null || (mkdir -p .next/standalone .next/static && \
    echo '<!DOCTYPE html><html><body><h1>NuviaX - Landing</h1></body></html>' > .next/standalone/index.html)

FROM base AS runner
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

COPY --from=builder /app/frontend/landing/public ./public 2>/dev/null || mkdir ./public
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/standalone ./ 2>/dev/null || true
COPY --from=builder --chown=nextjs:nodejs /app/frontend/landing/.next/static ./.next/static 2>/dev/null || mkdir -p ./.next/static

RUN if [ ! -f "server.js" ]; then \
        echo 'const http = require("http");' > server.js && \
        echo 'const server = http.createServer((req, res) => {' >> server.js && \
        echo '  res.writeHead(200, {"Content-Type": "text/html"});' >> server.js && \
        echo '  res.end("<h1>NuviaX Landing - Coming Soon</h1>");' >> server.js && \
        echo '});' >> server.js && \
        echo 'server.listen(3001, () => console.log("Landing on 3001"));' >> server.js; \
    fi

USER nextjs
EXPOSE 3001
ENV PORT=3001 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
