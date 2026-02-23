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

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	role := c.Query("role")

	users, err := h.userService.List(c.Context(), role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list users")
	}

	return response.Success(c, fiber.StatusOK, "users retrieved successfully", users)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req request.CreateUserRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	actorID := middleware.GetUserID(c)

	result, err := h.userService.Create(c.Context(), &req, actorID)
	if err != nil {
		switch err.Error() {
		case "email already registered":
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create user")
	}

	return response.Success(c, fiber.StatusCreated, "user created successfully", result)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	var req request.UpdateUserRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	actorID := middleware.GetUserID(c)

	result, err := h.userService.Update(c.Context(), id, &req, actorID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "email already registered":
			return response.Error(c, fiber.StatusConflict, err.Error())
		case "cannot change your own role":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update user")
	}

	return response.Success(c, fiber.StatusOK, "user updated successfully", result)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	actorID := middleware.GetUserID(c)

	if err := h.userService.Delete(c.Context(), id, actorID); err != nil {
		switch err.Error() {
		case "cannot delete yourself":
			return response.Error(c, fiber.StatusForbidden, err.Error())
		case "user not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "cannot delete user with associated data":
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete user")
	}

	return response.Success(c, fiber.StatusOK, "user deleted successfully", nil)
}
