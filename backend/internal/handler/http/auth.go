package http

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/username/pal-property-backend/internal/dto/response"
	"github.com/username/pal-property-backend/internal/service"
	"github.com/username/pal-property-backend/pkg/utils"
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

	resp := response.AuthResponse{
		User: response.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
		Token: "dummy-jwt-token",
	}

	return utils.SendResponse(c, fiber.StatusOK, resp)
}
