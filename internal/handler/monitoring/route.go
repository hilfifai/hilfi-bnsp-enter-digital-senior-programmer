// internal/handler/monitoring/route.go
package monitoring

import (
	"api-jellyfish/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterMonitoringRoutes registers the monitoring routes
func RegisterMonitoringRoutes(r *gin.RouterGroup, monitoringService service.IMonitoringService) {
	handler := NewMonitoringHandler(monitoringService)

	monitoring := r.Group("/monitoring")
	{
		// Health and metrics endpoints
		monitoring.GET("/health", handler.GetHealth)
		monitoring.GET("/metrics", handler.GetMetrics)
		monitoring.GET("/dashboard", handler.GetDashboard)

		// Alerts endpoints
		monitoring.GET("/alerts", handler.GetAlerts)
		monitoring.POST("/alerts/:id/read", handler.MarkAlertAsRead)

		// Audit logs endpoint
		monitoring.GET("/audit-logs", handler.GetAuditLogs)

		// Performance report endpoint
		monitoring.GET("/performance", handler.GetPerformanceReport)
	}
}
