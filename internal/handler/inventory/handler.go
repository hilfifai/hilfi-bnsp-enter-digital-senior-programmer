// internal/handler/inventory/handler.go
package inventory

import (
	dto "api-jellyfish/internal/dto"
	apperrors "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)

	CreateStockMovement(c *gin.Context)
	TransferStock(c *gin.Context)
	AdjustStock(c *gin.Context)
	PerformStockOpname(c *gin.Context)
	GetStockMovements(c *gin.Context)
	GetStockByProduct(c *gin.Context)
	GetStockByWarehouse(c *gin.Context)
	GetStockSummary(c *gin.Context)
	GetLowStockAlerts(c *gin.Context)
	GetStockHistory(c *gin.Context)
}

func NewHandler(svc service.IInventoryService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IInventoryService
}

// CreateStockMovement godoc
// @Summary Create stock movement
// @Description Create a new stock movement (IN, OUT, TRANSFER, ADJUSTMENT)
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movement body dto.StockMovementRequest true "Stock Movement"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/movements [post]
func (h *Handler) CreateStockMovement(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.StockMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock movement data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	movement, err := h.Service.CreateStockMovement(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create stock movement")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "Warehouse not found", err)
		case errors.Is(err, apperrors.ErrInsufficientStock):
			helper.APIResponse(c, http.StatusConflict, "Insufficient stock",
				map[string]string{"quantity": "Insufficient stock available."}, err)
		case errors.Is(err, apperrors.ErrInvalidMovementType):
			helper.APIBadRequest(c, "Invalid movement type", err)
		default:
			helper.APIInternalServerError(c, "Failed to create stock movement", err)
		}
		return
	}

	log.Info().Str("reference", movement.ReferenceNumber).Msg("Stock movement created successfully")
	helper.APICreateSuccess(c, "Stock movement created successfully", movement)
}

// TransferStock godoc
// @Summary Transfer stock between warehouses
// @Description Transfer stock from one warehouse to another
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transfer body dto.StockTransferRequest true "Stock Transfer"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/transfer [post]
func (h *Handler) TransferStock(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.StockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock transfer data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	movement, err := h.Service.TransferStock(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to transfer stock")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "Warehouse not found", err)
		case errors.Is(err, apperrors.ErrInsufficientStock):
			helper.APIResponse(c, http.StatusConflict, "Insufficient stock",
				map[string]string{"quantity": "Insufficient stock available."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to transfer stock", err)
		}
		return
	}

	log.Info().Str("reference", movement.ReferenceNumber).Msg("Stock transfer completed")
	helper.APICreateSuccess(c, "Stock transfer completed", movement)
}

// AdjustStock godoc
// @Summary Adjust stock quantity
// @Description Adjust stock quantity to a new value
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param adjustment body dto.StockAdjustmentRequest true "Stock Adjustment"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/adjustment [post]
func (h *Handler) AdjustStock(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock adjustment data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	movement, err := h.Service.AdjustStock(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to adjust stock")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "Warehouse not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to adjust stock", err)
		}
		return
	}

	log.Info().Str("reference", movement.ReferenceNumber).Msg("Stock adjustment completed")
	helper.APIUpdateSuccess(c, "Stock adjustment completed", movement)
}

// PerformStockOpname godoc
// @Summary Perform stock opname
// @Description Perform physical stock count and reconciliation
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param opname body dto.StockOpnameRequest true "Stock Opname"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/opname [post]
func (h *Handler) PerformStockOpname(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.StockOpnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock opname data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	movements, err := h.Service.PerformStockOpname(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to perform stock opname")
		helper.APIInternalServerError(c, "Failed to perform stock opname", err)
		return
	}

	log.Info().Int("count", len(movements)).Msg("Stock opname completed")
	helper.APICreateSuccess(c, "Stock opname completed", movements)
}

// GetStockMovements godoc
// @Summary Get stock movements
// @Description Get list of stock movements with filtering
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param product_id query string false "Filter by product ID"
// @Param movement_type query string false "Filter by movement type"
// @Param start_date query string false "Filter by start date"
// @Param end_date query string false "Filter by end date"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/movements [get]
func (h *Handler) GetStockMovements(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.StockReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Error().Err(err).Msg("Failed to bind query parameters")
		helper.APIBindingError(c, err)
		return
	}

	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}

	if err := validation.Validate(filter); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid filter parameters", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	movements, total, err := h.Service.GetStockMovements(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock movements")
		helper.APIInternalServerError(c, "Failed to get stock movements", err)
		return
	}

	response := map[string]interface{}{
		"movements": movements,
		"pagination": map[string]interface{}{
			"page":       filter.Page,
			"limit":      filter.Limit,
			"total":      total,
			"totalPages": (total + filter.Limit - 1) / filter.Limit,
		},
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock movements", response, nil)
}

// GetStockByProduct godoc
// @Summary Get stock by product
// @Description Get stock levels for a specific product across warehouses
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/inventory/product/{id} [get]
func (h *Handler) GetStockByProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	stocks, err := h.Service.GetStockByProduct(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock by product")
		helper.APIInternalServerError(c, "Failed to get stock", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock by product", stocks, nil)
}

// GetStockByWarehouse godoc
// @Summary Get stock by warehouse
// @Description Get stock levels for all products in a specific warehouse
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/inventory/warehouse/{id} [get]
func (h *Handler) GetStockByWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid warehouse ID", err)
		return
	}

	stocks, err := h.Service.GetStockByWarehouse(ctx, warehouseID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock by warehouse")
		helper.APIInternalServerError(c, "Failed to get stock", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock by warehouse", stocks, nil)
}

// GetStockSummary godoc
// @Summary Get stock summary
// @Description Get overall stock summary across all warehouses
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/summary [get]
func (h *Handler) GetStockSummary(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	summary, err := h.Service.GetStockSummary(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock summary")
		helper.APIInternalServerError(c, "Failed to get stock summary", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock summary", summary, nil)
}

// GetLowStockAlerts godoc
// @Summary Get low stock alerts
// @Description Get products with stock below minimum threshold
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/inventory/alerts/low-stock [get]
func (h *Handler) GetLowStockAlerts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	alerts, err := h.Service.GetLowStockAlerts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get low stock alerts")
		helper.APIInternalServerError(c, "Failed to get low stock alerts", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get low stock alerts", alerts, nil)
}

// GetStockHistory godoc
// @Summary Get stock history
// @Description Get stock movement history for a product
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product UUID"
// @Param warehouse_id query string false "Warehouse UUID"
// @Param days query int false "Number of days to look back" default(30)
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Router /api/v1/inventory/history/{product_id} [get]
func (h *Handler) GetStockHistory(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	warehouseID := uuid.Nil
	if warehouseStr := c.Query("warehouse_id"); warehouseStr != "" {
		warehouseID, err = uuid.Parse(warehouseStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid warehouse UUID")
			helper.APIBadRequest(c, "Invalid warehouse ID", err)
			return
		}
	}

	days := 30
	if daysStr := c.DefaultQuery("days", "30"); daysStr != "" {
		var daysVal int
		daysVal, err = parseInt(daysStr)
		if err != nil || daysVal < 1 {
			daysVal = 30
		}
		days = daysVal
	}

	history, err := h.Service.GetStockHistory(ctx, productID, warehouseID, days)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock history")
		helper.APIInternalServerError(c, "Failed to get stock history", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock history", history, nil)
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
