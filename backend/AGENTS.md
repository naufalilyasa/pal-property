# AGENTS.md — backend/

## OVERVIEW

Go 1.26 REST API. Fiber v3 + GORM + PostgreSQL + Redis with strict layering. Listings include Cloudinary-backed image management, and authorization now routes through a shared Casbin-backed authz layer.

## STRUCTURE

```text
backend/
├── cmd/
│   ├── property-service/main.go   # DI wiring + server start
│   └── migrate/main.go            # golang-migrate runner
├── internal/                      # See internal/AGENTS.md
├── pkg/                           # See pkg/AGENTS.md
├── db/migrations/                 # SQL files: NNNNNN_desc.{up,down}.sql
├── .env / .env-example            # local env guidance
└── .env.docker                    # docker-compose env file
```

## LAYER FLOW

```text
cmd/main.go -> router.go -> handler/ -> service/ -> repository/ -> domain/
```

## KEY PATTERNS

**DI wiring** (`cmd/property-service/main.go`):
```go
listingRepo := postgres.NewListingRepository(db)
authzSvc := authz.NewService(db)
listingAuthzSvc := service.NewAuthzService(authzSvc)
var imageStorage domain.ListingImageStorage
if config.Env.CloudinaryEnabled {
    imageStorage = cloudinary.NewAdapter(...)
}
listingSvc := service.NewListingServiceWithAuthz(listingRepo, listingAuthzSvc, imageStorage)
listingHandler := http.NewListingHandler(listingSvc)
```

**Authorization wiring**:
- `pkg/authz/` owns the embedded Casbin model and enforcer bootstrap.
- `middleware.Protected` still validates JWTs and reloads the current DB role on every request.
- `middleware.RequirePermission` handles coarse route checks; listing services keep owner/admin decisions close to loaded listing resources.

**Error handler**:
- Maps domain errors to HTTP status codes.
- Includes listing-image validation/config errors (`ErrInvalidImageFile`, `ErrImageOrderInvalid`, `ErrImageLimitReached`, `ErrImageStorageUnset`).

**Config sync**:
- Runtime authority is `pkg/config/config.go`.
- Any new env var still requires updates to `.env-example` and `.env.docker` when local startup depends on it.

## CURRENT FEATURE AREAS

- Auth + refresh rotation
- Casbin-backed RBAC + owner-aware listing authorization
- Category CRUD
- Listing CRUD + read filters
- Listing image upload/delete/set-primary/reorder

## ANTI-PATTERNS

- **NEVER** put business logic in `cmd/`.
- **NEVER** skip a layer.
- **NEVER** wire live Cloudinary into tests; use fake storage.
