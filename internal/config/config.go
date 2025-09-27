package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Service Identity
	ServiceName    string
	ServiceVersion string

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
}

func Load() *Config {
	return &Config{
		// Service Identity
		ServiceName:    getEnv("SERVICE_NAME", "audit-correlator"),
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),

		// Network Configuration
		HTTPPort: getEnvAsInt("HTTP_PORT", 8083),
		GRPCPort: getEnvAsInt("GRPC_PORT", 9093),

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
	}
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