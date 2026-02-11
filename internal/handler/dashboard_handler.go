package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.dashboardService.GetDashboard(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to get dashboard")
	}

	return response.Success(c, fiber.StatusOK, "dashboard retrieved successfully", result)
}
