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

type QCReportHandler struct {
	service *service.QCReportService
}

func NewQCReportHandler(service *service.QCReportService) *QCReportHandler {
	return &QCReportHandler{service: service}
}

func (h *QCReportHandler) Create(c *fiber.Ctx) error {
	var req request.CreateQCReportRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	rep, err := h.service.Create(c.Context(), &req, userID, role)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not a member of this project" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "qc report created", rep)
}

func (h *QCReportHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var projectID *uint64
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err == nil {
			projectID = &pid
		}
	}

	reports, err := h.service.List(c.Context(), userID, role, projectID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc reports retrieved", reports)
}

func (h *QCReportHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	rep, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "qc report not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc report retrieved", rep)
}

func (h *QCReportHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	var req request.UpdateQCReportRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	rep, err := h.service.Update(c.Context(), id, &req, userID, role)
	if err != nil {
		if err.Error() == "qc report not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not authorized to update this report" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc report updated", rep)
}

func (h *QCReportHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if err := h.service.Delete(c.Context(), id, userID, role); err != nil {
		if err.Error() == "qc report not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not authorized to delete this report" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc report deleted", nil)
}
