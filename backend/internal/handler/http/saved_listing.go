package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type SavedListingHandler struct {
	svc service.SavedListingService
}

func NewSavedListingHandler(svc service.SavedListingService) *SavedListingHandler {
	return &SavedListingHandler{svc: svc}
}

func (h *SavedListingHandler) List(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	page, limit := parsePagination(c.Query("page"), c.Query("limit"))
	filter := domain.SavedListingFilter{
		Page:  page,
		Limit: limit,
	}

	res, err := h.svc.ListByUserID(c.Context(), principal, filter)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, res)
}

func (h *SavedListingHandler) Contains(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	idsParam := c.Query("listing_ids")
	listingIDs, err := parseListingIDs(idsParam)
	if err != nil {
		return err
	}

	if len(listingIDs) > service.SavedListingContainsLimit {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("contains: at most %d listing ids are allowed", service.SavedListingContainsLimit))
	}

	res, err := h.svc.Contains(c.Context(), principal, listingIDs)
	if err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, savedListingContainsResponse{ListingIDs: res})
}

func (h *SavedListingHandler) Save(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	var req createSavedListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error())
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error())
	}

	if err := h.svc.Save(c.Context(), principal, req.ListingID); err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusCreated, savedListingToggleResponse{ListingID: req.ListingID, Saved: true})
}

func (h *SavedListingHandler) Remove(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	listingID, err := uuid.Parse(c.Params("listingId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
	}

	if err := h.svc.Remove(c.Context(), principal, listingID); err != nil {
		return err
	}

	return utils.SendResponse(c, fiber.StatusOK, savedListingToggleResponse{ListingID: listingID, Saved: false})
}

type createSavedListingRequest struct {
	ListingID uuid.UUID `json:"listing_id" validate:"required"`
}

type savedListingToggleResponse struct {
	ListingID uuid.UUID `json:"listing_id"`
	Saved     bool      `json:"saved"`
}

type savedListingContainsResponse struct {
	ListingIDs []uuid.UUID `json:"listing_ids"`
}

func parsePagination(pageParam, limitParam string) (int, int) {
	page := 1
	if pageParam != "" {
		if value, err := strconv.Atoi(pageParam); err == nil && value > 0 {
			page = value
		}
	}
	limit := 20
	if limitParam != "" {
		if value, err := strconv.Atoi(limitParam); err == nil && value > 0 {
			limit = value
		}
	}
	return page, limit
}

func parseListingIDs(param string) ([]uuid.UUID, error) {
	if strings.TrimSpace(param) == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "listing_ids is required")
	}

	parts := strings.Split(param, ",")
	ids := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		id, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid listing id in listing_ids")
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "listing_ids is required")
	}
	return ids, nil
}
