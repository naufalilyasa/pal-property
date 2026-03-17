# AGENTS.md — backend/internal/service/

## OVERVIEW

Business logic layer. Orchestrates repositories plus optional storage adapters and returns DTO-shaped responses for handlers.

## CURRENT PATTERN

```go
type listingService struct {
    repo    domain.ListingRepository
    storage domain.ListingImageStorage
}

func NewListingService(repo domain.ListingRepository, storage ...domain.ListingImageStorage) ListingService {
    svc := &listingService{repo: repo}
    if len(storage) > 0 {
        svc.storage = storage[0]
    }
    return svc
}
```

## RULES

- Depend on domain interfaces/entities, request/response DTOs, and small shared helpers from `pkg/`.
- Use domain errors for control flow; do not branch on ORM errors.
- Keep HTTP concerns out of services.
- For listing images, validate ownership/order in service code and delegate persistence/storage work to the repository/storage interfaces.

## TEST PATTERN

- Unit tests use mocks.
- Handler integration tests may instantiate `NewListingService(listingRepo, fakeStorage)` to cover multipart and image mutation routes without Cloudinary.

## ANTI-PATTERNS

- **NEVER** import concrete repository implementations here.
- **NEVER** read request cookies, headers, or Fiber context APIs here.
