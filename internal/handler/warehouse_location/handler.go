// internal/handler/warehouse_location/handler.go
package warehouse_location

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

	CreateLocation(c *gin.Context)
	UpdateLocation(c *gin.Context)
	DeleteLocation(c *gin.Context)
	GetLocationsByWarehouse(c *gin.Context)
	GetLocationByID(c *gin.Context)
	GetAvailableLocations(c *gin.Context)
}

func NewHandler(svc service.IWarehouseLocationService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IWarehouseLocationService
}

// CreateLocation godoc
// @Summary Create warehouse location
// @Description Membuat warehouse location baru
// @Tags Warehouse Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location body dto.CreateWarehouseLocationRequest true "Warehouse Location"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/location [post]
func (h *Handler) CreateLocation(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateWarehouseLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in CreateLocation")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	location, err := h.Service.CreateLocation(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating location")
		switch {
		case errors.Is(err, apperrors.ErrLocationCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Location is already in use.",
				map[string]string{"code": "Code Location is already in use."}, err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		case errors.Is(err, apperrors.ErrCreateLocation):
			helper.APIInternalServerError(c, "Failed create location", err)
		default:
			helper.APIInternalServerError(c, "Failed create location", err)
		}
		return
	}

	helper.APICreateSuccess(c, "Warehouse Location", location)
}

// UpdateLocation godoc
// @Summary Update warehouse location
// @Description Mengupdate data warehouse location
// @Tags Warehouse Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location UUID"
// @Param location body dto.UpdateWarehouseLocationRequest true "Warehouse Location"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/location/{id} [put]
func (h *Handler) UpdateLocation(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in UpdateLocation")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	var req dto.UpdateWarehouseLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in UpdateLocation")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	updatedLocation, err := h.Service.UpdateLocation(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error updating location")
		switch {
		case errors.Is(err, apperrors.ErrLocationCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Location is already in use.",
				map[string]string{"code": "Code Location is already in use."}, err)
		case errors.Is(err, apperrors.ErrLocationNotFound):
			helper.APINotFound(c, "location not found", err)
		default:
			helper.APIInternalServerError(c, "Failed update location", err)
		}
		return
	}

	helper.APIUpdateSuccess(c, "Warehouse Location updated successfully", updatedLocation)
}

// DeleteLocation godoc
// @Summary Delete warehouse location
// @Description Menghapus warehouse location berdasarkan ID
// @Tags Warehouse Location
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/location/{id} [delete]
func (h *Handler) DeleteLocation(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in DeleteLocation")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	if err := h.Service.DeleteLocation(ctx, id, userID); err != nil {
		log.Error().Err(err).Msg("Error deleting location")
		switch {
		case errors.Is(err, apperrors.ErrLocationNotFound):
			helper.APINotFound(c, "location not found", err)
		default:
			helper.APIInternalServerError(c, "Failed delete location", err)
		}
		return
	}

	helper.APIDeleteSuccess(c, "Warehouse Location deleted successfully")
}

// GetLocationsByWarehouse godoc
// @Summary Get locations by warehouse
// @Description get semua data location berdasarkan warehouse ID
// @Tags Warehouse Location
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/location [get]
func (h *Handler) GetLocationsByWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		helper.APIBadRequest(c, "warehouse_id is required", fmt.Errorf("warehouse_id is required"))
		return
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in GetLocationsByWarehouse")
		helper.APIBadRequest(c, "invalid warehouse_id UUID", err)
		return
	}

	locations, err := h.Service.GetLocationsByWarehouse(ctx, warehouseID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting locations by warehouse")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get data locations", err)
		}
		return
	}
	helper.APIResponse(c, http.StatusOK, "Success get locations by warehouse", locations, nil)
}

// GetLocationByID godoc
// @Summary Get location by ID
// @Description get location berdasarkan ID
// @Tags Warehouse Location
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/location/{id} [get]
func (h *Handler) GetLocationByID(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in GetLocationByID")
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	location, err := h.Service.GetLocationByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Error getting location by ID")
		switch {
		case errors.Is(err, apperrors.ErrLocationNotFound):
			helper.APINotFound(c, "location not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get location", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get location by ID", location, nil)
}

// GetAvailableLocations godoc
// @Summary Get available locations
// @Description get locations yang tersedia dengan kapasitas yang cukup
// @Tags Warehouse Location
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string true "Warehouse UUID"
// @Param required_capacity query int true "Required Capacity"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/location/available [get]
func (h *Handler) GetAvailableLocations(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		helper.APIBadRequest(c, "warehouse_id is required", fmt.Errorf("warehouse_id is required"))
		return
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in GetAvailableLocations")
		helper.APIBadRequest(c, "invalid warehouse_id UUID", err)
		return
	}

	requiredCapacity := c.Query("required_capacity")
	if requiredCapacity == "" {
		helper.APIBadRequest(c, "required_capacity is required", fmt.Errorf("required_capacity is required"))
		return
	}

	var capacity int
	if _, err := fmt.Sscanf(requiredCapacity, "%d", &capacity); err != nil {
		helper.APIBadRequest(c, "invalid required_capacity", err)
		return
	}

	locations, err := h.Service.GetAvailableLocations(ctx, warehouseID, capacity)
	if err != nil {
		log.Error().Err(err).Msg("Error getting available locations")
		switch {
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed get available locations", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get available locations", locations, nil)
}
