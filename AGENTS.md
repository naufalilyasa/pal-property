# AGENTS.md — pal-property

**Generated:** 2026-03-04
**Branch:** main

## OVERVIEW

Property listing platform (Indonesia). Go REST API backend, Next.js frontend, event-driven workers (Redpanda/Kafka). Auth + Listing CRUD are fully implemented end-to-end.

**Stack:** Go 1.26 + Fiber v3 (Sonic JSON) + GORM + PostgreSQL 17 | Next.js 16 + React 19 + Tailwind v4 | Redpanda (Kafka-compat) | Redis | Elasticsearch 8

**Config:** `caarlos0/env` v11 (struct tags, no Viper). All config lives in `backend/pkg/config/config.go` as a single `Config` struct.

**Implemented:** Auth (OAuth2/Google, JWT RS256, refresh rotation), Listing CRUD (create/read/update/delete/list/filter, soft delete, view count)
**Planned (not yet impl):** RBAC/Casbin, Redpanda producers/consumers, Elasticsearch indexing, Image upload (S3/R2)

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
cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v  # listing integration

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
- **JSON encoder**: Fiber is initialized with `sonic.Marshal` / `sonic.Unmarshal` — do not swap to `encoding/json`.
- **Context propagation**: always pass `c.UserContext()` from handler to service — never `context.Background()`.
- **Goroutine safety**: always copy variables before passing to goroutines (zero-allocation, avoid closure capture bugs).
- **Zap logger**: dev → stdout + file (`tmp/logs/app.log`); prod → stdout only. Tests → `logger.Log = zap.NewNop()`.
- **Rate limiter**: skipped automatically when `config.Env.AppEnv == "testing" || "development"` (see `router.go`).

## ANTI-PATTERNS (THIS PROJECT)

- **NEVER** import `gorm.io/gorm` in service layer — use `domain.ErrNotFound` etc.
- **NEVER** return raw `gorm.ErrRecordNotFound` from repository — translate to `domain.ErrNotFound`.
- **NEVER** store sensitive tokens (OAuth access/refresh) as plaintext — use `pkg/crypto.Encrypt`.
- **NEVER** use `float64` for money/price fields.
- **NEVER** add config fields without updating `.env-example`.
- **NEVER** bypass the layer boundary (handler → service → repository → domain).
- **NEVER** use `context.Background()` in handlers — always `c.UserContext()` so request cancellation propagates.
- **NEVER** pass goroutine captures without copying first — loop vars and Fiber ctx are unsafe across goroutine boundaries.
- **NEVER** spin up Redpanda/Kafka testcontainers in integration tests — mock the producer/consumer interface instead.
- **NEVER** use `viper` — project uses `caarlos0/env` for all config (struct tags on `pkg/config/config.go`).

## TESTING RULES

- **Pyramid**: unit tests (mocks via `testify/mock`) + integration tests (real DB via `testcontainers-go`).
- **Target coverage**: 70–80% meaningful coverage. No coverage-padding tests.
- **Unit tests**: flat `TestXxx` functions, package `*_test`, use `mocks.NewXxx(t)` constructors.
- **Integration tests**: `testify/suite`, spin up `postgres:17-alpine` (and `redis:8.2-alpine` if needed) via testcontainers.
- **Test setup required** in `SetupSuite`:
  ```go
  logger.Log = zap.NewNop()      // silence logs
  config.Env.AppEnv = "testing"  // bypass rate limiter
  ```
- **Kafka/Redpanda**: always mock in tests — do NOT use testcontainer for message broker.
- **Truncate between tests**: `SetupTest` must TRUNCATE all relevant tables + flush Redis to ensure test isolation.
- **Run integration tests**: `go test ./... -count=1 -run TestXxxSuite -v -timeout 120s`

## NEW FEATURE WORKFLOW

For every new backend feature, follow this order:
1. **Entity** — add/update struct in `domain/entity/`
2. **Domain interface** — add methods to `domain/xxx_repository.go`
3. **Repository** — implement in `repository/postgres/xxx.go`
4. **DTO** — add request/response in `dto/request/` and `dto/response/`
5. **Service** — implement business logic in `service/xxx_service.go`
6. **Handler** — add HTTP handler in `handler/http/xxx.go`
7. **Router** — register routes in `router/router.go`
8. **Config** — if new env var needed, update `pkg/config/config.go` + `.env-example`
9. **Unit tests** — `service/xxx_service_test.go` (mocks)
10. **Integration tests** — `handler/http/xxx_test.go` (testcontainers)

## NOTES

- `OAUTH_TOKEN_ENCRYPTION_KEY`: base64-encoded 32-byte AES key. Generate: `openssl rand -base64 32`
- JWT keys are RS256 PEM stored as base64 in env vars (`JWT_PRIVATE_KEY_BASE64`, `JWT_PUBLIC_KEY_BASE64`)
- Postgres exposed on host port **5433** (not 5432) to avoid conflicts
- workers/ and infra/ are empty — Redpanda and Elasticsearch are provisioned but not yet consumed
- No graceful shutdown implemented in main.go yet
- RBAC (Casbin) is planned but not yet implemented — current auth is simple role string on User entity (`role varchar default 'user'`)
