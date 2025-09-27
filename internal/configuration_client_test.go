//go:build unit

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure"
)

// TestConfigurationClient_RedPhase defines the expected behaviors for configuration service integration
// These tests will fail initially and drive our implementation (TDD Red-Green-Refactor)
func TestConfigurationClient_GetConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectedType infrastructure.ConfigValueType
		wantErr      bool
	}{
		{
			name:         "audit_retention_days",
			key:          "audit.retention_days",
			expectedType: infrastructure.ConfigValueTypeNumber,
			wantErr:      false,
		},
		{
			name:         "correlation_enabled",
			key:          "audit.correlation.enabled",
			expectedType: infrastructure.ConfigValueTypeBoolean,
			wantErr:      false,
		},
		{
			name:         "storage_backend",
			key:          "audit.storage.backend",
			expectedType: infrastructure.ConfigValueTypeString,
			wantErr:      false,
		},
		{
			name:    "invalid_key",
			key:     "nonexistent.key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := infrastructure.NewConfigurationClient(&config.Config{
				ConfigurationServiceURL: "http://localhost:8090",
				RequestTimeout:          5 * time.Second,
				CacheTTL:               5 * time.Minute,
			}, nil)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := client.Connect(ctx)
			if err != nil {
				t.Skip("Configuration service not available for test")
			}
			defer client.Disconnect(ctx)

			value, err := client.GetConfiguration(ctx, tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigurationClient.GetConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if value.Key != tt.key {
					t.Errorf("Expected key %s, got %s", tt.key, value.Key)
				}

				if value.Type != tt.expectedType {
					t.Errorf("Expected type %v, got %v", tt.expectedType, value.Type)
				}
			}
		})
	}
}

func TestConfigurationClient_Caching(t *testing.T) {
	t.Run("caches_configuration_values", func(t *testing.T) {
		t.Parallel()

		client := infrastructure.NewConfigurationClient(&config.Config{
			ConfigurationServiceURL: "http://localhost:8090",
			RequestTimeout:          5 * time.Second,
			CacheTTL:               300 * time.Second,
		}, nil)

		ctx := context.Background()
		err := client.Connect(ctx)
		if err != nil {
			t.Skip("Configuration service not available for test")
		}
		defer client.Disconnect(ctx)

		key := "audit.test_cache_key"

		// First call - should hit the service
		value1, err := client.GetConfiguration(ctx, key)
		if err != nil {
			t.Skip("Configuration key not available for test")
		}

		stats1 := client.GetCacheStats()

		// Second call - should hit the cache
		value2, err := client.GetConfiguration(ctx, key)
		if err != nil {
			t.Errorf("Unexpected error on cached call: %v", err)
		}

		stats2 := client.GetCacheStats()

		if value1.Value != value2.Value {
			t.Error("Cached value should match original value")
		}

		if stats2.CacheHits <= stats1.CacheHits {
			t.Error("Expected cache hits to increase")
		}
	})
}

func TestConfigurationClient_TypeConversions(t *testing.T) {
	tests := []struct {
		name        string
		configValue infrastructure.ConfigurationValue
		testFunc    func(t *testing.T, value infrastructure.ConfigurationValue)
	}{
		{
			name: "string_conversion",
			configValue: infrastructure.ConfigurationValue{
				Key:   "test.string",
				Value: "audit-correlator",
				Type:  infrastructure.ConfigValueTypeString,
			},
			testFunc: func(t *testing.T, value infrastructure.ConfigurationValue) {
				result := value.AsString()
				if result != "audit-correlator" {
					t.Errorf("Expected 'audit-correlator', got '%s'", result)
				}
			},
		},
		{
			name: "number_conversion",
			configValue: infrastructure.ConfigurationValue{
				Key:   "test.number",
				Value: "30",
				Type:  infrastructure.ConfigValueTypeNumber,
			},
			testFunc: func(t *testing.T, value infrastructure.ConfigurationValue) {
				result, err := value.AsInt()
				if err != nil {
					t.Errorf("AsInt() failed: %v", err)
				}
				if result != 30 {
					t.Errorf("Expected 30, got %d", result)
				}
			},
		},
		{
			name: "boolean_conversion",
			configValue: infrastructure.ConfigurationValue{
				Key:   "test.boolean",
				Value: "true",
				Type:  infrastructure.ConfigValueTypeBoolean,
			},
			testFunc: func(t *testing.T, value infrastructure.ConfigurationValue) {
				result, err := value.AsBool()
				if err != nil {
					t.Errorf("AsBool() failed: %v", err)
				}
				if !result {
					t.Error("Expected true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t, tt.configValue)
		})
	}
}

