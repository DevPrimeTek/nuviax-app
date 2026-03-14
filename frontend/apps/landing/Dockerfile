FROM node:20-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

FROM base AS deps
COPY package.json yarn.lock* ./
COPY apps/landing/package.json ./apps/landing/
COPY packages/styles/package.json ./packages/styles/ 2>/dev/null || true
RUN yarn install --frozen-lockfile

FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY . .
WORKDIR /app/apps/landing
RUN yarn build

FROM base AS runner
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

COPY --from=builder /app/apps/landing/public ./public 2>/dev/null || true
COPY --from=builder --chown=nextjs:nodejs /app/apps/landing/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/landing/.next/static ./.next/static

USER nextjs
EXPOSE 3001
ENV PORT=3001 HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
