package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/config"
)

// ServiceInfo represents service registration information
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
type ServiceDiscovery interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	RegisterService(ctx context.Context) error
	DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error)
	StartHeartbeat(ctx context.Context)
}

// RedisServiceDiscovery implements ServiceDiscovery using Redis
type RedisServiceDiscovery struct {
	config      *config.Config
	redisClient *redis.Client
	logger      *logrus.Logger

	// Service information
	serviceInfo ServiceInfo

	// Heartbeat management
	heartbeatInterval time.Duration
	heartbeatStop     chan struct{}
}

// NewServiceDiscovery creates a new Redis-based service discovery client
func NewServiceDiscovery(cfg *config.Config, logger *logrus.Logger) ServiceDiscovery {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	// Parse Redis URL
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.WithError(err).Error("Failed to parse Redis URL")
		// Use default values
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	client := redis.NewClient(opt)

	// Determine host IP
	host := getLocalIP()
	if host == "" {
		host = "localhost"
	}

	serviceInfo := ServiceInfo{
		Name:     cfg.ServiceName,
		Version:  cfg.ServiceVersion,
		Host:     host,
		GRPCPort: cfg.GRPCPort,
		HTTPPort: cfg.HTTPPort,
		Status:   "healthy",
		LastSeen: time.Now(),
	}

	return &RedisServiceDiscovery{
		config:            cfg,
		redisClient:       client,
		logger:            logger,
		serviceInfo:       serviceInfo,
		heartbeatInterval: cfg.HealthCheckInterval,
		heartbeatStop:     make(chan struct{}),
	}
}

// Connect establishes connection to Redis
func (r *RedisServiceDiscovery) Connect(ctx context.Context) error {
	// Test Redis connectivity
	_, err := r.redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.logger.Info("Connected to Redis service discovery")
	return nil
}

// Disconnect closes the Redis connection
func (r *RedisServiceDiscovery) Disconnect(ctx context.Context) error {
	// Stop heartbeat
	close(r.heartbeatStop)

	// Remove service registration
	serviceKey := fmt.Sprintf("services:%s:%s", r.serviceInfo.Name, r.getServiceID())
	err := r.redisClient.Del(ctx, serviceKey).Err()
	if err != nil {
		r.logger.WithError(err).Warn("Failed to deregister service")
	}

	// Close Redis connection
	err = r.redisClient.Close()
	if err != nil {
		r.logger.WithError(err).Warn("Failed to close Redis connection")
	}

	r.logger.Info("Disconnected from Redis service discovery")
	return nil
}

// RegisterService registers this service in the Redis registry
func (r *RedisServiceDiscovery) RegisterService(ctx context.Context) error {
	r.serviceInfo.LastSeen = time.Now()

	// Serialize service info
	serviceData, err := json.Marshal(r.serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Register in Redis with TTL
	serviceKey := fmt.Sprintf("services:%s:%s", r.serviceInfo.Name, r.getServiceID())
	ttl := r.heartbeatInterval * 3 // Allow 3 missed heartbeats before expiration

	err = r.redisClient.Set(ctx, serviceKey, serviceData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"service_name": r.serviceInfo.Name,
		"host":         r.serviceInfo.Host,
		"grpc_port":    r.serviceInfo.GRPCPort,
		"http_port":    r.serviceInfo.HTTPPort,
	}).Info("Service registered in discovery")

	return nil
}

// DiscoverServices finds all registered instances of a service
func (r *RedisServiceDiscovery) DiscoverServices(ctx context.Context, serviceName string) ([]ServiceInfo, error) {
	// Search for services by pattern
	pattern := fmt.Sprintf("services:%s:*", serviceName)
	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var services []ServiceInfo

	// Retrieve each service's information
	for _, key := range keys {
		serviceData, err := r.redisClient.Get(ctx, key).Result()
		if err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to get service data")
			continue
		}

		var serviceInfo ServiceInfo
		if err := json.Unmarshal([]byte(serviceData), &serviceInfo); err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to unmarshal service data")
			continue
		}

		services = append(services, serviceInfo)
	}

	r.logger.WithFields(logrus.Fields{
		"service_name":    serviceName,
		"instances_found": len(services),
	}).Debug("Discovered services")

	return services, nil
}

// StartHeartbeat starts the periodic heartbeat to maintain service registration
func (r *RedisServiceDiscovery) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(r.heartbeatInterval)
	defer ticker.Stop()

	r.logger.WithField("interval", r.heartbeatInterval).Info("Starting service discovery heartbeat")

	for {
		select {
		case <-ticker.C:
			if err := r.RegisterService(ctx); err != nil {
				r.logger.WithError(err).Error("Failed to send heartbeat")
			}
		case <-r.heartbeatStop:
			r.logger.Info("Stopping service discovery heartbeat")
			return
		case <-ctx.Done():
			r.logger.Info("Context cancelled, stopping heartbeat")
			return
		}
	}
}

// getServiceID generates a unique identifier for this service instance
func (r *RedisServiceDiscovery) getServiceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%s-%d", hostname, r.serviceInfo.Host, r.serviceInfo.GRPCPort)
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