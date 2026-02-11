package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func ParseAndValidate(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return fmt.Errorf("invalid request body")
	}

	if err := validate.Struct(out); err != nil {
		return formatValidationErrors(err)
	}

	return nil
}

func formatValidationErrors(err error) error {
	var messages []string

	for _, e := range err.(validator.ValidationErrors) {
		field := toSnakeCase(e.Field())
		switch e.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "email":
			messages = append(messages, fmt.Sprintf("%s must be a valid email", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		case "gt":
			messages = append(messages, fmt.Sprintf("%s must be greater than %s", field, e.Param()))
		case "oneof":
			messages = append(messages, fmt.Sprintf("%s must be one of: %s", field, e.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", field))
		}
	}

	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
