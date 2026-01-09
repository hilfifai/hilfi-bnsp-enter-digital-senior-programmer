package errors

import "errors"

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductSKUExists  = errors.New("product SKU already exists")
	ErrCreateProduct     = errors.New("failed to create product")
	ErrUpdateProduct     = errors.New("failed to update product")
	ErrDeleteProduct     = errors.New("failed to delete product")
	ErrGetProducts       = errors.New("failed to get products")
	ErrProductValidation = errors.New("product validation error")

	// Product category errors
	ErrProductCategoryNotFound   = errors.New("product category not found")
	ErrProductCategoryCodeExists = errors.New("product category code already exists")
	ErrCreateProductCategory     = errors.New("failed to create product category")
	ErrUpdateProductCategory     = errors.New("failed to update product category")
	ErrDeleteProductCategory     = errors.New("failed to delete product category")

	// Stock errors
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrStockNotFound         = errors.New("stock not found")
	ErrGetStock              = errors.New("failed to get stock")
	ErrUpdateStock           = errors.New("failed to update stock")
	ErrCreateStockMovement   = errors.New("failed to create stock movement")
	ErrInvalidStockOperation = errors.New("invalid stock operation")
	ErrUpdateProductStock    = errors.New("failed to update product stock")

	// Warehouse errors
	ErrWarehouseNotFound           = errors.New("warehouse not found")
	ErrWarehouseCodeExists         = errors.New("warehouse code already exists")
	ErrWarehouseValidation         = errors.New("warehouse validation error")
	ErrCreateWarehouse             = errors.New("failed to create warehouse")
	ErrLocationNotFound            = errors.New("location not found")
	ErrLocationCodeExists          = errors.New("location code already exists")
	ErrLocationValidation          = errors.New("location validation error")
	ErrCreateLocation              = errors.New("failed to create location")
	ErrGetLocation                 = errors.New("failed to get location")
	ErrMissingSourceWarehouse      = errors.New("source warehouse is required")
	ErrMissingDestinationWarehouse = errors.New("destination warehouse is required")

	// Transaction errors
	ErrBeginTransaction    = errors.New("failed to begin transaction")
	ErrCommitTransaction   = errors.New("failed to commit transaction")
	ErrRollbackTransaction = errors.New("failed to rollback transaction")

	// Movement errors
	ErrInvalidMovementType         = errors.New("invalid movement type")
	ErrMissingWarehouseForTransfer = errors.New("source and destination warehouse are required for transfer")
	ErrSameWarehouseTransfer       = errors.New("source and destination warehouse cannot be the same")

	// Stock operation errors
	ErrCreateStock         = errors.New("failed to create stock")
	ErrUpdateStockMovement = errors.New("failed to update stock movement")
	ErrUpdateLocationUsage = errors.New("failed to update location usage")

	// Validation errors
	ErrInvalidPrice      = errors.New("invalid price")
	ErrInvalidStock      = errors.New("invalid stock")
	ErrInvalidStockRange = errors.New("invalid stock range")

	// Pagination errors
	ErrInvalidPaginationParam = errors.New("invalid pagination parameter")
	ErrTooManyFilters         = errors.New("too many filters")
	ErrInvalidQueryString     = errors.New("invalid query string")
	ErrTooManySortFields      = errors.New("too many sort fields")

	// Supplier errors
	ErrSupplierNotFound   = errors.New("supplier not found")
	ErrSupplierCodeExists = errors.New("supplier code already exists")

	// Purchase order errors
	ErrPurchaseOrderNotFound     = errors.New("purchase order not found")
	ErrCreatePurchaseOrder       = errors.New("failed to create purchase order")
	ErrCreatePurchaseOrderItem   = errors.New("failed to create purchase order item")
	ErrPurchaseOrderNotApproved  = errors.New("purchase order not approved")
	ErrPurchaseOrderItemNotFound = errors.New("purchase order item not found")
	ErrUpdatePurchaseOrder       = errors.New("failed to update purchase order")
	ErrUpdatePurchaseOrderItem   = errors.New("failed to update purchase order item")
	ErrInvalidReceivedQuantity   = errors.New("invalid received quantity")
	ErrExceedRemainingQuantity   = errors.New("received quantity exceeds remaining quantity")

	// Goods receipt errors
	ErrCreateGoodsReceipt     = errors.New("failed to create goods receipt")
	ErrCreateGoodsReceiptItem = errors.New("failed to create goods receipt item")
	ErrUpdateGoodsReceipt     = errors.New("failed to update goods receipt")

	// Report errors
	ErrGenerateReport = errors.New("failed to generate report")

	// Role errors
	ErrRoleNotFound = errors.New("role not found")
	ErrCreateRole   = errors.New("failed to create role")
	ErrUpdateRole   = errors.New("failed to update role")
	ErrGetRole      = errors.New("failed to get role")

	// Customer errors
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrCustomerCodeExists = errors.New("customer code already exists")

	// Sales order errors
	ErrGenerateSONumber         = errors.New("failed to generate sales order number")
	ErrSalesOrderNotFound       = errors.New("sales order not found")
	ErrCreateSalesOrder         = errors.New("failed to create sales order")
	ErrCreateSalesOrderItem     = errors.New("failed to create sales order item")
	ErrSalesOrderNotApproved    = errors.New("sales order not approved")
	ErrSalesOrderItemNotFound   = errors.New("sales order item not found")
	ErrUpdateSalesOrder         = errors.New("failed to update sales order")
	ErrUpdateSalesOrderItem     = errors.New("failed to update sales order item")
	ErrInvalidDeliveredQuantity = errors.New("invalid delivered quantity")

	// Delivery order errors
	ErrGenerateDONumber        = errors.New("failed to generate delivery order number")
	ErrCreateDeliveryOrder     = errors.New("failed to create delivery order")
	ErrCreateDeliveryOrderItem = errors.New("failed to create delivery order item")
	ErrUpdateDeliveryOrder     = errors.New("failed to update delivery order")
	ErrDeliveryOrderNotFound   = errors.New("delivery order not found")

	// Sales return errors
	ErrGenerateReturnNumber    = errors.New("failed to generate return number")
	ErrCreateSalesReturn       = errors.New("failed to create sales return")
	ErrCreateSalesReturnItem   = errors.New("failed to create sales return item")
	ErrUpdateDeliveryOrderItem = errors.New("failed to update delivery order item")
	ErrInvalidReturnedQuantity = errors.New("invalid returned quantity")
	ErrExceedDeliveredQuantity = errors.New("returned quantity exceeds delivered quantity")

	ErrAssessmentValidation        = errors.New("Assessment validation failed")
	ErrAssessmentNotFound          = errors.New("Assessment not found")
	ErrAssessmentExists            = errors.New("Assessment already exists")
	ErrDuplicateAssessmentItemCode = errors.New("Assessment item code already exists")
	ErrAssessmentItemNotFound      = errors.New("Assessment item not found")
	ErrCreateAssessment            = errors.New("Failed to create assessment")
	ErrGetAssessment               = errors.New("Failed to get assessment")
	ErrUpdateAssessment            = errors.New("Failed to update assessment")
	ErrDeleteAssessment            = errors.New("Failed to delete assessment")
	ErrAssessmentCodeExists        = errors.New("Assessment code already exists")

	ErrCreateAssessmentItem     = errors.New("Failed to create assessment item")
	ErrGetAssessmentItem        = errors.New("Failed to get assessment item")
	ErrUpdateAssessmentItem     = errors.New("Failed to update assessment item")
	ErrDeleteAssessmentItem     = errors.New("Failed to delete assessment item")
	ErrAssessmentItemCodeExists = errors.New("Assessment item code already exists")
	ErrAssessmentItemValidation = errors.New("Assessment item validation failed")
)
