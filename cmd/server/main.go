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

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/handlers"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure"
	grpcpresentation "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/presentation/grpc"
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
	httpServer := setupHTTPServer(cfg, auditService, logger)

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


func setupHTTPServer(cfg *config.Config, auditService *services.AuditService, logger *logrus.Logger) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandlerWithAuditService(auditService, logger)
	auditHandler := handlers.NewAuditHandler(auditService, logger)

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
		Handler: router,
	}
}

