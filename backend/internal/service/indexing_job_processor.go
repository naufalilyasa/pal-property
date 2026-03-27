package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"go.uber.org/zap"
)

type IndexingJobProcessor interface {
	ProcessBatch(ctx context.Context, limit int) error
	Run(ctx context.Context, interval time.Duration, batchSize int) error
}

type indexingJobProcessor struct {
	jobs      domain.SearchIndexJobRepository
	projector domain.SearchProjector
	maxTries  int
}

func NewIndexingJobProcessor(jobs domain.SearchIndexJobRepository, projector domain.SearchProjector) (IndexingJobProcessor, error) {
	if jobs == nil {
		return nil, fmt.Errorf("indexing job processor: repository is required")
	}
	if projector == nil {
		return nil, fmt.Errorf("indexing job processor: projector is required")
	}
	return &indexingJobProcessor{jobs: jobs, projector: projector, maxTries: 5}, nil
}

func (p *indexingJobProcessor) ProcessBatch(ctx context.Context, limit int) error {
	now := time.Now().UTC()
	jobs, err := p.jobs.ClaimPending(ctx, now, limit)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := p.processJob(ctx, job, now); err != nil {
			return err
		}
	}
	return nil
}

func (p *indexingJobProcessor) Run(ctx context.Context, interval time.Duration, batchSize int) error {
	if interval <= 0 {
		interval = 3 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := p.ProcessBatch(ctx, batchSize); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			logger.Log.Warn("Indexing job batch failed", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *indexingJobProcessor) processJob(ctx context.Context, job *entity.SearchIndexJob, now time.Time) error {
	var payload domain.SearchIndexJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		nextAvailableAt := nextRetryTime(now, job.AttemptCount)
		markTerminal := job.AttemptCount+1 >= p.maxTries
		return p.jobs.MarkFailed(ctx, job.ID, nextAvailableAt, fmt.Sprintf("decode job payload: %v", err), markTerminal)
	}

	var err error
	switch payload.Event.AggregateType {
	case domain.AggregateTypeListing:
		err = p.projector.HandleListingEvent(ctx, listingEventFromJob(job, payload))
	case domain.AggregateTypeCategory:
		err = p.projector.HandleCategoryEvent(ctx, categoryEventFromJob(job, payload))
	default:
		err = fmt.Errorf("unsupported aggregate type: %s", payload.Event.AggregateType)
	}
	if err != nil {
		nextAvailableAt := nextRetryTime(now, job.AttemptCount)
		markTerminal := job.AttemptCount+1 >= p.maxTries
		return p.jobs.MarkFailed(ctx, job.ID, nextAvailableAt, err.Error(), markTerminal)
	}
	return p.jobs.MarkDone(ctx, job.ID, now)
}

func listingEventFromJob(job *entity.SearchIndexJob, payload domain.SearchIndexJobPayload) domain.ListingEvent {
	var eventPayload domain.ListingEventPayload
	_ = json.Unmarshal(payload.Event.Payload, &eventPayload)
	return domain.ListingEvent{
		Metadata: domain.EventMetadata{EventID: job.ID, EventType: payload.Event.EventType, AggregateType: payload.Event.AggregateType, AggregateID: payload.Event.AggregateID, Version: 1, OccurredAt: job.CreatedAt},
		Payload:  eventPayload,
	}
}

func categoryEventFromJob(job *entity.SearchIndexJob, payload domain.SearchIndexJobPayload) domain.CategoryEvent {
	var eventPayload domain.CategoryEventPayload
	_ = json.Unmarshal(payload.Event.Payload, &eventPayload)
	return domain.CategoryEvent{
		Metadata: domain.EventMetadata{EventID: job.ID, EventType: payload.Event.EventType, AggregateType: payload.Event.AggregateType, AggregateID: payload.Event.AggregateID, Version: 1, OccurredAt: job.CreatedAt},
		Payload:  eventPayload,
	}
}

func nextRetryTime(now time.Time, attemptCount int) time.Time {
	return now.Add(time.Duration(attemptCount+1) * time.Minute)
}
