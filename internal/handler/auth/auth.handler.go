package auth

import (
	dto "api-jellyfish/internal/dto/auth"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/validation"
	service "api-jellyfish/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService service.IAuthService
	auth        jwt.IJWTAuth
}

type IHandler interface {
	NewRoutes(e *gin.RouterGroup)
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	UserInfo(c *gin.Context)
}

func NewHandler(authService service.IAuthService, auth jwt.IJWTAuth) IHandler {
	return &Handler{authService: authService, auth: auth}
}

// @Summary Auth
// @Description Get Access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} types.ResponseAPI
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var loginRequest dto.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		helper.APIValidationError(c, "invalid request", err)
		return
	}

	if err := validation.Validate(loginRequest); err != nil {
		helper.APIValidationError(c, "invalid request", err)
		return
	}

	token, err := h.authService.Login(ctx, loginRequest)
	if err != nil {
		helper.APIResponse(c, http.StatusBadRequest, "Invalid username or password", nil, err)
		return
	}
	helper.APIResponse(c, http.StatusOK, "success", token, nil)
}

// @Summary Refresh Token
// @Description Get Refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} types.ResponseAPI
// @Router /api/v1/auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.APIValidationError(c, "invalid request", err)
		return
	}
	if err := validation.Validate(request); err != nil {
		helper.APIValidationError(c, "invalid request", err)
		return
	}
	token, err := h.authService.RefreshToken(ctx, request.RefreshToken)
	if err != nil {
		helper.APIResponse(c, http.StatusBadRequest, "invalid refresh token", nil, err)
		return
	}
	helper.APIResponse(c, http.StatusOK, "success", token, nil)
}

// @Summary User Info
// @Description Get User Info
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} types.ResponseAPI
// @Router /api/v1/auth/data [get]
// @Security BearerAuth
func (h *Handler) UserInfo(c *gin.Context) {
	user := jwt.GetUser(c)
	helper.APIResponse(c, http.StatusOK, "success", user, nil)
}
