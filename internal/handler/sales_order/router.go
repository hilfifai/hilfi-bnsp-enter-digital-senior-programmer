package sales_order

import "github.com/gin-gonic/gin"

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	salesOrders := e.Group("/sales-orders")
	salesOrders.Use(authMiddleware)
	{
		salesOrders.POST("", h.CreateSalesOrder)
		salesOrders.GET("", h.GetSalesOrders)
		salesOrders.GET("/:id", h.GetSalesOrder)
		salesOrders.PUT("/:id/approve", h.ApproveSalesOrder)
		salesOrders.PUT("/:id/cancel", h.CancelSalesOrder)
	}
}
