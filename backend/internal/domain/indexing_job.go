package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/datatypes"
)

const (
	SearchIndexJobStatusPending    = "pending"
	SearchIndexJobStatusProcessing = "processing"
	SearchIndexJobStatusDone       = "done"
	SearchIndexJobStatusFailed     = "failed"
)

type SearchIndexJobPayload struct {
	Event SearchIndexEventPayload `json:"event"`
}

type SearchIndexEventPayload struct {
	AggregateType string         `json:"aggregate_type"`
	AggregateID   uuid.UUID      `json:"aggregate_id"`
	EventType     string         `json:"event_type"`
	Payload       datatypes.JSON `json:"payload"`
}

type SearchIndexJobRepository interface {
	Enqueue(ctx context.Context, job *entity.SearchIndexJob) (*entity.SearchIndexJob, error)
	ClaimPending(ctx context.Context, now time.Time, limit int) ([]*entity.SearchIndexJob, error)
	MarkDone(ctx context.Context, id uuid.UUID, processedAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, nextAvailableAt time.Time, lastError string, markTerminal bool) error
}
