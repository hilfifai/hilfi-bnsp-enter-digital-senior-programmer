// internal/handler/delivery_order/handler.go
package delivery_order

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

	CreateDeliveryOrder(c *gin.Context)
	GetDeliveryOrders(c *gin.Context)
	GetDeliveryOrder(c *gin.Context)
	CreateSalesReturn(c *gin.Context)
}

type Handler struct {
	Service service.IDeliveryOrderService
}

func NewHandler(svc service.IDeliveryOrderService) IHandler {
	return &Handler{Service: svc}
}

// CreateDeliveryOrder godoc
// @Summary Create a delivery order
// @Description Create a delivery order for a sales order
// @Tags Delivery Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param delivery_order body dto.CreateDeliveryOrderRequest true "Delivery Order data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 409 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/delivery-orders [post]
func (h *Handler) CreateDeliveryOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid delivery order data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	deliveryOrder, err := h.Service.CreateDeliveryOrder(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create delivery order")
		switch {
		case errors.Is(err, apperrors.ErrSalesOrderNotFound):
			helper.APINotFound(c, "Sales order not found", err)
		case errors.Is(err, apperrors.ErrSalesOrderNotApproved):
			helper.APIResponse(c, http.StatusConflict, "Sales order not approved",
				map[string]string{"status": "Sales order must be approved."}, err)
		case errors.Is(err, apperrors.ErrInsufficientStock):
			helper.APIResponse(c, http.StatusConflict, "Insufficient stock",
				map[string]string{"stock": "Insufficient stock available."}, err)
		case errors.Is(err, apperrors.ErrInvalidDeliveredQuantity):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid delivered quantity",
				map[string]string{"quantity": "Delivered quantity must be greater than 0."}, err)
		case errors.Is(err, apperrors.ErrExceedRemainingQuantity):
			helper.APIResponse(c, http.StatusConflict, "Exceed remaining quantity",
				map[string]string{"quantity": "Delivered quantity exceeds remaining quantity."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to create delivery order", err)
		}
		return
	}

	log.Info().Str("do_number", deliveryOrder.DONumber).Msg("Delivery order created successfully")
	helper.APICreateSuccess(c, "Delivery order created successfully", deliveryOrder)
}

// GetDeliveryOrders godoc
// @Summary Get all delivery orders
// @Description Get all delivery orders with optional filters
// @Tags Delivery Order
// @Produce json
// @Security BearerAuth
// @Param sales_order_id query string false "Filter by sales order ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/delivery-orders [get]
func (h *Handler) GetDeliveryOrders(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.DeliveryOrderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Error().Err(err).Msg("Failed to bind query")
		helper.APIBindingError(c, err)
		return
	}

	orders, total, err := h.Service.GetDeliveryOrders(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get delivery orders")
		helper.APIInternalServerError(c, "Failed to get delivery orders", err)
		return
	}

	helper.APISuccess(c, "Delivery orders retrieved successfully", gin.H{
		"data":   orders,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetDeliveryOrder godoc
// @Summary Get delivery order by ID
// @Description Get delivery order by ID
// @Tags Delivery Order
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery Order ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/delivery-orders/{id} [get]
func (h *Handler) GetDeliveryOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid delivery order ID")
		helper.APIBadRequest(c, "Invalid delivery order ID", err)
		return
	}

	deliveryOrder, err := h.Service.GetDeliveryOrderByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get delivery order")
		switch {
		case errors.Is(err, apperrors.ErrDeliveryOrderNotFound):
			helper.APINotFound(c, "Delivery order not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to get delivery order", err)
		}
		return
	}

	helper.APISuccess(c, "Delivery order retrieved successfully", deliveryOrder)
}

// CreateSalesReturn godoc
// @Summary Create a sales return
// @Description Create a sales return
// @Tags Delivery Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param return body dto.SalesReturnRequest true "Sales Return data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/delivery-orders/returns [post]
func (h *Handler) CreateSalesReturn(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.SalesReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid sales return data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	salesReturn, err := h.Service.ProcessSalesReturn(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sales return")
		helper.APIInternalServerError(c, "Failed to create sales return", err)
		return
	}

	log.Info().Msg("Sales return created successfully")
	helper.APICreateSuccess(c, "Sales return created successfully", salesReturn)
}

// NewRoutes configures the routes for delivery order handler
func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	deliveryOrders := e.Group("/delivery-orders")
	deliveryOrders.Use(authMiddleware)
	{
		deliveryOrders.POST("", h.CreateDeliveryOrder)
		deliveryOrders.GET("", h.GetDeliveryOrders)
		deliveryOrders.GET("/:id", h.GetDeliveryOrder)

		// Returns
		returns := deliveryOrders.Group("/returns")
		returns.POST("", h.CreateSalesReturn)
	}
}
