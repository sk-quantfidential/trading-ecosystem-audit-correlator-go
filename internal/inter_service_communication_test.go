//go:build integration

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
)

// TestInterServiceCommunication_RedPhase defines the expected behaviors for inter-service communication
// These tests will fail initially and drive our implementation (TDD Red-Green-Refactor)
func TestInterServiceCommunication_RiskMonitorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("can_communicate_with_risk_monitor", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			ServiceName: "audit-correlator",
		}

		clientManager := NewInterServiceClientManager(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := clientManager.Initialize(ctx)
		if err != nil {
			t.Skip("Inter-service infrastructure not available for test")
		}
		defer clientManager.Cleanup(ctx)

		// Get risk monitor client
		riskClient, err := clientManager.GetRiskMonitorClient(ctx)
		if err != nil {
			t.Errorf("Failed to get risk monitor client: %v", err)
			return
		}

		// Test health check
		health, err := riskClient.HealthCheck(ctx)
		if err != nil {
			t.Errorf("Risk monitor health check failed: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("Expected healthy status, got %s", health.Status)
		}
	})
}

func TestInterServiceCommunication_TradingEngineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("can_communicate_with_trading_engine", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			ServiceName: "audit-correlator",
		}

		clientManager := NewInterServiceClientManager(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := clientManager.Initialize(ctx)
		if err != nil {
			t.Skip("Inter-service infrastructure not available for test")
		}
		defer clientManager.Cleanup(ctx)

		// Get trading engine client
		tradingClient, err := clientManager.GetTradingEngineClient(ctx)
		if err != nil {
			t.Errorf("Failed to get trading engine client: %v", err)
			return
		}

		// Test health check
		health, err := tradingClient.HealthCheck(ctx)
		if err != nil {
			t.Errorf("Trading engine health check failed: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("Expected healthy status, got %s", health.Status)
		}
	})
}

func TestInterServiceCommunication_ServiceDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("discovers_services_dynamically", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			ServiceName: "audit-correlator",
			RedisURL:    "redis://localhost:6379",
		}

		clientManager := NewInterServiceClientManager(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := clientManager.Initialize(ctx)
		if err != nil {
			t.Skip("Service discovery not available for test")
		}
		defer clientManager.Cleanup(ctx)

		// Discover available services
		services, err := clientManager.DiscoverServices(ctx)
		if err != nil {
			t.Errorf("Service discovery failed: %v", err)
		}

		// Should find at least one service (potentially ourselves)
		if len(services) == 0 {
			t.Log("No services discovered - this might be expected in test environment")
		}

		// Verify service info structure
		for _, service := range services {
			if service.Name == "" {
				t.Error("Service name should not be empty")
			}
			if service.GRPCPort == 0 {
				t.Error("Service gRPC port should be set")
			}
		}
	})
}

func TestInterServiceCommunication_ConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("reuses_connections_efficiently", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			ServiceName: "audit-correlator",
		}

		clientManager := NewInterServiceClientManager(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := clientManager.Initialize(ctx)
		if err != nil {
			t.Skip("Inter-service infrastructure not available for test")
		}
		defer clientManager.Cleanup(ctx)

		// Get the same client multiple times
		client1, err := clientManager.GetRiskMonitorClient(ctx)
		if err != nil {
			t.Skip("Risk monitor not available for test")
		}

		client2, err := clientManager.GetRiskMonitorClient(ctx)
		if err != nil {
			t.Errorf("Failed to get second client instance: %v", err)
		}

		// Verify connection reuse (implementation detail - would check internal state)
		stats := clientManager.GetConnectionStats()
		if stats.ActiveConnections == 0 {
			t.Error("Expected at least one active connection")
		}

		if stats.TotalConnections == 0 {
			t.Error("Expected total connections to be tracked")
		}

		// Both clients should work
		_, err = client1.HealthCheck(ctx)
		if err != nil {
			t.Skip("Health check not available for test")
		}

		_, err = client2.HealthCheck(ctx)
		if err != nil {
			t.Errorf("Second client health check failed: %v", err)
		}
	})
}

func TestInterServiceCommunication_ErrorHandling(t *testing.T) {
	t.Run("handles_service_unavailable_gracefully", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			ServiceName:    "audit-correlator",
			RequestTimeout: 1 * time.Second,
		}

		clientManager := NewInterServiceClientManager(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := clientManager.Initialize(ctx)
		if err != nil {
			t.Skip("Inter-service infrastructure not available for test")
		}
		defer clientManager.Cleanup(ctx)

		// Try to get a client for a non-existent service
		_, err = clientManager.GetClientByName(ctx, "non-existent-service")
		if err == nil {
			t.Error("Expected error when getting non-existent service client")
		}

		// Verify error type
		if !IsServiceUnavailableError(err) {
			t.Errorf("Expected ServiceUnavailableError, got %T", err)
		}
	})
}

// InterServiceClientManager interface that needs to be implemented
type InterServiceClientManager interface {
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	GetRiskMonitorClient(ctx context.Context) (RiskMonitorClient, error)
	GetTradingEngineClient(ctx context.Context) (TradingEngineClient, error)
	GetClientByName(ctx context.Context, serviceName string) (ServiceClient, error)
	DiscoverServices(ctx context.Context) ([]ServiceInfo, error)
	GetConnectionStats() ConnectionStats
}

type ServiceClient interface {
	HealthCheck(ctx context.Context) (HealthStatus, error)
}

type RiskMonitorClient interface {
	ServiceClient
	GetRiskMetrics(ctx context.Context) (RiskMetrics, error)
}

type TradingEngineClient interface {
	ServiceClient
	GetTradingStatus(ctx context.Context) (TradingStatus, error)
}

type HealthStatus struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Details     string    `json:"details"`
}

type RiskMetrics struct {
	TotalRiskEvents  int64     `json:"total_risk_events"`
	HighRiskAlerts   int64     `json:"high_risk_alerts"`
	LastUpdated      time.Time `json:"last_updated"`
}

type TradingStatus struct {
	ActiveStrategies int       `json:"active_strategies"`
	TotalTrades      int64     `json:"total_trades"`
	LastTradeTime    time.Time `json:"last_trade_time"`
}

type ConnectionStats struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalConnections  int64 `json:"total_connections"`
	FailedConnections int64 `json:"failed_connections"`
}

// Error handling
func IsServiceUnavailableError(err error) bool {
	// Implementation will check error type
	panic("TDD Red Phase: IsServiceUnavailableError not implemented yet")
}

// Constructor function that needs to be implemented
func NewInterServiceClientManager(cfg *config.Config) InterServiceClientManager {
	panic("TDD Red Phase: NewInterServiceClientManager not implemented yet")
}