package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

const chatSessionPrefix = "chat:session:"

type chatMemoryRepository struct {
	client   *redis.Client
	ttl      time.Duration
	maxTurns int64
}

func NewChatMemoryRepository(client *redis.Client, ttl time.Duration, maxTurns int) domain.ChatMemoryRepository {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if maxTurns <= 0 {
		maxTurns = 10
	}

	return &chatMemoryRepository{client: client, ttl: ttl, maxTurns: int64(maxTurns)}
}

func (r *chatMemoryRepository) AppendTurn(ctx context.Context, sessionID string, turn domain.ChatTurn) error {
	key, err := r.sessionKey(sessionID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(turn)
	if err != nil {
		return fmt.Errorf("marshal chat turn: %w", err)
	}

	pipe := r.client.TxPipeline()
	pipe.RPush(ctx, key, payload)
	pipe.LTrim(ctx, key, -r.maxTurns, -1)
	pipe.Expire(ctx, key, r.ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("append chat turn: %w", err)
	}

	return nil
}

func (r *chatMemoryRepository) GetTurns(ctx context.Context, sessionID string) ([]domain.ChatTurn, error) {
	key, err := r.sessionKey(sessionID)
	if err != nil {
		return nil, err
	}
	values, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("read chat turns: %w", err)
	}
	if len(values) == 0 {
		return nil, domain.ErrChatMemoryNotFound
	}

	turns := make([]domain.ChatTurn, 0, len(values))
	for _, value := range values {
		var turn domain.ChatTurn
		if err := json.Unmarshal([]byte(value), &turn); err != nil {
			return nil, fmt.Errorf("decode chat turn: %w", err)
		}
		turns = append(turns, turn)
	}

	if err := r.client.Expire(ctx, key, r.ttl).Err(); err != nil {
		return nil, fmt.Errorf("refresh chat ttl: %w", err)
	}

	return turns, nil
}

func (r *chatMemoryRepository) ClearSession(ctx context.Context, sessionID string) error {
	key, err := r.sessionKey(sessionID)
	if err != nil {
		return err
	}
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("clear chat session: %w", err)
	}
	return nil
}

func (r *chatMemoryRepository) sessionKey(sessionID string) (string, error) {
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return "", fmt.Errorf("session id is required")
	}
	return chatSessionPrefix + trimmed, nil
}
