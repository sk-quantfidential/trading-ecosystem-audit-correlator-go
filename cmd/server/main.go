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
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/adapters"
)

func main() {
	cfg := config.Load()

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize data adapter
	ctx := context.Background()
	dataAdapter, err := adapters.InitializeAndConnect(ctx, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize data adapter")
	}
	defer func() {
		if err := dataAdapter.Disconnect(ctx); err != nil {
			logger.WithError(err).Error("Failed to disconnect data adapter")
		}
	}()

	// Initialize service discovery using data adapter
	serviceDiscovery := infrastructure.NewDataAdapterServiceDiscovery(cfg, dataAdapter, logger)
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
		logger.WithError(err).Fatal("Failed to register service")
	}

	// Start heartbeat in background
	go serviceDiscovery.StartHeartbeat(ctx)

	auditService := services.NewAuditServiceWithDataAdapter(dataAdapter, logger)

	grpcServer := grpcpresentation.NewAuditGRPCServer(cfg, auditService, logger)
	httpServer := setupHTTPServer(cfg, auditService, logger, dataAdapter)

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


func setupHTTPServer(cfg *config.Config, auditService *services.AuditService, logger *logrus.Logger, dataAdapter adapters.DataAdapter) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())

	healthHandler := handlers.NewHealthHandler(logger)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Health)
		v1.GET("/ready", healthHandler.Ready)
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}
}

