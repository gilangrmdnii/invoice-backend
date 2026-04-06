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

type ProjectWorkerHandler struct {
	service *service.ProjectWorkerService
}

func NewProjectWorkerHandler(service *service.ProjectWorkerService) *ProjectWorkerHandler {
	return &ProjectWorkerHandler{service: service}
}

func (h *ProjectWorkerHandler) Create(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("projectId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var req request.CreateProjectWorkerRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	req.ProjectID = projectID

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	worker, err := h.service.Create(c.Context(), &req, userID, role)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not a member of this project" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "worker added", worker)
}

func (h *ProjectWorkerHandler) ListByProject(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("projectId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	workers, err := h.service.ListByProject(c.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "workers retrieved", workers)
}

func (h *ProjectWorkerHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	worker, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "worker not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "worker retrieved", worker)
}

func (h *ProjectWorkerHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	var req request.UpdateProjectWorkerRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	worker, err := h.service.Update(c.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "worker not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "worker updated", worker)
}

func (h *ProjectWorkerHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	userID := middleware.GetUserID(c)

	if err := h.service.Delete(c.Context(), id, userID); err != nil {
		if err.Error() == "worker not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "worker deleted", nil)
}
