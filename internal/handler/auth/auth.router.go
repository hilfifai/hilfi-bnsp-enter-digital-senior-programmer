package auth

import (
	"api-jellyfish/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup) {
	group := e.Group("/auth")

	group.
		POST("/login", h.Login).
		POST("/refresh-token", h.RefreshToken).
		GET("/data", middleware.AuthMiddleware(h.auth), h.UserInfo)
}
