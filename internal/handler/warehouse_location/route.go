// internal/handler/warehouse_location/route.go
package warehouse_location

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	location := e.Group("/location")
	location.Use(authMiddleware)

	location.POST("", h.CreateLocation)
	location.GET("", h.GetLocationsByWarehouse)
	location.GET("/:id", h.GetLocationByID)
	location.GET("/available", h.GetAvailableLocations)
	location.PUT("/:id", h.UpdateLocation)
	location.DELETE("/:id", h.DeleteLocation)
}
