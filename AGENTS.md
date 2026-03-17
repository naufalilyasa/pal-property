# AGENTS.md — pal-property

**Generated:** 2026-03-17
**Branch:** main

## OVERVIEW

Property listing platform for Indonesia. Active implementation is concentrated in the Go backend; the Next.js frontend is still scaffold-level, while workers/infra/deploy remain mostly placeholders.

**Stack:** Go 1.26 + Fiber v3 (Sonic JSON) + GORM + PostgreSQL 17 + Redis + Goth OAuth | Next.js 16 + React 19 + Tailwind v4 | Cloudinary for listing images | Redpanda and Elasticsearch provisioned for later work

**Implemented:** Auth (Google OAuth, JWT RS256, refresh rotation), Listing CRUD (create/read/update/delete/list/filter, soft delete, view count), Category management, Listing image management (upload/delete/set-primary/reorder with Cloudinary-backed storage)
**Planned:** RBAC/Casbin, Redpanda producers/consumers, Elasticsearch indexing, non-scaffold frontend screens

## STRUCTURE

```text
pal-property/
├── backend/                  # active Go service
│   ├── cmd/                  # property-service + migrate entrypoints
│   ├── internal/             # handlers, services, repositories, domain, DTOs, router
│   ├── pkg/                  # config, crypto, cloudinary, mediaasset, middleware, utils
│   ├── db/migrations/        # golang-migrate SQL files
│   └── postman_collection.json
├── frontend/                 # Next.js App Router scaffold
├── workers/                  # planned, currently empty
├── infra/                    # planned, currently empty
├── deploy/                   # planned, currently empty
└── docker-compose.yml        # local infra + backend container
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add backend feature | `backend/internal/` | flow = domain -> repository -> service -> handler -> router |
| Startup wiring | `backend/cmd/property-service/main.go` | DI + Fiber config + global error mapping |
| Listing image transport | `backend/internal/handler/http/listing.go` | multipart upload + JSON mutation endpoints |
| Listings + images business logic | `backend/internal/service/listing_service.go` | ownership, upload, delete, reorder, primary selection |
| Listing persistence | `backend/internal/repository/postgres/listing.go` | listing + listing_images queries and transactions |
| Config/env vars | `backend/pkg/config/config.go` | `config.LoadConfig()` populates global `config.Env` |
| Cloudinary adapter | `backend/pkg/cloudinary/adapter.go` | concrete storage implementation for listing images |

## COMMANDS

```bash
# Backend dev
cd backend && air
cd backend && go run ./cmd/property-service

# Migrations
cd backend && go run ./cmd/migrate/main.go
cd backend && go run ./cmd/migrate/main.go down

# Backend verification
cd backend && go test ./... -count=1
cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v
cd backend && go build ./...
cd backend && go vet ./...

# Frontend
cd frontend && npm run dev
cd frontend && npm run build
cd frontend && npm run lint
```

## KEY CONVENTIONS

- **Money**: always `int64`, unit = IDR.
- **UUIDs**: use `uuid.UUID`; entities generate IDs with `uuid.NewV7()` in hooks.
- **JSON**: Fiber app uses `sonic.Marshal` / `sonic.Unmarshal`.
- **Context propagation**: handlers pass `c.Context()` into services; do not use `context.Background()` in request flow.
- **Errors**: repositories translate `gorm.ErrRecordNotFound` to domain errors; the Fiber global handler maps domain errors to the standard JSON envelope.
- **Auth**: access + refresh tokens are httpOnly cookies with `SameSite=Lax`; refresh JTIs live in Redis.
- **Listing images**: tests use fake storage implementations; production uses the Cloudinary adapter behind the `domain.ListingImageStorage` interface.
- **Cloudinary config**: `CLOUDINARY_ENABLED=false` keeps image upload wiring optional; once enabled, all three credentials must be present together.

## TESTING RULES

- `backend/internal/service/*_test.go`: `package service_test`, `testify/mock`, flat `TestXxx` functions.
- `backend/internal/handler/http/*_test.go`: `testify/suite` + testcontainers Postgres, with `logger.Log = zap.NewNop()` in setup.
- Listing image handler coverage should inject fake storage into `service.NewListingService(listingRepo, fakeStorage)`; never hit live Cloudinary in tests.
- `config.Env.AppEnv = "testing"` keeps the rate limiter out of integration-test request paths.

## NOTES

- Host Postgres port is `5433`, not `5432`.
- `backend/postman_collection.json` covers Authentication, Categories, Listings, and listing-image routes.
- Frontend files still contain create-next-app placeholder content.
