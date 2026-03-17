package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
)

type ListingHandler struct {
	svc service.ListingService
}

func NewListingHandler(svc service.ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

func (h *ListingHandler) Create(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req request.CreateListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	res, err := h.svc.Create(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusCreated, res)
}

func (h *ListingHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	res, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) GetBySlug(c fiber.Ctx) error {
	slugStr := c.Params("slug")
	if slugStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}

	res, err := h.svc.GetBySlug(c.Context(), slugStr)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	userRole := "user"
	if role, ok := c.Locals("user_role").(string); ok {
		userRole = role
	}

	var req request.UpdateListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	res, err := h.svc.Update(c.Context(), id, userID, userRole, &req)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	userRole := "user"
	if role, ok := c.Locals("user_role").(string); ok {
		userRole = role
	}

	err = h.svc.Delete(c.Context(), id, userID, userRole)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, fiber.Map{"message": "listing deleted successfully"})
}

func (h *ListingHandler) UploadImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	res, err := h.svc.UploadImage(c.Context(), id, userID, h.userRole(c), file)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) DeleteImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid image id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	res, err := h.svc.DeleteImage(c.Context(), id, imageID, userID, h.userRole(c))
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) SetPrimaryImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid image id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	res, err := h.svc.SetPrimaryImage(c.Context(), id, imageID, userID, h.userRole(c))
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) ReorderImages(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req request.ReorderListingImagesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	res, err := h.svc.ReorderImages(c.Context(), id, userID, h.userRole(c), req.OrderedImageIDs)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) List(c fiber.Ctx) error {
	filter := h.parseFilter(c)
	res, err := h.svc.List(c.Context(), filter)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) ListByUserID(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	filter := h.parseFilter(c)
	res, err := h.svc.ListByUserID(c.Context(), userID, filter)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *ListingHandler) userRole(c fiber.Ctx) string {
	userRole := "user"
	if role, ok := c.Locals("user_role").(string); ok {
		userRole = role
	}

	return userRole
}

func (h *ListingHandler) parseFilter(c fiber.Ctx) domain.ListingFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filter := domain.ListingFilter{
		Status:       c.Query("status"),
		LocationCity: c.Query("city"),
		Page:         page,
		Limit:        limit,
	}

	if catID := c.Query("category_id"); catID != "" {
		if uid, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &uid
		}
	}

	if pMin := c.Query("price_min"); pMin != "" {
		if val, err := strconv.ParseInt(pMin, 10, 64); err == nil {
			filter.PriceMin = &val
		}
	}

	if pMax := c.Query("price_max"); pMax != "" {
		if val, err := strconv.ParseInt(pMax, 10, 64); err == nil {
			filter.PriceMax = &val
		}
	}

	return filter
}
