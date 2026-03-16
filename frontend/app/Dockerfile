# ═══════════════════════════════════════════════════════════
# NuviaX App - Dockerfile v9.0 FINAL
# FIX: COPY public/ doar dacă există
# ═══════════════════════════════════════════════════════════

FROM node:20-alpine AS deps
WORKDIR /app
RUN apk add --no-cache libc6-compat
COPY frontend/app/package*.json ./
RUN if [ -f package-lock.json ]; then \
      npm ci --legacy-peer-deps; \
    else \
      npm install --legacy-peer-deps --production=false; \
    fi

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
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy standalone output
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# CRITICAL FIX: Copy public doar dacă există în builder
# Folosim sh -c pentru a testa existența folderului
RUN --mount=type=bind,from=builder,source=/app,target=/tmp/build \
    if [ -d /tmp/build/public ] && [ "$(ls -A /tmp/build/public)" ]; then \
      cp -r /tmp/build/public ./public; \
    else \
      mkdir -p ./public; \
    fi && \
    chown -R nextjs:nodejs ./public

USER nextjs
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000', (r) => r.statusCode === 200 ? process.exit(0) : process.exit(1))"

CMD ["node", "server.js"]