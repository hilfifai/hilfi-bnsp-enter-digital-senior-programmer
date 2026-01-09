// internal/handler/role/handler.go
package role

import (
	dto "api-jellyfish/internal/dto/role"
	"api-jellyfish/internal/entity"
	apperrors "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)

	GetRoleByUser(c *gin.Context)
	GetRoleByID(c *gin.Context)
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	GetAllRoles(c *gin.Context)
}

func NewHandler(svc service.IRoleService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IRoleService
}

// GetRoleByUser godoc
// @Summary Get roles by user
// @Description Mengambil roles berdasarkan user ID
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/role/user/{userId} [get]
func (h *Handler) GetRoleByUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		helper.APIBadRequest(c, "invalid user UUID", err)
		return
	}

	roles, err := h.Service.GetRoleByUser(c, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrRoleNotFound):
			helper.APINotFound(c, "Roles not found untuk user ini", err)
		default:
			helper.APIInternalServerError(c, "Failed get data role", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get roles by user", roles, nil)
}

// GetRoleByID godoc
// @Summary Get role by ID
// @Description Mengambil role berdasarkan ID
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/role/{id} [get]
func (h *Handler) GetRoleByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	role, err := h.Service.GetRoleByID(c, id)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrRoleNotFound):
			helper.APINotFound(c, "Role not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get data role", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get role by ID", role, nil)
}

// CreateRole godoc
// @Summary Create role
// @Description Membuat role baru
// @Tags Role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role body role.CreateRoleRequest true "Role"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/role [post]
func (h *Handler) CreateRole(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in CreateRole")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	role := &entity.Role{
		RoleGroupID: uuid.MustParse(req.RoleGroupID),
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	}

	if err := h.Service.CreateRole(c, role, userID); err != nil {
		log.Error().Err(err).Msg("Error creating role")
		switch {
		case errors.Is(err, apperrors.ErrCreateRole):
			helper.APIInternalServerError(c, "Failed create role", err)
		default:
			helper.APIInternalServerError(c, "Failed create role", err)
		}
		return
	}

	helper.APICreateSuccess(c, "Role", role)
}

// UpdateRole godoc
// @Summary Update role
// @Description Mengupdate data role
// @Tags Role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Param role body role.UpdateRoleRequest true "Role"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/role/{id} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in UpdateRole")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in UpdateRole")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	role := &entity.Role{
		ID:          id,
		RoleGroupID: uuid.MustParse(req.RoleGroupID),
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	}

	if err := h.Service.UpdateRole(c, role, userID); err != nil {
		log.Error().Err(err).Msg("Error updating role")
		switch {
		case errors.Is(err, apperrors.ErrRoleNotFound):
			helper.APINotFound(c, "Role not found", err)
		case errors.Is(err, apperrors.ErrUpdateRole):
			helper.APIInternalServerError(c, "Failed update role", err)
		default:
			helper.APIInternalServerError(c, "Failed update role", err)
		}
		return
	}

	helper.APIUpdateSuccess(c, "Role updated successfully", nil)
}

// GetAllRoles godoc
// @Summary Get all roles
// @Description Mengambil semua data role yang aktif
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/role [get]
func (h *Handler) GetAllRoles(c *gin.Context) {
	roles, err := h.Service.GetAllRoles(c)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data role", err)
		return
	}
	helper.APIResponse(c, http.StatusOK, "Success get all roles", roles, nil)
}

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	role := e.Group("/role")
	role.Use(authMiddleware)
	{
		role.GET("/user/:userId", h.GetRoleByUser)
		role.GET("", h.GetAllRoles)
		role.GET("/:id", h.GetRoleByID)
		role.POST("", h.CreateRole)
		role.PUT("/:id", h.UpdateRole)
	}
}
