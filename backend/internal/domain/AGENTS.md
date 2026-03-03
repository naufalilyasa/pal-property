# AGENTS.md — backend/internal/domain/

## OVERVIEW

Core contracts of the application. Defines what exists (entities), what's possible (interfaces), and what can go wrong (errors). Everything else depends on this package — it depends on nothing internal.

## FILES

| File | Purpose |
|------|---------|
| `errors.go` | 5 sentinel errors for all layers |
| `repository.go` | `AuthRepository` interface |
| `cache_repository.go` | `CacheRepository` interface |
| `entity/` | GORM-tagged structs |
| `mocks/` | testify/mock generated implementations |

## SENTINEL ERRORS

```go
var (
    ErrNotFound          = errors.New("not found")
    ErrConflict          = errors.New("conflict")
    ErrInvalidCredential = errors.New("invalid credential")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
)
```

Map these to HTTP in the global error handler (`router.go`), NOT in handlers.

## ENTITY CONVENTIONS

**BaseEntity** (embed in all entities):
```go
type BaseEntity struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
// BeforeCreate auto-assigns uuid.NewV7()
```

**Rules**:
- Price/money fields → `int64` (IDR rupiah, no decimals)
- Nullable fields → pointer types (`*string`, `*uuid.UUID`)
- JSON columns → `datatypes.JSON` (from `gorm.io/datatypes`) — NOT `pgtype.JSONB`
- Sensitive fields (OAuth tokens) → stored encrypted, tagged `json:"-"` on raw value

## ADDING A NEW REPOSITORY INTERFACE

```go
// In repository.go — add to existing file or create new domain file
type ListingRepository interface {
    Create(ctx context.Context, listing *entity.Listing) (*entity.Listing, error)
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Listing, error)
    // ...
}
```

## ANTI-PATTERNS

- **NEVER** import any internal package here (no service, no handler, no pkg)
- **NEVER** import `gorm.io/gorm` here (except inside `entity/` structs for tags)
- **NEVER** add HTTP or DB logic here — interfaces only
