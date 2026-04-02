package redis

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestChatMemoryAppendAndReadTurns(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redisv9.NewClient(&redisv9.Options{Addr: mr.Addr()})
	repo := NewChatMemoryRepository(client, 15*time.Minute, 10)

	ctx := context.Background()
	require.NoError(t, repo.AppendTurn(ctx, "session-1", domain.ChatTurn{Role: "user", Message: "Halo"}))
	require.NoError(t, repo.AppendTurn(ctx, "session-1", domain.ChatTurn{Role: "assistant", Message: "Ada yang bisa saya bantu?"}))

	turns, err := repo.GetTurns(ctx, "session-1")
	require.NoError(t, err)
	require.Len(t, turns, 2)
	require.Equal(t, "user", turns[0].Role)
	require.Equal(t, "assistant", turns[1].Role)
}

func TestChatMemoryTrimsToMaxTurns(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redisv9.NewClient(&redisv9.Options{Addr: mr.Addr()})
	repo := NewChatMemoryRepository(client, 15*time.Minute, 2)

	ctx := context.Background()
	require.NoError(t, repo.AppendTurn(ctx, "session-2", domain.ChatTurn{Role: "user", Message: "1"}))
	require.NoError(t, repo.AppendTurn(ctx, "session-2", domain.ChatTurn{Role: "assistant", Message: "2"}))
	require.NoError(t, repo.AppendTurn(ctx, "session-2", domain.ChatTurn{Role: "user", Message: "3"}))

	turns, err := repo.GetTurns(ctx, "session-2")
	require.NoError(t, err)
	require.Len(t, turns, 2)
	require.Equal(t, "2", turns[0].Message)
	require.Equal(t, "3", turns[1].Message)
}

func TestChatMemoryReturnsNotFoundAfterExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redisv9.NewClient(&redisv9.Options{Addr: mr.Addr()})
	repo := NewChatMemoryRepository(client, 2*time.Second, 10)

	ctx := context.Background()
	require.NoError(t, repo.AppendTurn(ctx, "session-3", domain.ChatTurn{Role: "user", Message: "Halo"}))
	mr.FastForward(3 * time.Second)

	turns, err := repo.GetTurns(ctx, "session-3")
	require.ErrorIs(t, err, domain.ErrChatMemoryNotFound)
	require.Nil(t, turns)
}

func TestChatMemoryClearSession(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redisv9.NewClient(&redisv9.Options{Addr: mr.Addr()})
	repo := NewChatMemoryRepository(client, 15*time.Minute, 10)

	ctx := context.Background()
	require.NoError(t, repo.AppendTurn(ctx, "session-4", domain.ChatTurn{Role: "user", Message: "Halo"}))
	require.NoError(t, repo.ClearSession(ctx, "session-4"))

	turns, err := repo.GetTurns(ctx, "session-4")
	require.ErrorIs(t, err, domain.ErrChatMemoryNotFound)
	require.Nil(t, turns)
}
