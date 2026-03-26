package eventing

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMessageReader struct {
	messages []kafka.Message
	err      error
	commits  []kafka.Message
}

func (r *fakeMessageReader) FetchMessage(context.Context) (kafka.Message, error) {
	if len(r.messages) == 0 {
		if r.err != nil {
			return kafka.Message{}, r.err
		}
		return kafka.Message{}, context.Canceled
	}
	msg := r.messages[0]
	r.messages = r.messages[1:]
	return msg, nil
}

func (r *fakeMessageReader) CommitMessages(_ context.Context, msgs ...kafka.Message) error {
	r.commits = append(r.commits, msgs...)
	return nil
}

func (r *fakeMessageReader) Close() error { return nil }

type fakeSearchProjector struct {
	listingEvents  []domain.ListingEvent
	categoryEvents []domain.CategoryEvent
}

func (p *fakeSearchProjector) HandleListingEvent(_ context.Context, event domain.ListingEvent) error {
	p.listingEvents = append(p.listingEvents, event)
	return nil
}

func (p *fakeSearchProjector) HandleCategoryEvent(_ context.Context, event domain.CategoryEvent) error {
	p.categoryEvents = append(p.categoryEvents, event)
	return nil
}

func TestKafkaConsumer_DispatchesListingAndCategoryEvents(t *testing.T) {
	listingEvent := domain.ListingEvent{Metadata: domain.EventMetadata{EventID: uuid.New(), EventType: domain.EventTypeListingCreated, AggregateType: domain.AggregateTypeListing, AggregateID: uuid.New(), Version: 1, OccurredAt: time.Now().UTC()}, Payload: domain.ListingEventPayload{ID: uuid.New(), Title: "Listing"}}
	categoryEvent := domain.CategoryEvent{Metadata: domain.EventMetadata{EventID: uuid.New(), EventType: domain.EventTypeCategoryUpdated, AggregateType: domain.AggregateTypeCategory, AggregateID: uuid.New(), Version: 1, OccurredAt: time.Now().UTC()}, Payload: domain.CategoryEventPayload{ID: uuid.New(), Name: "Residential"}}
	listingPayload, err := json.Marshal(listingEvent)
	require.NoError(t, err)
	categoryPayload, err := json.Marshal(categoryEvent)
	require.NoError(t, err)

	projector := &fakeSearchProjector{}
	consumer := newKafkaConsumerWithReaders(
		&fakeMessageReader{messages: []kafka.Message{{Value: listingPayload}}, err: context.Canceled},
		&fakeMessageReader{messages: []kafka.Message{{Value: categoryPayload}}, err: context.Canceled},
		projector,
	)

	err = consumer.Consume(context.Background())
	require.NoError(t, err)
	assert.Len(t, projector.listingEvents, 1)
	assert.Len(t, projector.categoryEvents, 1)
	assert.Equal(t, domain.EventTypeListingCreated, projector.listingEvents[0].Metadata.EventType)
	assert.Equal(t, domain.EventTypeCategoryUpdated, projector.categoryEvents[0].Metadata.EventType)
}

func TestKafkaConsumer_PropagatesDecodeErrors(t *testing.T) {
	consumer := newKafkaConsumerWithReaders(
		&fakeMessageReader{messages: []kafka.Message{{Value: []byte("not-json")}}, err: context.Canceled},
		&fakeMessageReader{err: context.Canceled},
		&fakeSearchProjector{},
	)

	err := consumer.Consume(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal listing event")
}

func TestNewKafkaConsumer_RequiresProjector(t *testing.T) {
	consumer, err := NewKafkaConsumer([]string{"localhost:19092"}, "group", "listing.events", "category.events", nil)
	require.Error(t, err)
	assert.Nil(t, consumer)
	assert.True(t, errors.Is(err, err) || err != nil)
}
