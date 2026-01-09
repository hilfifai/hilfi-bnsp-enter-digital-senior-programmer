// internal/handler/supplier/route.go
package supplier

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {

	supplier := e.Group("/supplier")
	supplier.Use(authMiddleware)

	supplier.POST("", h.CreateSupplier)
	supplier.GET("", h.GetSuppliers)
	supplier.GET("/:id", h.GetSupplierByID)
	supplier.PUT("/:id", h.UpdateSupplier)
	supplier.DELETE("/:id", h.DeleteSupplier)
}
