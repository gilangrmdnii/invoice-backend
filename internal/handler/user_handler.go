package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	role := c.Query("role")

	var roles []string
	if role != "" {
		roles = []string{role}
	} else {
		roles = []string{"SPV", "FINANCE", "OWNER"}
	}

	users, err := h.userRepo.FindByRoles(c.Context(), roles)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list users")
	}

	return response.Success(c, fiber.StatusOK, "users retrieved successfully", users)
}
