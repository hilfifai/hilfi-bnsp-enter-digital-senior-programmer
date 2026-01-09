// internal/handler/sales_order/handler.go
package sales_order

import (
	"api-jellyfish/internal/dto"
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

	CreateSalesOrder(c *gin.Context)
	GetSalesOrders(c *gin.Context)
	GetSalesOrder(c *gin.Context)
	ApproveSalesOrder(c *gin.Context)
	CancelSalesOrder(c *gin.Context)
}

type Handler struct {
	Service service.ISalesOrderService
}

func NewHandler(svc service.ISalesOrderService) IHandler {
	return &Handler{Service: svc}
}

// CreateSalesOrder godoc
// @Summary Create a new sales order
// @Description Create a new sales order with items
// @Tags Sales Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sales_order body dto.CreateSalesOrderRequest true "Sales Order data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/sales-orders [post]
func (h *Handler) CreateSalesOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid sales order data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	salesOrder, err := h.Service.CreateSalesOrder(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sales order")
		switch {
		case errors.Is(err, apperrors.ErrCustomerNotFound):
			helper.APINotFound(c, "Customer not found", err)
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrInsufficientStock):
			helper.APIResponse(c, http.StatusConflict, "Insufficient stock",
				map[string]string{"stock": "Insufficient stock available."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to create sales order", err)
		}
		return
	}

	log.Info().Str("so_number", salesOrder.SONumber).Msg("Sales order created successfully")
	helper.APICreateSuccess(c, "Sales order created successfully", salesOrder)
}

// GetSalesOrders godoc
// @Summary Get all sales orders
// @Description Get all sales orders with optional filters
// @Tags Sales Order
// @Produce json
// @Security BearerAuth
// @Param customer_id query string false "Filter by customer ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/sales-orders [get]
func (h *Handler) GetSalesOrders(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.SalesOrderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Error().Err(err).Msg("Failed to bind query")
		helper.APIBindingError(c, err)
		return
	}

	orders, total, err := h.Service.GetSalesOrders(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales orders")
		helper.APIInternalServerError(c, "Failed to get sales orders", err)
		return
	}

	helper.APISuccess(c, "Sales orders retrieved successfully", gin.H{
		"data":   orders,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetSalesOrder godoc
// @Summary Get sales order by ID
// @Description Get sales order by ID
// @Tags Sales Order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/sales-orders/{id} [get]
func (h *Handler) GetSalesOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid sales order ID")
		helper.APIBadRequest(c, "Invalid sales order ID", err)
		return
	}

	salesOrder, err := h.Service.GetSalesOrderByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales order")
		helper.APINotFound(c, "Sales order not found", err)
		return
	}

	helper.APISuccess(c, "Sales order retrieved successfully", salesOrder)
}

// ApproveSalesOrder godoc
// @Summary Approve sales order
// @Description Approve sales order by ID
// @Tags Sales Order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/sales-orders/{id}/approve [put]
func (h *Handler) ApproveSalesOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid sales order ID")
		helper.APIBadRequest(c, "Invalid sales order ID", err)
		return
	}

	salesOrder, err := h.Service.ApproveSalesOrder(ctx, id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to approve sales order")
		helper.APIResponse(c, http.StatusBadRequest, "Failed to approve sales order", nil, err)
		return
	}

	log.Info().Str("so_number", salesOrder.SONumber).Msg("Sales order approved successfully")
	helper.APISuccess(c, "Sales order approved successfully", salesOrder)
}

// CancelSalesOrder godoc
// @Summary Cancel sales order
// @Description Cancel sales order by ID
// @Tags Sales Order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/sales-orders/{id}/cancel [put]
func (h *Handler) CancelSalesOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid sales order ID")
		helper.APIBadRequest(c, "Invalid sales order ID", err)
		return
	}

	salesOrder, err := h.Service.CancelSalesOrder(ctx, id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel sales order")
		helper.APIResponse(c, http.StatusBadRequest, "Failed to cancel sales order", nil, err)
		return
	}

	log.Info().Str("so_number", salesOrder.SONumber).Msg("Sales order cancelled successfully")
	helper.APISuccess(c, "Sales order cancelled successfully", salesOrder)
}
