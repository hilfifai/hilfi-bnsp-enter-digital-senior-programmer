// internal/handler/goods_receipt/route.go
package goods_receipt

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	goodsReceipt := e.Group("/goods-receipt")
	goodsReceipt.Use(authMiddleware)
	{
		goodsReceipt.POST("", h.CreateGoodsReceipt)
		goodsReceipt.GET("/:id", h.GetGoodsReceiptByID)
	}
}
