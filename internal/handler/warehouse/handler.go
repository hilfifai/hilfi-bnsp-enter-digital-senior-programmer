// internal/handler/warehouse/handler.go
package warehouse

import (
	dto "api-jellyfish/internal/dto"
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

	CreateWarehouse(c *gin.Context)
	UpdateWarehouse(c *gin.Context)
	DeleteWarehouse(c *gin.Context)
	GetAllWarehouses(c *gin.Context)
	GetWarehouseByID(c *gin.Context)
	GetWarehouseStats(c *gin.Context)
}

func NewHandler(svc service.IWarehouseService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IWarehouseService
}

// CreateWarehouse godoc
// @Summary Create warehouse
// @Description Membuat warehouse baru
// @Tags Warehouse
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param warehouse body dto.CreateWarehouseRequest true "Warehouse"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/warehouse [post]
func (h *Handler) CreateWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in CreateWarehouse")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	warehouse, err := h.Service.CreateWarehouse(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating warehouse")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Warehouse is already in use.",
				map[string]string{"code": "Code Warehouse is already in use."}, err)
		case errors.Is(err, apperrors.ErrCreateWarehouse):
			helper.APIInternalServerError(c, "Failed create warehouse", err)
		default:
			helper.APIInternalServerError(c, "Failed create warehouse", err)
		}
		return
	}

	helper.APICreateSuccess(c, "Warehouse", warehouse)
}

// UpdateWarehouse godoc
// @Summary Update warehouse
// @Description Mengupdate data warehouse
// @Tags Warehouse
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Param warehouse body dto.UpdateWarehouseRequest true "Warehouse"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/warehouse/{id} [put]
func (h *Handler) UpdateWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in UpdateWarehouse")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	var req dto.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in UpdateWarehouse")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	updatedWarehouse, err := h.Service.UpdateWarehouse(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error updating warehouse")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Warehouse is already in use.",
				map[string]string{"code": "Code Warehouse is already in use."}, err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed update warehouse", err)
		}
		return
	}

	helper.APIUpdateSuccess(c, "Warehouse updated successfully", updatedWarehouse)
}

// DeleteWarehouse godoc
// @Summary Delete warehouse
// @Description Menghapus warehouse berdasarkan ID
// @Tags Warehouse
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/warehouse/{id} [delete]
func (h *Handler) DeleteWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in DeleteWarehouse")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	if err := h.Service.DeleteWarehouse(ctx, id, userID); err != nil {
		log.Error().Err(err).Msg("Error deleting warehouse")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed delete warehouse", err)
		}
		return
	}

	helper.APIDeleteSuccess(c, "Warehouse deleted successfully")
}

// GetAllWarehouses godoc
// @Summary Get all warehouses
// @Description get semua data warehouse
// @Tags Warehouse
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/warehouse [get]
func (h *Handler) GetAllWarehouses(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	warehouses, err := h.Service.GetAllWarehouses(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting all warehouses")
		helper.APIInternalServerError(c, "Failed get data warehouse", err)
		return
	}
	helper.APIResponse(c, http.StatusOK, "Success get all warehouses", warehouses, nil)
}

// GetWarehouseByID godoc
// @Summary Get warehouse by ID
// @Description get warehouse berdasarkan ID
// @Tags Warehouse
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/warehouse/{id} [get]
func (h *Handler) GetWarehouseByID(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in GetWarehouseByID")
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	warehouse, err := h.Service.GetWarehouseByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Error getting warehouse by ID")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get data warehouse", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get warehouse by ID", warehouse, nil)
}

// GetWarehouseStats godoc
// @Summary Get warehouse stats
// @Description get statistik warehouse
// @Tags Warehouse
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/warehouse/{id}/stats [get]
func (h *Handler) GetWarehouseStats(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in GetWarehouseStats")
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	stats, err := h.Service.GetWarehouseStats(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Error getting warehouse stats")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get warehouse stats", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get warehouse stats", stats, nil)
}
