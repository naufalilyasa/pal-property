# AGENTS.md — backend/pkg/

## OVERVIEW

Shared, reusable utilities. No internal domain knowledge — importable by any layer including future workers.

## PACKAGES

| Package | Purpose | Key API |
|---------|---------|---------|
| `config/` | App config via `caarlos0/env` | `config.Env` (global), `config.LoadConfig()` |
| `crypto/` | AES-256-GCM encrypt/decrypt | `crypto.Encrypt(plain, key)`, `crypto.Decrypt(encoded, key)` |
| `utils/jwt/` | RS256 JWT generate/validate | `jwt.GenerateTokens(...)`, `jwt.ValidateToken(...)` |
| `logger/` | Zap structured logging | `logger.InitLogger()`, `logger.Log` |
| `kafka/` | Redpanda/Kafka producer | (future — scaffolded) |
| `middleware/` | Shared Fiber middleware | (future) |
| `validator/` | go-playground/validator setup | `validator.Validate` |

## CONFIG (`pkg/config/config.go`)

Uses `caarlos0/env/v11` — struct tags drive env parsing:
```go
type AppConfig struct {
    AppEnv  string `env:"APP_ENV" validate:"required"`
    Port    string `env:"PORT"    validate:"required"`
    // DB, Redis, OAuth, JWT, Crypto fields...
}
var Env AppConfig  // global singleton
```

**Loading**: `config.LoadConfig()` → `env.Parse()` → `validator.Struct()` → base64-decode JWT keys + AES key.

**Adding a new config field**:
1. Add field to `AppConfig` with `env:"VAR_NAME"` tag
2. Add `VAR_NAME=value` to `.env-example` (MANDATORY)
3. Add to `.env` and `.env.docker` for local dev

## CRYPTO (`pkg/crypto/aes.go`)

```go
// key MUST be exactly 32 bytes (base64-decoded from OAUTH_TOKEN_ENCRYPTION_KEY)
encrypted, err := crypto.Encrypt(plaintext, config.Env.OAuthTokenEncryptionKey)
decrypted, err := crypto.Decrypt(encoded, config.Env.OAuthTokenEncryptionKey)
```

Output format: base64url-encoded `nonce(12) + ciphertext + tag(16)`.
Use for: OAuth provider access/refresh tokens, any PII stored in DB.

## JWT (`pkg/utils/jwt/jwt.go`)

- Algorithm: RS256 (asymmetric — safe to share public key)
- Keys: base64 PEM in `JWT_PRIVATE_KEY_BASE64` / `JWT_PUBLIC_KEY_BASE64`
- Generates: `(accessToken, refreshToken, jti string, error)`
- JTI stored in Redis for refresh token rotation + revocation

## ANTI-PATTERNS

- **NEVER** import `internal/` packages from `pkg/` — pkg must stay domain-agnostic
- **NEVER** add config field without `.env-example` entry
- **NEVER** use `config.Env` directly in `pkg/` functions — pass config values as parameters
