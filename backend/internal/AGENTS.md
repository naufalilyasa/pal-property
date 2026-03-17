# AGENTS.md — backend/internal/

## OVERVIEW

Private application code for the backend module. Layer boundaries are strict, but listing images add one more domain port (`ListingImageStorage`) alongside the repository contracts.

## PACKAGES

| Package | Role | Notes |
|---------|------|-------|
| `domain/` | Contracts, entities, errors | includes `listing_repository.go` and `listing_image_storage.go` (uses provider-agnostic `pkg/mediaasset` types at the port boundary) |
| `handler/http/` | Fiber transport | JSON + multipart parsing, auth locals, response writing |
| `service/` | Business logic | listings, categories, auth, image orchestration |
| `repository/postgres/` | GORM persistence | listings, categories, auth, listing_images |
| `repository/redis/` | Refresh-token cache | auth/session support |
| `router/` | Route registration | public + protected routes |
| `dto/` | Request/response payloads | includes listing-image reorder DTOs |

## IMPORT GRAPH

```text
handler -> service (interfaces)
service -> domain interfaces/entities + selected pkg helpers
repository -> domain interfaces/entities + gorm/redis
domain -> entities/contracts/errors; may use provider-agnostic shared pkg types only through domain ports (for example `pkg/mediaasset` in `listing_image_storage.go`), and must never depend on concrete provider implementations
```

## CURRENT CONVENTIONS

- Handlers extract auth state from `c.Locals(...)` and pass `c.Context()` downstream.
- Listing-image upload stays in the handler layer only for multipart extraction via `c.FormFile("file")`.
- Services enforce ownership and image ordering rules.
- Repositories own transactional image mutations and read projections.

## ANTI-PATTERNS

- **NEVER** call repositories from handlers.
- **NEVER** import `gorm.io/gorm` in services for error handling.
- **NEVER** let domain-layer contracts depend on concrete Cloudinary code.
