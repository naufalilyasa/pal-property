# Plan: Category Management

**Generated:** 2026-03-04
**Module:** `github.com/naufalilyasa/pal-property-backend`
**Feature:** Category CRUD + seeder + RequireRole middleware + nested category in listing response

---

## Decisions Made (Do Not Re-ask)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Category FK on listing delete | `ON DELETE SET NULL` via migration | Listings keep existing, lose category ref |
| Slug on Update | **Locked after creation** — slug field excluded from UpdateCategoryRequest | Prevents broken URLs/bookmarks |
| `GET /api/categories/:slug` response | Category + Children + Parent | Enables breadcrumb on listing detail page |
| Seed UUIDs | **Fixed UUID literals** | Reproducible across envs, integration tests predictable |
| Tree query strategy | `Preload("Children").Where("parent_id IS NULL")` | Sufficient for 2-level tree, no CTE needed |
| CategoryResponse in ListingResponse | `CategoryShortResponse{id, name, slug, icon_url}` only — no children array | Keeps listing payload lean |

---

## Architecture Map

```
New files:
  backend/db/migrations/000004_category_fk_set_null_and_seed.up.sql
  backend/db/migrations/000004_category_fk_set_null_and_seed.down.sql
  backend/internal/domain/category_repository.go
  backend/internal/domain/mocks/CategoryRepository.go          ← mockery generated
  backend/internal/dto/request/category_request.go
  backend/internal/dto/response/category_response.go
  backend/internal/repository/postgres/category.go
  backend/internal/service/category_service.go
  backend/internal/service/category_service_test.go
  backend/internal/handler/http/category.go
  backend/internal/handler/http/category_test.go
  backend/pkg/middleware/role.go

Modified files:
  backend/internal/domain/entity/listing.go                    ← add BeforeCreate to Category
  backend/internal/dto/response/listing_response.go            ← add Category *CategoryShortResponse
  backend/internal/service/listing_service.go                  ← update mapToResponse to embed category
  backend/internal/router/router.go                            ← new routes + updated Register sig
  backend/cmd/property-service/main.go                         ← DI wiring for CategoryHandler
  backend/postman_collection.json                              ← add Category section
```

---

## - [ ] Task 1 — Add `BeforeCreate` Hook to `entity.Category`

**File:** `backend/internal/domain/entity/listing.go`
**Why:** Category has no `BeforeCreate` hook. GORM will insert `uuid.Nil` without it.
**What:** Add a `BeforeCreate` method directly on `Category` (do NOT embed BaseEntity — Category has no UpdatedAt/DeletedAt by design):

```go
func (c *Category) BeforeCreate(tx *gorm.DB) error {
    if c.ID == uuid.Nil {
        id, err := uuid.NewV7()
        if err != nil {
            return err
        }
        c.ID = id
    }
    return nil
}
```

**Import needed:** `gorm.io/gorm` (only in entity layer — entity is NOT service layer, this is fine).

**QA:** Write a quick unit assertion: `cat := entity.Category{Name:"X", Slug:"x"}; db.Create(&cat); assert cat.ID != uuid.Nil`.

---

## - [ ] Task 2 — Migration 000004: FK Change + Seed Data

**Files:**
- `backend/db/migrations/000004_category_fk_set_null_and_seed.up.sql`
- `backend/db/migrations/000004_category_fk_set_null_and_seed.down.sql`

### up.sql

**Part A — FK change (listings.category_id → ON DELETE SET NULL):**
```sql
ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_category_id_fkey,
    ADD CONSTRAINT listings_category_id_fkey
        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
```

**Part B — Seed categories (fixed UUIDs, ON CONFLICT DO NOTHING):**

Root categories:
```sql
-- Residensial
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000001', 'Residensial', 'residensial', NULL, NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Komersial
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000002', 'Komersial', 'komersial', NULL, NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Lainnya
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000003', 'Lainnya', 'lainnya', NULL, NULL, NOW())
ON CONFLICT (slug) DO NOTHING;
```

Children of Residensial (`parent_id = '11111111-...-0001'`):
```sql
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000011', 'Rumah',      'rumah',      '11111111-0000-7000-8000-000000000001', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000012', 'Apartemen',  'apartemen',  '11111111-0000-7000-8000-000000000001', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000013', 'Kos',        'kos',        '11111111-0000-7000-8000-000000000001', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000014', 'Villa',      'villa',      '11111111-0000-7000-8000-000000000001', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000015', 'Townhouse',  'townhouse',  '11111111-0000-7000-8000-000000000001', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;
```

Children of Komersial:
```sql
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000021', 'Ruko',       'ruko',       '11111111-0000-7000-8000-000000000002', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000022', 'Kantor',     'kantor',     '11111111-0000-7000-8000-000000000002', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000023', 'Gudang',     'gudang',     '11111111-0000-7000-8000-000000000002', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000024', 'Toko',       'toko',       '11111111-0000-7000-8000-000000000002', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;
```

Children of Lainnya:
```sql
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
  ('11111111-0000-7000-8000-000000000031', 'Tanah',      'tanah',      '11111111-0000-7000-8000-000000000003', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000032', 'Kavling',    'kavling',    '11111111-0000-7000-8000-000000000003', NULL, NOW()),
  ('11111111-0000-7000-8000-000000000033', 'Kebun',      'kebun',      '11111111-0000-7000-8000-000000000003', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;
```

### down.sql
```sql
-- Remove seed data
DELETE FROM categories WHERE id LIKE '11111111-0000-7000-8000-%';

-- Revert FK to no ON DELETE clause (RESTRICT is PG default)
ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_category_id_fkey,
    ADD CONSTRAINT listings_category_id_fkey
        FOREIGN KEY (category_id) REFERENCES categories(id);
```

**QA:** Run `go run ./cmd/migrate/main.go` locally. Verify `SELECT COUNT(*) FROM categories` = 16. Re-run — still 16 (idempotent via ON CONFLICT).

---

## - [ ] Task 3 — `CategoryRepository` Domain Interface

**File:** `backend/internal/domain/category_repository.go`

```go
package domain

import (
    "context"
    "github.com/google/uuid"
    "github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type CategoryRepository interface {
    FindAll(ctx context.Context) ([]entity.Category, error)
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
    FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
    Create(ctx context.Context, category *entity.Category) (*entity.Category, error)
    Update(ctx context.Context, category *entity.Category, fields []string) (*entity.Category, error)
    Delete(ctx context.Context, id uuid.UUID) error
    ExistsBySlug(ctx context.Context, slug string) (bool, error)
    CountListingsByCategory(ctx context.Context, id uuid.UUID) (int64, error)
    CountChildrenByParent(ctx context.Context, parentID uuid.UUID) (int64, error)
}
```

**QA:** `go build ./...` compiles clean after this file is created.

---

## - [ ] Task 4 — Generate `CategoryRepository` Mock

**Command (run from `backend/`):**
```bash
go run github.com/vektra/mockery/v2 \
  --name=CategoryRepository \
  --dir=internal/domain \
  --output=internal/domain/mocks \
  --outpkg=mocks
```

If mockery is not in PATH, use the binary from go tool: check `go.mod` for mockery version and use `go run github.com/vektra/mockery/v2@<version>`.

**Output:** `backend/internal/domain/mocks/CategoryRepository.go`

**QA:** File exists, compiles, exports `NewCategoryRepository(t)` constructor.

---

## - [ ] Task 5 — `CategoryResponse` and `CategoryShortResponse` DTOs

**File:** `backend/internal/dto/response/category_response.go`

```go
package response

import (
    "time"
    "github.com/google/uuid"
)

// CategoryShortResponse is embedded inside ListingResponse — lean, no children.
type CategoryShortResponse struct {
    ID      uuid.UUID `json:"id"`
    Name    string    `json:"name"`
    Slug    string    `json:"slug"`
    IconURL *string   `json:"icon_url"`
}

// CategoryResponse is returned for GET /api/categories and GET /api/categories/:slug.
type CategoryResponse struct {
    ID        uuid.UUID               `json:"id"`
    Name      string                  `json:"name"`
    Slug      string                  `json:"slug"`
    ParentID  *uuid.UUID              `json:"parent_id"`
    IconURL   *string                 `json:"icon_url"`
    CreatedAt time.Time               `json:"created_at"`
    Parent    *CategoryShortResponse  `json:"parent,omitempty"`
    Children  []CategoryShortResponse `json:"children,omitempty"`
}
```

**QA:** `go build ./...` compiles clean.

---

## - [ ] Task 6 — `CreateCategoryRequest` and `UpdateCategoryRequest` DTOs

**File:** `backend/internal/dto/request/category_request.go`

```go
package request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
    Name     string     `json:"name"     validate:"required,min=2,max=100"`
    ParentID *uuid.UUID `json:"parent_id"`
    IconURL  *string    `json:"icon_url"`
}

type UpdateCategoryRequest struct {
    Name    *string `json:"name"     validate:"omitempty,min=2,max=100"`
    IconURL *string `json:"icon_url"`
    // Slug is intentionally excluded — locked after creation
}
```

**QA:** `go build ./...` compiles clean.

---

## - [ ] Task 7 — `RequireRole` Middleware

**File:** `backend/pkg/middleware/role.go`

```go
package middleware

import (
    "github.com/gofiber/fiber/v3"
)

// RequireRole returns a middleware that checks c.Locals("user_role").
// MUST be used AFTER middleware.Protected (which sets user_role).
// Returns 403 Forbidden if role does not match any of the allowed roles.
func RequireRole(roles ...string) fiber.Handler {
    return func(c fiber.Ctx) error {
        role, ok := c.Locals("user_role").(string)
        if !ok || role == "" {
            return fiber.NewError(fiber.StatusForbidden, "forbidden")
        }
        for _, r := range roles {
            if role == r {
                return c.Next()
            }
        }
        return fiber.NewError(fiber.StatusForbidden, "forbidden")
    }
}
```

**QA:** Compile. Unit test (no fiber app needed — just pass a mock context or use fiber test helper):
- Role "admin" → calls Next()
- Role "user" → returns 403
- No role set in Locals → returns 403

---

## - [ ] Task 8 — `CategoryRepository` Postgres Implementation

**File:** `backend/internal/repository/postgres/category.go`

Follow `listing.go` patterns exactly: `db.WithContext(ctx)`, translate errors, check `RowsAffected`.

```go
package postgres

import (
    "context"
    "strings"
    "fmt"

    "github.com/google/uuid"
    "gorm.io/gorm"
    "github.com/naufalilyasa/pal-property-backend/internal/domain"
    "github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type categoryRepository struct { db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
    return &categoryRepository{db: db}
}

// FindAll returns root categories (parent_id IS NULL) with their Children preloaded.
func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
    var cats []entity.Category
    err := r.db.WithContext(ctx).
        Preload("Children").
        Where("parent_id IS NULL").
        Order("name ASC").
        Find(&cats).Error
    if err != nil {
        return nil, fmt.Errorf("category FindAll: %w", err)
    }
    return cats, nil
}

// FindByID preloads Parent and Children for breadcrumb + sub-navigation.
func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
    var cat entity.Category
    err := r.db.WithContext(ctx).
        Preload("Children").
        Preload("Parent").
        First(&cat, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrNotFound
        }
        return nil, fmt.Errorf("category FindByID: %w", err)
    }
    return &cat, nil
}

// FindBySlug preloads Parent and Children.
func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
    var cat entity.Category
    err := r.db.WithContext(ctx).
        Preload("Children").
        Preload("Parent").
        First(&cat, "slug = ?", slug).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrNotFound
        }
        return nil, fmt.Errorf("category FindBySlug: %w", err)
    }
    return &cat, nil
}

// Create inserts a new category. Translates unique constraint violation to ErrConflict.
func (r *categoryRepository) Create(ctx context.Context, cat *entity.Category) (*entity.Category, error) {
    err := r.db.WithContext(ctx).Create(cat).Error
    if err != nil {
        if strings.Contains(err.Error(), "23505") {
            return nil, domain.ErrConflict
        }
        return nil, fmt.Errorf("category Create: %w", err)
    }
    return cat, nil
}

// Update patches only the specified fields.
func (r *categoryRepository) Update(ctx context.Context, cat *entity.Category, fields []string) (*entity.Category, error) {
    err := r.db.WithContext(ctx).Model(cat).Select(fields).Updates(cat).Error
    if err != nil {
        if strings.Contains(err.Error(), "23505") {
            return nil, domain.ErrConflict
        }
        return nil, fmt.Errorf("category Update: %w", err)
    }
    return cat, nil
}

// Delete hard-deletes a category by ID. Service must pre-check children + listings.
func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&entity.Category{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("category Delete: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return domain.ErrNotFound
    }
    return nil
}

func (r *categoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&entity.Category{}).Where("slug = ?", slug).Count(&count).Error
    return count > 0, err
}

func (r *categoryRepository) CountListingsByCategory(ctx context.Context, id uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&entity.Listing{}).
        Where("category_id = ? AND deleted_at IS NULL", id).Count(&count).Error
    return count, err
}

func (r *categoryRepository) CountChildrenByParent(ctx context.Context, parentID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&entity.Category{}).
        Where("parent_id = ?", parentID).Count(&count).Error
    return count, err
}
```

**Note:** Add `"errors"` import alongside `"fmt"` for `errors.Is`.

**QA:** `go build ./...` compiles clean.

---

## - [ ] Task 9 — `CategoryService`

**File:** `backend/internal/service/category_service.go`

### Interface + Struct

```go
type CategoryService interface {
    List(ctx context.Context) ([]*response.CategoryResponse, error)
    GetByID(ctx context.Context, id uuid.UUID) (*response.CategoryResponse, error)
    GetBySlug(ctx context.Context, slug string) (*response.CategoryResponse, error)
    Create(ctx context.Context, req request.CreateCategoryRequest) (*response.CategoryResponse, error)
    Update(ctx context.Context, id uuid.UUID, req request.UpdateCategoryRequest) (*response.CategoryResponse, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
    repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) CategoryService {
    return &categoryService{repo: repo}
}
```

### Method Specs

**List:**
- Call `repo.FindAll(ctx)` → get root categories with `Children` preloaded
- Map each root to `CategoryResponse` with `Children []CategoryShortResponse`
- Children mapped from `cat.Children` (already preloaded by repo)

**GetByID / GetBySlug:**
- Call `repo.FindByID` / `repo.FindBySlug`
- Map to `CategoryResponse` including `Parent *CategoryShortResponse` and `Children []CategoryShortResponse`

**Create:**
- Generate slug via `slug.GenerateUnique(req.Name, func(s string) bool { exists, _ := repo.ExistsBySlug(ctx, s); return exists })`
- Build `entity.Category{Name: req.Name, Slug: slug, ParentID: req.ParentID, IconURL: req.IconURL}`
- Call `repo.Create(ctx, &cat)`
- Return mapped `CategoryResponse` (no Children/Parent yet — just created, return minimal)

**Update:**
- Fetch existing via `repo.FindByID(ctx, id)` → 404 if not found
- Patch fields: if `req.Name != nil` → update Name (slug remains unchanged — locked)
- if `req.IconURL != nil` → update IconURL
- Build `fields []string` of changed columns
- If len(fields) == 0 → return current category response (no-op)
- Call `repo.Update(ctx, &cat, fields)`
- Return mapped response

**Delete:**
- Fetch existing via `repo.FindByID(ctx, id)` → 404 if not found
- Check children: `repo.CountChildrenByParent(ctx, id)` → if > 0 return `domain.ErrConflict` with message "category has children"
- Check listings: `repo.CountListingsByCategory(ctx, id)` → if > 0 return `domain.ErrConflict` with message "category is used by listings"
- Call `repo.Delete(ctx, id)`

### mapToResponse helper:

```go
func mapCategoryToResponse(cat *entity.Category) *response.CategoryResponse {
    r := &response.CategoryResponse{
        ID:        cat.ID,
        Name:      cat.Name,
        Slug:      cat.Slug,
        ParentID:  cat.ParentID,
        IconURL:   cat.IconURL,
        CreatedAt: cat.CreatedAt,
    }
    if cat.Parent != nil {
        r.Parent = &response.CategoryShortResponse{
            ID:      cat.Parent.ID,
            Name:    cat.Parent.Name,
            Slug:    cat.Parent.Slug,
            IconURL: cat.Parent.IconURL,
        }
    }
    for _, child := range cat.Children {
        child := child // copy for safety
        r.Children = append(r.Children, response.CategoryShortResponse{
            ID:      child.ID,
            Name:    child.Name,
            Slug:    child.Slug,
            IconURL: child.IconURL,
        })
    }
    return r
}
```

**QA:** `go build ./...` clean.

---

## - [ ] Task 10 — `CategoryHandler`

**File:** `backend/internal/handler/http/category.go`

Follow `listing.go` pattern: extract from Locals, Bind().JSON(), validate, call service, use `utils.SendResponse`.

```go
type CategoryHandler struct {
    svc service.CategoryService
}

func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
    return &CategoryHandler{svc: svc}
}
```

### Endpoints:

| Method | Handler | Notes |
|--------|---------|-------|
| `List` | `GET /api/categories` | Public. No auth. Returns tree. |
| `GetBySlug` | `GET /api/categories/:slug` | Public. Returns category + parent + children. |
| `Create` | `POST /api/categories/` | Admin only. Bind + validate. Returns 201. |
| `Update` | `PUT /api/categories/:id` | Admin only. Parse UUID from param. Returns 200. |
| `Delete` | `DELETE /api/categories/:id` | Admin only. Parse UUID from param. Returns 200. |

**Error handling:**
- `domain.ErrNotFound` → 404
- `domain.ErrConflict` → 409
- Validation error → 400
- UUID parse error → 400 with "invalid category id"

**QA:** `go build ./...` clean.

---

## - [ ] Task 11 — Router Update

**File:** `backend/internal/router/router.go`

### A. Update `Register` function signature:

```go
func Register(
    app *fiber.App,
    db *gorm.DB,
    authHandler *http.AuthHandler,
    listingHandler *http.ListingHandler,
    categoryHandler *http.CategoryHandler,  // ← ADD
) {
```

### B. Add public category routes (inside `api` group):

```go
// Public category routes
api.Get("/categories", categoryHandler.List)
api.Get("/categories/:slug", categoryHandler.GetBySlug)
```

### C. Add admin category routes (new group, after existing listing routes):

```go
// Admin category routes — Protected FIRST, then RequireRole
categoryAdmin := api.Group("/categories",
    middleware.Protected(db),
    middleware.RequireRole("admin"),
)
categoryAdmin.Post("/", categoryHandler.Create)
categoryAdmin.Put("/:id", categoryHandler.Update)
categoryAdmin.Delete("/:id", categoryHandler.Delete)
```

**QA:** `go build ./...` clean. Route order: public GET routes must come before the admin group to avoid conflict.

---

## - [ ] Task 12 — DI Wiring in `main.go`

**File:** `backend/cmd/property-service/main.go`

```go
// Add after existing repo/service/handler setup:
categoryRepo := postgres.NewCategoryRepository(db)
categoryService := service.NewCategoryService(categoryRepo)
categoryHandler := handler.NewCategoryHandler(categoryService)

// Update router.Register call:
router.Register(app, db, authHandler, listingHandler, categoryHandler)
```

**QA:** `go run ./cmd/property-service` starts without error. `go build ./...` clean.

---

## - [ ] Task 13 — Update `ListingResponse` with Nested Category

**File:** `backend/internal/dto/response/listing_response.go`

Add field to `ListingResponse`:
```go
Category *CategoryShortResponse `json:"category,omitempty"`
```

Place after `CategoryID *uuid.UUID` line. Import `category_response.go` is in the same package — no import change needed.

**QA:** `go build ./...` clean.

---

## - [ ] Task 14 — Update `mapToResponse` in Listing Service

**File:** `backend/internal/service/listing_service.go`

In the `mapToResponse` function, add category mapping after existing field mappings:

```go
// After existing field assignments:
if l.Category != nil {
    resp.Category = &response.CategoryShortResponse{
        ID:      l.Category.ID,
        Name:    l.Category.Name,
        Slug:    l.Category.Slug,
        IconURL: l.Category.IconURL,
    }
}
```

**Note:** `l.Category` is already preloaded in `FindByID` and `FindBySlug` in the listing repo. For `List`, check if the listing repo's `List` method also preloads `Category` — if not, add `Preload("Category")` to the `List` query in `repository/postgres/listing.go`.

**QA:**
- `GET /api/listings/:id` response includes `"category": {"id":"...", "name":"Rumah", "slug":"rumah", "icon_url":null}` when listing has category_id set
- `GET /api/listings/:id` response includes `"category": null` when listing has no category

---

## - [ ] Task 15 — Category Service Unit Tests

**File:** `backend/internal/service/category_service_test.go`

**Package:** `service_test`
**Pattern:** Flat `TestXxx` functions. Use `mocks.NewCategoryRepository(t)`.

Test cases:

```
List_Success             → repo.FindAll returns 2 roots each with children → mapped correctly
List_Empty               → repo.FindAll returns [] → returns empty slice (not nil)
GetByID_Success          → parent + children in response
GetByID_NotFound         → repo returns ErrNotFound → service returns ErrNotFound
GetBySlug_Success        → with parent preloaded (parent != nil)
GetBySlug_NotFound       → ErrNotFound
Create_Success           → ExistsBySlug=false, repo.Create succeeds → returns response with auto-slug
Create_SlugCollision     → ExistsBySlug=true once, then false → slug gets suffix
Create_RepoError         → repo.Create returns error → propagated
Update_Success           → fetch ok, name changed, repo.Update called with ["name"]
Update_NoChanges         → all fields nil → repo.Update NOT called, returns current
Update_NotFound          → repo.FindByID returns ErrNotFound → 404
Delete_Success           → no children, no listings, repo.Delete called
Delete_HasChildren       → CountChildren > 0 → ErrConflict, repo.Delete NOT called
Delete_HasListings       → CountListings > 0 → ErrConflict, repo.Delete NOT called
Delete_NotFound          → repo.FindByID ErrNotFound → ErrNotFound
```

**QA:** `go test ./internal/service/... -v -count=1` all pass.

---

## - [ ] Task 16 — Category Handler Integration Tests

**File:** `backend/internal/handler/http/category_test.go`

**Package:** `http_test`
**Pattern:** `testify/suite`, testcontainers postgres:17-alpine. No Redis needed.
**Suite name:** `CategoryHandlerTestSuite`
**Runner:** `TestCategoryHandlerSuite(t *testing.T)`

### SetupSuite:
- Spin up `postgres:17-alpine` testcontainer
- Run GORM AutoMigrate for `entity.User`, `entity.Category`, `entity.Listing`, `entity.ListingImage`
- Create real layers: `categoryRepo → categoryService → categoryHandler`
- Spin up Fiber app with manual route registration (no `router.Register` — to avoid rate limiter)
- Set `logger.Log = zap.NewNop()` and `config.Env.AppEnv = "testing"`
- Generate RSA keypair and set `config.Env.Jwt*` vars for JWT minting

### SetupTest:
```sql
-- MUST truncate in FK-safe order
TRUNCATE listings CASCADE;
TRUNCATE categories CASCADE;
TRUNCATE users CASCADE;
```

### Helper methods:
- `createAdminUser() entity.User` — inserts user with role="admin"
- `createUser() entity.User` — inserts user with role="user"
- `mintJWT(userID uuid.UUID, role string) string` — returns access token string
- `createCategory(name, slug string, parentID *uuid.UUID) entity.Category` — direct DB insert
- `setCookie(req *http.Request, token string)` — adds `access_token` cookie

### Test cases:

```
GET /api/categories
  List_Empty              → 200, data=[] (empty slice not null)
  List_WithData           → seed 2 root + 3 children, 200, data has 2 roots each with children

GET /api/categories/:slug
  GetBySlug_Found         → 200, response has slug, children, parent fields
  GetBySlug_NotFound      → 404

POST /api/categories/
  Create_AsAdmin_201      → admin token, valid payload, 201, slug auto-generated
  Create_AsUser_403       → user token (role="user"), 403
  Create_NoAuth_401       → no cookie, 401
  Create_MissingName_400  → missing name field, 400
  Create_DuplicateSlug_409 → same name twice, second returns 409

PUT /api/categories/:id
  Update_AsAdmin_200      → change name, 200, name updated, slug unchanged
  Update_AsUser_403       → 403
  Update_NotFound_404     → non-existent UUID, 404

DELETE /api/categories/:id
  Delete_AsAdmin_200      → leaf category, 200
  Delete_HasChildren_409  → parent with children, 409
  Delete_HasListings_409  → category used by listing, 409
  Delete_AsUser_403       → 403
  Delete_NotFound_404     → non-existent UUID, 404

GET /api/listings/:id (regression — nested category)
  ListingWithCategory     → create listing with category_id set, GET listing, assert .data.category.slug not empty
  ListingWithoutCategory  → listing with no category_id, .data.category is null
```

**QA:** `go test ./internal/handler/http/... -run TestCategoryHandlerSuite -v -timeout 120s` all pass.

---

## - [ ] Task 17 — Update Postman Collection

**File:** `backend/postman_collection.json`

Add a new "Categories" section between "Authentication" and "Listings" with these requests:

1. `GET /api/categories` — List All (tree) — public
2. `GET /api/categories/:slug` — Get By Slug — public  
3. `POST /api/categories/` — Create (Admin) — with example body `{"name":"Kondominium","parent_id":"11111111-0000-7000-8000-000000000001"}`
4. `PUT /api/categories/:id` — Update (Admin) — body `{"name":"Kondominium Updated"}`
5. `DELETE /api/categories/:id` — Delete (Admin)

Add collection variable `categoryId` and `categorySlug`.

**QA:** JSON is valid (parse with `python3 -m json.tool postman_collection.json`).

---

## Final Verification Wave

Run these in sequence after all tasks complete:

```bash
# 1. Build clean
cd backend && go build ./... 2>&1

# 2. Vet
cd backend && go vet ./... 2>&1

# 3. Migrate (local DB must be running)
cd backend && go run ./cmd/migrate/main.go 2>&1

# 4. All unit tests (fast)
cd backend && go test ./internal/service/... -v -count=1 2>&1

# 5. Category integration tests
cd backend && go test ./internal/handler/http/... -run TestCategoryHandlerSuite -v -count=1 -timeout 120s 2>&1

# 6. Listing integration tests (regression — nested category)
cd backend && go test ./internal/handler/http/... -run TestListingHandlerSuite -v -count=1 -timeout 120s 2>&1

# 7. Full suite
cd backend && go test ./... -count=1 -timeout 180s 2>&1

# 8. Postman JSON valid
python3 -m json.tool backend/postman_collection.json > /dev/null && echo "VALID"
```

All 8 commands must exit 0. Zero test failures.

---

## Checklist of Files to Create/Modify

| Action | File |
|--------|------|
| **MODIFY** | `backend/internal/domain/entity/listing.go` — BeforeCreate on Category |
| **CREATE** | `backend/db/migrations/000004_category_fk_set_null_and_seed.up.sql` |
| **CREATE** | `backend/db/migrations/000004_category_fk_set_null_and_seed.down.sql` |
| **CREATE** | `backend/internal/domain/category_repository.go` |
| **CREATE** | `backend/internal/domain/mocks/CategoryRepository.go` |
| **CREATE** | `backend/internal/dto/request/category_request.go` |
| **CREATE** | `backend/internal/dto/response/category_response.go` |
| **CREATE** | `backend/pkg/middleware/role.go` |
| **CREATE** | `backend/internal/repository/postgres/category.go` |
| **CREATE** | `backend/internal/service/category_service.go` |
| **CREATE** | `backend/internal/service/category_service_test.go` |
| **CREATE** | `backend/internal/handler/http/category.go` |
| **CREATE** | `backend/internal/handler/http/category_test.go` |
| **MODIFY** | `backend/internal/router/router.go` — signature + new routes |
| **MODIFY** | `backend/cmd/property-service/main.go` — DI wiring |
| **MODIFY** | `backend/internal/dto/response/listing_response.go` — add Category field |
| **MODIFY** | `backend/internal/service/listing_service.go` — update mapToResponse |
| **MODIFY** | `backend/postman_collection.json` — add Categories section |
