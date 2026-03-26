package eventing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	listingWriter  *kafka.Writer
	categoryWriter *kafka.Writer
}

func NewKafkaPublisher(brokers []string, clientID, listingTopic, categoryTopic string) (*KafkaPublisher, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("eventing: at least one broker is required")
	}
	if listingTopic == "" || categoryTopic == "" {
		return nil, fmt.Errorf("eventing: listing and category topics are required")
	}
	baseTransport := &kafka.Transport{ClientID: clientID}
	return &KafkaPublisher{
		listingWriter:  &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: listingTopic, Balancer: &kafka.LeastBytes{}, Transport: baseTransport},
		categoryWriter: &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: categoryTopic, Balancer: &kafka.LeastBytes{}, Transport: baseTransport},
	}, nil
}

func (p *KafkaPublisher) PublishListingEvent(ctx context.Context, event domain.ListingEvent) error {
	if p == nil || p.listingWriter == nil {
		return nil
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventing: marshal listing event: %w", err)
	}
	return p.listingWriter.WriteMessages(ctx, kafka.Message{Key: []byte(event.Metadata.AggregateID.String()), Value: payload, Headers: []kafka.Header{{Key: "event_id", Value: []byte(event.Metadata.EventID.String())}, {Key: "event_type", Value: []byte(event.Metadata.EventType)}}})
}

func (p *KafkaPublisher) PublishCategoryEvent(ctx context.Context, event domain.CategoryEvent) error {
	if p == nil || p.categoryWriter == nil {
		return nil
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventing: marshal category event: %w", err)
	}
	return p.categoryWriter.WriteMessages(ctx, kafka.Message{Key: []byte(event.Metadata.AggregateID.String()), Value: payload, Headers: []kafka.Header{{Key: "event_id", Value: []byte(event.Metadata.EventID.String())}, {Key: "event_type", Value: []byte(event.Metadata.EventType)}}})
}

func (p *KafkaPublisher) Close() error {
	if p == nil {
		return nil
	}
	var err error
	if p.listingWriter != nil {
		if closeErr := p.listingWriter.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if p.categoryWriter != nil {
		if closeErr := p.categoryWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func newEventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
