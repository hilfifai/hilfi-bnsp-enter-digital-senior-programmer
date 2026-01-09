// internal/handler/warehouse/route.go
package warehouse

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	warehouse := e.Group("/warehouse")
	warehouse.Use(authMiddleware)

	warehouse.POST("", h.CreateWarehouse)
	warehouse.GET("", h.GetAllWarehouses)
	warehouse.GET("/:id", h.GetWarehouseByID)
	warehouse.GET("/:id/stats", h.GetWarehouseStats)
	warehouse.PUT("/:id", h.UpdateWarehouse)
	warehouse.DELETE("/:id", h.DeleteWarehouse)
}
