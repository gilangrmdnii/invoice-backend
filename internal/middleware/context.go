package middleware

import "github.com/gofiber/fiber/v2"

const (
	KeyUserID    = "userID"
	KeyUserEmail = "userEmail"
	KeyUserRole  = "userRole"
)

func GetUserID(c *fiber.Ctx) uint64 {
	val, _ := c.Locals(KeyUserID).(uint64)
	return val
}

func GetUserEmail(c *fiber.Ctx) string {
	val, _ := c.Locals(KeyUserEmail).(string)
	return val
}

func GetUserRole(c *fiber.Ctx) string {
	val, _ := c.Locals(KeyUserRole).(string)
	return val
}
