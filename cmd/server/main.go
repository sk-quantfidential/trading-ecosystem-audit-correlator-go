package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	auditv1connect "github.com/quantfidential/trading-ecosystem/audit-correlator-go/gen/go/audit/v1/auditv1connect"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/handlers"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure/observability"
	connectpresentation "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/presentation/connect"
	grpcpresentation "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/presentation/grpc"
	grpcservices "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/presentation/grpc/services"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
)

func main() {
	cfg := config.Load()

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Add instance context to all logs
	logger = logger.WithFields(logrus.Fields{
		"service_name":  cfg.ServiceName,
		"instance_name": cfg.ServiceInstanceName,
		"environment":   cfg.Environment,
	}).Logger

	logger.Info("Starting audit-correlator service")

	// Initialize data adapter at config level
	ctx := context.Background()
	if err := cfg.InitializeDataAdapter(ctx, logger); err != nil {
		logger.WithError(err).Warn("Failed to initialize DataAdapter - continuing with stub mode")
	}
	defer func() {
		if err := cfg.DisconnectDataAdapter(ctx); err != nil {
			logger.WithError(err).Error("Failed to disconnect DataAdapter")
		}
	}()

	// Initialize service discovery (will use DataAdapter if available)
	serviceDiscovery := infrastructure.NewServiceDiscovery(cfg, logger)
	if err := serviceDiscovery.Connect(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to connect service discovery")
	}
	defer func() {
		if err := serviceDiscovery.Disconnect(ctx); err != nil {
			logger.WithError(err).Error("Failed to disconnect service discovery")
		}
	}()

	// Register service and start heartbeat
	if err := serviceDiscovery.RegisterService(ctx); err != nil {
		logger.WithError(err).Warn("Failed to register service - continuing in stub mode")
	}

	// Start heartbeat in background
	go serviceDiscovery.StartHeartbeat(ctx)

	// Initialize audit service (will use DataAdapter if available)
	var auditService *services.AuditService
	if dataAdapter := cfg.GetDataAdapter(); dataAdapter != nil {
		auditService = services.NewAuditServiceWithDataAdapter(dataAdapter, logger)
	} else {
		auditService = services.NewAuditService(logger)
	}

	grpcServer := grpcpresentation.NewAuditGRPCServer(cfg, auditService, logger)
	httpServer := setupHTTPServer(cfg, auditService, grpcServer, logger)

	go func() {
		logger.WithField("port", cfg.GRPCPort).Info("Starting gRPC server")
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			logger.WithError(err).Fatal("Failed to listen on gRPC port")
		}
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithError(err).Fatal("Failed to start gRPC server")
		}
	}()

	go func() {
		logger.WithField("port", cfg.HTTPPort).Info("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server forced to shutdown")
	}

	grpcServer.GracefulStop()
	logger.Info("Servers shutdown complete")
}


func setupHTTPServer(cfg *config.Config, auditService *services.AuditService, grpcServer *grpcpresentation.AuditGRPCServer, logger *logrus.Logger) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())

	// Add CORS middleware for Connect protocol (browser requests)
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, X-Client, X-Client-Version")
		c.Header("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Initialize observability (Clean Architecture: port + adapter)
	constantLabels := map[string]string{
		"service":  cfg.ServiceName,
		"instance": cfg.ServiceInstanceName,
		"version":  cfg.ServiceVersion,
	}
	metricsPort := observability.NewPrometheusMetricsAdapter(constantLabels)

	// Add RED metrics middleware (Rate, Errors, Duration)
	router.Use(observability.REDMetricsMiddleware(metricsPort))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandlerWithConfig(cfg, auditService, logger)
	auditHandler := handlers.NewAuditHandler(auditService, logger)
	metricsHandler := handlers.NewMetricsHandler(metricsPort)

	// Register Connect protocol handlers (for browser gRPC-Web/Connect clients)
	registerConnectHandlers(router, grpcServer, logger)

	// Observability endpoints (separate from business logic)
	router.GET("/metrics", metricsHandler.Metrics)

	v1 := router.Group("/api/v1")
	{
		// Health endpoints
		v1.GET("/health", healthHandler.Health)
		v1.GET("/ready", healthHandler.Ready)

		// Audit endpoints
		audit := v1.Group("/audit")
		{
			audit.POST("/events", auditHandler.LogEvent)
			audit.GET("/correlations", auditHandler.CorrelateEvents)
			audit.GET("/events/trace/:trace_id", auditHandler.GetEventsByTraceID)
			audit.GET("/events/service", auditHandler.GetEventsByServiceType)
			audit.POST("/correlations", auditHandler.CreateCorrelation)
			audit.GET("/status", auditHandler.GetAuditStatus)
		}
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: h2c.NewHandler(router, &http2.Server{}), // Enable HTTP/2 for Connect protocol
	}
}

// registerConnectHandlers registers Connect protocol handlers for browser-based gRPC clients
func registerConnectHandlers(router *gin.Engine, grpcServer *grpcpresentation.AuditGRPCServer, logger *logrus.Logger) {
	// Create a new TopologyService instance (matches pattern from grpc/server.go)
	topologyService := services.NewTopologyService(logger)

	// Load topology configuration from file (if exists)
	configPath := "/app/config/topology.json"
	if err := topologyService.LoadConfigFromFile(configPath); err != nil {
		logger.WithError(err).Warn("Failed to load topology config, starting with empty topology")
	}

	topologyServer := grpcservices.NewTopologyServiceServer(topologyService, logger)

	// Create Connect adapter
	connectAdapter := connectpresentation.NewTopologyConnectAdapter(topologyServer)

	// Generate Connect HTTP handler
	path, handler := auditv1connect.NewTopologyServiceHandler(connectAdapter)

	// Register with Gin router
	// The path will be "/audit.v1.TopologyService/" and we need to handle all sub-paths
	router.Any(path+"*method", gin.WrapH(handler))

	logger.WithField("path", path).Info("Registered Connect protocol handlers for TopologyService")
}

