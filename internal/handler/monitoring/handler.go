// internal/handler/monitoring/handler.go
package monitoring

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IMonitoringHandler interface {
	GetHealth(c *gin.Context)
	GetMetrics(c *gin.Context)
	GetAlerts(c *gin.Context)
	GetAuditLogs(c *gin.Context)
	GetPerformanceReport(c *gin.Context)
	GetDashboard(c *gin.Context)
	MarkAlertAsRead(c *gin.Context)
}

type MonitoringHandler struct {
	monitoringService service.IMonitoringService
}

func NewMonitoringHandler(monitoringService service.IMonitoringService) IMonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// NewRoutes registers the monitoring routes
func NewRoutes(r *gin.RouterGroup, monitoringService service.IMonitoringService) {
	handler := NewMonitoringHandler(monitoringService)
	RegisterRoutes(r, handler)
}

// GetHealth returns the system health status
// @Summary Get system health status
// @Description Get the current health status of the system
// @Tags monitoring
// @Produce json
// @Success 200 {object} entity.SystemHealth
// @Router /api/v1/monitoring/health [get]
func (h *MonitoringHandler) GetHealth(c *gin.Context) {
	ctx := c.Request.Context()
	health, err := h.monitoringService.GetSystemHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get system health",
		})
		return
	}
	c.JSON(http.StatusOK, health)
}

// GetMetrics returns service metrics
// @Summary Get service metrics
// @Description Get the current metrics of the service
// @Tags monitoring
// @Produce json
// @Success 200 {object} entity.ServiceMetrics
// @Router /api/v1/monitoring/metrics [get]
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	metrics, err := h.monitoringService.GetServiceMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get service metrics",
		})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetAlerts returns inventory alerts
// @Summary Get inventory alerts
// @Description Get inventory alerts with optional filters
// @Tags monitoring
// @Produce json
// @Param type query string false "Alert type"
// @Param severity query string false "Alert severity"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alerts [get]
func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
	var filter dto.AlertFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filter parameters",
		})
		return
	}

	ctx := c.Request.Context()
	alerts, total, err := h.monitoringService.GetInventoryAlerts(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   alerts,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetAuditLogs returns audit logs
// @Summary Get audit logs
// @Description Get audit logs with optional filters
// @Tags monitoring
// @Produce json
// @Param user_id query string false "User ID"
// @Param action query string false "Action type"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/audit-logs [get]
func (h *MonitoringHandler) GetAuditLogs(c *gin.Context) {
	var filter dto.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filter parameters",
		})
		return
	}

	ctx := c.Request.Context()
	logs, total, err := h.monitoringService.GetAuditLogs(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetPerformanceReport returns performance report
// @Summary Get performance report
// @Description Get performance report for a specific period
// @Tags monitoring
// @Produce json
// @Param period query string false "Period (hourly, daily, weekly, monthly)"
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} entity.PerformanceReport
// @Router /api/v1/monitoring/performance [get]
func (h *MonitoringHandler) GetPerformanceReport(c *gin.Context) {
	var req dto.PerformanceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid parameters",
		})
		return
	}

	ctx := c.Request.Context()
	report, err := h.monitoringService.GetPerformanceReport(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get performance report",
		})
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetDashboard returns the monitoring dashboard
// @Summary Get monitoring dashboard
// @Description Get the complete monitoring dashboard data
// @Tags monitoring
// @Produce json
// @Success 200 {object} dto.MonitoringDashboardResponse
// @Router /api/v1/monitoring/dashboard [get]
func (h *MonitoringHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	dashboard, err := h.monitoringService.GetMonitoringDashboard(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get dashboard",
		})
		return
	}
	c.JSON(http.StatusOK, dashboard)
}

// MarkAlertAsRead marks an alert as read
// @Summary Mark alert as read
// @Description Mark a specific alert as read
// @Tags monitoring
// @Accept json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/monitoring/alerts/{id}/read [post]
func (h *MonitoringHandler) MarkAlertAsRead(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	parsedID, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.monitoringService.MarkAlertAsRead(ctx, parsedID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark alert as read",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert marked as read",
	})
}

// Internal function to register routes for the handler
func RegisterRoutes(r *gin.RouterGroup, h IMonitoringHandler) {
	monitoring := r.Group("/monitoring")
	{
		monitoring.GET("/health", h.GetHealth)
		monitoring.GET("/metrics", h.GetMetrics)
		monitoring.GET("/dashboard", h.GetDashboard)
		monitoring.GET("/alerts", h.GetAlerts)
		monitoring.POST("/alerts/:id/read", h.MarkAlertAsRead)
		monitoring.GET("/audit-logs", h.GetAuditLogs)
		monitoring.GET("/performance", h.GetPerformanceReport)
	}
}
