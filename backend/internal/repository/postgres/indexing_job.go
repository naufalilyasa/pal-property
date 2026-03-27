package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type searchIndexJobRepository struct {
	db *gorm.DB
}

func NewSearchIndexJobRepository(db *gorm.DB) domain.SearchIndexJobRepository {
	return &searchIndexJobRepository{db: db}
}

func (r *searchIndexJobRepository) Enqueue(ctx context.Context, job *entity.SearchIndexJob) (*entity.SearchIndexJob, error) {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("enqueue search index job: %w", err)
	}
	return job, nil
}

func (r *searchIndexJobRepository) ClaimPending(ctx context.Context, now time.Time, limit int) ([]*entity.SearchIndexJob, error) {
	if limit <= 0 {
		limit = 100
	}
	var jobs []*entity.SearchIndexJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND available_at <= ?", domain.SearchIndexJobStatusPending, now).
			Order("created_at ASC").
			Limit(limit).
			Find(&jobs).Error; err != nil {
			return fmt.Errorf("claim pending search index jobs: %w", err)
		}
		if len(jobs) == 0 {
			return nil
		}
		ids := make([]uuid.UUID, 0, len(jobs))
		for _, job := range jobs {
			ids = append(ids, job.ID)
		}
		if err := tx.Model(&entity.SearchIndexJob{}).
			Where("id IN ?", ids).
			Updates(map[string]any{"status": domain.SearchIndexJobStatusProcessing, "updated_at": now}).Error; err != nil {
			return fmt.Errorf("mark search index jobs processing: %w", err)
		}
		for _, job := range jobs {
			job.Status = domain.SearchIndexJobStatusProcessing
			job.UpdatedAt = now
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *searchIndexJobRepository) MarkDone(ctx context.Context, id uuid.UUID, processedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&entity.SearchIndexJob{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": domain.SearchIndexJobStatusDone, "processed_at": processedAt, "updated_at": processedAt}).Error; err != nil {
		return fmt.Errorf("mark search index job done: %w", err)
	}
	return nil
}

func (r *searchIndexJobRepository) MarkFailed(ctx context.Context, id uuid.UUID, nextAvailableAt time.Time, lastError string, markTerminal bool) error {
	status := domain.SearchIndexJobStatusPending
	if markTerminal {
		status = domain.SearchIndexJobStatusFailed
	}
	if err := r.db.WithContext(ctx).Model(&entity.SearchIndexJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"available_at":  nextAvailableAt,
			"last_error":    lastError,
			"updated_at":    nextAvailableAt,
		}).Error; err != nil {
		return fmt.Errorf("mark search index job failed: %w", err)
	}
	return nil
}
