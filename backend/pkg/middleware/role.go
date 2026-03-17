package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// RequireRole returns a middleware that checks c.Locals("user_role").
// MUST be composed AFTER middleware.Protected (which sets user_role).
// Returns 403 Forbidden if role does not match any of the allowed roles.
func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || role == "" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}
}
