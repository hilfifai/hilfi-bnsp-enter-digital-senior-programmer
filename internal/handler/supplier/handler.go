// internal/handler/supplier/handler.go
package supplier

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

	CreateSupplier(c *gin.Context)
	GetSupplierByID(c *gin.Context)
	GetSuppliers(c *gin.Context)
	UpdateSupplier(c *gin.Context)
	DeleteSupplier(c *gin.Context)
}

func NewHandler(svc service.ISupplierService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.ISupplierService
}

// CreateSupplier godoc
// @Summary Create supplier
// @Description Membuat supplier baru
// @Tags Supplier
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param supplier body dto.CreateSupplierRequest true "Supplier"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/supplier [post]
func (h *Handler) CreateSupplier(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	supplier, err := h.Service.CreateSupplier(c, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating supplier")
		switch {
		case errors.Is(err, apperrors.ErrSupplierCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Supplier is already in use.",
				map[string]string{"code": "Code Supplier is already in use."}, err)
		default:
			helper.APIInternalServerError(c, "Failed create supplier", err)
		}
		return
	}

	helper.APICreateSuccess(c, "Supplier", supplier)
}

// GetSupplierByID godoc
// @Summary Get supplier by ID
// @Description Mendapatkan supplier berdasarkan ID
// @Tags Supplier
// @Produce json
// @Security BearerAuth
// @Param id path string true "Supplier UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/supplier/{id} [get]
func (h *Handler) GetSupplierByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	supplier, err := h.Service.GetSupplierByID(c, id)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data supplier", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get supplier by ID", supplier, nil)
}

// GetSuppliers godoc
// @Summary Get all suppliers
// @Description Mendapatkan semua supplier
// @Tags Supplier
// @Produce json
// @Security BearerAuth
// @Param active_only query bool false "Active only"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/supplier [get]
func (h *Handler) GetSuppliers(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"

	suppliers, err := h.Service.GetSuppliers(c, activeOnly)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data supplier", err)
		return
	}
	helper.APIResponse(c, http.StatusOK, "Success get all suppliers", suppliers, nil)
}

// UpdateSupplier godoc
// @Summary Update supplier
// @Description Mengupdate data supplier
// @Tags Supplier
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Supplier UUID"
// @Param supplier body dto.UpdateSupplierRequest true "Supplier"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/supplier/{id} [put]
func (h *Handler) UpdateSupplier(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in UpdateSupplier")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	var req dto.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error in UpdateSupplier")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	updatedSupplier, err := h.Service.UpdateSupplier(c, id, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error updating supplier")
		switch {
		case errors.Is(err, apperrors.ErrSupplierCodeExists):
			helper.APIResponse(c, http.StatusBadRequest, "Validation Failed: Code Supplier is already in use.",
				map[string]string{"code": "Code Supplier is already in use."}, err)
		case errors.Is(err, apperrors.ErrSupplierNotFound):
			helper.APINotFound(c, "supplier not found", err)
		default:
			helper.APIInternalServerError(c, "Failed update supplier", err)
		}
		return
	}

	helper.APIUpdateSuccess(c, "Supplier updated successfully", updatedSupplier)
}

// DeleteSupplier godoc
// @Summary Delete supplier
// @Description Menghapus supplier berdasarkan ID
// @Tags Supplier
// @Produce json
// @Security BearerAuth
// @Param id path string true "Supplier UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/supplier/{id} [delete]
func (h *Handler) DeleteSupplier(c *gin.Context) {
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	if err := h.Service.DeleteSupplier(c, id, userID); err != nil {
		helper.APIInternalServerError(c, "Failed delete supplier", err)
		return
	}

	helper.APIDeleteSuccess(c, "Supplier deleted successfully")
}
