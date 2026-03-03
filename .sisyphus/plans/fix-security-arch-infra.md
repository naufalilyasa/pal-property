# Fix: Backend Security, Architecture & Infrastructure Issues (#2, #3, #4, #5, #7, #8)

## Context

Project: pal-property (property listing platform)
Backend: Go 1.26, Fiber v3, GORM v1.31, PostgreSQL 17
Module: `github.com/naufalilyasa/pal-property-backend`

## Issues Being Fixed

| # | Issue | Files Affected |
|---|-------|---------------|
| 2 | OAuth tokens stored plaintext in DB | `entity/user.go`, `repository/postgres/auth.go`, `service/auth_service.go`, `pkg/crypto/` (new) |
| 3 | gorm.ErrRecordNotFound leaks into service layer | `domain/errors.go` (new), `repository/postgres/auth.go`, `service/auth_service.go` |
| 4 | CORS missing AllowCredentials + AllowMethods | `internal/router/router.go` |
| 5 | Kafka+Zookeeper → Redpanda | `docker-compose.yml` (root), `pkg/kafka/` (verify no cp-kafka specific APIs) |
| 7 | Price as float64 → int64 (rupiah, no decimal) | `entity/listing.go` |
| 8 | Viper → caarlos0/env | `pkg/config/config.go`, `go.mod` |

---

## Task 1 — Create domain/errors.go (prerequisite for Fix #3)

**File**: `backend/internal/domain/errors.go` (CREATE NEW)

```go
package domain

import "errors"

// Sentinel errors for the domain layer.
// Repository implementations MUST translate ORM/driver-specific errors
// (e.g. gorm.ErrRecordNotFound) into these domain errors so that the
// service layer stays completely decoupled from persistence internals.
var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("record already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)
```

**QA**: Run `go build ./internal/domain/...` — should compile with zero errors.

---

## Task 2 — Create pkg/crypto/aes.go (prerequisite for Fix #2)

**File**: `backend/pkg/crypto/aes.go` (CREATE NEW)

This package provides AES-256-GCM authenticated encryption for sensitive strings stored in the database (OAuth provider tokens).

```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plaintext using AES-256-GCM.
// key MUST be exactly 32 bytes (256-bit). Pass via config (OAUTH_TOKEN_ENCRYPTION_KEY env var).
// Returns a base64url-encoded string: nonce (12 bytes) + ciphertext + auth tag.
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64url-encoded AES-256-GCM ciphertext produced by Encrypt.
func Decrypt(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes for AES-256")
	}
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
```

**QA**: Run `go build ./pkg/crypto/...` — should compile with zero errors.

---

## Task 3 — Fix #8: Replace Viper with caarlos0/env in config.go

### 3a. Install dependency

```bash
cd backend && go get github.com/caarlos0/env/v11@latest && go get -u && go mod tidy
```

Remove Viper and all its deps that are no longer needed:
```bash
go mod edit -droprequire github.com/spf13/viper
go mod edit -droprequire github.com/spf13/cast
go mod edit -droprequire github.com/spf13/afero
go mod edit -droprequire github.com/spf13/pflag
go mod edit -droprequire github.com/fsnotify/fsnotify
go mod edit -droprequire github.com/mitchellh/mapstructure
go mod tidy
```

### 3b. Rewrite config.go

**File**: `backend/pkg/config/config.go` (FULL REWRITE)

```go
package config

import (
	"encoding/base64"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

// AppConfig holds all application configuration loaded from environment variables.
// caarlos0/env reads struct tags directly — no mapstructure workaround needed.
type AppConfig struct {
	AppEnv   string `env:"APP_ENV" envDefault:"development"`
	Port     string `env:"PORT"    envDefault:"8080"`
	AppName  string `env:"APP_NAME" envDefault:"pal-property"`

	// Database
	DBHost     string `env:"DB_HOST"     validate:"required"`
	DBUser     string `env:"DB_USER"     validate:"required"`
	DBPassword string `env:"DB_PASSWORD" validate:"required"`
	DBName     string `env:"DB_NAME"     validate:"required"`
	DBPort     string `env:"DB_PORT"     envDefault:"5432"`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`

	// Redis
	RedisAddr     string `env:"REDIS_ADDR"     validate:"required"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB"       envDefault:"0"`

	// CORS
	CorsAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`

	// Rate Limiting
	RateLimitMax int `env:"RATE_LIMIT_MAX" envDefault:"100"`
	RateLimitExp int `env:"RATE_LIMIT_EXP" envDefault:"60"` // seconds

	// OAuth
	OAuthClientID     string `env:"OAUTH_CLIENT_ID"     validate:"required"`
	OAuthClientSecret string `env:"OAUTH_CLIENT_SECRET" validate:"required"`
	OAuthCallbackURL  string `env:"OAUTH_CALLBACK_URL"  validate:"required"`

	// JWT — RS256, keys stored base64-encoded
	JwtPrivateKeyBase64   string `env:"JWT_PRIVATE_KEY_BASE64" validate:"required"`
	JwtPublicKeyBase64    string `env:"JWT_PUBLIC_KEY_BASE64"  validate:"required"`
	JwtAccessExpiration   int    `env:"JWT_ACCESS_EXPIRATION"  envDefault:"900"`     // seconds, default 15m
	JwtRefreshExpiration  int    `env:"JWT_REFRESH_EXPIRATION" envDefault:"604800"`  // seconds, default 7d

	// Encryption key for OAuth provider tokens stored in DB (AES-256 = 32 bytes, base64-encoded)
	OAuthTokenEncryptionKeyBase64 string `env:"OAUTH_TOKEN_ENCRYPTION_KEY" validate:"required"`

	// Parsed (not from env directly — populated in LoadConfig)
	JwtPrivateKeyPEM []byte `env:"-"`
	JwtPublicKeyPEM  []byte `env:"-"`
	OAuthTokenEncryptionKey []byte `env:"-"` // decoded 32-byte key
}

// Env is the global config singleton populated by LoadConfig.
var Env AppConfig

// LoadConfig parses environment variables into AppConfig, validates required fields,
// and decodes base64 PEM keys.
func LoadConfig() error {
	cfg := AppConfig{}

	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("config: failed to parse env vars: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("config: validation failed: %w", err)
	}

	// Decode base64 JWT keys
	privKey, err := base64.StdEncoding.DecodeString(cfg.JwtPrivateKeyBase64)
	if err != nil {
		return fmt.Errorf("config: invalid JWT_PRIVATE_KEY_BASE64: %w", err)
	}
	cfg.JwtPrivateKeyPEM = privKey

	pubKey, err := base64.StdEncoding.DecodeString(cfg.JwtPublicKeyBase64)
	if err != nil {
		return fmt.Errorf("config: invalid JWT_PUBLIC_KEY_BASE64: %w", err)
	}
	cfg.JwtPublicKeyPEM = pubKey

	// Decode AES encryption key (must decode to exactly 32 bytes)
	encKey, err := base64.StdEncoding.DecodeString(cfg.OAuthTokenEncryptionKeyBase64)
	if err != nil {
		return fmt.Errorf("config: invalid OAUTH_TOKEN_ENCRYPTION_KEY: %w", err)
	}
	if len(encKey) != 32 {
		return fmt.Errorf("config: OAUTH_TOKEN_ENCRYPTION_KEY must decode to exactly 32 bytes (got %d)", len(encKey))
	}
	cfg.OAuthTokenEncryptionKey = encKey

	Env = cfg
	return nil
}
```

**Important**: After rewriting config.go, search for all usages of the OLD field names from Viper/mapstructure era and update them:
- `config.Env.ClientID` → `config.Env.OAuthClientID`
- `config.Env.ClientSecret` → `config.Env.OAuthClientSecret`  
- `config.Env.CallbackURL` → `config.Env.OAuthCallbackURL`
- `config.Env.RateLimitExp` was `time.Duration` → now `int` (seconds). Update usages in limiter middleware accordingly.
- `config.Env.JwtAccessExpiration` / `JwtRefreshExpiration` were `time.Duration` → now `int` (seconds). Update JWT generation usages.

Search with: `grep -r "config\.Env\." backend/ --include="*.go"`

After finding all usages, update each call site. The JWT pkg likely did `config.Env.JwtAccessExpiration` as duration — now pass `time.Duration(config.Env.JwtAccessExpiration) * time.Second`.

**QA**: `go build ./pkg/config/...` must compile.

---

## Task 4 — Fix #3: Fix repository layer (translate GORM errors → domain errors)

**File**: `backend/internal/repository/postgres/auth.go` (EDIT)

Replace all `return nil, err` after `db.First()` / `db.Where().First()` calls with domain error translation.

Pattern to apply to EVERY `.First()` call in this file:
```go
// BEFORE (leaks gorm error):
if err := r.db.Where(...).First(&entity).Error; err != nil {
    return nil, err
}

// AFTER (domain error):
if err := r.db.Where(...).First(&entity).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, domain.ErrNotFound
    }
    return nil, err
}
```

Imports to add: `"errors"`, `"github.com/naufalilyasa/pal-property-backend/internal/domain"`
Imports to KEEP: `"gorm.io/gorm"` (still needed for `gorm.ErrRecordNotFound` check and `*gorm.DB` type — the translation happens HERE in the repository, which is correct)

Apply to all 3 query methods: `FindOAuthAccount`, `FindUserByEmail`, `FindUserByID`.

**QA**: `go build ./internal/repository/...` must compile.

---

## Task 5 — Fix #3 (part 2): Remove gorm import from service layer

**File**: `backend/internal/service/auth_service.go` (EDIT)

1. Remove import `"gorm.io/gorm"`
2. Replace `errors.Is(err, gorm.ErrRecordNotFound)` → `errors.Is(err, domain.ErrNotFound)`

Specific change in `CompleteAuth`:
```go
// BEFORE:
if !errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, fmt.Errorf(...)
}

// AFTER:
if !errors.Is(err, domain.ErrNotFound) {
    return nil, fmt.Errorf(...)
}
```

Ensure import `"github.com/naufalilyasa/pal-property-backend/internal/domain"` is present (it likely already is).

**QA**: `go build ./internal/service/...` — must compile with NO gorm import in this file.

---

## Task 6 — Fix #2: Encrypt OAuth tokens in repository

### 6a. Update OAuthAccount entity (user.go) — add doc comment

**File**: `backend/internal/domain/entity/user.go` (EDIT — doc comment only, no type change)

Add comment above AccessToken and RefreshToken fields:
```go
// AccessToken and RefreshToken are AES-256-GCM encrypted before storage.
// Use pkg/crypto.Decrypt(value, key) to read the plaintext.
AccessToken  *string `gorm:"type:text" json:"-"`
RefreshToken *string `gorm:"type:text" json:"-"`
```

### 6b. Add OAUTH_TOKEN_ENCRYPTION_KEY to .env-example

**File**: `backend/.env-example` (EDIT — append)

```
# AES-256 encryption key for OAuth provider tokens (base64-encoded 32 bytes)
# Generate: openssl rand -base64 32
OAUTH_TOKEN_ENCRYPTION_KEY=
```

### 6c. Encrypt tokens in repository CreateUserWithOAuth

**File**: `backend/internal/repository/postgres/auth.go` (EDIT)

In `CreateUserWithOAuth`, BEFORE calling `r.db.Create(account)`, encrypt the tokens:

```go
import "github.com/naufalilyasa/pal-property-backend/pkg/crypto"
import "github.com/naufalilyasa/pal-property-backend/pkg/config"

// In CreateUserWithOAuth, after account.UserID = user.ID:
if account.AccessToken != nil {
    encrypted, err := crypto.Encrypt(*account.AccessToken, config.Env.OAuthTokenEncryptionKey)
    if err != nil {
        return nil, fmt.Errorf("repository: failed to encrypt access token: %w", err)
    }
    account.AccessToken = &encrypted
}
if account.RefreshToken != nil {
    encrypted, err := crypto.Encrypt(*account.RefreshToken, config.Env.OAuthTokenEncryptionKey)
    if err != nil {
        return nil, fmt.Errorf("repository: failed to encrypt refresh token: %w", err)
    }
    account.RefreshToken = &encrypted
}
```

### 6d. Decrypt tokens in FindOAuthAccount

**File**: `backend/internal/repository/postgres/auth.go` (EDIT, same file as 6c — batch in same edit call)

In `FindOAuthAccount`, AFTER successful `.First()`, decrypt the tokens before returning:

```go
// After successful db.First(&account):
if account.AccessToken != nil {
    decrypted, err := crypto.Decrypt(*account.AccessToken, config.Env.OAuthTokenEncryptionKey)
    if err != nil {
        return nil, fmt.Errorf("repository: failed to decrypt access token: %w", err)
    }
    account.AccessToken = &decrypted
}
if account.RefreshToken != nil {
    decrypted, err := crypto.Decrypt(*account.RefreshToken, config.Env.OAuthTokenEncryptionKey)
    if err != nil {
        return nil, fmt.Errorf("repository: failed to decrypt refresh token: %w", err)
    }
    account.RefreshToken = &decrypted
}
return &account, nil
```

**QA**: `go build ./internal/repository/...` must compile.

---

## Task 7 — Fix #4: CORS AllowCredentials in router.go

**File**: `backend/internal/router/router.go` (EDIT)

Find the `cors.New(cors.Config{...})` call and update it:

```go
// BEFORE:
app.Use(cors.New(cors.Config{
    AllowOrigins: config.Env.CorsAllowedOrigins,
}))

// AFTER:
app.Use(cors.New(cors.Config{
    AllowOrigins:     config.Env.CorsAllowedOrigins,
    AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

> **IMPORTANT**: `AllowCredentials: true` requires `AllowOrigins` to be an explicit origin (NOT `*`). The current config reads from `CorsAllowedOrigins` env var which should already be set to `http://localhost:3000` — this is correct. Do NOT use wildcard `*` when credentials are enabled.

**QA**: `go build ./internal/router/...` must compile.

---

## Task 8 — Fix #5: Replace Kafka+Zookeeper with Redpanda in docker-compose.yml

**File**: `docker-compose.yml` (root of repo) (EDIT)

Remove the entire `zookeeper` service block and the `kafka` service block. Replace with a single `redpanda` service:

```yaml
  redpanda:
    image: redpandadata/redpanda:v24.3.1
    container_name: pal-redpanda
    command:
      - redpanda
      - start
      - --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092
      - --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
      - --pandaproxy-addr internal://0.0.0.0:8082,external://0.0.0.0:18082
      - --advertise-pandaproxy-addr internal://redpanda:8082,external://localhost:18082
      - --schema-registry-addr internal://0.0.0.0:8081,external://0.0.0.0:18081
      - --rpc-addr redpanda:33145
      - --advertise-rpc-addr redpanda:33145
      - --mode dev-container
      - --smp 1
      - --memory 512M
      - --overprovisioned
      - --default-log-level=warn
    ports:
      - "19092:19092"   # Kafka-compatible external port
      - "18082:18082"   # Pandaproxy (REST proxy)
      - "18081:18081"   # Schema Registry
      - "9644:9644"     # Admin API
    volumes:
      - redpanda_data:/var/lib/redpanda/data
    healthcheck:
      test: ["CMD-SHELL", "rpk cluster health | grep -E 'Healthy:.+true' || exit 1"]
      interval: 15s
      timeout: 10s
      retries: 5
      start_period: 30s
```

Also update the `volumes:` section at the bottom — remove `zookeeper_data` and `kafka_data`, add `redpanda_data`.

Update backend service `depends_on` — replace `kafka` with `redpanda`.

Also check `backend/pkg/kafka/` — the Kafka client library (`segmentio/kafka-go` or `confluent-kafka-go` or `IBM/sarama`) must remain compatible. Redpanda is 100% Kafka protocol-compatible so NO changes to the Kafka client code needed. Just verify which library is used:
- If `segmentio/kafka-go`: change broker address from `kafka:9092` → `redpanda:9092` in config/env
- If `IBM/sarama` or `confluent-kafka-go`: same — just update the broker address

Update `.env-example` and `.env.docker`:
```
# BEFORE:
KAFKA_BROKERS=kafka:9092

# AFTER:
KAFKA_BROKERS=redpanda:9092
```

**QA**: `docker compose config` must pass (valid YAML). `docker compose up redpanda -d` should start cleanly.

---

## Task 9 — Fix #7: Price float64 → int64 in listing.go

**File**: `backend/internal/domain/entity/listing.go` (EDIT)

```go
// BEFORE:
Price float64 `gorm:"not null" json:"price"`

// AFTER:
// Price is stored in the smallest currency unit (Indonesian Rupiah, no decimal).
// Example: Rp 500.000.000 is stored as 500000000.
Price int64 `gorm:"not null" json:"price"`
```

**Note**: If any DTO or response struct also declares price as `float64`, update those too.
Search: `grep -rn "Price" backend/ --include="*.go"` — check all hits.

**QA**: `go build ./internal/domain/...` must compile.

---

## Task 10 — Add OAUTH_TOKEN_ENCRYPTION_KEY to .env files

**File**: `backend/.env` (EDIT — append)

Generate a real key and add:
```bash
# Generate key: openssl rand -base64 32
OAUTH_TOKEN_ENCRYPTION_KEY=<generate with: openssl rand -base64 32>
```

Do the same for `backend/.env.docker`.

**QA**: Manually verify the key decodes to 32 bytes: `echo "$OAUTH_TOKEN_ENCRYPTION_KEY" | base64 -d | wc -c` → must print `32`.

---

## Task 11 — go mod tidy + Full Build Verification

```bash
cd backend
go mod tidy
go build ./...
go vet ./...
```

If any compile error, trace to the file and fix. Most likely issues:
- `config.Env.RateLimitExp` usages: was `time.Duration`, now `int`. Update: `time.Duration(config.Env.RateLimitExp) * time.Second`
- `config.Env.JwtAccessExpiration` / `JwtRefreshExpiration`: same pattern
- Any file that imported `github.com/spf13/viper` directly (unlikely, but check)

**QA**: `go build ./... && go vet ./...` must both exit 0.

---

## Task 12 — Run existing tests

```bash
cd backend && go test ./... -count=1 -timeout 60s
```

If tests use config, they may need `OAUTH_TOKEN_ENCRYPTION_KEY` set. Update test setup accordingly.

**QA**: All existing tests must pass (or pre-existing failures only).

---

## Final Verification Wave

- [ ] `domain/errors.go` exists with 5 sentinel errors
- [ ] `pkg/crypto/aes.go` exists with Encrypt/Decrypt using stdlib only (no external dep)
- [ ] `pkg/config/config.go` uses `caarlos0/env`, no Viper import anywhere in pkg/config/
- [ ] `repository/postgres/auth.go` translates to `domain.ErrNotFound`, no raw gorm errors returned
- [ ] `service/auth_service.go` has NO `gorm.io/gorm` import
- [ ] `router.go` CORS has `AllowCredentials: true`, `AllowMethods`, `AllowHeaders`
- [ ] `docker-compose.yml` has NO zookeeper, NO cp-kafka — only redpanda single service
- [ ] `entity/listing.go` Price field is `int64`
- [ ] `go build ./...` exits 0
- [ ] `go vet ./...` exits 0
- [ ] `.env-example` has `OAUTH_TOKEN_ENCRYPTION_KEY` entry
