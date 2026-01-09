package product

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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)
	CreateProduct(c *gin.Context)
	GetProduct(c *gin.Context)
	GetProducts(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
	UpdateStock(c *gin.Context)
	GetLowStockProducts(c *gin.Context)
	SearchProducts(c *gin.Context)
}

type Handler struct {
	Service service.IProductService
}

func NewHandler(svc service.IProductService) IHandler {
	return &Handler{Service: svc}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with all required information
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body dto.CreateProductRequest true "Product data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 409 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid product data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	product, err := h.Service.CreateProduct(c, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create product")
		switch {
		case errors.Is(err, apperrors.ErrProductSKUExists):
			helper.APIResponse(c, http.StatusConflict, "SKU already exists",
				map[string]string{"sku": "Product SKU is already in use."}, err)
		case errors.Is(err, apperrors.ErrProductCategoryNotFound):
			helper.APINotFound(c, "Product category not found", err)
		case errors.Is(err, apperrors.ErrInvalidPrice):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid price value",
				map[string]string{"price": "Price must be greater than or equal to 0."}, err)
		case errors.Is(err, apperrors.ErrInvalidStock):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid stock value",
				map[string]string{"stock": "Stock values must be positive."}, err)
		case errors.Is(err, apperrors.ErrInvalidStockRange):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid stock range",
				map[string]string{"min_stock": "Minimum stock cannot be greater than maximum stock."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to create product", err)
		}
		return
	}

	log.Info().Str("product_id", product.ID.String()).Str("sku", product.SKU).Msg("Product created successfully")
	helper.APICreateSuccess(c, "Product created successfully", product)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get detailed information about a specific product
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	product, err := h.Service.GetProductByID(c, id)
	if err != nil {
		log.Error().Err(err).Str("product_id", id.String()).Msg("Failed to get product")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to get product", err)
		}
		return
	}

	helper.APIResponse(c, http.StatusOK, "Product retrieved successfully", product, nil)
}

// GetProducts godoc
// @Summary Get list of products with filtering and pagination
// @Description Get paginated list of products with optional filters
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param category_id query string false "Filter by category ID"
// @Param sku query string false "Filter by SKU"
// @Param name query string false "Filter by product name"
// @Param barcode query string false "Filter by barcode"
// @Param is_active query boolean false "Filter by active status"
// @Param low_stock_only query boolean false "Show only low stock products"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sort_by query string false "Sort field" Enums(name, sku, created_at, current_stock, unit_price)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Router /api/v1/products [get]
func (h *Handler) GetProducts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var filter dto.ProductFilterDTO
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Error().Err(err).Msg("Failed to bind query parameters")
		helper.APIBindingError(c, err)
		return
	}

	// Set default values
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
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

	products, total, err := h.Service.GetProducts(c, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get products")
		helper.APIInternalServerError(c, "Failed to get products", err)
		return
	}

	response := map[string]interface{}{
		"products": products,
		"pagination": map[string]interface{}{
			"page":       filter.Page,
			"limit":      filter.Limit,
			"total":      total,
			"totalPages": (total + filter.Limit - 1) / filter.Limit,
		},
	}

	log.Info().Int("count", len(products)).Int("total", total).Msg("Products retrieved successfully")
	helper.APIResponse(c, http.StatusOK, "Products retrieved successfully", response, nil)
}

// UpdateProduct godoc
// @Summary Update product information
// @Description Update an existing product's information
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Param product body dto.UpdateProductRequest true "Updated product data"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid product data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	product, err := h.Service.UpdateProduct(c, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("product_id", id.String()).Msg("Failed to update product")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrProductCategoryNotFound):
			helper.APINotFound(c, "Product category not found", err)
		case errors.Is(err, apperrors.ErrInvalidPrice):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid price value",
				map[string]string{"price": "Price must be greater than or equal to 0."}, err)
		case errors.Is(err, apperrors.ErrInvalidStock):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid stock value",
				map[string]string{"stock": "Stock values must be positive."}, err)
		case errors.Is(err, apperrors.ErrInvalidStockRange):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid stock range",
				map[string]string{"min_stock": "Minimum stock cannot be greater than maximum stock."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to update product", err)
		}
		return
	}

	log.Info().Str("product_id", id.String()).Msg("Product updated successfully")
	helper.APIUpdateSuccess(c, "Product updated successfully", product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Soft delete a product (mark as inactive)
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	err = h.Service.DeleteProduct(c, id, userID)
	if err != nil {
		log.Error().Err(err).Str("product_id", id.String()).Msg("Failed to delete product")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		default:
			helper.APIInternalServerError(c, "Failed to delete product", err)
		}
		return
	}

	log.Info().Str("product_id", id.String()).Msg("Product deleted successfully")
	helper.APIDeleteSuccess(c, "Product deleted successfully")
}

// UpdateStock godoc
// @Summary Update product stock
// @Description Update stock quantity for a product in a specific warehouse
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Param stock body dto.BulkUpdateStockRequest true "Stock update data"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 409 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products/{id}/stock [put]
func (h *Handler) UpdateStock(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid product ID", err)
		return
	}

	var req dto.BulkUpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	req.ProductID = productID

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid stock update data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	err = h.Service.UpdateStock(c, req, userID)
	if err != nil {
		log.Error().Err(err).Str("product_id", productID.String()).Msg("Failed to update stock")
		switch {
		case errors.Is(err, apperrors.ErrProductNotFound):
			helper.APINotFound(c, "Product not found", err)
		case errors.Is(err, apperrors.ErrWarehouseNotFound):
			helper.APINotFound(c, "Warehouse not found", err)
		case errors.Is(err, apperrors.ErrInsufficientStock):
			helper.APIResponse(c, http.StatusConflict, "Insufficient stock",
				map[string]string{"quantity": "Insufficient stock available."}, err)
		case errors.Is(err, apperrors.ErrInvalidStockOperation):
			helper.APIResponse(c, http.StatusBadRequest, "Invalid stock operation",
				map[string]string{"type": "Stock operation type must be INCREMENT, DECREMENT, or SET."}, err)
		default:
			helper.APIInternalServerError(c, "Failed to update stock", err)
		}
		return
	}

	log.Info().Str("product_id", productID.String()).
		Str("warehouse_id", req.WarehouseID.String()).
		Int("quantity", req.Quantity).
		Msg("Stock updated successfully")

	helper.APIResponse(c, http.StatusOK, "Stock updated successfully", nil, nil)
}

// GetLowStockProducts godoc
// @Summary Get low stock products
// @Description Get list of products with stock below minimum threshold
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products/low-stock [get]
func (h *Handler) GetLowStockProducts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	products, err := h.Service.GetLowStockProducts(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get low stock products")
		helper.APIInternalServerError(c, "Failed to get low stock products", err)
		return
	}

	log.Info().Int("count", len(products)).Msg("Low stock products retrieved")
	helper.APIResponse(c, http.StatusOK, "Low stock products retrieved", products, nil)
}

// SearchProducts godoc
// @Summary Search products
// @Description Search products by name, SKU, or barcode
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param query query string true "Search query"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/products/search [get]
func (h *Handler) SearchProducts(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	query := c.Query("query")
	if query == "" {
		helper.APIBadRequest(c, "Search query is required", nil)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := h.Service.SearchProducts(c, query, limit)
	if err != nil {
		log.Error().Err(err).Str("query", query).Msg("Failed to search products")
		helper.APIInternalServerError(c, "Failed to search products", err)
		return
	}

	log.Info().Str("query", query).Int("results", len(products)).Msg("Products search completed")
	helper.APIResponse(c, http.StatusOK, "Search completed", products, nil)
}
