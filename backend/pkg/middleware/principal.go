package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

func CurrentPrincipal(c fiber.Ctx) (authz.Principal, error) {
	if principal, ok := c.Locals(authz.PrincipalContextKey).(authz.Principal); ok {
		return principal, nil
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return authz.Principal{}, fmt.Errorf("authz: missing user_id local")
	}

	role, ok := c.Locals("user_role").(string)
	if !ok {
		return authz.Principal{}, fmt.Errorf("authz: missing user_role local")
	}

	return authz.NewPrincipal(userID, role)
}
