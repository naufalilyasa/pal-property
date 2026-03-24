package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigHonorsDBSSLMode(t *testing.T) {
	t.Cleanup(func() { Env = AppConfig{} })

	setRequiredConfigEnv(t)

	t.Setenv("DB_SSL_MODE", "require-full")

	require.NoError(t, LoadConfig())
	require.Equal(t, "require-full", Env.DBSSLMode)
}

func TestLoadConfigParsesEventingAndSearchConfig(t *testing.T) {
	t.Cleanup(func() { Env = AppConfig{} })
	setRequiredConfigEnv(t)

	t.Setenv("KAFKA_BROKERS", "redpanda:9092,redpanda:19092")
	t.Setenv("KAFKA_GROUP_ID", "listing-indexer")
	t.Setenv("KAFKA_CLIENT_ID", "property-service")
	t.Setenv("KAFKA_TOPIC_LISTING_EVENTS", "listing.events.v1")
	t.Setenv("KAFKA_TOPIC_CATEGORY_EVENTS", "category.events.v1")
	t.Setenv("ELASTIC_ADDRESS", "http://elasticsearch:9200")
	t.Setenv("ELASTIC_USERNAME", "elastic")
	t.Setenv("ELASTIC_PASSWORD", "password")
	t.Setenv("ELASTIC_INDEX_LISTINGS", "listings_v1")

	require.NoError(t, LoadConfig())
	require.Equal(t, []string{"redpanda:9092", "redpanda:19092"}, Env.KafkaBrokers)
	require.Equal(t, "listing-indexer", Env.KafkaGroupID)
	require.Equal(t, "property-service", Env.KafkaClientID)
	require.Equal(t, "listing.events.v1", Env.KafkaTopicListingEvents)
	require.Equal(t, "category.events.v1", Env.KafkaTopicCategoryEvents)
	require.Equal(t, "http://elasticsearch:9200", Env.ElasticAddress)
	require.Equal(t, "elastic", Env.ElasticUsername)
	require.Equal(t, "password", Env.ElasticPassword)
	require.Equal(t, "listings_v1", Env.ElasticListingsIndex)
}

func TestLoadConfigRejectsPartialElasticCredentials(t *testing.T) {
	t.Cleanup(func() { Env = AppConfig{} })
	setRequiredConfigEnv(t)

	t.Setenv("ELASTIC_USERNAME", "elastic")

	err := LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "ELASTIC_USERNAME and ELASTIC_PASSWORD must be set together")
}

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()

	for key, value := range map[string]string{
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
	} {
		t.Setenv(key, value)
	}
}
