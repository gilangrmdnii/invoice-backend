package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

func RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := GetUserRole(c)

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, "insufficient permissions")
	}
}
