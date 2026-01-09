package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"

	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to create test product
func createTestProduct() *entity.Product {
	ID := uuid.New()
	sku := "TEST-SKU-001"
	categoryID := uuid.New()
	description := "Test product description"
	return &entity.Product{
		ID:           ID,
		SKU:          sku,
		Name:         "Test Product",
		Description:  &description,
		CategoryID:   categoryID,
		UnitPrice:    decimal.NewFromFloat(15.00),
		CostPrice:    decimal.NewFromFloat(10.00),
		CurrentStock: 100,
		MinStock:     10,
		MaxStock:     500,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Helper function to create test category
func createTestCategory() *entity.ProductCategory {
	ID := uuid.New()
	desc := "Test category description"
	return &entity.ProductCategory{
		ID:          ID,
		Code:        "CAT-001",
		Name:        "Test Category",
		Description: &desc,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// MockProductCategoryRepository is a mock implementation of IProductCategoryRepository
type MockProductCategoryRepository struct {
	mock.Mock
}

func (m *MockProductCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ProductCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductCategory), args.Error(1)
}

func (m *MockProductCategoryRepository) FindByCode(ctx context.Context, code string) (*entity.ProductCategory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductCategory), args.Error(1)
}

func (m *MockProductCategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockProductCategoryRepository) Update(ctx context.Context, id uuid.UUID, category *entity.ProductCategory) error {
	args := m.Called(ctx, id, category)
	return args.Error(0)
}

func (m *MockProductCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductCategoryRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.ProductCategory, error) {
	args := m.Called(ctx, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.ProductCategory), args.Error(1)
}

func (m *MockProductCategoryRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, code, excludeID)
	return args.Bool(0), args.Error(1)
}

// MockInventoryRepository is a mock implementation of IInventoryRepository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) GetStockByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error) {
	args := m.Called(ctx, productID, warehouseID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.InventoryStock), args.Error(1)
}

func (m *MockInventoryRepository) GetStockByProductAndWarehouseTx(ctx context.Context, tx interface{}, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error) {
	args := m.Called(ctx, tx, productID, warehouseID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.InventoryStock), args.Error(1)
}

func (m *MockInventoryRepository) CreateStock(ctx context.Context, stock *entity.InventoryStock) error {
	args := m.Called(ctx, stock)
	return args.Error(0)
}

func (m *MockInventoryRepository) CreateStockTx(ctx context.Context, tx interface{}, stock *entity.InventoryStock) error {
	args := m.Called(ctx, tx, stock)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStock(ctx context.Context, id uuid.UUID, stock *entity.InventoryStock) error {
	args := m.Called(ctx, id, stock)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStockTx(ctx context.Context, tx interface{}, id uuid.UUID, stock *entity.InventoryStock) error {
	args := m.Called(ctx, tx, id, stock)
	return args.Error(0)
}

func (m *MockInventoryRepository) CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStockMovement(ctx context.Context, tx interface{}, movement *entity.StockMovement) error {
	args := m.Called(ctx, tx, movement)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetProductStockByProductID(ctx context.Context, productID uuid.UUID) ([]entity.InventoryStock, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.InventoryStock), args.Error(1)
}

func (m *MockInventoryRepository) GetWarehouseByID(ctx context.Context, warehouseID uuid.UUID) (*entity.Warehouse, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockInventoryRepository) BeginTx(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func (m *MockInventoryRepository) GetLowStockCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockInventoryRepository) GetTotalProducts(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockInventoryRepository) GetStockReport(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.StockReport), args.Error(1)
}

func (m *MockInventoryRepository) GetLowStockReport(ctx context.Context) ([]entity.StockReport, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.StockReport), args.Error(1)
}

func (m *MockInventoryRepository) GetMovementReport(ctx context.Context, filter dto.MovementReportFilter) ([]entity.StockMovement, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.StockMovement), args.Int(1), args.Error(2)
}

func (m *MockInventoryRepository) GetInventoryValuation(ctx context.Context, warehouseID *uuid.UUID) ([]entity.InventoryValuation, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.InventoryValuation), args.Error(1)
}

func (m *MockInventoryRepository) GetAllWarehouses(ctx context.Context) ([]entity.Warehouse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Warehouse), args.Error(1)
}

// MockProductRepository is a mock implementation of IProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Create(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Update(ctx context.Context, id uuid.UUID, product *entity.Product) error {
	args := m.Called(ctx, id, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) FindAll(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindActive(ctx context.Context) ([]entity.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) IsSKUExists(ctx context.Context, sku string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, sku, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductRepository) GetWithStockInfo(ctx context.Context, id uuid.UUID) (*entity.ProductWithStock, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductWithStock), args.Error(1)
}

func (m *MockProductRepository) GetCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockProductRepository) GetLowStockCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// MockWarehouseRepository is a mock implementation of IWarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Create(ctx context.Context, warehouse *entity.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Warehouse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindByCode(ctx context.Context, code string) (*entity.Warehouse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindAll(ctx context.Context) ([]entity.Warehouse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindActive(ctx context.Context) ([]entity.Warehouse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) Update(ctx context.Context, id uuid.UUID, warehouse *entity.Warehouse) error {
	args := m.Called(ctx, id, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWarehouseRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, code, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWarehouseRepository) GetWarehouseStats(ctx context.Context, id uuid.UUID) (*entity.WarehouseStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseStats), args.Error(1)
}

// MockWarehouseLocationRepository is a mock implementation of IWarehouseLocationRepository
type MockWarehouseLocationRepository struct {
	mock.Mock
}

func (m *MockWarehouseLocationRepository) Create(ctx context.Context, location *entity.WarehouseLocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockWarehouseLocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.WarehouseLocation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseLocation), args.Error(1)
}

func (m *MockWarehouseLocationRepository) FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WarehouseLocation), args.Error(1)
}

func (m *MockWarehouseLocationRepository) FindActiveByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WarehouseLocation), args.Error(1)
}

func (m *MockWarehouseLocationRepository) FindByCodeAndWarehouse(ctx context.Context, code string, warehouseID uuid.UUID) (*entity.WarehouseLocation, error) {
	args := m.Called(ctx, code, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseLocation), args.Error(1)
}

func (m *MockWarehouseLocationRepository) Update(ctx context.Context, id uuid.UUID, location *entity.WarehouseLocation) error {
	args := m.Called(ctx, id, location)
	return args.Error(0)
}

func (m *MockWarehouseLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWarehouseLocationRepository) UpdateUsage(ctx context.Context, id uuid.UUID, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockWarehouseLocationRepository) IsCodeExists(ctx context.Context, code string, warehouseID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, code, warehouseID, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWarehouseLocationRepository) GetAvailableLocations(ctx context.Context, warehouseID uuid.UUID, requiredCapacity int) ([]entity.WarehouseLocation, error) {
	args := m.Called(ctx, warehouseID, requiredCapacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WarehouseLocation), args.Error(1)
}

// InventoryServiceInstance returns a testable InventoryService with mocks
func InventoryServiceInstance() (*InventoryService, *MockInventoryRepository, *MockProductRepository, *MockWarehouseRepository, *MockWarehouseLocationRepository) {
	mockInventoryRepo := new(MockInventoryRepository)
	mockProductRepo := new(MockProductRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)
	mockLocationRepo := new(MockWarehouseLocationRepository)

	inventoryRepoInterface := interface{}(mockInventoryRepo).(repository.IInventoryRepository)
	productRepoInterface := interface{}(mockProductRepo).(repository.IProductRepository)
	warehouseRepoInterface := interface{}(mockWarehouseRepo).(repository.IWarehouseRepository)
	locationRepoInterface := interface{}(mockLocationRepo).(repository.IWarehouseLocationRepository)

	svc := &InventoryService{
		inventoryRepo: inventoryRepoInterface,
		productRepo:   productRepoInterface,
		warehouseRepo: warehouseRepoInterface,
		locationRepo:  locationRepoInterface,
	}

	return svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo
}

// Helper function to create test warehouse
func createTestWarehouse() *entity.Warehouse {
	ID := uuid.New()
	address := "123 Test St"
	return &entity.Warehouse{
		ID:        ID,
		Code:      "WH-001",
		Name:      "Test Warehouse",
		Address:   &address,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper function to create test warehouse location
func createTestLocation() *entity.WarehouseLocation {
	ID := uuid.New()
	desc := "Test location description"
	warehouseID := uuid.New()
	return &entity.WarehouseLocation{
		ID:          ID,
		Code:        "LOC-001",
		Name:        "Test Location",
		Description: &desc,
		WarehouseID: warehouseID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Helper function to create test stock movement request
func createTestStockMovementRequest() dto.StockMovementRequest {
	warehouseID := uuid.New()
	return dto.StockMovementRequest{
		MovementType:  "IN",
		ProductID:     uuid.New(),
		ToWarehouseID: &warehouseID,
		Quantity:      100,
		UnitPrice:     decimal.NewFromFloat(10.00),
		Notes:         strPtr("Test movement"),
		MovementDate:  time.Now(),
	}
}

func TestInventoryService_ValidateStockMovement_StockIn_Success(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	warehouse := createTestWarehouse()
	req := createTestStockMovementRequest()
	req.MovementType = "IN"

	mockWarehouseRepo.On("FindByID", ctx, *req.ToWarehouseID).Return(warehouse, nil)

	err := svc.validateStockMovement(ctx, req)

	assert.NoError(t, err)
	mockWarehouseRepo.AssertExpectations(t)
	_ = mockInventoryRepo
	_ = mockProductRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_StockIn_MissingDestination(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	req := createTestStockMovementRequest()
	req.MovementType = "IN"
	req.ToWarehouseID = nil

	err := svc.validateStockMovement(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errpkg.ErrMissingDestinationWarehouse, err)
	_ = mockInventoryRepo
	_ = mockProductRepo
	_ = mockWarehouseRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_StockOut_Success(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	warehouse := createTestWarehouse()
	warehouseID := uuid.New()
	productID := uuid.New()

	req := dto.StockMovementRequest{
		MovementType:    "OUT",
		ProductID:       productID,
		FromWarehouseID: &warehouseID,
		Quantity:        50,
		UnitPrice:       decimal.NewFromFloat(10.00),
	}

	stock := &entity.InventoryStock{
		ID:                uuid.New(),
		ProductID:         productID,
		WarehouseID:       warehouseID,
		AvailableQuantity: 100,
	}

	mockWarehouseRepo.On("FindByID", ctx, warehouseID).Return(warehouse, nil)
	mockInventoryRepo.On("GetStockByProductAndWarehouse", ctx, productID, warehouseID, (*uuid.UUID)(nil)).Return(stock, nil)

	err := svc.validateStockMovement(ctx, req)

	assert.NoError(t, err)
	mockWarehouseRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockProductRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_StockOut_InsufficientStock(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	warehouse := createTestWarehouse()
	warehouseID := uuid.New()
	productID := uuid.New()

	req := dto.StockMovementRequest{
		MovementType:    "OUT",
		ProductID:       productID,
		FromWarehouseID: &warehouseID,
		Quantity:        150, // More than available
		UnitPrice:       decimal.NewFromFloat(10.00),
	}

	stock := &entity.InventoryStock{
		ID:                uuid.New(),
		ProductID:         productID,
		WarehouseID:       warehouseID,
		AvailableQuantity: 100,
	}

	mockWarehouseRepo.On("FindByID", ctx, warehouseID).Return(warehouse, nil)
	mockInventoryRepo.On("GetStockByProductAndWarehouse", ctx, productID, warehouseID, (*uuid.UUID)(nil)).Return(stock, nil)

	err := svc.validateStockMovement(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errpkg.ErrInsufficientStock, err)
	mockWarehouseRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockProductRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_Transfer_MissingWarehouses(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	req := dto.StockMovementRequest{
		MovementType: "TRANSFER",
		ProductID:    uuid.New(),
		Quantity:     50,
		UnitPrice:    decimal.NewFromFloat(10.00),
	}

	err := svc.validateStockMovement(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errpkg.ErrMissingWarehouseForTransfer, err)
	_ = mockInventoryRepo
	_ = mockProductRepo
	_ = mockWarehouseRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_Transfer_SameWarehouse(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	warehouseID := uuid.New()
	req := dto.StockMovementRequest{
		MovementType:    "TRANSFER",
		ProductID:       uuid.New(),
		FromWarehouseID: &warehouseID,
		ToWarehouseID:   &warehouseID, // Same warehouse
		Quantity:        50,
		UnitPrice:       decimal.NewFromFloat(10.00),
	}

	err := svc.validateStockMovement(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errpkg.ErrSameWarehouseTransfer, err)
	_ = mockInventoryRepo
	_ = mockProductRepo
	_ = mockWarehouseRepo
	_ = mockLocationRepo
}

func TestInventoryService_ValidateStockMovement_InvalidPrice(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	warehouseID := uuid.New()
	req := dto.StockMovementRequest{
		MovementType:  "IN",
		ProductID:     uuid.New(),
		ToWarehouseID: &warehouseID,
		Quantity:      50,
		UnitPrice:     decimal.NewFromFloat(-10.00), // Negative price
	}

	// Even with invalid price, FindByID is still called
	warehouse := createTestWarehouse()
	mockWarehouseRepo.On("FindByID", ctx, *req.ToWarehouseID).Return(warehouse, nil)

	err := svc.validateStockMovement(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errpkg.ErrInvalidPrice, err)
	mockWarehouseRepo.AssertExpectations(t)
	_ = mockInventoryRepo
	_ = mockProductRepo
	_ = mockLocationRepo
}

func TestInventoryService_GetStockByProduct_Success(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	productID := uuid.New()
	stocks := []entity.InventoryStock{
		{ID: uuid.New(), ProductID: productID, Quantity: 50},
		{ID: uuid.New(), ProductID: productID, Quantity: 30},
	}

	mockInventoryRepo.On("GetProductStockByProductID", ctx, productID).Return(stocks, nil)

	result, err := svc.GetStockByProduct(ctx, productID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockInventoryRepo.AssertExpectations(t)
	_ = mockProductRepo
	_ = mockWarehouseRepo
	_ = mockLocationRepo
}

func TestInventoryService_AdjustStock_ProductNotFound(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	productID := uuid.New()
	warehouseID := uuid.New()

	req := dto.StockAdjustmentRequest{
		ProductID:   productID,
		WarehouseID: warehouseID,
		NewQuantity: 100,
		Reason:      "Stock adjustment",
	}

	mockProductRepo.On("FindByID", ctx, productID).Return(nil, errpkg.ErrProductNotFound)

	result, err := svc.AdjustStock(ctx, req, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product not found")
	mockProductRepo.AssertExpectations(t)
	_ = mockInventoryRepo
	_ = mockWarehouseRepo
	_ = mockLocationRepo
}

func TestInventoryService_AdjustStock_WarehouseNotFound(t *testing.T) {
	svc, mockInventoryRepo, mockProductRepo, mockWarehouseRepo, mockLocationRepo := InventoryServiceInstance()
	ctx := context.Background()

	productID := uuid.New()
	warehouseID := uuid.New()

	req := dto.StockAdjustmentRequest{
		ProductID:   productID,
		WarehouseID: warehouseID,
		NewQuantity: 100,
		Reason:      "Stock adjustment",
	}

	product := createTestProduct()
	mockProductRepo.On("FindByID", ctx, productID).Return(product, nil)
	mockWarehouseRepo.On("FindByID", ctx, warehouseID).Return(nil, errpkg.ErrWarehouseNotFound)

	result, err := svc.AdjustStock(ctx, req, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "warehouse not found")
	mockProductRepo.AssertExpectations(t)
	mockWarehouseRepo.AssertExpectations(t)
	_ = mockInventoryRepo
	_ = mockLocationRepo
}
