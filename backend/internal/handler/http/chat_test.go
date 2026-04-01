package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	responsedto "github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChatService struct {
	req requestdto.ChatRequest
	res *responsedto.ChatResponse
	err error
}

func (f *fakeChatService) Respond(_ context.Context, req requestdto.ChatRequest) (*responsedto.ChatResponse, error) {
	f.req = req
	return f.res, f.err
}

func TestChatHandler_CreateMessage_Success(t *testing.T) {
	fake := &fakeChatService{res: &responsedto.ChatResponse{SessionID: "ses-1", Answer: "Ada satu rumah aktif.", Grounding: responsedto.ChatGrounding{ListingSlugs: []string{"rumah-aktif"}}}}
	app := fiber.New()
	app.Post("/api/chat/messages", handler.NewChatHandler(fake).CreateMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/messages", bytes.NewBufferString(`{"session_id":"ses-1","message":"Ada rumah aktif?"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ses-1", fake.req.SessionID)
	assert.Equal(t, "Ada rumah aktif?", fake.req.Message)
}

func TestChatHandler_CreateMessage_InvalidBody(t *testing.T) {
	fake := &fakeChatService{}
	app := fiber.New()
	app.Post("/api/chat/messages", handler.NewChatHandler(fake).CreateMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/messages", bytes.NewBufferString(`{"session_id":"","message":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
