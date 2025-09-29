package infrastructure

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
)

// ServiceInfo represents service registration information
// TODO: This will be replaced with audit-data-adapter-go models in Phase 2
type ServiceInfo struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Host     string    `json:"host"`
	GRPCPort int       `json:"grpc_port"`
	HTTPPort int       `json:"http_port"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// ServiceDiscovery interface for service registry integration
// TODO: This will be replaced with audit-data-adapter-go DataAdapter.ServiceDiscoveryRepository in Phase 2
type ServiceDiscovery interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	RegisterService(ctx context.Context) error
	DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error)
	StartHeartbeat(ctx context.Context)
}

// AdapterServiceDiscovery implements ServiceDiscovery using audit-data-adapter-go DataAdapter
// This is a temporary stub that will be properly implemented in Phase 2
type AdapterServiceDiscovery struct {
	config *config.Config
	logger *logrus.Logger

	// Service information
	serviceInfo ServiceInfo

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

	return &AdapterServiceDiscovery{
		config:            cfg,
		logger:            logger,
		heartbeatInterval: 30 * time.Second,
		heartbeatStop:     make(chan struct{}),
		serviceInfo: ServiceInfo{
			Name:     "audit-correlator",
			Version:  "1.0.0",
			Status:   "healthy",
			LastSeen: time.Now(),
		},
	}
}

// Connect establishes connection to service discovery backend
// TODO: Initialize audit-data-adapter-go DataAdapter in Phase 2
func (sd *AdapterServiceDiscovery) Connect(ctx context.Context) error {
	sd.logger.Info("ServiceDiscovery: Connect stub called - will use DataAdapter in Phase 2")
	// Stub implementation - no actual connection yet
	return nil
}

// Disconnect closes connection to service discovery backend
// TODO: Disconnect audit-data-adapter-go DataAdapter in Phase 2
func (sd *AdapterServiceDiscovery) Disconnect(ctx context.Context) error {
	sd.logger.Info("ServiceDiscovery: Disconnect stub called - will use DataAdapter in Phase 2")
	// Signal heartbeat to stop
	select {
	case sd.heartbeatStop <- struct{}{}:
	default:
	}
	return nil
}

// RegisterService registers this service instance
// TODO: Use DataAdapter.ServiceDiscoveryRepository.RegisterService() in Phase 2
func (sd *AdapterServiceDiscovery) RegisterService(ctx context.Context) error {
	sd.logger.WithField("service", sd.serviceInfo.Name).Info("ServiceDiscovery: RegisterService stub called - will use DataAdapter in Phase 2")
	// Stub implementation - no actual registration yet
	return nil
}

// DiscoverServices finds services by name
// TODO: Use DataAdapter.ServiceDiscoveryRepository.DiscoverServices() in Phase 2
func (sd *AdapterServiceDiscovery) DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error) {
	sd.logger.WithField("serviceName", serviceName).Info("ServiceDiscovery: DiscoverServices stub called - will use DataAdapter in Phase 2")
	// Stub implementation - return empty list for now
	return []ServiceInfo{}, nil
}

// StartHeartbeat begins heartbeat process
// TODO: Use DataAdapter.ServiceDiscoveryRepository for heartbeat in Phase 2
func (sd *AdapterServiceDiscovery) StartHeartbeat(ctx context.Context) {
	sd.logger.Info("ServiceDiscovery: StartHeartbeat stub called - will use DataAdapter in Phase 2")

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
				sd.logger.Debug("ServiceDiscovery: Heartbeat stub tick - will use DataAdapter in Phase 2")
				// Stub implementation - no actual heartbeat yet
			}
		}
	}()
}