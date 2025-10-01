package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
)

type HealthHandler struct {
	logger       *logrus.Logger
	auditService *services.AuditService
}

func NewHealthHandler(logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

func NewHealthHandlerWithAuditService(auditService *services.AuditService, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger:       logger,
		auditService: auditService,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "audit-correlator",
		"version": "1.0.0",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := gin.H{
		"service": "ready",
	}

	// Add audit service health status if available
	if h.auditService != nil {
		auditStatus := h.auditService.GetHealthStatus()
		for key, value := range auditStatus {
			checks[key] = value
		}
	} else {
		checks["audit_service"] = "not_configured"
	}

	status := http.StatusOK

	// Check if any component is unhealthy
	for _, value := range checks {
		if value == "unavailable" || value == "error" {
			status = http.StatusServiceUnavailable
			break
		}
	}

	c.JSON(status, gin.H{
		"status": func() string {
			if status == http.StatusOK {
				return "ready"
			}
			return "not_ready"
		}(),
		"checks": checks,
	})
}