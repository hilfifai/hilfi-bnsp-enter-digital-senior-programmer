package product

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	products := e.Group("/products")
	products.Use(authMiddleware)

	products.POST("", h.CreateProduct)
	products.GET("", h.GetProducts)
	products.GET("/low-stock", h.GetLowStockProducts)
	products.GET("/search", h.SearchProducts)
	products.GET("/:id", h.GetProduct)
	products.PUT("/:id", h.UpdateProduct)
	products.DELETE("/:id", h.DeleteProduct)
	products.PUT("/:id/stock", h.UpdateStock)
}
