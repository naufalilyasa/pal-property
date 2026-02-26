package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
)

// Protected protects routes by checking for a valid access_token cookie
func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Get access_token from cookie
		tokenString := c.Cookies("access_token")
		if tokenString == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing access token")
		}

		// 2. Validate token and extract User ID
		userID, err := jwt.ValidateAccessToken(tokenString)
		if err != nil {
			// Frontend will receive 401 and should attempt to call /auth/refresh
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired access token: "+err.Error())
		}

		// 3. Set User ID in context locals
		c.Locals("user_id", userID)

		// 4. Continue to next handler
		return c.Next()
	}
}
