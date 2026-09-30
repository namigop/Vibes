# syntax=docker/dockerfile:1.6

# ---- dev stage (runs air for hot reload) ----
FROM golang:1.26-alpine

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOFLAGS="-mod=mod"

# Install git (needed for go install) and build essentials (air watches + builds Go)
RUN apk add --no-cache git ca-certificates

# Install air for hot reload.
#
# NOTE: the project was renamed from github.com/cosmtrek/air to
# github.com/air-verse/air. The old path still resolves via a GitHub redirect,
# so `go install github.com/cosmtrek/air@...` fails with a confusing
# "module declares its path as: github.com/air-verse/air" error.
# v1.67.4 requires Go >= 1.26, which the base image above satisfies.
RUN go install github.com/air-verse/air@v1.67.4

WORKDIR /src

# Cache module downloads. We copy only go.mod here so the layer is reusable
# even when go.sum hasn't been generated yet (e.g. fresh checkout before
# `go mod tidy`). The later `COPY . .` brings in go.sum if it exists locally;
# otherwise the module cache populated by `go mod download` is used (GOFLAGS
# is set to -mod=mod above).
COPY go.mod ./
RUN go mod download

# Source is mounted as a volume in docker-compose for hot reload
COPY . .

EXPOSE 8080

# Default command runs air; the binary location in .air.toml is ./tmp/server
CMD ["air", "-c", ".air.toml"]
