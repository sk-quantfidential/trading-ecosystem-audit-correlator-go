package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
)

type AuditHandler struct {
	auditService *services.AuditService
	logger       *logrus.Logger
}

func NewAuditHandler(auditService *services.AuditService, logger *logrus.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// LogEvent handles audit event logging via HTTP
func (h *AuditHandler) LogEvent(c *gin.Context) {
	var req struct {
		EventType string `json:"event_type" binding:"required"`
		Source    string `json:"source" binding:"required"`
		Message   string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auditService.LogEvent(req.EventType, req.Source, req.Message); err != nil {
		h.logger.WithError(err).Error("Failed to log audit event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log audit event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Audit event logged successfully",
	})
}

// CorrelateEvents handles event correlation requests
func (h *AuditHandler) CorrelateEvents(c *gin.Context) {
	timeWindow := c.Query("time_window")
	if timeWindow == "" {
		timeWindow = "1h" // Default to 1 hour
	}

	correlations, err := h.auditService.CorrelateEvents(timeWindow)
	if err != nil {
		h.logger.WithError(err).Error("Failed to correlate events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to correlate events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"correlations": correlations,
		"count":        len(correlations),
		"time_window":  timeWindow,
	})
}

// GetEventsByTraceID retrieves events for a specific trace ID
func (h *AuditHandler) GetEventsByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trace_id parameter is required"})
		return
	}

	events, err := h.auditService.GetEventsByTraceID(traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get events by trace ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"trace_id": traceID,
		"events":   events,
		"count":    len(events),
	})
}

// GetEventsByServiceType retrieves events for a service and event type
func (h *AuditHandler) GetEventsByServiceType(c *gin.Context) {
	serviceName := c.Query("service_name")
	eventType := c.Query("event_type")
	timeWindowStr := c.Query("time_window")

	if serviceName == "" || eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_name and event_type parameters are required"})
		return
	}

	// Parse time window (default to 1 hour)
	timeWindow := 1 * time.Hour
	if timeWindowStr != "" {
		if parsed, err := time.ParseDuration(timeWindowStr); err == nil {
			timeWindow = parsed
		}
	}

	events, err := h.auditService.GetEventsByServiceType(serviceName, eventType, timeWindow)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get events by service type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"service_name": serviceName,
		"event_type":   eventType,
		"time_window":  timeWindow.String(),
		"events":       events,
		"count":        len(events),
	})
}

// CreateCorrelation creates a correlation between events
func (h *AuditHandler) CreateCorrelation(c *gin.Context) {
	var req struct {
		SourceEventID   string  `json:"source_event_id" binding:"required"`
		TargetEventID   string  `json:"target_event_id" binding:"required"`
		CorrelationType string  `json:"correlation_type" binding:"required"`
		Confidence      float64 `json:"confidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default confidence to 1.0 if not provided
	if req.Confidence == 0 {
		req.Confidence = 1.0
	}

	// Validate confidence range
	if req.Confidence < 0 || req.Confidence > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Confidence must be between 0 and 1"})
		return
	}

	if err := h.auditService.CreateCorrelation(req.SourceEventID, req.TargetEventID, req.CorrelationType, req.Confidence); err != nil {
		h.logger.WithError(err).Error("Failed to create correlation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create correlation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Correlation created successfully",
	})
}

// GetAuditStatus returns the audit service health and statistics
func (h *AuditHandler) GetAuditStatus(c *gin.Context) {
	status := h.auditService.GetHealthStatus()

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"audit_service": status,
	})
}