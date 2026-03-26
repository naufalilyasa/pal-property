package eventing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/segmentio/kafka-go"
)

type messageReader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
	Close() error
}

type KafkaConsumer struct {
	listingReader  messageReader
	categoryReader messageReader
	projector      domain.SearchProjector
}

func NewKafkaConsumer(brokers []string, groupID, listingTopic, categoryTopic string, projector domain.SearchProjector) (*KafkaConsumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("eventing: at least one broker is required")
	}
	if groupID == "" {
		return nil, fmt.Errorf("eventing: consumer group id is required")
	}
	if listingTopic == "" || categoryTopic == "" {
		return nil, fmt.Errorf("eventing: listing and category topics are required")
	}
	if projector == nil {
		return nil, fmt.Errorf("eventing: search projector is required")
	}
	return newKafkaConsumerWithReaders(
		kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: groupID, Topic: listingTopic}),
		kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: groupID, Topic: categoryTopic}),
		projector,
	), nil
}

func newKafkaConsumerWithReaders(listingReader, categoryReader messageReader, projector domain.SearchProjector) *KafkaConsumer {
	return &KafkaConsumer{listingReader: listingReader, categoryReader: categoryReader, projector: projector}
}

func (c *KafkaConsumer) Consume(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("eventing: consumer is nil")
	}
	errCh := make(chan error, 2)
	doneCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := c.consumeListingEvents(ctx); err != nil && !isContextDone(err) {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := c.consumeCategoryEvents(ctx); err != nil && !isContextDone(err) {
			errCh <- err
		}
	}()
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	select {
	case err := <-errCh:
		return err
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return nil
	}
}

func (c *KafkaConsumer) consumeListingEvents(ctx context.Context) error {
	for {
		msg, err := c.listingReader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var event domain.ListingEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("eventing: unmarshal listing event: %w", err)
		}
		if err := c.projector.HandleListingEvent(ctx, event); err != nil {
			return fmt.Errorf("eventing: handle listing event: %w", err)
		}
		if err := c.listingReader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("eventing: commit listing event: %w", err)
		}
	}
}

func (c *KafkaConsumer) consumeCategoryEvents(ctx context.Context) error {
	for {
		msg, err := c.categoryReader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var event domain.CategoryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("eventing: unmarshal category event: %w", err)
		}
		if err := c.projector.HandleCategoryEvent(ctx, event); err != nil {
			return fmt.Errorf("eventing: handle category event: %w", err)
		}
		if err := c.categoryReader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("eventing: commit category event: %w", err)
		}
	}
}

func (c *KafkaConsumer) Close() error {
	if c == nil {
		return nil
	}
	var err error
	if c.listingReader != nil {
		if closeErr := c.listingReader.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if c.categoryReader != nil {
		if closeErr := c.categoryReader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func isContextDone(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}
