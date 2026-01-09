// internal/service/monitoring.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"context"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IMonitoringService interface {
	GetSystemHealth(ctx context.Context) (*entity.SystemHealth, error)
	GetServiceMetrics(ctx context.Context) (*entity.ServiceMetrics, error)
	GetInventoryAlerts(ctx context.Context, filter dto.AlertFilter) ([]entity.InventoryAlert, int, error)
	GetAuditLogs(ctx context.Context, filter dto.AuditLogFilter) ([]entity.AuditLog, int, error)
	GetPerformanceReport(ctx context.Context, req dto.PerformanceReportRequest) (*entity.PerformanceReport, error)
	GetMonitoringDashboard(ctx context.Context) (*dto.MonitoringDashboardResponse, error)
	MarkAlertAsRead(ctx context.Context, alertID uuid.UUID) error
}

type MonitoringService struct{}

func NewMonitoringService() IMonitoringService {
	return &MonitoringService{}
}

func (s *MonitoringService) GetSystemHealth(ctx context.Context) (*entity.SystemHealth, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	components := []entity.ComponentHealth{
		{
			Name:      "database",
			Status:    "up",
			LatencyMs: 5,
			LastCheck: time.Now(),
		},
		{
			Name:      "cache",
			Status:    "up",
			LatencyMs: 1,
			LastCheck: time.Now(),
		},
		{
			Name:      "message_queue",
			Status:    "up",
			LatencyMs: 2,
			LastCheck: time.Now(),
		},
	}

	resourceUsage := entity.ResourceUsage{
		CPUUsage:      getCPUUsage(),
		MemoryUsage:   float64(m.Alloc) / float64(m.Sys) * 100,
		MemoryUsedMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB: float64(m.Sys) / 1024 / 1024,
		DiskUsage:     45.5,
		DiskUsedGB:    12.5,
		DiskTotalGB:   100,
		ActiveConns:   10,
		RequestPerSec: 25.5,
		AvgResponseMs: 45.2,
	}

	health := &entity.SystemHealth{
		Status:        "healthy",
		Version:       "1.0.0",
		Uptime:        getUptime(),
		Timestamp:     time.Now(),
		Components:    components,
		ResourceUsage: resourceUsage,
	}

	return health, nil
}

func (s *MonitoringService) GetServiceMetrics(ctx context.Context) (*entity.ServiceMetrics, error) {
	metrics := &entity.ServiceMetrics{
		ServiceName:     "inventory-api",
		RequestCount:    15000,
		ErrorCount:      45,
		SuccessRate:     99.7,
		AvgLatencyMs:    45.5,
		P95LatencyMs:    120.5,
		P99LatencyMs:    250.0,
		ThroughputRps:   25.5,
		ActiveInstances: 1,
		LastUpdated:     time.Now(),
		EndpointMetrics: []entity.EndpointMetrics{
			{
				Endpoint:     "/api/v1/products",
				Method:       "GET",
				RequestCount: 5000,
				ErrorCount:   10,
				AvgLatencyMs: 25.5,
				MinLatencyMs: 5.0,
				MaxLatencyMs: 150.0,
			},
			{
				Endpoint:     "/api/v1/inventory",
				Method:       "POST",
				RequestCount: 3000,
				ErrorCount:   5,
				AvgLatencyMs: 35.2,
				MinLatencyMs: 10.0,
				MaxLatencyMs: 200.0,
			},
			{
				Endpoint:     "/api/v1/orders",
				Method:       "GET",
				RequestCount: 7000,
				ErrorCount:   30,
				AvgLatencyMs: 55.8,
				MinLatencyMs: 15.0,
				MaxLatencyMs: 300.0,
			},
		},
	}

	return metrics, nil
}

func (s *MonitoringService) GetInventoryAlerts(ctx context.Context, filter dto.AlertFilter) ([]entity.InventoryAlert, int, error) {
	alerts := []entity.InventoryAlert{
		{
			ID:          uuid.New(),
			Type:        "low_stock",
			Severity:    "warning",
			ProductSKU:  "SKU-001",
			ProductName: "Product A",
			CurrentQty:  5,
			Threshold:   10,
			Message:     "Stock level is below minimum threshold",
			IsRead:      false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        "out_of_stock",
			Severity:    "critical",
			ProductSKU:  "SKU-002",
			ProductName: "Product B",
			CurrentQty:  0,
			Threshold:   5,
			Message:     "Product is out of stock",
			IsRead:      false,
			CreatedAt:   time.Now(),
		},
	}

	return alerts, len(alerts), nil
}

func (s *MonitoringService) GetAuditLogs(ctx context.Context, filter dto.AuditLogFilter) ([]entity.AuditLog, int, error) {
	logs := []entity.AuditLog{
		{
			ID:         uuid.New(),
			UserID:     uuid.New(),
			UserName:   "John Doe",
			Action:     "CREATE",
			Resource:   "product",
			ResourceID: "prod-123",
			Details:    "Created new product",
			IPAddress:  "192.168.1.100",
			CreatedAt:  time.Now(),
		},
	}

	return logs, len(logs), nil
}

func (s *MonitoringService) GetPerformanceReport(ctx context.Context, req dto.PerformanceReportRequest) (*entity.PerformanceReport, error) {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if req.StartDate != nil {
		startTime = *req.StartDate
	}
	if req.EndDate != nil {
		endTime = *req.EndDate
	}

	report := &entity.PerformanceReport{
		ReportID:      uuid.New(),
		Period:        req.Period,
		StartTime:     startTime,
		EndTime:       endTime,
		TotalRequests: 15000,
		TotalErrors:   45,
		ErrorRate:     0.3,
		AvgLatencyMs:  45.5,
		MaxLatencyMs:  500.0,
		MinLatencyMs:  5.0,
		ThroughputRps: 25.5,
		Availability:  99.97,
		GeneratedAt:   time.Now(),
	}

	return report, nil
}

func (s *MonitoringService) GetMonitoringDashboard(ctx context.Context) (*dto.MonitoringDashboardResponse, error) {
	health, err := s.GetSystemHealth(ctx)
	if err != nil {
		return nil, err
	}

	metrics, err := s.GetServiceMetrics(ctx)
	if err != nil {
		return nil, err
	}

	alerts, _, err := s.GetInventoryAlerts(ctx, dto.AlertFilter{Limit: 5})
	if err != nil {
		return nil, err
	}

	alertResponses := make([]dto.InventoryAlertResponse, len(alerts))
	for i, alert := range alerts {
		alertResponses[i] = dto.InventoryAlertResponse{
			ID:          alert.ID,
			Type:        alert.Type,
			Severity:    alert.Severity,
			ProductSKU:  alert.ProductSKU,
			ProductName: alert.ProductName,
			CurrentQty:  alert.CurrentQty,
			Threshold:   alert.Threshold,
			Message:     alert.Message,
			IsRead:      alert.IsRead,
			CreatedAt:   alert.CreatedAt,
			ResolvedAt:  alert.ResolvedAt,
		}
	}

	dashboard := &dto.MonitoringDashboardResponse{
		SystemHealth: dto.SystemHealthResponse{
			Status:     health.Status,
			Version:    health.Version,
			Uptime:     health.Uptime,
			Timestamp:  health.Timestamp,
			Components: []dto.ComponentHealthDTO{},
			ResourceUsage: dto.ResourceUsageDTO{
				CPUUsage:      health.ResourceUsage.CPUUsage,
				MemoryUsage:   health.ResourceUsage.MemoryUsage,
				MemoryUsedMB:  health.ResourceUsage.MemoryUsedMB,
				MemoryTotalMB: health.ResourceUsage.MemoryTotalMB,
				DiskUsage:     health.ResourceUsage.DiskUsage,
				DiskUsedGB:    health.ResourceUsage.DiskUsedGB,
				DiskTotalGB:   health.ResourceUsage.DiskTotalGB,
				ActiveConns:   health.ResourceUsage.ActiveConns,
				RequestPerSec: health.ResourceUsage.RequestPerSec,
				AvgResponseMs: health.ResourceUsage.AvgResponseMs,
			},
		},
		ServiceMetrics: dto.ServiceMetricsResponse{
			ServiceName:     metrics.ServiceName,
			RequestCount:    metrics.RequestCount,
			ErrorCount:      metrics.ErrorCount,
			SuccessRate:     metrics.SuccessRate,
			AvgLatencyMs:    metrics.AvgLatencyMs,
			P95LatencyMs:    metrics.P95LatencyMs,
			P99LatencyMs:    metrics.P99LatencyMs,
			ThroughputRps:   metrics.ThroughputRps,
			ActiveInstances: metrics.ActiveInstances,
			LastUpdated:     metrics.LastUpdated,
			Endpoints:       []dto.EndpointMetricsDTO{},
		},
		RecentAlerts:     alertResponses,
		RecentActivities: []dto.ActivityLogResponse{},
		LastUpdated:      time.Now(),
	}

	return dashboard, nil
}

func (s *MonitoringService) MarkAlertAsRead(ctx context.Context, alertID uuid.UUID) error {
	return nil
}

var serviceStartTime = time.Now()

func getUptime() string {
	duration := time.Since(serviceStartTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	return formatDuration(days, hours, minutes)
}

func formatDuration(days, hours, minutes int) string {
	return decimal.NewFromInt(int64(days)).String() + "d " +
		decimal.NewFromInt(int64(hours)).String() + "h " +
		decimal.NewFromInt(int64(minutes)).String() + "m"
}

func getCPUUsage() float64 {
	return 35.5
}
