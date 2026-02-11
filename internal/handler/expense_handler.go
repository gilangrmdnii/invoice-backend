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

type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

func NewExpenseHandler(expenseService *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var req request.CreateExpenseRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.expenseService.Create(c.Context(), &req, userID, role)
	if err != nil {
		switch err.Error() {
		case "project not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not a member of this project":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create expense")
	}

	return response.Success(c, fiber.StatusCreated, "expense created successfully", result)
}

func (h *ExpenseHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	expenses, err := h.expenseService.List(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list expenses")
	}

	return response.Success(c, fiber.StatusOK, "expenses retrieved successfully", expenses)
}

func (h *ExpenseHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid expense id")
	}

	expense, err := h.expenseService.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "expense not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get expense")
	}

	return response.Success(c, fiber.StatusOK, "expense retrieved successfully", expense)
}

func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid expense id")
	}

	var req request.UpdateExpenseRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	result, err := h.expenseService.Update(c.Context(), id, &req, userID, role)
	if err != nil {
		switch err.Error() {
		case "expense not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not authorized to update this expense":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		case "only pending expenses can be updated":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update expense")
	}

	return response.Success(c, fiber.StatusOK, "expense updated successfully", result)
}

func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid expense id")
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if err := h.expenseService.Delete(c.Context(), id, userID, role); err != nil {
		switch err.Error() {
		case "expense not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "not authorized to delete this expense":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		case "only pending expenses can be deleted":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete expense")
	}

	return response.Success(c, fiber.StatusOK, "expense deleted successfully", nil)
}

func (h *ExpenseHandler) Approve(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid expense id")
	}

	var req request.ApproveExpenseRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	result, err := h.expenseService.Approve(c.Context(), id, userID, req.Notes)
	if err != nil {
		switch err.Error() {
		case "expense not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "expense is not pending":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to approve expense")
	}

	return response.Success(c, fiber.StatusOK, "expense approved successfully", result)
}

func (h *ExpenseHandler) Reject(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid expense id")
	}

	var req request.ApproveExpenseRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)

	result, err := h.expenseService.Reject(c.Context(), id, userID, req.Notes)
	if err != nil {
		switch err.Error() {
		case "expense not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "expense is not pending":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to reject expense")
	}

	return response.Success(c, fiber.StatusOK, "expense rejected successfully", result)
}
