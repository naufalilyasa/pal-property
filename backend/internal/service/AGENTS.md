# AGENTS.md — backend/internal/service/

## OVERVIEW

Business logic layer. Orchestrates repositories and cache. Has zero knowledge of HTTP, GORM, or Redis.

## PATTERN

```go
type authService struct {
    repo  domain.AuthRepository   // interface
    cache domain.CacheRepository  // interface
}

func (s *authService) CompleteAuth(ctx context.Context, gothUser *goth.User) (*entity.User, error) {
    // 1. Try find existing OAuth account
    _, err := s.repo.FindOAuthAccount(ctx, gothUser.Provider, gothUser.UserID)
    if err != nil && !errors.Is(err, domain.ErrNotFound) {
        return nil, err  // unexpected error
    }
    // 2. Branch on domain.ErrNotFound — no gorm anywhere
    if errors.Is(err, domain.ErrNotFound) {
        // ... create new user
    }
    // ...
}
```

## RULES

- **Only imports**: `domain` package interfaces, `domain/entity`, `pkg/utils/jwt`, stdlib
- **Error checking**: `errors.Is(err, domain.ErrNotFound)` — NEVER `gorm.ErrRecordNotFound`
- **No gorm import** — the compiler enforces this at build time
- **Return domain errors** to handler — let global error handler map to HTTP

## TEST PATTERN

Unit tests with `testify/mock`:
```go
func TestCompleteAuth_NewUser(t *testing.T) {
    mockRepo := &mocks.AuthRepository{}
    mockCache := &mocks.CacheRepository{}
    mockRepo.On("FindOAuthAccount", ...).Return(nil, domain.ErrNotFound)
    mockRepo.On("FindUserByEmail", ...).Return(nil, domain.ErrNotFound)
    mockRepo.On("CreateUserWithOAuth", ...).Return(&entity.User{...}, nil)
    svc := NewAuthService(mockRepo, mockCache)
    user, err := svc.CompleteAuth(ctx, &gothUser)
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## ANTI-PATTERNS

- **NEVER** `import "gorm.io/gorm"` — build will pass but it's an architecture violation
- **NEVER** call `config.Env` directly — receive config values via constructor
- **NEVER** handle HTTP concerns (status codes, cookies) — that's handler territory
