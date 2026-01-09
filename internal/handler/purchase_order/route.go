// internal/handler/purchase_order/route.go
package purchase_order

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {

	purchaseOrder := e.Group("/purchase-order")
	purchaseOrder.Use(authMiddleware)

	purchaseOrder.POST("", h.CreatePurchaseOrder)
	purchaseOrder.GET("", h.GetPurchaseOrders)
	purchaseOrder.GET("/:id", h.GetPurchaseOrderByID)
	purchaseOrder.POST("/:id/approve", h.ApprovePurchaseOrder)
	purchaseOrder.POST("/:id/cancel", h.CancelPurchaseOrder)
}
