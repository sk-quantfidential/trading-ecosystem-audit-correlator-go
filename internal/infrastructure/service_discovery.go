package infrastructure

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/adapters"
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/models"
)

// ServiceInfo is an alias for backward compatibility
// Uses audit-data-adapter-go models for consistency
type ServiceInfo = models.ServiceRegistration

// ServiceDiscovery interface for service registry integration
type ServiceDiscovery interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	RegisterService(ctx context.Context) error
	DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error)
	StartHeartbeat(ctx context.Context)
}

// AdapterServiceDiscovery implements ServiceDiscovery using audit-data-adapter-go DataAdapter
type AdapterServiceDiscovery struct {
	config      *config.Config
	dataAdapter adapters.DataAdapter
	logger      *logrus.Logger

	// Service information
	serviceInfo *models.ServiceRegistration

	// Heartbeat management
	heartbeatInterval time.Duration
	heartbeatStop     chan struct{}
}

// NewServiceDiscovery creates a new DataAdapter-based service discovery client
func NewServiceDiscovery(cfg *config.Config, logger *logrus.Logger) ServiceDiscovery {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	// Determine host IP
	host := getLocalIP()
	if host == "" {
		host = "localhost"
	}

	// Create service registration
	serviceInfo := &models.ServiceRegistration{
		ID:       cfg.ServiceInstanceName, // Use instance name directly as ID
		Name:     cfg.ServiceName,
		Version:  cfg.ServiceVersion,
		Host:     host,
		GRPCPort: cfg.GRPCPort,
		HTTPPort: cfg.HTTPPort,
		Status:   "healthy",
		Metadata: map[string]string{
			"environment":   cfg.Environment,
			"log_level":     cfg.LogLevel,
			"service_type":  cfg.ServiceName,         // Service type (e.g., "audit-correlator")
			"instance_name": cfg.ServiceInstanceName, // Instance identifier (e.g., "audit-correlator")
		},
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	return &AdapterServiceDiscovery{
		config:            cfg,
		dataAdapter:       cfg.GetDataAdapter(),
		logger:            logger,
		heartbeatInterval: cfg.HealthCheckInterval,
		heartbeatStop:     make(chan struct{}),
		serviceInfo:       serviceInfo,
	}
}

// Connect establishes connection to service discovery backend
func (sd *AdapterServiceDiscovery) Connect(ctx context.Context) error {
	if sd.dataAdapter == nil {
		sd.logger.Info("ServiceDiscovery: No DataAdapter available - using stub mode")
		return nil
	}

	// DataAdapter connection is handled at the config level
	sd.logger.Info("ServiceDiscovery: Connected via DataAdapter")
	return nil
}

// Disconnect closes connection to service discovery backend
func (sd *AdapterServiceDiscovery) Disconnect(ctx context.Context) error {
	// Signal heartbeat to stop
	select {
	case sd.heartbeatStop <- struct{}{}:
	default:
	}

	if sd.dataAdapter == nil {
		sd.logger.Info("ServiceDiscovery: Disconnect in stub mode")
		return nil
	}

	// Unregister service
	if err := sd.dataAdapter.UnregisterService(ctx, sd.serviceInfo.ID); err != nil {
		sd.logger.WithError(err).Warn("Failed to unregister service")
	}

	sd.logger.Info("ServiceDiscovery: Disconnected from DataAdapter")
	return nil
}

// RegisterService registers this service instance
func (sd *AdapterServiceDiscovery) RegisterService(ctx context.Context) error {
	if sd.dataAdapter == nil {
		sd.logger.WithField("service", sd.serviceInfo.Name).Info("ServiceDiscovery: RegisterService in stub mode")
		return nil
	}

	// Update last seen timestamp
	sd.serviceInfo.LastSeen = time.Now()

	// Register via DataAdapter
	if err := sd.dataAdapter.RegisterService(ctx, sd.serviceInfo); err != nil {
		sd.logger.WithError(err).Error("Failed to register service via DataAdapter")
		return fmt.Errorf("failed to register service: %w", err)
	}

	sd.logger.WithFields(logrus.Fields{
		"service_id":   sd.serviceInfo.ID,
		"service_name": sd.serviceInfo.Name,
		"host":         sd.serviceInfo.Host,
		"grpc_port":    sd.serviceInfo.GRPCPort,
		"http_port":    sd.serviceInfo.HTTPPort,
	}).Info("Service registered successfully via DataAdapter")

	return nil
}

// DiscoverServices finds services by name
func (sd *AdapterServiceDiscovery) DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error) {
	if sd.dataAdapter == nil {
		sd.logger.WithField("serviceName", serviceName).Info("ServiceDiscovery: DiscoverServices in stub mode")
		return []ServiceInfo{}, nil
	}

	// Discover services via DataAdapter
	services, err := sd.dataAdapter.GetServicesByName(ctx, serviceName)
	if err != nil {
		sd.logger.WithError(err).Error("Failed to discover services via DataAdapter")
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	// Convert to ServiceInfo slice (they're aliases, so this is a direct conversion)
	var result []ServiceInfo
	for _, service := range services {
		result = append(result, *service)
	}

	sd.logger.WithFields(logrus.Fields{
		"service_name":    serviceName,
		"instances_found": len(result),
	}).Debug("Discovered services via DataAdapter")

	return result, nil
}

// StartHeartbeat begins heartbeat process
func (sd *AdapterServiceDiscovery) StartHeartbeat(ctx context.Context) {
	sd.logger.WithField("interval", sd.heartbeatInterval).Info("ServiceDiscovery: Starting heartbeat")

	go func() {
		ticker := time.NewTicker(sd.heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				sd.logger.Info("ServiceDiscovery: Heartbeat stopped due to context cancellation")
				return
			case <-sd.heartbeatStop:
				sd.logger.Info("ServiceDiscovery: Heartbeat stopped due to stop signal")
				return
			case <-ticker.C:
				if sd.dataAdapter == nil {
					sd.logger.Debug("ServiceDiscovery: Heartbeat tick in stub mode")
					continue
				}

				// Update heartbeat via DataAdapter
				if err := sd.dataAdapter.UpdateHeartbeat(ctx, sd.serviceInfo.ID); err != nil {
					sd.logger.WithError(err).Error("Failed to update heartbeat via DataAdapter")
				} else {
					sd.logger.Debug("ServiceDiscovery: Heartbeat updated via DataAdapter")
				}
			}
		}
	}()
}

// getLocalIP attempts to determine the local IP address
func getLocalIP() string {
	// Try to get hostname first
	hostname, err := os.Hostname()
	if err == nil && hostname != "" && hostname != "localhost" {
		return hostname
	}

	// Fallback to localhost
	return "localhost"
}

// getServiceInstanceID generates a unique instance identifier
func getServiceInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}