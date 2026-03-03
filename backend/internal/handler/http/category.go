package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
)

type CategoryHandler struct {
	svc service.CategoryService
}

func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c fiber.Ctx) error {
	res, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, res)
}


func (h *CategoryHandler) GetBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}

	res, err := h.svc.GetBySlug(c.Context(), slug)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, res)
}


func (h *CategoryHandler) Create(c fiber.Ctx) error {
	var req request.CreateCategoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusCreated, res)
}



func (h *CategoryHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
	}

	var req request.UpdateCategoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	res, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, res)
}


func (h *CategoryHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
	}

	err = h.svc.Delete(c.Context(), id)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, fiber.Map{"message": "category deleted successfully"})
}

