package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
)

type RegionHandler struct {
	service service.RegionLookupService
}

func NewRegionHandler(service service.RegionLookupService) *RegionHandler {
	return &RegionHandler{service: service}
}

func (h *RegionHandler) ListProvinces(c fiber.Ctx) error {
	regions, err := h.service.ListProvinces(c.Context())
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, regions)
}

func (h *RegionHandler) ListCities(c fiber.Ctx) error {
	regions, err := h.service.ListCities(c.Context(), c.Query("province_code"))
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, regions)
}

func (h *RegionHandler) ListDistricts(c fiber.Ctx) error {
	regions, err := h.service.ListDistricts(c.Context(), c.Query("city_code"))
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, regions)
}

func (h *RegionHandler) ListVillages(c fiber.Ctx) error {
	regions, err := h.service.ListVillages(c.Context(), c.Query("district_code"))
	if err != nil {
		return err
	}
	return utils.SendResponse(c, fiber.StatusOK, regions)
}
