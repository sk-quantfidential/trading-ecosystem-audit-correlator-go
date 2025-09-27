package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
)

// ConfigValueType represents the type of a configuration value
type ConfigValueType int

const (
	ConfigValueTypeString ConfigValueType = iota
	ConfigValueTypeNumber
	ConfigValueTypeBoolean
	ConfigValueTypeJSON
)

// ConfigurationValue represents a configuration value with metadata
type ConfigurationValue struct {
	Key         string          `json:"key"`
	Value       string          `json:"value"`
	Type        ConfigValueType `json:"type"`
	Environment string          `json:"environment"`
	LastUpdated time.Time       `json:"last_updated"`
}

// AsString returns the configuration value as a string
func (cv ConfigurationValue) AsString() string {
	return cv.Value
}

// AsInt converts and returns the configuration value as an integer
func (cv ConfigurationValue) AsInt() (int, error) {
	return strconv.Atoi(cv.Value)
}

// AsBool converts and returns the configuration value as a boolean
func (cv ConfigurationValue) AsBool() (bool, error) {
	return strconv.ParseBool(cv.Value)
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	CacheHits   int64   `json:"cache_hits"`
	CacheMisses int64   `json:"cache_misses"`
	CacheSize   int     `json:"cache_size"`
	HitRate     float64 `json:"hit_rate"`
}

// ConfigurationClient interface for configuration service integration
type ConfigurationClient interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	GetConfiguration(ctx context.Context, key string) (ConfigurationValue, error)
	GetCacheStats() CacheStats
}

// HTTPConfigurationClient implements ConfigurationClient using HTTP
type HTTPConfigurationClient struct {
	config     *config.Config
	httpClient *http.Client
	logger     *logrus.Logger

	// Cache management
	cache      map[string]cachedValue
	cacheMutex sync.RWMutex
	cacheStats CacheStats
	statsMutex sync.RWMutex
}

type cachedValue struct {
	value     ConfigurationValue
	expiresAt time.Time
}

// configResponse represents the HTTP response from configuration service
type configResponse struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	LastUpdated string `json:"last_updated"`
}

// NewConfigurationClient creates a new HTTP-based configuration client
func NewConfigurationClient(cfg *config.Config, logger *logrus.Logger) ConfigurationClient {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	return &HTTPConfigurationClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		logger: logger,
		cache:  make(map[string]cachedValue),
	}
}

// Connect establishes connection to the configuration service
func (c *HTTPConfigurationClient) Connect(ctx context.Context) error {
	// Test connectivity with a health check
	healthURL := fmt.Sprintf("%s/health", c.config.ConfigurationServiceURL)

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to configuration service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("configuration service health check failed: status %d", resp.StatusCode)
	}

	c.logger.Info("Connected to configuration service")
	return nil
}

// Disconnect closes the connection to the configuration service
func (c *HTTPConfigurationClient) Disconnect(ctx context.Context) error {
	// Clear cache
	c.cacheMutex.Lock()
	c.cache = make(map[string]cachedValue)
	c.cacheMutex.Unlock()

	c.logger.Info("Disconnected from configuration service")
	return nil
}

// GetConfiguration retrieves a configuration value with caching
func (c *HTTPConfigurationClient) GetConfiguration(ctx context.Context, key string) (ConfigurationValue, error) {
	// Check cache first
	if cachedVal, found := c.getCachedValue(key); found {
		c.incrementCacheHits()
		return cachedVal, nil
	}

	c.incrementCacheMisses()

	// Fetch from service
	configURL := fmt.Sprintf("%s/api/v1/configuration/%s", c.config.ConfigurationServiceURL, key)

	req, err := http.NewRequestWithContext(ctx, "GET", configURL, nil)
	if err != nil {
		return ConfigurationValue{}, fmt.Errorf("failed to create configuration request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ConfigurationValue{}, fmt.Errorf("failed to fetch configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ConfigurationValue{}, fmt.Errorf("configuration key not found: %s", key)
	}

	if resp.StatusCode != http.StatusOK {
		return ConfigurationValue{}, fmt.Errorf("configuration service error: status %d", resp.StatusCode)
	}

	var configResp configResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return ConfigurationValue{}, fmt.Errorf("failed to decode configuration response: %w", err)
	}

	// Convert response to ConfigurationValue
	configValue := c.convertResponse(configResp)

	// Cache the value
	c.cacheValue(key, configValue)

	return configValue, nil
}

// GetCacheStats returns current cache performance statistics
func (c *HTTPConfigurationClient) GetCacheStats() CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	c.cacheMutex.RLock()
	cacheSize := len(c.cache)
	c.cacheMutex.RUnlock()

	stats := c.cacheStats
	stats.CacheSize = cacheSize

	total := stats.CacheHits + stats.CacheMisses
	if total > 0 {
		stats.HitRate = float64(stats.CacheHits) / float64(total)
	}

	return stats
}

// getCachedValue retrieves a value from cache if not expired
func (c *HTTPConfigurationClient) getCachedValue(key string) (ConfigurationValue, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	cached, exists := c.cache[key]
	if !exists {
		return ConfigurationValue{}, false
	}

	if time.Now().After(cached.expiresAt) {
		// Value expired, remove from cache
		delete(c.cache, key)
		return ConfigurationValue{}, false
	}

	return cached.value, true
}

// cacheValue stores a value in cache with TTL
func (c *HTTPConfigurationClient) cacheValue(key string, value ConfigurationValue) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[key] = cachedValue{
		value:     value,
		expiresAt: time.Now().Add(c.config.CacheTTL),
	}
}

// incrementCacheHits increments the cache hit counter
func (c *HTTPConfigurationClient) incrementCacheHits() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.cacheStats.CacheHits++
}

// incrementCacheMisses increments the cache miss counter
func (c *HTTPConfigurationClient) incrementCacheMisses() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.cacheStats.CacheMisses++
}

// convertResponse converts the HTTP response to ConfigurationValue
func (c *HTTPConfigurationClient) convertResponse(resp configResponse) ConfigurationValue {
	configType := c.parseConfigType(resp.Type)

	lastUpdated, _ := time.Parse(time.RFC3339, resp.LastUpdated)

	return ConfigurationValue{
		Key:         resp.Key,
		Value:       resp.Value,
		Type:        configType,
		Environment: resp.Environment,
		LastUpdated: lastUpdated,
	}
}

// parseConfigType converts string type to ConfigValueType
func (c *HTTPConfigurationClient) parseConfigType(typeStr string) ConfigValueType {
	switch strings.ToLower(typeStr) {
	case "string":
		return ConfigValueTypeString
	case "number", "int", "integer":
		return ConfigValueTypeNumber
	case "boolean", "bool":
		return ConfigValueTypeBoolean
	case "json", "object":
		return ConfigValueTypeJSON
	default:
		return ConfigValueTypeString
	}
}