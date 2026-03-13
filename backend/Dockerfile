# ════════════════════════════════════════════════════════════════
# Stage 1 — Builder
# ════════════════════════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

ARG BUILD_VERSION=dev

# Build tools
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Download deps first (Docker cache layer)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source
COPY . .

# Compile: stripped binary, no debug symbols, reproducible
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.Version=${BUILD_VERSION}" \
    -trimpath \
    -o /app/server \
    ./cmd/server

# Verify binary
RUN test -f /app/server && chmod +x /app/server && \
    echo "Binary size: $(du -sh /app/server | cut -f1)"

# ════════════════════════════════════════════════════════════════
# Stage 2 — Runtime (distroless: no shell, no tools, ~8MB)
# ════════════════════════════════════════════════════════════════
FROM gcr.io/distroless/static-debian12:nonroot

# Timezone + TLS certs from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary only — no source code, no Go runtime
COPY --from=builder /app/server /server

# Non-root user (uid 65532)
USER nonroot:nonroot

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=3 \
  CMD ["/server", "-health"] || exit 1

ENTRYPOINT ["/server"]
