package config

import (
	"encoding/base64"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// AppConfig holds all application configuration loaded from environment variables.
type AppConfig struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	Port    string `env:"PORT"    envDefault:"8080"`
	AppName string `env:"APP_NAME" envDefault:"pal-property"`

	// Database
	DBHost     string `env:"DB_HOST"     validate:"required"`
	DBUser     string `env:"DB_USER"     validate:"required"`
	DBPassword string `env:"DB_PASSWORD" validate:"required"`
	DBName     string `env:"DB_NAME"     validate:"required"`
	DBPort     string `env:"DB_PORT"     envDefault:"5432"`
	DBSSLMode  string `env:"DB_SSLMODE"    envDefault:"disable"`

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
	JwtPrivateKeyBase64  string `env:"JWT_PRIVATE_KEY_BASE64" validate:"required"`
	JwtPublicKeyBase64   string `env:"JWT_PUBLIC_KEY_BASE64"  validate:"required"`
	JwtAccessExpiration  int    `env:"JWT_ACCESS_EXPIRATION"  envDefault:"900"`    // seconds
	JwtRefreshExpiration int    `env:"JWT_REFRESH_EXPIRATION" envDefault:"604800"` // seconds

	// Encryption key for OAuth provider tokens stored in DB (AES-256 = 32 bytes, base64-encoded)
	OAuthTokenEncryptionKeyBase64 string `env:"OAUTH_TOKEN_ENCRYPTION_KEY" validate:"required"`

	// Cloudinary
	CloudinaryEnabled   bool   `env:"CLOUDINARY_ENABLED"    envDefault:"false"`
	CloudinaryCloudName string `env:"CLOUDINARY_CLOUD_NAME"`
	CloudinaryAPIKey    string `env:"CLOUDINARY_API_KEY"`
	CloudinaryAPISecret string `env:"CLOUDINARY_API_SECRET"`

	// Parsed (not from env directly — populated in LoadConfig)
	JwtPrivateKeyPEM        []byte `env:"-"`
	JwtPublicKeyPEM         []byte `env:"-"`
	OAuthTokenEncryptionKey []byte `env:"-"` // decoded 32-byte key
}

// Env is the global config singleton populated by LoadConfig.
var Env AppConfig

// LoadConfig parses environment variables into AppConfig, validates required fields,
// and decodes base64 PEM keys.
func LoadConfig() error {
	_ = godotenv.Load() // Load .env file if it exists, ignore error; does NOT override container env vars

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

	if err := validateCloudinaryConfig(cfg); err != nil {
		return err
	}

	Env = cfg
	return nil
}

func validateCloudinaryConfig(cfg AppConfig) error {
	provided := 0
	for _, value := range []string{cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret} {
		if value != "" {
			provided++
		}
	}

	if !cfg.CloudinaryEnabled {
		if provided > 0 && provided < 3 {
			return fmt.Errorf("config: CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, and CLOUDINARY_API_SECRET must all be set together")
		}
		return nil
	}

	if provided < 3 {
		return fmt.Errorf("config: CLOUDINARY_ENABLED requires CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, and CLOUDINARY_API_SECRET")
	}

	return nil
}
