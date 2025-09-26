package grpc

import (
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
)

// AuditGRPCServer implements the gRPC server interface with enhanced health service
type AuditGRPCServer struct {
	config      *config.Config
	server      *grpc.Server
	healthSrv   *health.Server
	auditSvc    *services.AuditService
	logger      *logrus.Logger

	// Metrics tracking
	startTime         time.Time
	activeConnections int64
	totalRequests     int64
	mu                sync.RWMutex
}

// ServerMetrics represents the current server metrics
type ServerMetrics struct {
	ActiveConnections int64             `json:"active_connections"`
	TotalRequests     int64             `json:"total_requests"`
	ServiceStatus     map[string]string `json:"service_status"`
	Uptime            time.Duration     `json:"uptime"`
}

// NewAuditGRPCServer creates a new gRPC server instance with health service
func NewAuditGRPCServer(cfg *config.Config, auditService *services.AuditService, logger *logrus.Logger) *AuditGRPCServer {
	if logger == nil {
		// Create a default logger if none provided (for testing)
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel) // Quiet logging for tests
	}

	if auditService == nil {
		// Create a default audit service if none provided (for testing)
		auditService = services.NewAuditService(logger)
	}

	server := &AuditGRPCServer{
		config:    cfg,
		server:    grpc.NewServer(),
		healthSrv: health.NewServer(),
		auditSvc:  auditService,
		logger:    logger,
		startTime: time.Now(),
	}

	// Register health service
	grpc_health_v1.RegisterHealthServer(server.server, server.healthSrv)

	// Set initial health status
	server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	server.healthSrv.SetServingStatus(cfg.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	return server
}

// Serve starts the gRPC server on the provided listener
func (s *AuditGRPCServer) Serve(lis net.Listener) error {
	s.logger.WithField("address", lis.Addr().String()).Info("Starting gRPC server")

	return s.server.Serve(lis)
}

// GracefulStop gracefully stops the gRPC server
func (s *AuditGRPCServer) GracefulStop() {
	s.logger.Info("Gracefully stopping gRPC server")

	// Update health status to not serving
	s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.healthSrv.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Graceful stop
	s.server.GracefulStop()
	s.logger.Info("gRPC server stopped")
}

// GetMetrics returns current server metrics
func (s *AuditGRPCServer) GetMetrics() ServerMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]string)
	status["health_service"] = "serving"
	status["audit_service"] = "serving"

	return ServerMetrics{
		ActiveConnections: s.activeConnections,
		TotalRequests:     s.totalRequests,
		ServiceStatus:     status,
		Uptime:            time.Since(s.startTime),
	}
}

// incrementConnections tracks new connections
func (s *AuditGRPCServer) incrementConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeConnections++
}

// decrementConnections tracks closed connections
func (s *AuditGRPCServer) decrementConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeConnections--
}

// incrementRequests tracks processed requests
func (s *AuditGRPCServer) incrementRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalRequests++
}