# AGENTS.md — backend/internal/domain/

## OVERVIEW

Core contracts for the backend: entities, repository/storage interfaces, filters, and sentinel errors.

## FILES

| File | Purpose |
|------|---------|
| `errors.go` | Shared sentinel errors, including listing-image errors |
| `listing_repository.go` | listing + listing image persistence contract (`ListingRepository`) |
| `listing_image_storage.go` | provider-agnostic listing image storage port (`ListingImageStorage`) using `pkg/mediaasset` types |
| `category_repository.go` | category repository contract |
| `repository.go` | auth repository contract (`AuthRepository`) |
| `cache_repository.go` | refresh-token cache contract |
| `entity/` | GORM-tagged entity structs |
| `mocks/` | generated test doubles |

## SENTINEL ERRORS

- `ErrNotFound`
- `ErrConflict`
- `ErrInvalidCredential`
- `ErrUnauthorized`
- `ErrForbidden`
- `ErrInvalidImageFile`
- `ErrImageLimitReached`
- `ErrImageOrderInvalid`
- `ErrImageStorageUnset`

Map these in the global Fiber error handler, not in handlers.

## ENTITY + CONTRACT RULES

- Price fields stay `int64`.
- Optional DB fields use pointer types.
- `Listing.Images` projections should reflect active, ordered rows only.
- Domain storage ports stay provider-agnostic; Cloudinary details belong in `pkg/`, not here.

## ANTI-PATTERNS

- **NEVER** import non-domain `internal/` packages here; domain contracts may reference `internal/domain/entity` when the interface shape requires entities.
- **NEVER** add HTTP or provider-specific logic here.
