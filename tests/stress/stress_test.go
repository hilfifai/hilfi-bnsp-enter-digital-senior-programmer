package stress

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"api-jellyfish/internal/config"
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"api-jellyfish/internal/repository"
	"api-jellyfish/internal/service"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Test configuration
var (
	testDB       *sqlx.DB
	testConfig   *config.AppConfig
	testSvc      service.IProductService
	testCategory *entity.ProductCategory
	testCtx      context.Context
)

// Metrics for stress testing
type StressMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDurationNs int64
	RequestsPerSec  float64
	AvgLatencyMs    float64
	MinLatencyNs    int64
	MaxLatencyNs    int64
	Errors          map[string]int64
	ErrorsMu        sync.Mutex
}

func init() {
	testCtx = context.Background()
}

// SetupTestEnvironment prepares the test environment
func SetupTestEnvironment(t *testing.T) func() {
	t.Helper()

	var cleanupFunc func()

	t.Log("Setting up stress test environment...")

	// Load .env file from project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Log("Could not get current working directory, using environment variables")
	} else {
		// Try multiple possible locations for .env file
		envPaths := []string{
			filepath.Join(cwd, "..", "..", ".env"),
			filepath.Join(cwd, "..", "..", "sistem_inventaris", ".env"),
			filepath.Join(cwd, ".env"),
		}

		for _, envPath := range envPaths {
			if _, err := os.Stat(envPath); err == nil {
				err = godotenv.Load(envPath)
				if err == nil {
					t.Logf("Loaded .env file from: %s", envPath)
					break
				}
			}
		}
	}

	// Load configuration
	config.LoadConfig()
	cfg := &config.Config

	if cfg.Database.Host == "" {
		t.Skip("Skipping stress test: database not configured")
	}

	// Connect to database
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Pass,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		Driver:   database.POSTGRES, // Set driver type explicitly
	}

	db, err := database.Setup(dbConfig)
	if err != nil {
		t.Skip("Skipping stress test: database connection failed - ", err.Error())
	}

	sqlxDB, _ := db.GetSqlxDB()
	testDB = sqlxDB

	// Initialize repositories
	productRepo := repository.NewProductRepository(testDB)
	categoryRepo := repository.NewProductCategoryRepository(testDB)
	inventoryRepo := repository.NewInventoryRepository(testDB)

	// Initialize services
	testSvc = service.NewProductService(productRepo, categoryRepo, inventoryRepo)

	// Create or get test category
	testCategory = createTestCategory(t, categoryRepo)

	cleanupFunc = func() {
		t.Log("Cleaning up stress test environment...")
		// Cleanup test data
		cleanupTestData(t, testDB)
		// Close database connection
		db.Close()
	}

	t.Log("Stress test environment ready")
	return cleanupFunc
}

func createTestCategory(t *testing.T, categoryRepo repository.IProductCategoryRepository) *entity.ProductCategory {
	t.Helper()

	code := fmt.Sprintf("STRESS-TEST-%d", time.Now().Unix())
	category, err := categoryRepo.FindByCode(testCtx, code)
	if err == nil {
		return category
	}

	// Create new category with all required fields
	newCategory := &entity.ProductCategory{
		ID:        uuid.New(),
		Code:      code,
		Name:      "Stress Test Category",
		IsActive:  true,
		CreatedBy: uuid.New(), // Required field
		UpdatedBy: uuid.New(), // Required field
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = categoryRepo.Create(testCtx, newCategory)
	if err != nil {
		t.Logf("Failed to create test category: %v", err)
		// Try to find any active category as fallback
		categories, findErr := categoryRepo.FindAll(testCtx, true)
		if findErr == nil && len(categories) > 0 {
			t.Logf("Using fallback category: %s", categories[0].Name)
			return &categories[0]
		}
		t.Skip("Skipping stress test: cannot create test category and no fallback available")
	}

	return newCategory
}

func cleanupTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Delete test products
	_, err := db.ExecContext(testCtx, "DELETE FROM products WHERE sku LIKE 'STRESS-TEST-%' OR sku LIKE 'BENCHMARK-%' OR sku LIKE 'ENDURANCE-%' OR sku LIKE 'METRICS-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test products: %v", err)
	}

	// Delete test categories
	_, err = db.ExecContext(testCtx, "DELETE FROM product_categories WHERE code LIKE 'STRESS-TEST-%' OR code LIKE 'BENCHMARK-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test categories: %v", err)
	}
}

// =============================================================================
// BENCHMARK STRESS TESTS
// =============================================================================

// BenchmarkProductCreation measures product creation performance
func BenchmarkProductCreation(b *testing.B) {
	config.LoadConfig()
	cfg := &config.Config

	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Pass,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		Driver:   database.POSTGRES,
	}

	db, err := database.Setup(dbConfig)
	if err != nil {
		b.Skip("Skipping benchmark: database connection failed")
	}
	defer db.Close()

	sqlxDB, _ := db.GetSqlxDB()

	productRepo := repository.NewProductRepository(sqlxDB)
	categoryRepo := repository.NewProductCategoryRepository(sqlxDB)
	inventoryRepo := repository.NewInventoryRepository(sqlxDB)
	svc := service.NewProductService(productRepo, categoryRepo, inventoryRepo)

	// Create test category
	category := &entity.ProductCategory{
		ID:       uuid.New(),
		Code:     fmt.Sprintf("BENCHMARK-%d", time.Now().UnixNano()),
		Name:     "Benchmark Category",
		IsActive: true,
	}
	categoryRepo.Create(context.Background(), category)
	defer categoryRepo.Delete(context.Background(), category.ID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := generateCreateProductRequest(category.ID)
			svc.CreateProduct(context.Background(), req, uuid.New())
		}
	})
}

// BenchmarkProductRead measures product read performance
func BenchmarkProductRead(b *testing.B) {
	config.LoadConfig()
	cfg := &config.Config

	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Pass,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		Driver:   database.POSTGRES,
	}

	db, err := database.Setup(dbConfig)
	if err != nil {
		b.Skip("Skipping benchmark: database connection failed")
	}
	defer db.Close()

	sqlxDB, _ := db.GetSqlxDB()

	productRepo := repository.NewProductRepository(sqlxDB)
	categoryRepo := repository.NewProductCategoryRepository(sqlxDB)
	inventoryRepo := repository.NewInventoryRepository(sqlxDB)
	svc := service.NewProductService(productRepo, categoryRepo, inventoryRepo)

	// Create test category
	category := &entity.ProductCategory{
		ID:       uuid.New(),
		Code:     fmt.Sprintf("BENCHMARK-READ-%d", time.Now().UnixNano()),
		Name:     "Benchmark Read Category",
		IsActive: true,
	}
	categoryRepo.Create(context.Background(), category)
	defer categoryRepo.Delete(context.Background(), category.ID)

	// Create test products
	var productIDs []uuid.UUID
	for i := 0; i < 100; i++ {
		req := generateCreateProductRequest(category.ID)
		product, _ := svc.CreateProduct(context.Background(), req, uuid.New())
		if product != nil {
			productIDs = append(productIDs, product.ID)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, id := range productIDs {
				svc.GetProductByID(context.Background(), id)
			}
		}
	})
}

// =============================================================================
// CONCURRENT STRESS TESTS
// =============================================================================

// TestProductService_StressConcurrentCreate tests concurrent product creation
func TestProductService_StressConcurrentCreate(t *testing.T) {
	cleanup := SetupTestEnvironment(t)
	defer cleanup()

	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	numGoroutines := 50
	productsPerGoroutine := 20

	var wg sync.WaitGroup
	var metrics StressMetrics
	metrics.Errors = make(map[string]int64)
	metrics.MinLatencyNs = int64(time.Hour) // Initialize with max value

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < productsPerGoroutine; j++ {
				req := generateCreateProductRequest(testCategory.ID)
				req.SKU = fmt.Sprintf("STRESS-TEST-%d-%d-%d", goroutineID, j, time.Now().UnixNano())

				opStart := time.Now()
				_, err := testSvc.CreateProduct(testCtx, req, uuid.New())
				latency := time.Since(opStart)
				latencyNs := latency.Nanoseconds()

				atomic.AddInt64(&metrics.TotalRequests, 1)
				atomic.AddInt64(&metrics.TotalDurationNs, latencyNs)

				if err != nil {
					atomic.AddInt64(&metrics.FailedRequests, 1)
					metrics.ErrorsMu.Lock()
					metrics.Errors[err.Error()]++
					metrics.ErrorsMu.Unlock()
				} else {
					atomic.AddInt64(&metrics.SuccessRequests, 1)
				}

				// Update min latency
				min := atomic.LoadInt64(&metrics.MinLatencyNs)
				if latencyNs < min {
					atomic.CompareAndSwapInt64(&metrics.MinLatencyNs, min, latencyNs)
				}
				// Update max latency
				max := atomic.LoadInt64(&metrics.MaxLatencyNs)
				if latencyNs > max {
					atomic.CompareAndSwapInt64(&metrics.MaxLatencyNs, max, latencyNs)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate metrics
	metrics.RequestsPerSec = float64(metrics.TotalRequests) / totalDuration.Seconds()
	if metrics.TotalRequests > 0 {
		metrics.AvgLatencyMs = float64(atomic.LoadInt64(&metrics.TotalDurationNs)) / float64(metrics.TotalRequests) / 1e6
	}

	// Print results
	t.Logf("=== Stress Test Results: Concurrent Create ===")
	t.Logf("Total Requests: %d", metrics.TotalRequests)
	t.Logf("Success: %d", metrics.SuccessRequests)
	t.Logf("Failed: %d", metrics.FailedRequests)
	t.Logf("Duration: %v", totalDuration)
	t.Logf("Requests/sec: %.2f", metrics.RequestsPerSec)
	t.Logf("Avg Latency: %.2f ms", metrics.AvgLatencyMs)
	t.Logf("Min Latency: %v", time.Duration(atomic.LoadInt64(&metrics.MinLatencyNs)))
	t.Logf("Max Latency: %v", time.Duration(atomic.LoadInt64(&metrics.MaxLatencyNs)))

	// Assertions
	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	assert.Greater(t, successRate, float64(95), "Success rate should be > 95%%")
	assert.Greater(t, metrics.RequestsPerSec, float64(10), "Should handle > 10 req/s")
}

// TestProductService_StressConcurrentRead tests concurrent product reads
func TestProductService_StressConcurrentRead(t *testing.T) {
	cleanup := SetupTestEnvironment(t)
	defer cleanup()

	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// First, create test products
	numProducts := 100
	var productIDs []uuid.UUID

	for i := 0; i < numProducts; i++ {
		req := generateCreateProductRequest(testCategory.ID)
		req.SKU = fmt.Sprintf("STRESS-READ-TEST-%d-%d", i, time.Now().UnixNano())
		product, err := testSvc.CreateProduct(testCtx, req, uuid.New())
		if err == nil && product != nil {
			productIDs = append(productIDs, product.ID)
		}
	}

	if len(productIDs) == 0 {
		t.Skip("Skipping stress test: could not create test products")
	}

	numGoroutines := 100
	readsPerGoroutine := 50

	var wg sync.WaitGroup
	var metrics StressMetrics
	metrics.Errors = make(map[string]int64)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < readsPerGoroutine; j++ {
				productID := productIDs[j%len(productIDs)]

				opStart := time.Now()
				_, err := testSvc.GetProductByID(testCtx, productID)
				latency := time.Since(opStart)

				atomic.AddInt64(&metrics.TotalRequests, 1)
				atomic.AddInt64(&metrics.TotalDurationNs, latency.Nanoseconds())

				if err != nil {
					atomic.AddInt64(&metrics.FailedRequests, 1)
					metrics.ErrorsMu.Lock()
					metrics.Errors[err.Error()]++
					metrics.ErrorsMu.Unlock()
				} else {
					atomic.AddInt64(&metrics.SuccessRequests, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate metrics
	metrics.RequestsPerSec = float64(metrics.TotalRequests) / totalDuration.Seconds()
	if metrics.TotalRequests > 0 {
		metrics.AvgLatencyMs = float64(atomic.LoadInt64(&metrics.TotalDurationNs)) / float64(metrics.TotalRequests) / 1e6
	}

	// Print results
	t.Logf("=== Stress Test Results: Concurrent Read ===")
	t.Logf("Total Requests: %d", metrics.TotalRequests)
	t.Logf("Success: %d", metrics.SuccessRequests)
	t.Logf("Failed: %d", metrics.FailedRequests)
	t.Logf("Duration: %v", totalDuration)
	t.Logf("Requests/sec: %.2f", metrics.RequestsPerSec)
	t.Logf("Avg Latency: %.2f ms", metrics.AvgLatencyMs)

	// Assertions
	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	assert.Greater(t, successRate, float64(99), "Read success rate should be > 99%%")
	assert.Greater(t, metrics.RequestsPerSec, float64(100), "Should handle > 100 req/s")
}

// TestProductService_StressMixedOperations tests mixed read/write operations
func TestProductService_StressMixedOperations(t *testing.T) {
	cleanup := SetupTestEnvironment(t)
	defer cleanup()

	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create initial products
	numInitialProducts := 50
	var productIDs []uuid.UUID

	for i := 0; i < numInitialProducts; i++ {
		req := generateCreateProductRequest(testCategory.ID)
		req.SKU = fmt.Sprintf("STRESS-MIXED-%d-%d", i, time.Now().UnixNano())
		product, err := testSvc.CreateProduct(testCtx, req, uuid.New())
		if err == nil && product != nil {
			productIDs = append(productIDs, product.ID)
		}
	}

	numGoroutines := 30
	opsPerGoroutine := 30

	var wg sync.WaitGroup
	var metrics StressMetrics
	metrics.Errors = make(map[string]int64)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				opType := j % 3 // 0=create, 1=read, 2=read list

				switch opType {
				case 0: // Create
					req := generateCreateProductRequest(testCategory.ID)
					req.SKU = fmt.Sprintf("STRESS-MIXED-CREATE-%d-%d-%d", goroutineID, j, time.Now().UnixNano())
					_, err := testSvc.CreateProduct(testCtx, req, uuid.New())
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}

				case 1: // Read single
					if len(productIDs) > 0 {
						productID := productIDs[j%len(productIDs)]
						_, err := testSvc.GetProductByID(testCtx, productID)
						if err != nil {
							atomic.AddInt64(&metrics.FailedRequests, 1)
						} else {
							atomic.AddInt64(&metrics.SuccessRequests, 1)
						}
					}

				case 2: // Read list
					filter := dto.ProductFilterDTO{
						Page:  1,
						Limit: 10,
					}
					_, _, err := testSvc.GetProducts(testCtx, filter)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}

				atomic.AddInt64(&metrics.TotalRequests, 1)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate metrics
	metrics.RequestsPerSec = float64(metrics.TotalRequests) / totalDuration.Seconds()

	// Print results
	t.Logf("=== Stress Test Results: Mixed Operations ===")
	t.Logf("Total Requests: %d", metrics.TotalRequests)
	t.Logf("Success: %d", metrics.SuccessRequests)
	t.Logf("Failed: %d", metrics.FailedRequests)
	t.Logf("Duration: %v", totalDuration)
	t.Logf("Requests/sec: %.2f", metrics.RequestsPerSec)

	// Assertions
	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	assert.Greater(t, successRate, float64(95), "Success rate should be > 95%%")
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func generateCreateProductRequest(categoryID uuid.UUID) dto.CreateProductRequest {
	desc := "Stress test product description"
	return dto.CreateProductRequest{
		SKU:         fmt.Sprintf("STRESS-TEST-%d", time.Now().UnixNano()),
		Name:        fmt.Sprintf("Stress Test Product %d", time.Now().UnixNano()),
		Description: &desc,
		CategoryID:  categoryID,
		UnitPrice:   decimal.NewFromFloat(99.99),
		CostPrice:   decimal.NewFromFloat(50.00),
		MinStock:    10,
		MaxStock:    1000,
		IsActive:    true,
		Attributes: []dto.ProductAttributeDTO{
			{Key: "color", Value: "red", Label: "Color"},
		},
	}
}

// =============================================================================
// ENDURANCE TEST
// =============================================================================

// TestProductService_EnduranceTest tests system stability over extended period
func TestProductService_EnduranceTest(t *testing.T) {
	cleanup := SetupTestEnvironment(t)
	defer cleanup()

	if testing.Short() {
		t.Skip("Skipping endurance test in short mode")
	}

	// Run for 1 minute with periodic operations
	duration := 1 * time.Minute
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var totalOps int64
	var failedOps int64

	startTime := time.Now()
	for {
		select {
		case <-ticker.C:
			// Perform random operations
			req := generateCreateProductRequest(testCategory.ID)
			req.SKU = fmt.Sprintf("ENDURANCE-%d", time.Now().UnixNano())
			_, err := testSvc.CreateProduct(testCtx, req, uuid.New())

			atomic.AddInt64(&totalOps, 1)
			if err != nil {
				atomic.AddInt64(&failedOps, 1)
				t.Logf("Endurance test ERROR: %v", err)
			}

			// Log progress
			elapsed := time.Since(startTime)
			t.Logf("Endurance test progress: %v elapsed, %d ops, %d failed",
				elapsed, atomic.LoadInt64(&totalOps), atomic.LoadInt64(&failedOps))

		case <-time.After(duration):
			elapsed := time.Since(startTime)
			total := atomic.LoadInt64(&totalOps)
			failed := atomic.LoadInt64(&failedOps)

			t.Logf("=== Endurance Test Results ===")
			t.Logf("Duration: %v", elapsed)
			t.Logf("Total Operations: %d", total)
			t.Logf("Failed Operations: %d", failed)
			t.Logf("Ops/minute: %.2f", float64(total)/elapsed.Minutes())

			// Assertions
			failureRate := float64(failed) / float64(total) * 100
			assert.Less(t, failureRate, float64(1), "Failure rate should be < 1%%")

			return
		}
	}
}

// =============================================================================
// TARGET METRICS VERIFICATION
// =============================================================================

// TestTargetMetricsVerification verifies the system meets target metrics
func TestTargetMetricsVerification(t *testing.T) {
	cleanup := SetupTestEnvironment(t)
	defer cleanup()

	targets := struct {
		CreateOpsPerSec float64
		ReadOpsPerSec   float64
		SuccessRate     float64
		AvgLatencyMs    float64
	}{
		CreateOpsPerSec: 10,
		ReadOpsPerSec:   100,
		SuccessRate:     95,
		AvgLatencyMs:    500,
	}

	t.Logf("=== Target Metrics Verification ===")
	t.Logf("Target Create Ops/sec: %.2f", targets.CreateOpsPerSec)
	t.Logf("Target Read Ops/sec: %.2f", targets.ReadOpsPerSec)
	t.Logf("Target Success Rate: %.2f%%", targets.SuccessRate)
	t.Logf("Target Avg Latency: %.2f ms", targets.AvgLatencyMs)

	// Quick verification test
	numGoroutines := 10
	opsPerGoroutine := 20

	var wg sync.WaitGroup
	var totalDurationNs int64
	var successCount int64
	var totalCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				req := generateCreateProductRequest(testCategory.ID)
				req.SKU = fmt.Sprintf("METRICS-TEST-%d-%d-%d", goroutineID, j, time.Now().UnixNano())

				opStart := time.Now()
				_, err := testSvc.CreateProduct(testCtx, req, uuid.New())
				duration := time.Since(opStart)

				atomic.AddInt64(&totalCount, 1)
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				}

				atomic.AddInt64(&totalDurationNs, duration.Nanoseconds())
			}
		}(i)
	}

	wg.Wait()

	actualOpsPerSec := float64(totalCount) / (float64(totalDurationNs) / 1e9)
	actualSuccessRate := float64(successCount) / float64(totalCount) * 100
	actualAvgLatencyMs := float64(totalDurationNs) / float64(totalCount) / 1e6

	t.Logf("=== Actual Metrics ===")
	t.Logf("Actual Create Ops/sec: %.2f (target: %.2f)", actualOpsPerSec, targets.CreateOpsPerSec)
	t.Logf("Actual Success Rate: %.2f%% (target: %.2f%%)", actualSuccessRate, targets.SuccessRate)
	t.Logf("Actual Avg Latency: %.2f ms (target: %.2f ms)", actualAvgLatencyMs, targets.AvgLatencyMs)

	// Verify targets
	assert.GreaterOrEqual(t, actualOpsPerSec, targets.CreateOpsPerSec,
		"Should meet create ops/sec target")
	assert.GreaterOrEqual(t, actualSuccessRate, targets.SuccessRate,
		"Should meet success rate target")
	assert.LessOrEqual(t, actualAvgLatencyMs, targets.AvgLatencyMs,
		"Should meet latency target")
}

// =============================================================================
// RUN INSTRUCTIONS
// =============================================================================

/*
How to run stress tests:

1. Make sure the database is running:
   docker-compose up -d

2. Run migrations:
   make migrate

3. Run all stress tests:
   go test -v ./tests/stress/... -timeout 30m

4. Run specific stress test:
   go test -v ./tests/stress/... -run TestProductService_StressConcurrentCreate -timeout 10m

5. Run benchmark tests:
   go test -bench=. -benchmem ./tests/stress/... -timeout 10m

6. Run with race detector:
   go test -race -v ./tests/stress/... -timeout 30m

Expected metrics for stress tests:
- Product creation: >10 ops/second
- Product reads: >100 ops/second
- Stock updates: >50 ops/second
- Success rate: >95%
- Average latency: <500ms
*/
