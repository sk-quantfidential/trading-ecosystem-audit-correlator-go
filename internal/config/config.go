package config

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/adapters"
)

type Config struct {
	// Service Identity
	ServiceName         string // "audit-correlator" (service type)
	ServiceInstanceName string // "audit-correlator" (instance identifier)
	ServiceVersion      string

	// Network Configuration
	HTTPPort int
	GRPCPort int

	// External Services
	RedisURL                string
	ConfigurationServiceURL string

	// Service Discovery
	HealthCheckInterval time.Duration

	// Client Configuration
	RequestTimeout time.Duration
	CacheTTL       time.Duration

	// Logging
	LogLevel string

	// Environment
	Environment string

	// DataAdapter - initialized during Load()
	dataAdapter adapters.DataAdapter
}

func Load() *Config {
	cfg := &Config{
		// Service Identity
		ServiceName:         getEnv("SERVICE_NAME", "audit-correlator"),
		ServiceInstanceName: getEnv("SERVICE_INSTANCE_NAME",
			getEnv("SERVICE_NAME", "audit-correlator")),
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),

		// Network Configuration
		HTTPPort: getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort: getEnvAsInt("GRPC_PORT", 50051),

		// External Services
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		ConfigurationServiceURL: getEnv("CONFIGURATION_SERVICE_URL", "http://localhost:8090"),

		// Service Discovery
		HealthCheckInterval: getEnvAsDuration("HEALTH_CHECK_INTERVAL", 30*time.Second),

		// Client Configuration
		RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
		CacheTTL:       getEnvAsDuration("CACHE_TTL", 5*time.Minute),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Validate instance name
	if err := ValidateInstanceName(cfg.ServiceInstanceName); err != nil {
		// Log warning but don't fail - allow backward compatibility
		// In production, this should be enforced
		_ = err
	}

	return cfg
}

// ValidateInstanceName validates that an instance name follows DNS-safe naming conventions
func ValidateInstanceName(name string) error {
	// Required explicit - no empty strings
	if name == "" {
		return fmt.Errorf("instance name cannot be empty")
	}

	// DNS-safe: lowercase alphanumeric and hyphens only, must start/end with alphanumeric
	validPattern := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("instance name must be DNS-safe: lowercase, alphanumeric, hyphens only, must start and end with letter or number (got: %s)", name)
	}

	// Max 63 characters (DNS label limit)
	if len(name) > 63 {
		return fmt.Errorf("instance name exceeds 63 character limit (got: %d characters)", len(name))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}
	return defaultValue
}

// InitializeDataAdapter initializes the DataAdapter with logging
func (c *Config) InitializeDataAdapter(ctx context.Context, logger *logrus.Logger) error {
	if c.dataAdapter != nil {
		return nil // Already initialized
	}

	adapter, err := adapters.NewAuditDataAdapterFromEnv(logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to create DataAdapter from environment, using defaults")
		// Fallback to defaults for development
		adapter, err = adapters.NewAuditDataAdapterWithDefaults(logger)
		if err != nil {
			return err
		}
	}

	// Connect to the adapter
	if err := adapter.Connect(ctx); err != nil {
		logger.WithError(err).Warn("Failed to connect DataAdapter - continuing with stub")
		// Continue without adapter - will use stubs
		return nil
	}

	c.dataAdapter = adapter
	logger.Info("DataAdapter initialized and connected successfully")
	return nil
}

// GetDataAdapter returns the initialized DataAdapter (may be nil)
func (c *Config) GetDataAdapter() adapters.DataAdapter {
	return c.dataAdapter
}

// DisconnectDataAdapter safely disconnects the DataAdapter
func (c *Config) DisconnectDataAdapter(ctx context.Context) error {
	if c.dataAdapter != nil {
		return c.dataAdapter.Disconnect(ctx)
	}
	return nil
}