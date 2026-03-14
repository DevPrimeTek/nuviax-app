FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install deps — monorepo aware
FROM base AS deps
COPY package.json yarn.lock* ./
COPY frontend/app/package.json ./frontend/app/
COPY packages/styles/package.json ./packages/styles/ 2>/dev/null || true
RUN yarn install --frozen-lockfile

# Build
FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY . .
WORKDIR /app/frontend/app
RUN yarn build

# Production runner
FROM base AS runner
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

COPY --from=builder /app/frontend/app/public ./public 2>/dev/null || true
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/frontend/app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
