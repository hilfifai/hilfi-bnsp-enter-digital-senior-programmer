// internal/handler/customer/handler.go
package customer

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

	CreateCustomer(c *gin.Context)
	GetCustomers(c *gin.Context)
	GetCustomer(c *gin.Context)
	UpdateCustomer(c *gin.Context)
	DeleteCustomer(c *gin.Context)
}

type Handler struct {
	Service service.ICustomerService
}

func NewHandler(svc service.ICustomerService) IHandler {
	return &Handler{Service: svc}
}

// CreateCustomer godoc
// @Summary Create a new customer
// @Description Create a new customer
// @Tags Customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param customer body dto.CreateCustomerRequest true "Customer data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/customers [post]
func (h *Handler) CreateCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid customer data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	customer, err := h.Service.CreateCustomer(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create customer")
		switch {
		case errors.Is(err, apperrors.ErrCustomerCodeExists):
			helper.APIResponse(c, http.StatusConflict, "Customer code already exists", nil, err)
		default:
			helper.APIInternalServerError(c, "Failed to create customer", err)
		}
		return
	}

	log.Info().Str("customer_code", customer.Code).Msg("Customer created successfully")
	helper.APICreateSuccess(c, "Customer created successfully", customer)
}

// GetCustomers godoc
// @Summary Get all customers
// @Description Get all customers with optional active filter
// @Tags Customer
// @Produce json
// @Security BearerAuth
// @Param active_only query bool false "Filter active only"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/customers [get]
func (h *Handler) GetCustomers(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	activeOnly := c.Query("active_only") == "true"

	customers, err := h.Service.GetCustomers(ctx, activeOnly)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get customers")
		helper.APIInternalServerError(c, "Failed to get customers", err)
		return
	}

	helper.APISuccess(c, "Customers retrieved successfully", customers)
}

// GetCustomer godoc
// @Summary Get customer by ID
// @Description Get customer by ID
// @Tags Customer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/customers/{id} [get]
func (h *Handler) GetCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid customer ID")
		helper.APIBadRequest(c, "Invalid customer ID", err)
		return
	}

	customer, err := h.Service.GetCustomerByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get customer")
		helper.APINotFound(c, "Customer not found", err)
		return
	}

	helper.APISuccess(c, "Customer retrieved successfully", customer)
}

// UpdateCustomer godoc
// @Summary Update customer
// @Description Update customer by ID
// @Tags Customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Param customer body dto.UpdateCustomerRequest true "Customer data"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/customers/{id} [put]
func (h *Handler) UpdateCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid customer ID")
		helper.APIBadRequest(c, "Invalid customer ID", err)
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid customer data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	customer, err := h.Service.UpdateCustomer(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update customer")
		switch {
		case errors.Is(err, apperrors.ErrCustomerCodeExists):
			helper.APIResponse(c, http.StatusConflict, "Customer code already exists", nil, err)
		case errors.Is(err, apperrors.ErrCustomerNotFound):
			helper.APINotFound(c, "Customer not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to update customer", err)
		}
		return
	}

	log.Info().Str("customer_id", id.String()).Msg("Customer updated successfully")
	helper.APISuccess(c, "Customer updated successfully", customer)
}

// DeleteCustomer godoc
// @Summary Delete customer
// @Description Delete customer by ID (soft delete)
// @Tags Customer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/customers/{id} [delete]
func (h *Handler) DeleteCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid customer ID")
		helper.APIBadRequest(c, "Invalid customer ID", err)
		return
	}

	if err := h.Service.DeleteCustomer(ctx, id); err != nil {
		log.Error().Err(err).Msg("Failed to delete customer")
		switch {
		case errors.Is(err, apperrors.ErrCustomerNotFound):
			helper.APINotFound(c, "Customer not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to delete customer", err)
		}
		return
	}

	log.Info().Str("customer_id", id.String()).Msg("Customer deleted successfully")
	helper.APIDeleteSuccess(c, "Customer deleted successfully")
}

// NewRoutes configures the routes for customer handler
func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	customers := e.Group("/customers")
	customers.Use(authMiddleware)
	{
		customers.POST("", h.CreateCustomer)
		customers.GET("", h.GetCustomers)
		customers.GET("/:id", h.GetCustomer)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
	}
}
