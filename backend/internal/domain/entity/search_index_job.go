package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SearchIndexJob struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	AggregateType string         `gorm:"type:varchar(32);not null;index:idx_search_index_jobs_aggregate" json:"aggregate_type"`
	AggregateID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_search_index_jobs_aggregate" json:"aggregate_id"`
	EventType     string         `gorm:"type:varchar(64);not null" json:"event_type"`
	Payload       datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`
	Status        string         `gorm:"type:varchar(20);not null;default:'pending';index:idx_search_index_jobs_status_available,priority:1" json:"status"`
	AttemptCount  int            `gorm:"not null;default:0" json:"attempt_count"`
	AvailableAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_search_index_jobs_status_available,priority:2" json:"available_at"`
	LastError     *string        `gorm:"type:text" json:"last_error"`
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	ProcessedAt   *time.Time     `json:"processed_at"`
}

func (j *SearchIndexJob) BeforeCreate(_ *gorm.DB) error {
	if j.ID != uuid.Nil {
		return nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	j.ID = id
	return nil
}
