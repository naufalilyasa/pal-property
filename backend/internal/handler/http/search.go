package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type SearchHandler struct {
	svc service.SearchReadService
}

func NewSearchHandler(svc service.SearchReadService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

func (h *SearchHandler) SearchListings(c fiber.Ctx) error {
	req := request.SearchListingsRequest{
		Query:            c.Query("q"),
		TransactionType:  c.Query("transaction_type"),
		LocationProvince: c.Query("location_province"),
		LocationCity:     c.Query("location_city"),
		Sort:             c.Query("sort"),
	}
	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil {
		req.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit", "20")); err == nil {
		req.Limit = limit
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		uid, err := uuid.Parse(categoryID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
		}
		req.CategoryID = &uid
	}
	if min := c.Query("price_min"); min != "" {
		val, err := strconv.ParseInt(min, 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid price_min")
		}
		req.PriceMin = &val
	}
	if max := c.Query("price_max"); max != "" {
		val, err := strconv.ParseInt(max, 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid price_max")
		}
		req.PriceMax = &val
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error())
	}
	res, err := h.svc.SearchListings(c.Context(), req)
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, res)
}
