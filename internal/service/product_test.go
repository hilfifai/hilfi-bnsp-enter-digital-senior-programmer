package service

import (
	"context"
	"testing"
	"time"

	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ProductServiceInstance returns a testable ProductService with mocks
func ProductServiceInstance() (*ProductService, *MockProductRepository, *MockProductCategoryRepository, *MockInventoryRepository) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockProductCategoryRepository)
	mockInventoryRepo := new(MockInventoryRepository)

	productRepoInterface := interface{}(mockProductRepo).(repository.IProductRepository)
	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	inventoryRepoInterface := interface{}(mockInventoryRepo).(repository.IInventoryRepository)

	svc := NewProductService(productRepoInterface, categoryRepoInterface, inventoryRepoInterface).(*ProductService)

	return svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo
}

// Helper function to create test product request
func createTestCreateProductRequest() dto.CreateProductRequest {
	desc := "Test product description"
	return dto.CreateProductRequest{
		SKU:         "TEST-SKU-001",
		Name:        "Test Product",
		Description: &desc,
		CategoryID:  uuid.New(),
		UnitPrice:   decimal.NewFromFloat(99.99),
		CostPrice:   decimal.NewFromFloat(50.00),
		MinStock:    10,
		MaxStock:    100,
		IsActive:    true,
	}
}

func TestProductService_CreateProduct_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	req := createTestCreateProductRequest()
	userID := uuid.New()
	category := createTestCategory()

	mockCategoryRepo.On("FindByID", ctx, req.CategoryID).Return(category, nil)
	mockProductRepo.On("IsSKUExists", ctx, req.SKU, uuid.Nil).Return(false, nil)
	mockProductRepo.On("Create", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

	product, err := svc.CreateProduct(ctx, req, userID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, req.SKU, product.SKU)
	assert.Equal(t, req.Name, product.Name)
	assert.Equal(t, req.UnitPrice, product.UnitPrice)
	mockCategoryRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	_ = mockInventoryRepo
}

func TestProductService_CreateProduct_DuplicateSKU(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	req := createTestCreateProductRequest()
	userID := uuid.New()

	mockProductRepo.On("IsSKUExists", ctx, req.SKU, uuid.Nil).Return(true, nil)

	product, err := svc.CreateProduct(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Nil(t, product)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
	_ = mockInventoryRepo
}

func TestProductService_CreateProduct_CategoryNotFound(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	req := createTestCreateProductRequest()
	userID := uuid.New()

	mockProductRepo.On("IsSKUExists", ctx, req.SKU, uuid.Nil).Return(false, nil)
	mockCategoryRepo.On("FindByID", ctx, req.CategoryID).Return(nil, nil)

	product, err := svc.CreateProduct(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category not found")
	assert.Nil(t, product)
	mockProductRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
	_ = mockInventoryRepo
}

func TestProductService_CreateProduct_InvalidPrice(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	req := createTestCreateProductRequest()
	req.UnitPrice = decimal.NewFromFloat(-10) // Negative price
	userID := uuid.New()

	mockProductRepo.On("IsSKUExists", ctx, req.SKU, uuid.Nil).Return(false, nil)
	mockCategoryRepo.On("FindByID", ctx, req.CategoryID).Return(createTestCategory(), nil)

	product, err := svc.CreateProduct(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	assert.Nil(t, product)
	mockProductRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
	_ = mockInventoryRepo
}

func TestProductService_CreateProduct_InvalidStockRange(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	req := createTestCreateProductRequest()
	req.MinStock = 100 // Min stock greater than max stock
	req.MaxStock = 50
	userID := uuid.New()

	mockProductRepo.On("IsSKUExists", ctx, req.SKU, uuid.Nil).Return(false, nil)
	mockCategoryRepo.On("FindByID", ctx, req.CategoryID).Return(createTestCategory(), nil)

	product, err := svc.CreateProduct(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stock")
	assert.Nil(t, product)
	mockProductRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
	_ = mockInventoryRepo
}

func TestProductService_GetProductByID_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	product := createTestProduct()

	mockProductRepo.On("FindByID", ctx, product.ID).Return(product, nil)
	mockInventoryRepo.On("GetProductStockByProductID", ctx, product.ID).Return([]entity.InventoryStock{}, nil)

	result, err := svc.GetProductByID(ctx, product.ID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, product.ID, result.Product.ID)
	assert.Equal(t, product.SKU, result.Product.SKU)
	mockProductRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

func TestProductService_GetProductByID_NotFound(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	productID := uuid.New()

	mockProductRepo.On("FindByID", ctx, productID).Return(nil, errpkg.ErrProductNotFound)

	result, err := svc.GetProductByID(ctx, productID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
	_ = mockInventoryRepo
}

func TestProductService_GetProducts_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	products := []entity.Product{*createTestProduct()}
	filter := dto.ProductFilterDTO{
		Page:  1,
		Limit: 10,
	}

	product := createTestProduct()
	mockProductRepo.On("FindAll", ctx, mock.AnythingOfType("entity.ProductFilter")).Return(products, nil)
	mockProductRepo.On("GetCount", ctx).Return(1, nil)
	mockProductRepo.On("FindByID", ctx, mock.Anything).Return(product, nil)
	mockInventoryRepo.On("GetProductStockByProductID", ctx, mock.Anything).Return([]entity.InventoryStock{}, nil)

	results, count, err := svc.GetProducts(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, results, 1)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	existingProduct := createTestProduct()
	category := createTestCategory()

	desc := "Updated description"
	req := dto.UpdateProductRequest{
		Name:        "Updated Product Name",
		Description: &desc,
		CategoryID:  category.ID,
		UnitPrice:   decimal.NewFromFloat(149.99),
		CostPrice:   decimal.NewFromFloat(75.00),
		MinStock:    5,
		MaxStock:    200,
		IsActive:    true,
	}

	mockProductRepo.On("FindByID", ctx, existingProduct.ID).Return(existingProduct, nil)
	mockCategoryRepo.On("FindByID", ctx, category.ID).Return(category, nil)
	mockProductRepo.On("Update", ctx, existingProduct.ID, mock.AnythingOfType("*entity.Product")).Return(nil)

	result, err := svc.UpdateProduct(ctx, existingProduct.ID, req, uuid.New())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	mockProductRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
	_ = mockInventoryRepo
}

func TestProductService_UpdateProduct_NotFound(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	productID := uuid.New()
	req := dto.UpdateProductRequest{
		Name:       "Updated Product Name",
		CategoryID: uuid.New(),
		UnitPrice:  decimal.NewFromFloat(149.99),
		CostPrice:  decimal.NewFromFloat(75.00),
		MinStock:   5,
		MaxStock:   200,
		IsActive:   true,
	}

	mockProductRepo.On("FindByID", ctx, productID).Return(nil, errpkg.ErrProductNotFound)

	result, err := svc.UpdateProduct(ctx, productID, req, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
	_ = mockInventoryRepo
}

func TestProductService_DeleteProduct_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	product := createTestProduct()

	mockProductRepo.On("FindByID", ctx, product.ID).Return(product, nil)
	mockProductRepo.On("Delete", ctx, product.ID).Return(nil)

	err := svc.DeleteProduct(ctx, product.ID, uuid.New())

	assert.NoError(t, err)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
	_ = mockInventoryRepo
}

func TestProductService_DeleteProduct_NotFound(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	productID := uuid.New()

	mockProductRepo.On("FindByID", ctx, productID).Return(nil, errpkg.ErrProductNotFound)

	err := svc.DeleteProduct(ctx, productID, uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
	_ = mockInventoryRepo
}

func TestProductService_UpdateStock_Increment_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	product := createTestProduct()
	warehouseID := uuid.New()
	locationID := uuid.New()

	currentStock := &entity.InventoryStock{
		ID:          uuid.New(),
		ProductID:   product.ID,
		WarehouseID: warehouseID,
		LocationID:  &locationID,
		Quantity:    50,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := dto.BulkUpdateStockRequest{
		ProductID:   product.ID,
		WarehouseID: warehouseID,
		LocationID:  &locationID,
		Quantity:    10,
		Type:        "INCREMENT",
		Reason:      "Stock replenishment",
	}

	mockProductRepo.On("FindByID", ctx, product.ID).Return(product, nil)
	mockInventoryRepo.On("GetWarehouseByID", ctx, warehouseID).Return(&entity.Warehouse{ID: warehouseID}, nil)
	mockInventoryRepo.On("GetStockByProductAndWarehouse", ctx, product.ID, warehouseID, &locationID).Return(currentStock, nil)
	mockInventoryRepo.On("UpdateStock", ctx, mock.Anything, mock.AnythingOfType("*entity.InventoryStock")).Return(nil)
	mockProductRepo.On("UpdateStock", ctx, product.ID, 10).Return(nil)
	mockInventoryRepo.On("CreateStockMovement", ctx, mock.AnythingOfType("*entity.StockMovement")).Return(nil)

	err := svc.UpdateStock(ctx, req, uuid.New())

	assert.NoError(t, err)
	mockProductRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

func TestProductService_UpdateStock_Decrement_InsufficientStock(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	product := createTestProduct()
	warehouseID := uuid.New()
	locationID := uuid.New()

	currentStock := &entity.InventoryStock{
		ID:          uuid.New(),
		ProductID:   product.ID,
		WarehouseID: warehouseID,
		LocationID:  &locationID,
		Quantity:    5, // Only 5 in stock
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := dto.BulkUpdateStockRequest{
		ProductID:   product.ID,
		WarehouseID: warehouseID,
		LocationID:  &locationID,
		Quantity:    10, // Trying to decrement 10
		Type:        "DECREMENT",
		Reason:      "Sales order fulfillment",
	}

	mockProductRepo.On("FindByID", ctx, product.ID).Return(product, nil)
	mockInventoryRepo.On("GetWarehouseByID", ctx, warehouseID).Return(&entity.Warehouse{ID: warehouseID}, nil)
	mockInventoryRepo.On("GetStockByProductAndWarehouse", ctx, product.ID, warehouseID, &locationID).Return(currentStock, nil)

	err := svc.UpdateStock(ctx, req, uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
	mockProductRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

func TestProductService_UpdateStock_InvalidOperation(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	product := createTestProduct()
	warehouseID := uuid.New()

	req := dto.BulkUpdateStockRequest{
		ProductID:   product.ID,
		WarehouseID: warehouseID,
		Quantity:    10,
		Type:        "INVALID", // Invalid operation type
		Reason:      "Test",
	}

	mockProductRepo.On("FindByID", ctx, product.ID).Return(product, nil)
	mockInventoryRepo.On("GetWarehouseByID", ctx, warehouseID).Return(&entity.Warehouse{ID: warehouseID}, nil)
	mockInventoryRepo.On("GetStockByProductAndWarehouse", ctx, product.ID, warehouseID, (*uuid.UUID)(nil)).Return(&entity.InventoryStock{}, errpkg.ErrStockNotFound)

	err := svc.UpdateStock(ctx, req, uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	_ = mockCategoryRepo
}

func TestProductService_GetLowStockProducts_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	lowStockProduct := createTestProduct()
	lowStockProduct.CurrentStock = 5 // Below min stock of 10

	products := []entity.Product{*lowStockProduct}

	mockProductRepo.On("FindAll", ctx, mock.AnythingOfType("entity.ProductFilter")).Return(products, nil)
	mockProductRepo.On("FindByID", ctx, mock.Anything).Return(lowStockProduct, nil)
	mockInventoryRepo.On("GetProductStockByProductID", ctx, lowStockProduct.ID).Return([]entity.InventoryStock{}, nil)

	result, err := svc.GetLowStockProducts(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

func TestProductService_SearchProducts_Success(t *testing.T) {
	svc, mockProductRepo, mockCategoryRepo, mockInventoryRepo := ProductServiceInstance()
	ctx := context.Background()

	products := []entity.Product{*createTestProduct()}

	mockProductRepo.On("FindAll", ctx, mock.AnythingOfType("entity.ProductFilter")).Return(products, nil)
	mockProductRepo.On("FindByID", ctx, mock.Anything).Return(createTestProduct(), nil)
	mockInventoryRepo.On("GetProductStockByProductID", ctx, mock.Anything).Return([]entity.InventoryStock{}, nil)

	result, err := svc.SearchProducts(ctx, "test", 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockProductRepo.AssertExpectations(t)
	_ = mockCategoryRepo
}

// ProductCategoryService tests
func TestProductCategoryService_CreateCategory_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	userID := uuid.New()
	req := dto.CreateProductCategoryRequest{
		Code:        "CAT-001",
		Name:        "Test Category",
		Description: strPtr("Test description"),
		IsActive:    true,
	}

	mockCategoryRepo.On("IsCodeExists", ctx, req.Code, uuid.Nil).Return(false, nil)
	mockCategoryRepo.On("Create", ctx, mock.AnythingOfType("*entity.ProductCategory")).Return(nil)

	category, err := svc.CreateCategory(ctx, req, userID)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, req.Code, category.Code)
	assert.Equal(t, req.Name, category.Name)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_CreateCategory_DuplicateCode(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	userID := uuid.New()
	req := dto.CreateProductCategoryRequest{
		Code:        "CAT-001",
		Name:        "Test Category",
		Description: strPtr("Test description"),
		IsActive:    true,
	}

	mockCategoryRepo.On("IsCodeExists", ctx, req.Code, uuid.Nil).Return(true, nil)

	category, err := svc.CreateCategory(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category code already exists")
	assert.Nil(t, category)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_CreateCategory_ParentNotFound(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	userID := uuid.New()
	parentID := uuid.New()
	req := dto.CreateProductCategoryRequest{
		Code:             "CAT-001",
		Name:             "Test Category",
		Description:      strPtr("Test description"),
		ParentCategoryID: &parentID,
		IsActive:         true,
	}

	mockCategoryRepo.On("IsCodeExists", ctx, req.Code, uuid.Nil).Return(false, nil)
	mockCategoryRepo.On("FindByID", ctx, parentID).Return(nil, errpkg.ErrProductCategoryNotFound)

	category, err := svc.CreateCategory(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent category not found")
	assert.Nil(t, category)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_GetCategoryByID_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	expectedCategory := createTestCategory()

	mockCategoryRepo.On("FindByID", ctx, expectedCategory.ID).Return(expectedCategory, nil)

	category, err := svc.GetCategoryByID(ctx, expectedCategory.ID)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, expectedCategory.ID, category.ID)
	assert.Equal(t, expectedCategory.Code, category.Code)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_GetCategoryByID_NotFound(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	categoryID := uuid.New()

	mockCategoryRepo.On("FindByID", ctx, categoryID).Return(nil, errpkg.ErrProductCategoryNotFound)

	category, err := svc.GetCategoryByID(ctx, categoryID)

	assert.Error(t, err)
	assert.Nil(t, category)
	assert.Contains(t, err.Error(), "category not found")
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_GetCategoryByCode_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	expectedCategory := createTestCategory()

	mockCategoryRepo.On("FindByCode", ctx, expectedCategory.Code).Return(expectedCategory, nil)

	category, err := svc.GetCategoryByCode(ctx, expectedCategory.Code)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, expectedCategory.Code, category.Code)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_GetCategories_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	categories := []entity.ProductCategory{*createTestCategory()}

	mockCategoryRepo.On("FindAll", ctx, true).Return(categories, nil)

	result, err := svc.GetCategories(ctx, true)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_UpdateCategory_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	existingCategory := createTestCategory()
	userID := uuid.New()

	req := dto.UpdateProductCategoryRequest{
		Name:        "Updated Category",
		Description: strPtr("Updated description"),
		IsActive:    true,
	}

	mockCategoryRepo.On("FindByID", ctx, existingCategory.ID).Return(existingCategory, nil)
	mockCategoryRepo.On("Update", ctx, existingCategory.ID, mock.AnythingOfType("*entity.ProductCategory")).Return(nil)

	result, err := svc.UpdateCategory(ctx, existingCategory.ID, req, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_UpdateCategory_NotFound(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	categoryID := uuid.New()
	userID := uuid.New()

	req := dto.UpdateProductCategoryRequest{
		Name:        "Updated Category",
		Description: strPtr("Updated description"),
		IsActive:    true,
	}

	mockCategoryRepo.On("FindByID", ctx, categoryID).Return(nil, errpkg.ErrProductCategoryNotFound)

	result, err := svc.UpdateCategory(ctx, categoryID, req, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "category not found")
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_DeleteCategory_Success(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	category := createTestCategory()

	mockCategoryRepo.On("FindByID", ctx, category.ID).Return(category, nil)
	mockCategoryRepo.On("Delete", ctx, category.ID).Return(nil)

	err := svc.DeleteCategory(ctx, category.ID)

	assert.NoError(t, err)
	mockCategoryRepo.AssertExpectations(t)
}

func TestProductCategoryService_DeleteCategory_NotFound(t *testing.T) {
	mockCategoryRepo := new(MockProductCategoryRepository)

	categoryRepoInterface := interface{}(mockCategoryRepo).(repository.IProductCategoryRepository)
	svc := &ProductCategoryService{
		categoryRepo: categoryRepoInterface,
	}

	ctx := context.Background()
	categoryID := uuid.New()

	mockCategoryRepo.On("FindByID", ctx, categoryID).Return(nil, errpkg.ErrProductCategoryNotFound)

	err := svc.DeleteCategory(ctx, categoryID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category not found")
	mockCategoryRepo.AssertExpectations(t)
}
