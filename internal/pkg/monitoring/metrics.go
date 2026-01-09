// internal/pkg/monitoring/metrics.go
package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Business metrics
	productsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created",
		},
	)

	ordersProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total number of orders processed",
		},
		[]string{"type", "status"},
	)

	stockMovements = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stock_movements_total",
			Help: "Total number of stock movements",
		},
		[]string{"type", "direction"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"query_type"},
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, path, getStatusGroup(status)).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordProductCreation records a product creation metric
func RecordProductCreation() {
	productsCreated.Inc()
}

// RecordOrderProcessed records an order processed metric
func RecordOrderProcessed(orderType, status string) {
	ordersProcessed.WithLabelValues(orderType, status).Inc()
}

// RecordStockMovement records a stock movement metric
func RecordStockMovement(movementType, direction string) {
	stockMovements.WithLabelValues(movementType, direction).Inc()
}

// RecordDBQuery records a database query duration
func RecordDBQuery(queryType string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}

func getStatusGroup(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "other"
	}
}
