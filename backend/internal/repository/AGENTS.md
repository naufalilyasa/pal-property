# AGENTS.md — backend/internal/repository/

## OVERVIEW

Persistence layer. `postgres/` owns relational queries and transactions, while `redis/` handles refresh-token cache state.

## POSTGRES RULES

- Use `WithContext(ctx)` on every query.
- Translate `gorm.ErrRecordNotFound` into domain errors.
- Wrap unexpected DB errors with operation context.
- Keep listing-image mutations transactional (`DeleteImage`, `SetPrimaryImage`, `ReorderImages`).
- Read projections for listings should preload active images in deterministic `sort_order ASC` order.

## BOUNDARIES

- Cloudinary upload/delete is not a repository concern; that stays behind `domain.ListingImageStorage`.
- Repositories may persist provider metadata fields already present on `entity.ListingImage`, but they should not call Cloudinary APIs directly.

## ANTI-PATTERNS

- **NEVER** return raw GORM errors to service code.
- **NEVER** move business authorization checks here.
