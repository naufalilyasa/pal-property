package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigHonorsDBSSLMode(t *testing.T) {
	t.Cleanup(func() { Env = AppConfig{} })

	envVars := map[string]string{
		"DB_HOST":                    "db.local",
		"DB_USER":                    "user",
		"DB_PASSWORD":                "password",
		"DB_NAME":                    "pal",
		"REDIS_ADDR":                 "redis:6379",
		"OAUTH_CLIENT_ID":            "client-id",
		"OAUTH_CLIENT_SECRET":        "secret",
		"OAUTH_CALLBACK_URL":         "https://example.com/oauth",
		"JWT_PRIVATE_KEY_BASE64":     "cHJpdmF0ZQ==",
		"JWT_PUBLIC_KEY_BASE64":      "cHVibGlj",
		"OAUTH_TOKEN_ENCRYPTION_KEY": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}

	for key, value := range envVars {
		t.Setenv(key, value)
	}

	t.Setenv("DB_SSL_MODE", "require-full")

	require.NoError(t, LoadConfig())
	require.Equal(t, "require-full", Env.DBSSLMode)
}
