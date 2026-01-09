// internal/dto/monitoring.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// SystemHealthResponse represents the response for system health check
type SystemHealthResponse struct {
	Status        string               `json:"status"`
	Version       string               `json:"version"`
	Uptime        string               `json:"uptime"`
	Timestamp     time.Time            `json:"timestamp"`
	Components    []ComponentHealthDTO `json:"components"`
	ResourceUsage ResourceUsageDTO     `json:"resource_usage"`
}

// ComponentHealthDTO represents the health of a component
type ComponentHealthDTO struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// ResourceUsageDTO represents resource usage metrics
type ResourceUsageDTO struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	MemoryUsedMB  float64 `json:"memory_used_mb"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	DiskUsage     float64 `json:"disk_usage"`
	DiskUsedGB    float64 `json:"disk_used_gb"`
	DiskTotalGB   float64 `json:"disk_total_gb"`
	ActiveConns   int     `json:"active_connections"`
	RequestPerSec float64 `json:"requests_per_sec"`
	AvgResponseMs float64 `json:"avg_response_ms"`
}

// ServiceMetricsResponse represents the response for service metrics
type ServiceMetricsResponse struct {
	ServiceName     string               `json:"service_name"`
	RequestCount    int64                `json:"request_count"`
	ErrorCount      int64                `json:"error_count"`
	SuccessRate     float64              `json:"success_rate"`
	AvgLatencyMs    float64              `json:"avg_latency_ms"`
	P95LatencyMs    float64              `json:"p95_latency_ms"`
	P99LatencyMs    float64              `json:"p99_latency_ms"`
	ThroughputRps   float64              `json:"throughput_rps"`
	ActiveInstances int                  `json:"active_instances"`
	LastUpdated     time.Time            `json:"last_updated"`
	Endpoints       []EndpointMetricsDTO `json:"endpoints"`
}

// EndpointMetricsDTO represents metrics for an endpoint
type EndpointMetricsDTO struct {
	Endpoint     string  `json:"endpoint"`
	Method       string  `json:"method"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	MinLatencyMs float64 `json:"min_latency_ms"`
	MaxLatencyMs float64 `json:"max_latency_ms"`
}

// InventoryAlertResponse represents an inventory alert response
type InventoryAlertResponse struct {
	ID          uuid.UUID  `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	ProductSKU  string     `json:"product_sku,omitempty"`
	ProductName string     `json:"product_name,omitempty"`
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	CurrentQty  int        `json:"current_qty"`
	Threshold   int        `json:"threshold"`
	Message     string     `json:"message"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// AlertFilter represents filters for querying alerts
type AlertFilter struct {
	Type     string `json:"type,omitempty" query:"type"`
	Severity string `json:"severity,omitempty" query:"severity"`
	IsRead   *bool  `json:"is_read,omitempty" query:"is_read"`
	Limit    int    `json:"limit" query:"limit"`
	Offset   int    `json:"offset" query:"offset"`
}

// AuditLogResponse represents an audit log response
type AuditLogResponse struct {
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

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty" query:"user_id"`
	Action    string     `json:"action,omitempty" query:"action"`
	Resource  string     `json:"resource,omitempty" query:"resource"`
	StartDate *time.Time `json:"start_date,omitempty" query:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty" query:"end_date"`
	Limit     int        `json:"limit" query:"limit"`
	Offset    int        `json:"offset" query:"offset"`
}

// ActivityLogResponse represents an activity log response
type ActivityLogResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Metadata    string    `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PerformanceReportRequest represents a request for performance report
type PerformanceReportRequest struct {
	Period    string     `json:"period" query:"period"` // hourly, daily, weekly, monthly
	StartDate *time.Time `json:"start_date,omitempty" query:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty" query:"end_date"`
}

// PerformanceReportResponse represents a performance report response
type PerformanceReportResponse struct {
	ReportID      uuid.UUID `json:"report_id"`
	Period        string    `json:"period"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalRequests int64     `json:"total_requests"`
	TotalErrors   int64     `json:"total_errors"`
	ErrorRate     float64   `json:"error_rate"`
	AvgLatencyMs  float64   `json:"avg_latency_ms"`
	MaxLatencyMs  float64   `json:"max_latency_ms"`
	MinLatencyMs  float64   `json:"min_latency_ms"`
	ThroughputRps float64   `json:"throughput_rps"`
	Availability  float64   `json:"availability"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// MonitoringDashboardResponse represents the monitoring dashboard data
type MonitoringDashboardResponse struct {
	SystemHealth     SystemHealthResponse     `json:"system_health"`
	ServiceMetrics   ServiceMetricsResponse   `json:"service_metrics"`
	RecentAlerts     []InventoryAlertResponse `json:"recent_alerts"`
	RecentActivities []ActivityLogResponse    `json:"recent_activities"`
	LastUpdated      time.Time                `json:"last_updated"`
}
