//go:build unit

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure"
)

// TestServiceDiscovery_RedPhase defines the expected behaviors for service discovery integration
// These tests will fail initially and drive our implementation (TDD Red-Green-Refactor)
func TestServiceDiscovery_Connect(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "successful_connection",
			config: &config.Config{
				ServiceName:         "audit-correlator",
				ServiceVersion:      "1.0.0",
				RedisURL:            "redis://localhost:6379",
				GRPCPort:            9093,
				HTTPPort:            8083,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid_redis_url",
			config: &config.Config{
				ServiceName:         "audit-correlator",
				ServiceVersion:      "1.0.0",
				RedisURL:            "invalid://url",
				GRPCPort:            9093,
				HTTPPort:            8083,
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sd := infrastructure.NewServiceDiscovery(tt.config, nil)
			err := sd.Connect(context.Background())

			if tt.name == "successful_connection" && err != nil {
				t.Skip("Redis not available for test")
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceDiscovery.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				defer sd.Disconnect(context.Background())
			}
		})
	}
}

func TestServiceDiscovery_RegisterService(t *testing.T) {
	t.Run("registers_audit_correlator_service", func(t *testing.T) {
		t.Parallel()

		sd := infrastructure.NewServiceDiscovery(&config.Config{
			ServiceName:         "audit-correlator",
			ServiceVersion:      "1.0.0",
			RedisURL:            "redis://localhost:6379",
			GRPCPort:            50051,
			HTTPPort:            8080,
			HealthCheckInterval: 100 * time.Millisecond,
		}, nil)

		ctx := context.Background()
		err := sd.Connect(ctx)
		if err != nil {
			t.Skip("Redis not available for test")
		}
		defer sd.Disconnect(ctx)

		err = sd.RegisterService(ctx)
		if err != nil {
			t.Errorf("ServiceDiscovery.RegisterService() error = %v", err)
		}

		// Verify service can be discovered
		services, err := sd.DiscoverServices(ctx, "audit-correlator")
		if err != nil {
			t.Errorf("ServiceDiscovery.DiscoverServices() error = %v", err)
		}

		if len(services) == 0 {
			t.Error("Expected to find registered audit-correlator service")
		}
	})
}

func TestServiceDiscovery_HealthCheck(t *testing.T) {
	t.Run("maintains_service_heartbeat", func(t *testing.T) {
		t.Parallel()

		sd := infrastructure.NewServiceDiscovery(&config.Config{
			ServiceName:         "audit-correlator",
			ServiceVersion:      "1.0.0",
			RedisURL:            "redis://localhost:6379",
			GRPCPort:            9093,
			HTTPPort:            8083,
			HealthCheckInterval: 100 * time.Millisecond,
		}, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err := sd.Connect(ctx)
		if err != nil {
			t.Skip("Redis not available for test")
		}
		defer sd.Disconnect(ctx)

		err = sd.RegisterService(ctx)
		if err != nil {
			t.Errorf("ServiceDiscovery.RegisterService() error = %v", err)
		}

		// Start heartbeat
		go sd.StartHeartbeat(ctx)

		// Wait for multiple heartbeats
		time.Sleep(300 * time.Millisecond)

		// Verify service is still healthy
		services, err := sd.DiscoverServices(ctx, "audit-correlator")
		if err != nil {
			t.Errorf("ServiceDiscovery.DiscoverServices() error = %v", err)
		}

		if len(services) == 0 {
			t.Error("Expected audit-correlator service to remain healthy")
		}
	})
}
