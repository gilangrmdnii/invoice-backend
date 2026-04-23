package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
	"github.com/gilangrmdnii/invoice-backend/pkg/validator"
)

type FinanceReportHandler struct {
	service *service.FinanceReportService
}

func NewFinanceReportHandler(service *service.FinanceReportService) *FinanceReportHandler {
	return &FinanceReportHandler{service: service}
}

func (h *FinanceReportHandler) Get(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("projectId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	rep, err := h.service.Get(c.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "finance report retrieved", rep)
}

func (h *FinanceReportHandler) Upsert(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("projectId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var req request.UpsertFinanceReportRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	rep, err := h.service.Upsert(c.Context(), projectID, userID, &req)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "finance report saved", rep)
}
