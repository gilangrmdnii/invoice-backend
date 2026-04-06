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

type QCDocumentHandler struct {
	service *service.QCDocumentService
}

func NewQCDocumentHandler(service *service.QCDocumentService) *QCDocumentHandler {
	return &QCDocumentHandler{service: service}
}

func (h *QCDocumentHandler) Create(c *fiber.Ctx) error {
	var req request.CreateQCDocumentRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	doc, err := h.service.Create(c.Context(), &req, userID, role)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not a member of this project" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "qc document created", doc)
}

func (h *QCDocumentHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var projectID *uint64
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err == nil {
			projectID = &pid
		}
	}

	docs, err := h.service.List(c.Context(), userID, role, projectID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc documents retrieved", docs)
}

func (h *QCDocumentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	doc, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "qc document not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc document retrieved", doc)
}

func (h *QCDocumentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	var req request.UpdateQCDocumentRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	doc, err := h.service.Update(c.Context(), id, &req, userID, role)
	if err != nil {
		if err.Error() == "qc document not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not authorized to update this document" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc document updated", doc)
}

func (h *QCDocumentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if err := h.service.Delete(c.Context(), id, userID, role); err != nil {
		if err.Error() == "qc document not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		if err.Error() == "not authorized to delete this document" {
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "qc document deleted", nil)
}
