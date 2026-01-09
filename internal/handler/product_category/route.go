// internal/handler/product_category/route.go
package product_category

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	categories := e.Group("/product-categories")
	categories.Use(authMiddleware)

	categories.POST("", h.CreateCategory)
	categories.GET("", h.GetCategories)
	categories.GET("/:id", h.GetCategoryByID)
	categories.PUT("/:id", h.UpdateCategory)
	categories.DELETE("/:id", h.DeleteCategory)
}
