# AGENTS.md — pal-property

**Generated:** 2026-03-03
**Branch:** main

## OVERVIEW

Property listing platform (Indonesia). Go REST API backend, Next.js frontend, event-driven workers (Redpanda/Kafka). Still in early development — only auth is implemented end-to-end.

**Stack:** Go 1.26 + Fiber v3 + GORM + PostgreSQL 17 | Next.js 16 + React 19 + Tailwind v4 | Redpanda (Kafka-compat) | Redis | Elasticsearch 8

## STRUCTURE

```
pal-property/
├── backend/           # Go REST API (only active service)
│   ├── cmd/           # Entrypoints (property-service, migrate)
│   ├── internal/      # Domain logic (NOT importable externally)
│   ├── pkg/           # Shared utilities (config, crypto, jwt, kafka, logger)
│   ├── db/migrations/ # golang-migrate SQL files
│   └── certs/         # RSA keys for JWT (RS256)
├── frontend/          # Next.js App Router (bare scaffold)
├── workers/           # scraper + syndicator (planned, empty)
├── infra/             # k8s manifests + kafka config (planned, empty)
├── deploy/            # deploy scripts (empty)
└── docker-compose.yml # Full local dev stack
```

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add new feature | `backend/internal/` → domain → service → handler → router |
| New config var | `backend/pkg/config/config.go` + `backend/.env-example` |
| New migration | `backend/db/migrations/` (golang-migrate naming: `NNNNNN_desc.up.sql`) |
| Auth logic | `backend/internal/service/auth_service.go` |
| JWT utilities | `backend/pkg/utils/jwt/jwt.go` |
| Encrypt sensitive data | `backend/pkg/crypto/aes.go` |
| CORS/middleware | `backend/internal/router/router.go` |
| Domain contracts | `backend/internal/domain/` |

## COMMANDS

```bash
# Dev
cd backend && air                              # hot reload (requires Air)
cd backend && go run ./cmd/property-service    # without hot reload

# Migrations
cd backend && go run ./cmd/migrate/main.go

# Test
cd backend && go test ./... -count=1
cd backend && go test ./... -count=1 -run TestAuthHandlerSuite -v  # integration

# Build
cd backend && go build ./...
cd backend && go vet ./...

# Infra (local dev)
docker compose up                             # all services
docker compose up postgres redis              # only DB + cache
```

## INFRA SERVICES (docker-compose)

| Service | Image | Port | Note |
|---------|-------|------|------|
| postgres | 17.8-alpine | 5433:5432 | host port 5433 (not 5432) |
| redis | 8.2-alpine | 6379 | |
| redpanda | v25.3.9 | 9092 | Kafka-compatible, no Zookeeper |
| elasticsearch | 8.19.11 | 9200 | xpack.security disabled (dev only) |
| backend | local build | 8080 | |

## KEY CONVENTIONS

- **Price/money**: always `int64`, unit = Indonesian Rupiah (IDR). No float, no decimal lib.
- **UUIDs**: always `uuid.UUID` (google/uuid), generated via `uuid.NewV7()` in `BeforeCreate`.
- **Sensitive DB columns**: encrypt before write, decrypt after read (AES-256-GCM via `pkg/crypto`).
- **Error response shape**: `{"success": false, "message": "...", "data": null, "trace_id": "uuid"}`.
- **Auth tokens**: httpOnly cookies (`access_token`, `refresh_token`), SameSite:Lax.

## ANTI-PATTERNS (THIS PROJECT)

- **NEVER** import `gorm.io/gorm` in service layer — use `domain.ErrNotFound` etc.
- **NEVER** return raw `gorm.ErrRecordNotFound` from repository — translate to `domain.ErrNotFound`.
- **NEVER** store sensitive tokens (OAuth access/refresh) as plaintext — use `pkg/crypto.Encrypt`.
- **NEVER** use `float64` for money/price fields.
- **NEVER** add config fields without updating `.env-example`.
- **NEVER** bypass the layer boundary (handler → service → repository → domain).

## NOTES

- `OAUTH_TOKEN_ENCRYPTION_KEY`: base64-encoded 32-byte AES key. Generate: `openssl rand -base64 32`
- JWT keys are RS256 PEM stored as base64 in env vars (`JWT_PRIVATE_KEY_BASE64`, `JWT_PUBLIC_KEY_BASE64`)
- Postgres exposed on host port **5433** (not 5432) to avoid conflicts
- workers/ and infra/ are empty — Redpanda and Elasticsearch are provisioned but not yet consumed
- No graceful shutdown implemented in main.go yet
