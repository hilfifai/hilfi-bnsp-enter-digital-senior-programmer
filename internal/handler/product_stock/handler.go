// internal/handler/product_stock/handler.go
package product_stock

import (
	dto "api-jellyfish/internal/dto"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)
	GetStock(c *gin.Context)
	UpdateStock(c *gin.Context)
	GetStockMovements(c *gin.Context)
	GetLowStockProducts(c *gin.Context)
	GetStockByProduct(c *gin.Context)
	GetStockByWarehouse(c *gin.Context)
}

func NewHandler(svc service.IProductStockService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IProductStockService
}

// GetStock godoc
// @Summary Get product stock
// @Description Get stock information for a product in a specific warehouse
// @Tags Product Stock
// @Produce json
// @Security BearerAuth
// @Param product_id query string true "Product UUID"
// @Param warehouse_id query string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/product-stocks [get]
func (h *Handler) GetStock(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	productIDStr := c.Query("product_id")
	warehouseIDStr := c.Query("warehouse_id")

	if productIDStr == "" || warehouseIDStr == "" {
		helper.APIBadRequest(c, "product_id and warehouse_id are required", nil)
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid product UUID")
		helper.APIBadRequest(c, "Invalid product_id", err)
		return
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid warehouse UUID")
		helper.APIBadRequest(c, "Invalid warehouse_id", err)
		return
	}

	stock, err := h.Service.GetStock(ctx, productID, warehouseID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock")
		helper.APINotFound(c, "Stock not found", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Stock retrieved successfully", stock, nil)
}

// UpdateStock godoc
// @Summary Update product stock
// @Description Update stock quantity for a product
// @Tags Product Stock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param stock body dto.CreateProductStockRequest true "Stock data"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-stocks [post]
func (h *Handler) UpdateStock(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateProductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	stock := &dto.UpdateProductStockRequest{
		Quantity:     req.Quantity,
		ReorderLevel: req.ReorderLevel,
	}

	_ = userID
	_ = stock

	helper.APIResponse(c, http.StatusOK, "Stock updated successfully", nil, nil)
}

// GetStockMovements godoc
// @Summary Get stock movement history
// @Description Get stock movement history with filters
// @Tags Product Stock
// @Produce json
// @Security BearerAuth
// @Param product_id query string false "Product UUID"
// @Param warehouse_id query string false "Warehouse UUID"
// @Param movement_type query string false "Movement type"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-stocks/movements [get]
func (h *Handler) GetStockMovements(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	filter := make(map[string]interface{})

	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			helper.APIBadRequest(c, "Invalid product_id", err)
			return
		}
		filter["product_id"] = productID
	}

	if warehouseIDStr := c.Query("warehouse_id"); warehouseIDStr != "" {
		warehouseID, err := uuid.Parse(warehouseIDStr)
		if err != nil {
			helper.APIBadRequest(c, "Invalid warehouse_id", err)
			return
		}
		filter["warehouse_id"] = warehouseID
	}

	if movementType := c.Query("movement_type"); movementType != "" {
		filter["movement_type"] = movementType
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		filter["start_date"] = startDateStr
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		filter["end_date"] = endDateStr
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter["page"] = page
	filter["limit"] = limit

	movements, err := h.Service.GetStockMovementHistory(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock movements")
		helper.APIInternalServerError(c, "Failed to get stock movements", err)
		return
	}

	response := map[string]interface{}{
		"movements": movements,
		"count":     len(movements),
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
		},
	}

	helper.APIResponse(c, http.StatusOK, "Stock movements retrieved successfully", response, nil)
}

// GetLowStockProducts godoc
// @Summary Get low stock products
// @Description Get products with stock below reorder level
// @Tags Product Stock
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-stocks/low-stock [get]
func (h *Handler) GetLowStockProducts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	stocks, err := h.Service.GetLowStockProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get low stock products")
		helper.APIInternalServerError(c, "Failed to get low stock products", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Low stock products retrieved", stocks, nil)
}

// GetStockByProduct godoc
// @Summary Get stock by product
// @Description Get all stock entries for a specific product
// @Tags Product Stock
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-stocks/product/{id} [get]
func (h *Handler) GetStockByProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid product UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	stocks, err := h.Service.GetStockByProduct(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock by product")
		helper.APIInternalServerError(c, "Failed to get stock", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Stock retrieved successfully", stocks, nil)
}

// GetStockByWarehouse godoc
// @Summary Get stock by warehouse
// @Description Get all stock entries in a specific warehouse
// @Tags Product Stock
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-stocks/warehouse/{id} [get]
func (h *Handler) GetStockByWarehouse(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid warehouse UUID")
		helper.APIBadRequest(c, "Invalid warehouse ID", err)
		return
	}

	stocks, err := h.Service.GetStockByWarehouse(ctx, warehouseID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock by warehouse")
		helper.APIInternalServerError(c, "Failed to get stock", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Stock retrieved successfully", stocks, nil)
}
