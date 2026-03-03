# AGENTS.md — backend/internal/

## OVERVIEW

All domain-specific Go code. Strictly private to the backend module. Seven packages with hard layer boundaries.

## PACKAGES

| Package | Role | Key Files |
|---------|------|-----------|
| `domain/` | Contracts + entities + errors | `repository.go`, `cache_repository.go`, `errors.go`, `entity/` |
| `handler/http/` | Fiber HTTP handlers | One file per domain (e.g., `auth.go`) |
| `service/` | Business logic | One file per domain + `_test.go` |
| `repository/postgres/` | GORM implementations | One file per domain |
| `repository/redis/` | Redis cache impl | `cache.go` |
| `router/` | Route + middleware registration | `router.go` |
| `dto/` | Request/response structs | `request/` + `response/` |
| `middleware/` | Custom Fiber middleware | (currently empty) |

## STRICT RULES

**Import graph** (violations are bugs, not style):
```
handler → service (via interface)
service → domain interfaces only (NO gorm, NO redis)
repository → domain interfaces + gorm/redis
domain → stdlib only
```

**Error translation** (mandatory in every repository method):
```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, domain.ErrNotFound
}
```

**Service error checking**:
```go
if errors.Is(err, domain.ErrNotFound) { ... }  // CORRECT
if errors.Is(err, gorm.ErrRecordNotFound) { ... }  // FORBIDDEN in service
```

## MOCK GENERATION

Mocks live in `domain/mocks/`. Generated via mockery v2:
```bash
mockery --name=AuthRepository --dir=internal/domain --output=internal/domain/mocks
```

## ADDING NEW HANDLER

```go
type ListingHandler struct { svc service.ListingService }

func (h *ListingHandler) Create(c fiber.Ctx) error {
    var req dto.CreateListingRequest
    if err := c.Bind().JSON(&req); err != nil { return err }
    // call service, return response
}
```
- Extract user ID from `c.Locals("user_id").(uuid.UUID)`
- Never return raw errors — map to HTTP status + message
