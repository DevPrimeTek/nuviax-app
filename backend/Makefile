# NUViaX Backend — Makefile

.PHONY: build run test lint deps clean docker-build

## ── Local development ──────────────────────────────────────

# Download and tidy dependencies
deps:
	go mod tidy
	go mod verify

# Build binary locally
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# Run locally (needs .env in repo root or ENV vars set)
run:
	go run ./cmd/server

# Run tests
test:
	go test ./... -race -count=1

# Run tests with coverage
test-cover:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## ── Docker ──────────────────────────────────────────────────

# Build Docker image locally
docker-build:
	docker build \
		--build-arg BUILD_VERSION=local-$(shell git rev-parse --short HEAD) \
		-t devprimetek/nuviax-api:local \
		.

# Run API + dependencies locally via Docker Compose
docker-dev:
	docker compose -f ../../infra/docker-compose.yml up -d nuviax_db nuviax_redis
	go run ./cmd/server

## ── Database ────────────────────────────────────────────────

# Connect to local DB
db-shell:
	docker exec -it nuviax_db psql -U nuviax -d nuviax

# Run init schema manually
db-init:
	docker exec -i nuviax_db psql -U nuviax -d nuviax < ../../infra/init-db.sql

## ── Keys generation ─────────────────────────────────────────

# Generate RSA-4096 JWT keys
gen-keys:
	@mkdir -p .keys
	@openssl genrsa -out .keys/jwt_private.pem 4096
	@openssl rsa -in .keys/jwt_private.pem -pubout -out .keys/jwt_public.pem
	@echo "Private key (base64):"
	@cat .keys/jwt_private.pem | base64 -w 0
	@echo ""
	@echo "Public key (base64):"
	@cat .keys/jwt_public.pem | base64 -w 0
	@echo ""

# Generate AES-256 encryption key
gen-enc-key:
	@openssl rand -hex 32

clean:
	rm -rf bin/ coverage.out coverage.html .keys/
