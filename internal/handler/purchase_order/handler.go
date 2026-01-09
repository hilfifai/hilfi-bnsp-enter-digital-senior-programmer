// internal/handler/purchase_order/handler.go
package purchase_order

import (
	dto "api-jellyfish/internal/dto"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/validation"
	"api-jellyfish/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc)

	CreatePurchaseOrder(c *gin.Context)
	GetPurchaseOrderByID(c *gin.Context)
	GetPurchaseOrders(c *gin.Context)
	ApprovePurchaseOrder(c *gin.Context)
	CancelPurchaseOrder(c *gin.Context)
}

func NewHandler(svc service.IPurchaseOrderService) IHandler {
	return &Handler{Service: svc}
}

type Handler struct {
	Service service.IPurchaseOrderService
}

// CreatePurchaseOrder godoc
// @Summary Create purchase order
// @Description Membuat purchase order baru
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param purchase_order body dto.CreatePurchaseOrderRequest true "Purchase Order"
// @Success 201 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/purchase-order [post]
func (h *Handler) CreatePurchaseOrder(c *gin.Context) {
	userID := jwt.GetUser(c).UserID

	var req dto.CreatePurchaseOrderRequest
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

	purchaseOrder, err := h.Service.CreatePurchaseOrder(c, req, userID)
	if err != nil {
		helper.APIInternalServerError(c, "Failed create purchase order", err)
		return
	}

	helper.APICreateSuccess(c, "PurchaseOrder", purchaseOrder)
}

// GetPurchaseOrderByID godoc
// @Summary Get purchase order by ID
// @Description Mendapatkan purchase order berdasarkan ID
// @Tags PurchaseOrder
// @Produce json
// @Security BearerAuth
// @Param id path string true "Purchase Order UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 404 {object} types.ResponseAPI
// @Router /api/v1/purchase-order/{id} [get]
func (h *Handler) GetPurchaseOrderByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		helper.APIBadRequest(c, "invalid UUID", err)
		return
	}

	purchaseOrder, err := h.Service.GetPurchaseOrderByID(c, id)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data purchase order", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get purchase order by ID", purchaseOrder, nil)
}

// GetPurchaseOrders godoc
// @Summary Get all purchase orders
// @Description Mendapatkan semua purchase order
// @Tags PurchaseOrder
// @Produce json
// @Security BearerAuth
// @Param supplier_id query string false "Supplier ID"
// @Param status query string false "Status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/purchase-order [get]
func (h *Handler) GetPurchaseOrders(c *gin.Context) {
	var filter dto.PurchaseOrderFilter

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		supplierID, err := uuid.Parse(supplierIDStr)
		if err == nil {
			filter.SupplierID = &supplierID
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	fmt.Sscanf(c.Query("limit"), "%d", &filter.Limit)
	fmt.Sscanf(c.Query("offset"), "%d", &filter.Offset)

	purchaseOrders, total, err := h.Service.GetPurchaseOrders(c, filter)
	if err != nil {
		helper.APIInternalServerError(c, "Failed get data purchase order", err)
		return
	}

	helper.APIResponse(c, http.StatusOK, "Success get all purchase orders", map[string]interface{}{
		"data":  purchaseOrders,
		"total": total,
	}, nil)
}

// ApprovePurchaseOrder godoc
// @Summary Approve purchase order
// @Description Menyetujui purchase order
// @Tags PurchaseOrder
// @Produce json
// @Security BearerAuth
// @Param id path string true "Purchase Order UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/purchase-order/{id}/approve [post]
func (h *Handler) ApprovePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in ApprovePurchaseOrder")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	purchaseOrder, err := h.Service.ApprovePurchaseOrder(c, id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error approving purchase order")
		helper.APIInternalServerError(c, "Failed approve purchase order", err)
		return
	}

	helper.APIUpdateSuccess(c, "Purchase order approved successfully", purchaseOrder)
}

// CancelPurchaseOrder godoc
// @Summary Cancel purchase order
// @Description Membatalkan purchase order
// @Tags PurchaseOrder
// @Produce json
// @Security BearerAuth
// @Param id path string true "Purchase Order UUID"
// @Success 200 {object} types.ResponseAPI
// @Failure 400 {object} types.ResponseAPI
// @Failure 401 {object} types.ResponseAPI
// @Failure 500 {object} types.ResponseAPI
// @Router /api/v1/purchase-order/{id}/cancel [post]
func (h *Handler) CancelPurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)
	userID := jwt.GetUser(c).UserID

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID in CancelPurchaseOrder")
		helper.APIBadRequest(c, "", fmt.Errorf("invalid UUID"))
		return
	}

	purchaseOrder, err := h.Service.CancelPurchaseOrder(c, id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error cancelling purchase order")
		helper.APIInternalServerError(c, "Failed cancel purchase order", err)
		return
	}

	helper.APIUpdateSuccess(c, "Purchase order cancelled successfully", purchaseOrder)
}
