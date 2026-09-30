# Build stage: compile a static binary with CGO disabled.
FROM golang:1.27-alpine AS build

WORKDIR /src

# Dependencies first so the module cache survives source edits.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/server

# Runtime stage: just the binary and the CA bundle.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 app

COPY --from=build /out/server /usr/local/bin/server

USER app
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/api/v1/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
