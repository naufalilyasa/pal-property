package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type ListingHandler struct {
	svc service.ListingService
}

func NewListingHandler(svc service.ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

func (h *ListingHandler) Create(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	var req request.CreateListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error())
	}

	res, err := h.svc.Create(c.Context(), principal, &req)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	var req request.UpdateListingRequest
	if err = c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error())
	}

	res, err := h.svc.Update(c.Context(), id, principal, &req)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	err = h.svc.Delete(c.Context(), id, principal)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	res, err := h.svc.UploadImage(c.Context(), id, principal, file)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	res, err := h.svc.DeleteImage(c.Context(), id, imageID, principal)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	res, err := h.svc.SetPrimaryImage(c.Context(), id, imageID, principal)
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

	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	var req request.ReorderListingImagesRequest
	if err = c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}

	res, err := h.svc.ReorderImages(c.Context(), id, principal, req.OrderedImageIDs)
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
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	filter := h.parseFilter(c)
	res, err := h.svc.ListByUserID(c.Context(), principal, filter)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
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
		Status:           c.Query("status"),
		TransactionType:  c.Query("transaction_type"),
		LocationProvince: c.Query("location_province"),
		LocationCity:     firstQueryValue(c, "location_city", "city"),
		CertificateType:  c.Query("certificate_type"),
		Condition:        c.Query("condition"),
		Furnishing:       c.Query("furnishing"),
		Page:             page,
		Limit:            limit,
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

	if bedrooms := c.Query("bedroom_count"); bedrooms != "" {
		if val, err := strconv.Atoi(bedrooms); err == nil {
			filter.BedroomCount = &val
		}
	}

	if bathrooms := c.Query("bathroom_count"); bathrooms != "" {
		if val, err := strconv.Atoi(bathrooms); err == nil {
			filter.BathroomCount = &val
		}
	}

	if landMin := c.Query("land_area_min"); landMin != "" {
		if val, err := strconv.Atoi(landMin); err == nil {
			filter.LandAreaMin = &val
		}
	}

	if landMax := c.Query("land_area_max"); landMax != "" {
		if val, err := strconv.Atoi(landMax); err == nil {
			filter.LandAreaMax = &val
		}
	}

	if buildingMin := c.Query("building_area_min"); buildingMin != "" {
		if val, err := strconv.Atoi(buildingMin); err == nil {
			filter.BuildingAreaMin = &val
		}
	}

	if buildingMax := c.Query("building_area_max"); buildingMax != "" {
		if val, err := strconv.Atoi(buildingMax); err == nil {
			filter.BuildingAreaMax = &val
		}
	}

	return filter
}

func firstQueryValue(c fiber.Ctx, keys ...string) string {
	for _, key := range keys {
		if value := c.Query(key); value != "" {
			return value
		}
	}
	return ""
}
