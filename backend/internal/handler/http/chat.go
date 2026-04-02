package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type ChatHandler struct {
	svc service.ChatService
}

func NewChatHandler(svc service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) CreateMessage(c fiber.Ctx) error {
	var req request.ChatRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error())
	}

	res, err := h.svc.Respond(c.Context(), req)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, res)
}
