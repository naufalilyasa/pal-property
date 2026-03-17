package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"gorm.io/gorm"
)

// Protected protects routes by checking for a valid access_token cookie
func Protected(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Get access_token from cookie
		tokenString := c.Cookies("access_token")
		if tokenString == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing access token")
		}

		// 2. Validate token and extract User ID
		userID, err := jwt.ValidateAccessToken(tokenString)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired access token: "+err.Error())
		}

		// 3. Fetch user from DB to get role (in production this should be cached)
		var user struct{ Role string }
		if err := db.Table("users").Select("role").Where("id = ?", userID).First(&user).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found")
		}

		// 4. Set User ID and Role in context locals
		c.Locals("user_id", userID)
		c.Locals("user_role", user.Role)

		// 5. Continue to next handler
		return c.Next()
	}
}
