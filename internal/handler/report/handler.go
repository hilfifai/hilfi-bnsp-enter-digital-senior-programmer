// internal/handler/report/handler.go
package report

import (
	dto "api-jellyfish/internal/dto"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)

	GetStockReport(c *gin.Context)
	GetSalesReport(c *gin.Context)
	GetInventoryValuation(c *gin.Context)
	GetProfitLossReport(c *gin.Context)
	GetTopProducts(c *gin.Context)
	GetLowStockReport(c *gin.Context)
	GetMovementReport(c *gin.Context)
}

func NewHandler(svc service.IReportService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IReportService
}

// GetStockReport godoc
// @Summary Get stock report
// @Description Get stock report with filtering options
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param category_id query string false "Filter by category ID"
// @Param low_stock query bool false "Filter by low stock only"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/stock [get]
func (h *Handler) GetStockReport(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.StockReportFilter

	// Handle warehouse_id manually to support both single UUID and array format
	if warehouseIDStr := c.Query("warehouse_id"); warehouseIDStr != "" {
		// Remove [] brackets if present
		warehouseIDStr = strings.TrimPrefix(warehouseIDStr, "[")
		warehouseIDStr = strings.TrimSuffix(warehouseIDStr, "]")
		if warehouseID, err := uuid.Parse(warehouseIDStr); err == nil {
			filter.WarehouseID = &warehouseID
		}
	}

	// Handle category_id manually
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		categoryIDStr = strings.TrimPrefix(categoryIDStr, "[")
		categoryIDStr = strings.TrimSuffix(categoryIDStr, "]")
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filter.CategoryID = &categoryID
		}
	}

	// Handle product_id manually
	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productIDStr = strings.TrimPrefix(productIDStr, "[")
		productIDStr = strings.TrimSuffix(productIDStr, "]")
		if productID, err := uuid.Parse(productIDStr); err == nil {
			filter.ProductID = &productID
		}
	}

	report, err := h.Service.GetStockReport(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stock report")
		helper.APIInternalServerError(c, "Failed to get stock report", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get stock report", report, nil)
}

// GetSalesReport godoc
// @Summary Get sales report
// @Description Get sales report with date range and other filters
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param customer_id query string false "Filter by customer ID"
// @Param product_id query string false "Filter by product ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/sales [get]
func (h *Handler) GetSalesReport(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.SalesReportFilter

	// Handle start_date manually to support date-only format
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	// Handle end_date manually to support date-only format
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	// Handle customer_id manually
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerIDStr = strings.TrimPrefix(customerIDStr, "[")
		customerIDStr = strings.TrimSuffix(customerIDStr, "]")
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			filter.CustomerID = &customerID
		}
	}

	// Handle product_id manually
	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productIDStr = strings.TrimPrefix(productIDStr, "[")
		productIDStr = strings.TrimSuffix(productIDStr, "]")
		if productID, err := uuid.Parse(productIDStr); err == nil {
			filter.ProductID = &productID
		}
	}

	report, err := h.Service.GetSalesReport(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales report")
		helper.APIInternalServerError(c, "Failed to get sales report", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get sales report", report, nil)
}

// GetInventoryValuation godoc
// @Summary Get inventory valuation report
// @Description Get inventory valuation by warehouse
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Warehouse ID (optional)"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/inventory-valuation [get]
func (h *Handler) GetInventoryValuation(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	report, err := h.Service.GetInventoryValuation(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get inventory valuation")
		helper.APIInternalServerError(c, "Failed to get inventory valuation", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get inventory valuation", report, nil)
}

// GetProfitLossReport godoc
// @Summary Get profit and loss report
// @Description Get profit and loss report for a date range
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/profit-loss [get]
func (h *Handler) GetProfitLossReport(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	if startDateStr == "" || endDateStr == "" {
		helper.APIBadRequest(c, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid start date format")
		helper.APIBadRequest(c, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid end date format")
		helper.APIBadRequest(c, "Invalid end date format", err)
		return
	}

	report, err := h.Service.GetProfitLossReport(ctx, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get profit loss report")
		helper.APIInternalServerError(c, "Failed to get profit loss report", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get profit loss report", report, nil)
}

// GetTopProducts godoc
// @Summary Get top selling products
// @Description Get top selling products by period
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of products to return" default(10)
// @Param period query string false "Period (week, month, quarter, year)" default(month)
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/top-products [get]
func (h *Handler) GetTopProducts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	limit := 10
	if limitStr := c.DefaultQuery("limit", "10"); limitStr != "" {
		var err error
		limit, err = parseInt(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}
	}

	period := c.DefaultQuery("period", "month")

	report, err := h.Service.GetTopProducts(ctx, limit, period)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top products")
		helper.APIInternalServerError(c, "Failed to get top products", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get top products", report, nil)
}

// GetLowStockReport godoc
// @Summary Get low stock report
// @Description Get products with stock below minimum threshold
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/low-stock [get]
func (h *Handler) GetLowStockReport(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	report, err := h.Service.GetLowStockReport(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get low stock report")
		helper.APIInternalServerError(c, "Failed to get low stock report", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get low stock report", report, nil)
}

// GetMovementReport godoc
// @Summary Get stock movement report
// @Description Get stock movement report with filtering options
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param product_id query string false "Filter by product ID"
// @Param movement_type query string false "Filter by movement type"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/reports/movements [get]
func (h *Handler) GetMovementReport(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.MovementReportFilter

	// Handle start_date manually to support date-only format
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	// Handle end_date manually to support date-only format
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	// Handle warehouse_id manually
	if warehouseIDStr := c.Query("warehouse_id"); warehouseIDStr != "" {
		warehouseIDStr = strings.TrimPrefix(warehouseIDStr, "[")
		warehouseIDStr = strings.TrimSuffix(warehouseIDStr, "]")
		if warehouseID, err := uuid.Parse(warehouseIDStr); err == nil {
			filter.WarehouseID = &warehouseID
		}
	}

	// Handle product_id manually
	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productIDStr = strings.TrimPrefix(productIDStr, "[")
		productIDStr = strings.TrimSuffix(productIDStr, "]")
		if productID, err := uuid.Parse(productIDStr); err == nil {
			filter.ProductID = &productID
		}
	}

	// Handle movement_type manually
	if movementType := c.Query("movement_type"); movementType != "" {
		filter.MovementType = &movementType
	}

	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}

	report, total, err := h.Service.GetMovementReport(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get movement report")
		helper.APIInternalServerError(c, "Failed to get movement report", err)
		return
	}

	response := map[string]interface{}{
		"movements": report,
		"pagination": map[string]interface{}{
			"page":       filter.Page,
			"limit":      filter.Limit,
			"total":      total,
			"totalPages": (total + filter.Limit - 1) / filter.Limit,
		},
	}

	helper.APIResponse(c, http.StatusOK, "Success get movement report", response, nil)
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
