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

type BudgetRequestHandler struct {
	budgetRequestService *service.BudgetRequestService
}

func NewBudgetRequestHandler(budgetRequestService *service.BudgetRequestService) *BudgetRequestHandler {
	return &BudgetRequestHandler{budgetRequestService: budgetRequestService}
}

func (h *BudgetRequestHandler) Create(c *fiber.Ctx) error {
	var req request.CreateBudgetRequestRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.budgetRequestService.Create(c.Context(), &req, userID, role)
	if err != nil {
		switch err.Error() {
		case "project not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not a member of this project":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create budget request")
	}

	return response.Success(c, fiber.StatusCreated, "budget request created successfully", result)
}

func (h *BudgetRequestHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	requests, err := h.budgetRequestService.List(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list budget requests")
	}

	return response.Success(c, fiber.StatusOK, "budget requests retrieved successfully", requests)
}

func (h *BudgetRequestHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid budget request id")
	}

	br, err := h.budgetRequestService.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "budget request not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get budget request")
	}

	return response.Success(c, fiber.StatusOK, "budget request retrieved successfully", br)
}

func (h *BudgetRequestHandler) Approve(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid budget request id")
	}

	userID := middleware.GetUserID(c)

	result, err := h.budgetRequestService.Approve(c.Context(), id, userID)
	if err != nil {
		switch err.Error() {
		case "budget request not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "budget request is not pending":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to approve budget request")
	}

	return response.Success(c, fiber.StatusOK, "budget request approved successfully", result)
}

func (h *BudgetRequestHandler) Reject(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid budget request id")
	}

	userID := middleware.GetUserID(c)

	result, err := h.budgetRequestService.Reject(c.Context(), id, userID)
	if err != nil {
		switch err.Error() {
		case "budget request not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "budget request is not pending":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to reject budget request")
	}

	return response.Success(c, fiber.StatusOK, "budget request rejected successfully", result)
}
