// internal/handler/goods_receipt/handler.go
package goods_receipt

import (
	dto "api-jellyfish/internal/dto"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)

	CreateGoodsReceipt(c *gin.Context)
	GetGoodsReceiptByID(c *gin.Context)
}

func NewHandler(svc service.IGoodsReceiptService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IGoodsReceiptService
}

// CreateGoodsReceipt godoc
// @Summary Create goods receipt
// @Description Membuat goods receipt baru
// @Tags GoodsReceipt
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param goods_receipt body dto.GoodsReceiptRequest true "Goods Receipt"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/goods-receipt [post]
func (h *Handler) CreateGoodsReceipt(c *gin.Context) {
	userID := jwt.GetUser(c).UserID

	var req dto.GoodsReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.APIBindingError(c, err)
		return
	}

	if err := validation.Validate(req); err != nil {
		if validation.IsValidationError(err) {
			helper.APIValidationError(c, "", err)
			return
		}
		helper.APISystemValidationError(c, "", err)
		return
	}

	goodsReceipt, err := h.Service.CreateGoodsReceipt(c, req, userID)
	if err != nil {
		helper.APIInternalServerError(c, "Failed create goods receipt", err)
		return
	}

	helper.APICreateSuccess(c, "GoodsReceipt", goodsReceipt)
}

// GetGoodsReceiptByID godoc
// @Summary Get goods receipt by ID
// @Description Mendapatkan goods receipt berdasarkan ID
// @Tags GoodsReceipt
// @Produce json
// @Security BearerAuth
// @Param id path string true "Goods Receipt UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/goods-receipt/{id} [get]
func (h *Handler) GetGoodsReceiptByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	goodsReceipt, err := h.Service.GetGoodsReceiptByID(c, id)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data goods receipt", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get goods receipt by ID", goodsReceipt, nil)
}
