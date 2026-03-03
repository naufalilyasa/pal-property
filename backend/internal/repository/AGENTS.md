# AGENTS.md — backend/internal/repository/

## OVERVIEW

Data access layer. Two sub-packages: `postgres/` (GORM) and `redis/` (go-redis). Implements domain interfaces. Responsible for: error translation, crypto, query execution.

## POSTGRES PATTERN (`postgres/`)

```go
type authRepository struct{ db *gorm.DB }

func (r *authRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrNotFound  // MANDATORY translation
        }
        return nil, fmt.Errorf("find user by id: %w", err)
    }
    return &user, nil
}
```

**Rules**:
- `WithContext(ctx)` on EVERY query
- Translate `gorm.ErrRecordNotFound` → `domain.ErrNotFound` in EVERY find method
- Wrap other errors: `fmt.Errorf("operation name: %w", err)`
- Sensitive fields encrypted HERE (not in service): `crypto.Encrypt(token, key)` before save, `crypto.Decrypt` after load

## CRYPTO IN REPOSITORY

OAuth tokens are encrypted/decrypted at the repository boundary:
```go
// Before save (CreateUserWithOAuth):
if account.AccessToken != nil {
    enc, err := crypto.Encrypt(*account.AccessToken, config.Env.OAuthTokenEncryptionKey)
    account.AccessToken = &enc
}

// After load (FindOAuthAccount):
if account.AccessToken != nil {
    dec, err := crypto.Decrypt(*account.AccessToken, config.Env.OAuthTokenEncryptionKey)
    account.AccessToken = &dec
}
```

## REDIS PATTERN (`redis/`)

```go
type cacheRepository struct{ client *redis.Client }

func (r *cacheRepository) SaveRefreshTokenJTI(ctx context.Context, jti string, userID uuid.UUID, exp time.Duration) error {
    return r.client.Set(ctx, jti, userID.String(), exp).Err()
}
```

## ANTI-PATTERNS

- **NEVER** return `gorm.ErrRecordNotFound` — always translate
- **NEVER** put business logic here — only data access
- **NEVER** access `config.Env` here — receive key as constructor parameter
- **NEVER** use `db.Save()` for partial updates without understanding GORM behavior (prefer `db.Updates()`)
