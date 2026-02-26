package http

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) BeginAuth(c fiber.Ctx) error {
	provider := c.Params("provider")
	if provider == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider is required")
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()
		gothic.BeginAuthHandler(w, r)
	})

	return adaptor.HTTPHandler(handler)(c)
}

func (h *AuthHandler) Callback(c fiber.Ctx) error {
	provider := c.Params("provider")
	if provider == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider is required")
	}

	var gothUser goth.User
	var errAuth error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothUser, errAuth = gothic.CompleteUserAuth(w, r)
	})

	if err := adaptor.HTTPHandler(handler)(c); err != nil {
		return err
	}
	if errAuth != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errAuth.Error())
	}

	user, err := h.service.CompleteAuth(c.Context(), provider, gothUser)
	if err != nil {
		return err // Bubbled up to global Fiber error handler
	}

	tokens, err := h.service.LoginUser(c.Context(), user)
	if err != nil {
		return err // Bubbled up to global Fiber error handler
	}

	isSecure := config.Env.AppEnv == "production"

	// Set Access Token Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(config.Env.JwtAccessExpiration.Seconds()),
	})

	// Set Refresh Token Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(config.Env.JwtRefreshExpiration.Seconds()),
	})

	// Redirect back to frontend
	return c.Redirect().To("http://localhost:3000/dashboard")
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	// Extract user ID from locals (set by middleware)
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user id type in context")
	}

	// Call service
	userResponse, err := h.service.GetMe(c.Context(), userID)
	if err != nil {
		return err // Let global error handler handle it
	}

	return utils.SendResponse(c, fiber.StatusOK, userResponse)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	// 1. Get refresh token from cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing refresh token")
	}

	// 2. Call service to validate and rotate tokens
	tokens, err := h.service.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		// Clear cookies if refresh fails
		c.ClearCookie("access_token", "refresh_token")
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	isSecure := config.Env.AppEnv == "production"

	// 3. Set new Access Token Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(config.Env.JwtAccessExpiration.Seconds()),
	})

	// 4. Set new Refresh Token Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(config.Env.JwtRefreshExpiration.Seconds()),
	})

	return utils.SendResponse(c, fiber.StatusOK, fiber.Map{"message": "token refreshed successfully"})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		// Invalidate session in Redis
		_ = h.service.Logout(c.Context(), refreshToken)
	}

	// Clear cookies
	c.ClearCookie("access_token", "refresh_token")

	return utils.SendResponse(c, fiber.StatusOK, fiber.Map{"message": "logged out successfully"})
}
