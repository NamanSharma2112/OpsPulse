# Build the dashboard with Next's standalone output so the runtime image
# carries only the server and the files it actually needs.
FROM node:22-alpine AS deps

WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/web/package.json apps/web/
COPY packages/types/package.json packages/types/
RUN npm ci

FROM node:22-alpine AS build

WORKDIR /app
# npm hoists workspace dependencies to the root node_modules and links the
# workspaces from there, so this single copy covers both packages.
COPY --from=deps /app/node_modules ./node_modules
COPY package.json package-lock.json ./
COPY packages/ ./packages/
COPY apps/web/ ./apps/web/

ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_OUTPUT=standalone
RUN npm run build --workspace @opspulse/web

FROM node:22-alpine AS runtime

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN apk add --no-cache curl && addgroup -g 10001 nodejs && adduser -D -u 10001 -G nodejs nextjs

# Standalone output already includes the trimmed node_modules it needs.
COPY --from=build --chown=nextjs:nodejs /app/apps/web/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/apps/web/.next/static ./apps/web/.next/static

USER nextjs
EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
    CMD curl -fsS http://localhost:3000/ >/dev/null || exit 1

CMD ["node", "apps/web/server.js"]
