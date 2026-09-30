# Build stage.
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --no-audit --no-fund

COPY . .

# NEXT_PUBLIC_* values are inlined at build time, so the API base URL has to be
# present here rather than only at runtime.
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_TELEMETRY_DISABLED=1

RUN npm run build

# Runtime stage: only the standalone server output plus static assets.
FROM node:22-alpine AS run

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# `output: "standalone"` is required in next.config.ts for this to exist.
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public

EXPOSE 3000

CMD ["node", "server.js"]
