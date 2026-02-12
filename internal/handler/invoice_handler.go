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

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
	var req request.CreateInvoiceRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	result, err := h.invoiceService.Create(c.Context(), &req, userID)
	if err != nil {
		switch err.Error() {
		case "project not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not a member of this project":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create invoice")
	}

	return response.Success(c, fiber.StatusCreated, "invoice created successfully", result)
}

func (h *InvoiceHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	invoices, err := h.invoiceService.List(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list invoices")
	}

	return response.Success(c, fiber.StatusOK, "invoices retrieved successfully", invoices)
}

func (h *InvoiceHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid invoice id")
	}

	invoice, err := h.invoiceService.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "invoice not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get invoice")
	}

	return response.Success(c, fiber.StatusOK, "invoice retrieved successfully", invoice)
}

func (h *InvoiceHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid invoice id")
	}

	var req request.UpdateInvoiceRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	result, err := h.invoiceService.Update(c.Context(), id, &req, userID)
	if err != nil {
		switch err.Error() {
		case "invoice not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not authorized to update this invoice":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update invoice")
	}

	return response.Success(c, fiber.StatusOK, "invoice updated successfully", result)
}

func (h *InvoiceHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid invoice id")
	}

	userID := middleware.GetUserID(c)

	if err := h.invoiceService.Delete(c.Context(), id, userID); err != nil {
		switch err.Error() {
		case "invoice not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not authorized to delete this invoice":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete invoice")
	}

	return response.Success(c, fiber.StatusOK, "invoice deleted successfully", nil)
}
