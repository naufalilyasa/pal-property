package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CacheRepository interface {
	SaveRefreshTokenJTI(ctx context.Context, jti string, userID uuid.UUID, expiration time.Duration) error
}
