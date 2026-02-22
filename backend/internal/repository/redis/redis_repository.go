package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/username/pal-property-backend/internal/domain"
)

type cacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) domain.CacheRepository {
	return &cacheRepository{client: client}
}

func (r *cacheRepository) SaveRefreshTokenJTI(ctx context.Context, jti string, userID uuid.UUID, expiration time.Duration) error {
	// Store the jti as key, and user ID as value
	return r.client.Set(ctx, "refresh_token:"+jti, userID.String(), expiration).Err()
}
