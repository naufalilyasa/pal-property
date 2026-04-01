package domain

import "context"

type ChatTurn struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type ChatMemoryRepository interface {
	AppendTurn(ctx context.Context, sessionID string, turn ChatTurn) error
	GetTurns(ctx context.Context, sessionID string) ([]ChatTurn, error)
	ClearSession(ctx context.Context, sessionID string) error
}
