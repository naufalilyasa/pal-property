package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

func RequirePermission(authzService *authz.Service, resource string, action string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if authzService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "authorization service unavailable")
		}

		principal, err := CurrentPrincipal(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		allowed, err := authzService.Enforce(authz.Request{
			Principal: principal,
			Resource:  resource,
			Action:    action,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}

		if !allowed {
			return domain.ErrForbidden
		}

		return c.Next()
	}
}

func IsPermissionDenied(err error) bool {
	return errors.Is(err, domain.ErrForbidden)
}
