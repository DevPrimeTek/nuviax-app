# ═══════════════════════════════════════════════════════════
# NuviaX App - Production Dockerfile v8.2
# Deploy: deploy/Deploy-NuviaX-v8.2.ps1
# ═══════════════════════════════════════════════════════════

FROM node:20-alpine AS deps
WORKDIR /app
RUN apk add --no-cache libc6-compat
COPY frontend/app/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps --production=false; fi

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/app/ ./
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build
RUN test -d .next/standalone || (echo "ERROR: .next/standalone missing!" && exit 1)
RUN test -d .next/static || (echo "ERROR: .next/static missing!" && exit 1)

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
USER nextjs
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 CMD node -e "require('http').get('http://localhost:3000', (r) => r.statusCode === 200 ? process.exit(0) : process.exit(1))"
CMD ["node", "server.js"]