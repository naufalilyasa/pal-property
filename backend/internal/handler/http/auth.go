package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/username/pal-property-backend/internal/dto/response"
	"github.com/username/pal-property-backend/internal/service"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) BeginAuth(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	// Let Goth handle redirect
	h.service.BeginAuth(c, provider)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	user, err := h.service.CompleteAuth(c, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create simplified response
	resp := response.AuthResponse{
		User: response.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
		Token: "dummy-jwt-token", // In real app, generate JWT here
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
