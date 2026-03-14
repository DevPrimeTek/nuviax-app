# ════════════════════════════════════════════════════════════════
# Stage 1 — Builder
# ════════════════════════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

ARG BUILD_VERSION=dev

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.Version=${BUILD_VERSION}" \
    -trimpath \
    -o /app/server \
    ./cmd/server

RUN test -f /app/server && chmod +x /app/server

# ════════════════════════════════════════════════════════════════
# Stage 2 — Runtime (distroless)
# ════════════════════════════════════════════════════════════════
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]
