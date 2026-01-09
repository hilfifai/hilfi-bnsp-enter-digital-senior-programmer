// internal/handler/product_category/handler.go
package product_category

import (
	dto "api-jellyfish/internal/dto"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)
	CreateCategory(c *gin.Context)
	GetCategoryByID(c *gin.Context)
	GetCategories(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
}

func NewHandler(svc service.IProductCategoryService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IProductCategoryService
}

// CreateCategory godoc
// @Summary Create product category
// @Description Create a new product category
// @Tags Product Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body dto.CreateProductCategoryRequest true "Category data"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 409 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	var req dto.CreateProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid category data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	category, err := h.Service.CreateCategory(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create category")
		helper.APIResponse(c, http.StatusConflict, "Failed to create category", map[string]string{"error": err.Error()}, err)
		return
	}

	log.Info().Str("category_id", category.ID.String()).Str("code", category.Code).Msg("Category created successfully")
	helper.APICreateSuccess(c, "Category created successfully", category)
}

// GetCategoryByID godoc
// @Summary Get category by ID
// @Description Get detailed information about a specific category
// @Tags Product Categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/product-categories/{id} [get]
func (h *Handler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid category ID", err)
		return
	}

	category, err := h.Service.GetCategoryByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("category_id", id.String()).Msg("Failed to get category")
		helper.APINotFound(c, "Category not found", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Category retrieved successfully", category, nil)
}

// GetCategories godoc
// @Summary Get all product categories
// @Description Get list of all product categories
// @Tags Product Categories
// @Produce json
// @Security BearerAuth
// @Param active_only query bool false "Active only"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-categories [get]
func (h *Handler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	activeOnly := c.Query("active_only") == "true"

	categories, err := h.Service.GetCategories(ctx, activeOnly)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get categories")
		helper.APIInternalServerError(c, "Failed to get categories", err)
		return
	}

	response := map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	}

	log.Info().Int("count", len(categories)).Msg("Categories retrieved successfully")
	helper.APIResponse(c, http.StatusOK, "Categories retrieved successfully", response, nil)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update an existing category
// @Tags Product Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category UUID"
// @Param category body dto.UpdateProductCategoryRequest true "Updated category data"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid category ID", err)
		return
	}

	var req dto.UpdateProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		log.Error().Err(err).Msg("Validation error")
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "Invalid category data", err)
			return
		}
		helper.APISystemValidationError(c, "System validation error", err)
		return
	}

	category, err := h.Service.UpdateCategory(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("category_id", id.String()).Msg("Failed to update category")
		helper.APIInternalServerError(c, "Failed to update category", err)
		return
	}

	log.Info().Str("category_id", id.String()).Msg("Category updated successfully")
	helper.APIUpdateSuccess(c, "Category updated successfully", category)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category (soft delete)
// @Tags Product Categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/product-categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID")
		helper.APIBadRequest(c, "Invalid category ID", err)
		return
	}

	if err := h.Service.DeleteCategory(ctx, id); err != nil {
		log.Error().Err(err).Str("category_id", id.String()).Msg("Failed to delete category")
		helper.APIInternalServerError(c, "Failed to delete category", err)
		return
	}

	log.Info().Str("category_id", id.String()).Msg("Category deleted successfully")
	helper.APIDeleteSuccess(c, "Category deleted successfully")
}
