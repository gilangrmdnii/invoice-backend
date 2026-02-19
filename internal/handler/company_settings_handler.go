package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
	"github.com/gilangrmdnii/invoice-backend/pkg/validator"
)

type CompanySettingsHandler struct {
	service *service.CompanySettingsService
}

func NewCompanySettingsHandler(service *service.CompanySettingsService) *CompanySettingsHandler {
	return &CompanySettingsHandler{service: service}
}

func (h *CompanySettingsHandler) Get(c *fiber.Ctx) error {
	result, err := h.service.Get(c.Context())
	if err != nil {
		if err.Error() == "company settings not found" {
			return response.Success(c, fiber.StatusOK, "company settings not configured", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get company settings")
	}

	return response.Success(c, fiber.StatusOK, "company settings retrieved successfully", result)
}

func (h *CompanySettingsHandler) Upsert(c *fiber.Ctx) error {
	var req request.UpsertCompanySettingsRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.Upsert(c.Context(), &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to save company settings")
	}

	return response.Success(c, fiber.StatusOK, "company settings saved successfully", result)
}
