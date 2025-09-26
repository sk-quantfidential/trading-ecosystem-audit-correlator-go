package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
)

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Details     string    `json:"details"`
}

// RiskMetrics represents risk monitoring metrics
type RiskMetrics struct {
	TotalRiskEvents int64     `json:"total_risk_events"`
	HighRiskAlerts  int64     `json:"high_risk_alerts"`
	LastUpdated     time.Time `json:"last_updated"`
}

// TradingStatus represents trading engine status
type TradingStatus struct {
	ActiveStrategies int       `json:"active_strategies"`
	TotalTrades      int64     `json:"total_trades"`
	LastTradeTime    time.Time `json:"last_trade_time"`
}

// ConnectionStats represents connection pool statistics
type ConnectionStats struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalConnections  int64 `json:"total_connections"`
	FailedConnections int64 `json:"failed_connections"`
}

// ServiceClient interface for all service clients
type ServiceClient interface {
	HealthCheck(ctx context.Context) (HealthStatus, error)
}

// RiskMonitorClient interface for risk monitor service
type RiskMonitorClient interface {
	ServiceClient
	GetRiskMetrics(ctx context.Context) (RiskMetrics, error)
}

// TradingEngineClient interface for trading engine service
type TradingEngineClient interface {
	ServiceClient
	GetTradingStatus(ctx context.Context) (TradingStatus, error)
}

// InterServiceClientManager interface for managing all gRPC clients
type InterServiceClientManager interface {
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	GetRiskMonitorClient(ctx context.Context) (RiskMonitorClient, error)
	GetTradingEngineClient(ctx context.Context) (TradingEngineClient, error)
	GetClientByName(ctx context.Context, serviceName string) (ServiceClient, error)
	DiscoverServices(ctx context.Context) ([]ServiceInfo, error)
	GetConnectionStats() ConnectionStats
}

// ServiceUnavailableError represents an error when a service is not available
type ServiceUnavailableError struct {
	ServiceName string
	Reason      string
}

func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service '%s' is unavailable: %s", e.ServiceName, e.Reason)
}

// IsServiceUnavailableError checks if an error is a ServiceUnavailableError
func IsServiceUnavailableError(err error) bool {
	_, ok := err.(*ServiceUnavailableError)
	return ok
}

// DefaultInterServiceClientManager implements InterServiceClientManager
type DefaultInterServiceClientManager struct {
	config           *config.Config
	serviceDiscovery ServiceDiscovery
	configClient     ConfigurationClient
	logger           *logrus.Logger

	// Connection management
	connections      map[string]*grpc.ClientConn
	connectionsMutex sync.RWMutex
	stats            ConnectionStats
	statsMutex       sync.RWMutex
}

// NewInterServiceClientManager creates a new inter-service client manager
func NewInterServiceClientManager(cfg *config.Config, logger *logrus.Logger) InterServiceClientManager {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	return &DefaultInterServiceClientManager{
		config:           cfg,
		serviceDiscovery: NewServiceDiscovery(cfg, logger),
		configClient:     NewConfigurationClient(cfg, logger),
		logger:           logger,
		connections:      make(map[string]*grpc.ClientConn),
	}
}

// Initialize sets up the client manager and its dependencies
func (m *DefaultInterServiceClientManager) Initialize(ctx context.Context) error {
	// Initialize service discovery
	if err := m.serviceDiscovery.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to service discovery: %w", err)
	}

	// Initialize configuration client
	if err := m.configClient.Connect(ctx); err != nil {
		m.logger.WithError(err).Warn("Failed to connect to configuration service, continuing without it")
	}

	m.logger.Info("Inter-service client manager initialized")
	return nil
}

// Cleanup closes all connections and cleans up resources
func (m *DefaultInterServiceClientManager) Cleanup(ctx context.Context) error {
	m.connectionsMutex.Lock()
	defer m.connectionsMutex.Unlock()

	// Close all gRPC connections
	for serviceName, conn := range m.connections {
		if err := conn.Close(); err != nil {
			m.logger.WithError(err).WithField("service", serviceName).Warn("Failed to close gRPC connection")
		}
	}
	m.connections = make(map[string]*grpc.ClientConn)

	// Cleanup service discovery
	if err := m.serviceDiscovery.Disconnect(ctx); err != nil {
		m.logger.WithError(err).Warn("Failed to disconnect from service discovery")
	}

	// Cleanup configuration client
	if err := m.configClient.Disconnect(ctx); err != nil {
		m.logger.WithError(err).Warn("Failed to disconnect from configuration service")
	}

	m.logger.Info("Inter-service client manager cleaned up")
	return nil
}

// GetRiskMonitorClient returns a client for the risk monitor service
func (m *DefaultInterServiceClientManager) GetRiskMonitorClient(ctx context.Context) (RiskMonitorClient, error) {
	conn, err := m.getServiceConnection(ctx, "risk-monitor")
	if err != nil {
		return nil, err
	}

	return &DefaultRiskMonitorClient{
		conn:   conn,
		logger: m.logger,
	}, nil
}

// GetTradingEngineClient returns a client for the trading engine service
func (m *DefaultInterServiceClientManager) GetTradingEngineClient(ctx context.Context) (TradingEngineClient, error) {
	conn, err := m.getServiceConnection(ctx, "trading-system-engine")
	if err != nil {
		return nil, err
	}

	return &DefaultTradingEngineClient{
		conn:   conn,
		logger: m.logger,
	}, nil
}

// GetClientByName returns a generic client for any service by name
func (m *DefaultInterServiceClientManager) GetClientByName(ctx context.Context, serviceName string) (ServiceClient, error) {
	conn, err := m.getServiceConnection(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return &DefaultServiceClient{
		conn:        conn,
		serviceName: serviceName,
		logger:      m.logger,
	}, nil
}

// DiscoverServices returns all discovered services
func (m *DefaultInterServiceClientManager) DiscoverServices(ctx context.Context) ([]ServiceInfo, error) {
	// Get all services from discovery
	allServices := make([]ServiceInfo, 0)

	serviceNames := []string{"risk-monitor", "trading-system-engine", "test-coordinator"}
	for _, serviceName := range serviceNames {
		services, err := m.serviceDiscovery.DiscoverServices(ctx, serviceName)
		if err != nil {
			m.logger.WithError(err).WithField("service", serviceName).Warn("Failed to discover service")
			continue
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

// GetConnectionStats returns current connection statistics
func (m *DefaultInterServiceClientManager) GetConnectionStats() ConnectionStats {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	m.connectionsMutex.RLock()
	activeConnections := int64(len(m.connections))
	m.connectionsMutex.RUnlock()

	stats := m.stats
	stats.ActiveConnections = activeConnections
	return stats
}

// getServiceConnection gets or creates a gRPC connection to a service
func (m *DefaultInterServiceClientManager) getServiceConnection(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	m.connectionsMutex.Lock()
	defer m.connectionsMutex.Unlock()

	// Check if connection already exists
	if conn, exists := m.connections[serviceName]; exists {
		return conn, nil
	}

	// Discover service endpoint
	services, err := m.serviceDiscovery.DiscoverServices(ctx, serviceName)
	if err != nil {
		m.incrementFailedConnections()
		return nil, &ServiceUnavailableError{
			ServiceName: serviceName,
			Reason:      fmt.Sprintf("service discovery failed: %v", err),
		}
	}

	if len(services) == 0 {
		m.incrementFailedConnections()
		return nil, &ServiceUnavailableError{
			ServiceName: serviceName,
			Reason:      "no instances found",
		}
	}

	// Use the first available service instance
	service := services[0]
	endpoint := fmt.Sprintf("%s:%d", service.Host, service.GRPCPort)

	// Create gRPC connection
	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(m.config.RequestTimeout),
	)
	if err != nil {
		m.incrementFailedConnections()
		return nil, &ServiceUnavailableError{
			ServiceName: serviceName,
			Reason:      fmt.Sprintf("connection failed: %v", err),
		}
	}

	// Store connection
	m.connections[serviceName] = conn
	m.incrementTotalConnections()

	m.logger.WithFields(logrus.Fields{
		"service":  serviceName,
		"endpoint": endpoint,
	}).Info("Established gRPC connection")

	return conn, nil
}

// incrementTotalConnections increments the total connections counter
func (m *DefaultInterServiceClientManager) incrementTotalConnections() {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()
	m.stats.TotalConnections++
}

// incrementFailedConnections increments the failed connections counter
func (m *DefaultInterServiceClientManager) incrementFailedConnections() {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()
	m.stats.FailedConnections++
}

// DefaultServiceClient implements ServiceClient for any service
type DefaultServiceClient struct {
	conn        *grpc.ClientConn
	serviceName string
	logger      *logrus.Logger
}

// HealthCheck performs a health check on the service
func (c *DefaultServiceClient) HealthCheck(ctx context.Context) (HealthStatus, error) {
	healthClient := grpc_health_v1.NewHealthClient(c.conn)

	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: c.serviceName,
	})
	if err != nil {
		return HealthStatus{
			Status:      "unhealthy",
			LastChecked: time.Now(),
			Details:     fmt.Sprintf("health check failed: %v", err),
		}, err
	}

	status := "unhealthy"
	if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
		status = "healthy"
	}

	return HealthStatus{
		Status:      status,
		LastChecked: time.Now(),
		Details:     "Health check successful",
	}, nil
}

// DefaultRiskMonitorClient implements RiskMonitorClient
type DefaultRiskMonitorClient struct {
	conn   *grpc.ClientConn
	logger *logrus.Logger
}

// HealthCheck performs a health check on the risk monitor service
func (c *DefaultRiskMonitorClient) HealthCheck(ctx context.Context) (HealthStatus, error) {
	baseClient := &DefaultServiceClient{
		conn:        c.conn,
		serviceName: "risk-monitor",
		logger:      c.logger,
	}
	return baseClient.HealthCheck(ctx)
}

// GetRiskMetrics retrieves risk metrics from the risk monitor service
func (c *DefaultRiskMonitorClient) GetRiskMetrics(ctx context.Context) (RiskMetrics, error) {
	// This would use the actual protobuf client when available
	// For now, return mock data to satisfy the interface
	return RiskMetrics{
		TotalRiskEvents: 100,
		HighRiskAlerts:  5,
		LastUpdated:     time.Now(),
	}, nil
}

// DefaultTradingEngineClient implements TradingEngineClient
type DefaultTradingEngineClient struct {
	conn   *grpc.ClientConn
	logger *logrus.Logger
}

// HealthCheck performs a health check on the trading engine service
func (c *DefaultTradingEngineClient) HealthCheck(ctx context.Context) (HealthStatus, error) {
	baseClient := &DefaultServiceClient{
		conn:        c.conn,
		serviceName: "trading-system-engine",
		logger:      c.logger,
	}
	return baseClient.HealthCheck(ctx)
}

// GetTradingStatus retrieves trading status from the trading engine service
func (c *DefaultTradingEngineClient) GetTradingStatus(ctx context.Context) (TradingStatus, error) {
	// This would use the actual protobuf client when available
	// For now, return mock data to satisfy the interface
	return TradingStatus{
		ActiveStrategies: 3,
		TotalTrades:      150,
		LastTradeTime:    time.Now(),
	}, nil
}