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

type InvoicePaymentHandler struct {
	paymentService *service.InvoicePaymentService
}

func NewInvoicePaymentHandler(paymentService *service.InvoicePaymentService) *InvoicePaymentHandler {
	return &InvoicePaymentHandler{paymentService: paymentService}
}

func (h *InvoicePaymentHandler) Create(c *fiber.Ctx) error {
	var req request.CreateInvoicePaymentRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	result, err := h.paymentService.Create(c.Context(), &req, userID)
	if err != nil {
		switch err.Error() {
		case "invoice not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "invoice must be approved before recording payments":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		case "invoice is already fully paid":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		if len(err.Error()) > 14 && err.Error()[:14] == "payment amount" {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create payment")
	}

	return response.Success(c, fiber.StatusCreated, "payment recorded successfully", result)
}

func (h *InvoicePaymentHandler) ListByInvoice(c *fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("invoiceId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid invoice id")
	}

	payments, err := h.paymentService.ListByInvoice(c.Context(), invoiceID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list payments")
	}

	return response.Success(c, fiber.StatusOK, "payments retrieved successfully", payments)
}

func (h *InvoicePaymentHandler) Delete(c *fiber.Ctx) error {
	paymentID, err := strconv.ParseUint(c.Params("paymentId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid payment id")
	}

	invoiceID, err := strconv.ParseUint(c.Params("invoiceId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid invoice id")
	}

	userID := middleware.GetUserID(c)

	if err := h.paymentService.Delete(c.Context(), paymentID, invoiceID, userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete payment")
	}

	return response.Success(c, fiber.StatusOK, "payment deleted successfully", nil)
}
