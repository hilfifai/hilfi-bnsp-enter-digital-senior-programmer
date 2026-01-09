// internal/entity/monitoring.go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// SystemHealth represents the overall system health status
type SystemHealth struct {
	Status        string            `json:"status"` // healthy, degraded, unhealthy
	Version       string            `json:"version"`
	Uptime        string            `json:"uptime"`
	Timestamp     time.Time         `json:"timestamp"`
	Components    []ComponentHealth `json:"components"`
	ResourceUsage ResourceUsage     `json:"resource_usage"`
}

// ComponentHealth represents the health of individual system components
type ComponentHealth struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // up, down, degraded
	LatencyMs int64     `json:"latency_ms"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// ResourceUsage represents system resource usage metrics
type ResourceUsage struct {
	CPUUsage      float64 `json:"cpu_usage"`    // percentage 0-100
	MemoryUsage   float64 `json:"memory_usage"` // percentage 0-100
	MemoryUsedMB  float64 `json:"memory_used_mb"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	DiskUsage     float64 `json:"disk_usage"` // percentage 0-100
	DiskUsedGB    float64 `json:"disk_used_gb"`
	DiskTotalGB   float64 `json:"disk_total_gb"`
	ActiveConns   int     `json:"active_connections"`
	RequestPerSec float64 `json:"requests_per_sec"`
	AvgResponseMs float64 `json:"avg_response_ms"`
}

// ServiceMetrics represents metrics for a specific service
type ServiceMetrics struct {
	ServiceName     string            `json:"service_name"`
	RequestCount    int64             `json:"request_count"`
	ErrorCount      int64             `json:"error_count"`
	SuccessRate     float64           `json:"success_rate"` // percentage
	AvgLatencyMs    float64           `json:"avg_latency_ms"`
	P95LatencyMs    float64           `json:"p95_latency_ms"`
	P99LatencyMs    float64           `json:"p99_latency_ms"`
	ThroughputRps   float64           `json:"throughput_rps"`
	ActiveInstances int               `json:"active_instances"`
	LastUpdated     time.Time         `json:"last_updated"`
	EndpointMetrics []EndpointMetrics `json:"endpoint_metrics"`
}

// EndpointMetrics represents metrics for a specific endpoint
type EndpointMetrics struct {
	Endpoint     string  `json:"endpoint"`
	Method       string  `json:"method"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	MinLatencyMs float64 `json:"min_latency_ms"`
	MaxLatencyMs float64 `json:"max_latency_ms"`
}

// InventoryAlert represents an inventory-related alert
type InventoryAlert struct {
	ID          uuid.UUID  `json:"id"`
	Type        string     `json:"type"`     // low_stock, out_of_stock, expiring, overstock
	Severity    string     `json:"severity"` // critical, warning, info
	ProductID   uuid.UUID  `json:"product_id,omitempty"`
	ProductSKU  string     `json:"product_sku,omitempty"`
	ProductName string     `json:"product_name,omitempty"`
	WarehouseID uuid.UUID  `json:"warehouse_id,omitempty"`
	CurrentQty  int        `json:"current_qty"`
	Threshold   int        `json:"threshold"`
	Message     string     `json:"message"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	UserName   string    `json:"user_name"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Metadata    string    `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PerformanceReport represents a performance report
type PerformanceReport struct {
	ReportID      uuid.UUID `json:"report_id"`
	Period        string    `json:"period"` // hourly, daily, weekly, monthly
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalRequests int64     `json:"total_requests"`
	TotalErrors   int64     `json:"total_errors"`
	ErrorRate     float64   `json:"error_rate"`
	AvgLatencyMs  float64   `json:"avg_latency_ms"`
	MaxLatencyMs  float64   `json:"max_latency_ms"`
	MinLatencyMs  float64   `json:"min_latency_ms"`
	ThroughputRps float64   `json:"throughput_rps"`
	Availability  float64   `json:"availability"` // percentage
	GeneratedAt   time.Time `json:"generated_at"`
}
