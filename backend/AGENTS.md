# AGENTS.md — backend/

## OVERVIEW

Go 1.26 REST API. Fiber v3 + GORM + PostgreSQL. Strict layered architecture — no shortcuts between layers.

## STRUCTURE

```
backend/
├── cmd/
│   ├── property-service/main.go   # DI wiring + server start (NO business logic here)
│   └── migrate/main.go            # golang-migrate runner
├── internal/                      # See internal/AGENTS.md
├── pkg/                           # See pkg/AGENTS.md
├── db/migrations/                 # SQL files: NNNNNN_desc.{up,down}.sql
├── certs/                         # RSA PEM keys (base64-encoded in env)
├── .air.toml                      # Hot reload config
├── .env / .env-example            # App config
└── .env.docker                    # Config for docker compose run
```

## LAYER FLOW

```
cmd/main.go  →  router.go  →  handler/  →  service/  →  repository/  →  domain/
(DI only)       (routes)      (HTTP)       (biz)        (DB/Cache)      (interfaces + entities)
```

## KEY PATTERNS

**DI wiring** (in `cmd/property-service/main.go`):
```go
repo := postgres.NewAuthRepository(db)
cache := redis.NewCacheRepository(rdb)
svc := service.NewAuthService(repo, cache)
handler := http.NewAuthHandler(svc)
router.Register(app, handler, cfg)
```

**Error handler** (Fiber global):
- Catches domain errors → maps to HTTP status
- Response always: `{success, message, data, trace_id}`

**Migrations**: `db/migrations/` — golang-migrate naming convention:
- `000001_init_schema.up.sql` / `000001_init_schema.down.sql`
- Run via `go run ./cmd/migrate/main.go`

## ADDING A NEW DOMAIN (e.g., `listing`)

1. `internal/domain/entity/listing.go` — entity struct
2. `internal/domain/repository.go` — add interface methods
3. `internal/repository/postgres/listing.go` — implement
4. `internal/service/listing_service.go` — implement
5. `internal/handler/http/listing.go` — Fiber handlers
6. `internal/router/router.go` — register routes
7. `internal/dto/request/listing.go` + `dto/response/listing.go` — DTOs

## ANTI-PATTERNS

- **NEVER** put business logic in `cmd/` — only wiring
- **NEVER** skip a layer (e.g., handler calling repo directly)
- **NEVER** add a new feature without its domain interface in `internal/domain/`
