package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/pkg/jwt"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

func AuthRequired(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "invalid authorization format")
		}

		claims, err := jwt.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals(KeyUserID, claims.UserID)
		c.Locals(KeyUserEmail, claims.Email)
		c.Locals(KeyUserRole, claims.Role)

		return c.Next()
	}
}
