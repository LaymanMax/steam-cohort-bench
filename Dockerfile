# ----------------------------------------------------
# Stage 1: Build React Frontend
# ----------------------------------------------------
FROM node:22-bookworm-slim AS web-builder

WORKDIR /app/web
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

# ----------------------------------------------------
# Stage 2: Build Go Backend with CGO (DuckDB)
# ----------------------------------------------------
FROM golang:1.24-bookworm AS go-builder

WORKDIR /app

# Install build essentials for CGO
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build server and ingest binaries with CGO enabled
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o /app/bin/server ./cmd/server && \
    go build -ldflags="-s -w" -o /app/bin/ingest ./cmd/ingest

# ----------------------------------------------------
# Stage 3: Final Production Runtime Image
# ----------------------------------------------------
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled binaries and assets
COPY --from=go-builder /app/bin/server /app/bin/server
COPY --from=go-builder /app/bin/ingest /app/bin/ingest
COPY --from=web-builder /app/web/dist /app/web/dist

# Default directories for dataset mounts
RUN mkdir -p /app/data/raw /app/data/processed /app/data/manifests

ENV PORT=8080 \
    HOST=0.0.0.0 \
    DATA_DIR=/app/data \
    STATIC_DIR=/app/web/dist

EXPOSE 8080

CMD ["/app/bin/server"]
