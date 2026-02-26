package http

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
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
