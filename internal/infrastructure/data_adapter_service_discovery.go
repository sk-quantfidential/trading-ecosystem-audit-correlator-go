package infrastructure

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/adapters"
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/models"
)

// DataAdapterServiceDiscovery implements ServiceDiscovery using the audit data adapter
type DataAdapterServiceDiscovery struct {
	config      *config.Config
	dataAdapter adapters.DataAdapter
	logger      *logrus.Logger

	// Service information
	serviceInfo *models.ServiceRegistration

	// Heartbeat management
	heartbeatInterval time.Duration
	heartbeatStop     chan struct{}
}

// NewDataAdapterServiceDiscovery creates a new service discovery using the data adapter
func NewDataAdapterServiceDiscovery(cfg *config.Config, dataAdapter adapters.DataAdapter, logger *logrus.Logger) ServiceDiscovery {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	// Determine host IP
	host := getLocalIP()
	if host == "" {
		host = "localhost"
	}

	// Generate service ID
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	serviceID := fmt.Sprintf("%s-%s-%d", hostname, host, cfg.GRPCPort)

	serviceInfo := &models.ServiceRegistration{
		ID:       serviceID,
		Name:     cfg.ServiceName,
		Version:  cfg.ServiceVersion,
		Host:     host,
		GRPCPort: cfg.GRPCPort,
		HTTPPort: cfg.HTTPPort,
		Status:   "healthy",
		Metadata: map[string]string{
			"environment": cfg.Environment,
			"type":        "audit-correlator",
		},
	}

	return &DataAdapterServiceDiscovery{
		config:            cfg,
		dataAdapter:       dataAdapter,
		logger:            logger,
		serviceInfo:       serviceInfo,
		heartbeatInterval: cfg.HealthCheckInterval,
		heartbeatStop:     make(chan struct{}),
	}
}

// Connect establishes connection to the data adapter
func (d *DataAdapterServiceDiscovery) Connect(ctx context.Context) error {
	// The data adapter should already be connected
	if err := d.dataAdapter.Health(ctx); err != nil {
		return fmt.Errorf("data adapter health check failed: %w", err)
	}

	d.logger.Info("Connected to data adapter service discovery")
	return nil
}

// Disconnect closes the connection and cleans up
func (d *DataAdapterServiceDiscovery) Disconnect(ctx context.Context) error {
	// Stop heartbeat
	close(d.heartbeatStop)

	// Unregister service
	err := d.dataAdapter.UnregisterService(ctx, d.serviceInfo.ID)
	if err != nil {
		d.logger.WithError(err).Warn("Failed to unregister service")
	}

	d.logger.Info("Disconnected from data adapter service discovery")
	return nil
}

// RegisterService registers this service in the data adapter
func (d *DataAdapterServiceDiscovery) RegisterService(ctx context.Context) error {
	d.serviceInfo.LastSeen = time.Now()

	err := d.dataAdapter.RegisterService(ctx, d.serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"service_name": d.serviceInfo.Name,
		"service_id":   d.serviceInfo.ID,
		"host":         d.serviceInfo.Host,
		"grpc_port":    d.serviceInfo.GRPCPort,
		"http_port":    d.serviceInfo.HTTPPort,
	}).Info("Service registered in discovery")

	return nil
}

// DiscoverServices finds all registered instances of a service
func (d *DataAdapterServiceDiscovery) DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error) {
	registrations, err := d.dataAdapter.GetServicesByName(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	// Convert to ServiceInfo format for backward compatibility
	services := make([]ServiceInfo, len(registrations))
	for i, reg := range registrations {
		services[i] = ServiceInfo{
			Name:     reg.Name,
			Version:  reg.Version,
			Host:     reg.Host,
			GRPCPort: reg.GRPCPort,
			HTTPPort: reg.HTTPPort,
			Status:   reg.Status,
			LastSeen: reg.LastSeen,
		}
	}

	d.logger.WithFields(logrus.Fields{
		"service_name":    serviceName,
		"instances_found": len(services),
	}).Debug("Discovered services")

	return services, nil
}

// StartHeartbeat starts the periodic heartbeat to maintain service registration
func (d *DataAdapterServiceDiscovery) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(d.heartbeatInterval)
	defer ticker.Stop()

	d.logger.WithField("interval", d.heartbeatInterval).Info("Starting service discovery heartbeat")

	for {
		select {
		case <-ticker.C:
			if err := d.RegisterService(ctx); err != nil {
				d.logger.WithError(err).Error("Failed to send heartbeat")
			}
		case <-d.heartbeatStop:
			d.logger.Info("Stopping service discovery heartbeat")
			return
		case <-ctx.Done():
			d.logger.Info("Context cancelled, stopping heartbeat")
			return
		}
	}
}

// GetHealthyServices retrieves only healthy services with a specific name
func (d *DataAdapterServiceDiscovery) GetHealthyServices(ctx context.Context, serviceName string) ([]ServiceInfo, error) {
	registrations, err := d.dataAdapter.GetHealthyServices(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy services: %w", err)
	}

	// Convert to ServiceInfo format for backward compatibility
	services := make([]ServiceInfo, len(registrations))
	for i, reg := range registrations {
		services[i] = ServiceInfo{
			Name:     reg.Name,
			Version:  reg.Version,
			Host:     reg.Host,
			GRPCPort: reg.GRPCPort,
			HTTPPort: reg.HTTPPort,
			Status:   reg.Status,
			LastSeen: reg.LastSeen,
		}
	}

	d.logger.WithFields(logrus.Fields{
		"service_name":      serviceName,
		"healthy_instances": len(services),
	}).Debug("Retrieved healthy services")

	return services, nil
}

// UpdateHeartbeat updates the last seen timestamp for this service
func (d *DataAdapterServiceDiscovery) UpdateHeartbeat(ctx context.Context) error {
	return d.dataAdapter.UpdateHeartbeat(ctx, d.serviceInfo.ID)
}

// CleanupStaleServices removes services that haven't been seen for the specified TTL
func (d *DataAdapterServiceDiscovery) CleanupStaleServices(ctx context.Context, ttl time.Duration) (int64, error) {
	return d.dataAdapter.CleanupStaleServices(ctx, ttl)
}

// GetServiceMetrics retrieves metrics for this service
func (d *DataAdapterServiceDiscovery) GetServiceMetrics(ctx context.Context) (*models.ServiceMetrics, error) {
	return d.dataAdapter.GetServiceMetrics(ctx, d.serviceInfo.Name)
}

// UpdateServiceMetrics updates metrics for this service
func (d *DataAdapterServiceDiscovery) UpdateServiceMetrics(ctx context.Context, metrics *models.ServiceMetrics) error {
	metrics.ServiceName = d.serviceInfo.Name
	metrics.InstanceID = d.serviceInfo.ID
	return d.dataAdapter.UpdateServiceMetrics(ctx, metrics)
}

// getLocalIP attempts to determine the local IP address
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}