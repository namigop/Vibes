# syntax=docker/dockerfile:1.6

FROM node:20-alpine

WORKDIR /app

# Install deps separately for better layer caching.
COPY package.json package-lock.json* ./
RUN npm install --no-audit --no-fund

# Source is mounted as a volume in docker-compose for hot reload.
COPY . .

ENV NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
EXPOSE 3000

CMD ["npm", "run", "dev"]
