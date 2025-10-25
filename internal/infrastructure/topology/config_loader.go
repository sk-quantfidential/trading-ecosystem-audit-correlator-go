package topology

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

// TopologyConfig represents the JSON configuration format
type TopologyConfig struct {
	Version     string       `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Nodes       []NodeConfig `json:"nodes"`
	Edges       []EdgeConfig `json:"edges"`
}

// NodeConfig represents a node in the configuration
type NodeConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ServiceType string            `json:"service_type"`
	Category    string            `json:"category"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Endpoints   map[string]string `json:"endpoints"`
	Health      HealthConfig      `json:"health"`
}

// EdgeConfig represents an edge in the configuration
type EdgeConfig struct {
	ID           string        `json:"id"`
	SourceID     string        `json:"source_id"`
	TargetID     string        `json:"target_id"`
	Protocol     string        `json:"protocol"`
	Relationship string        `json:"relationship"`
	Status       string        `json:"status"`
	Metrics      MetricsConfig `json:"metrics"`
}

// HealthConfig represents health metrics
type HealthConfig struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      float64 `json:"memory_mb"`
	TotalRequests int64   `json:"total_requests"`
	TotalErrors   int64   `json:"total_errors"`
	ErrorRate     float64 `json:"error_rate"`
}

// MetricsConfig represents edge metrics
type MetricsConfig struct {
	LatencyP50Ms  float64 `json:"latency_p50_ms"`
	LatencyP99Ms  float64 `json:"latency_p99_ms"`
	ThroughputRps float64 `json:"throughput_rps"`
	ErrorRate     float64 `json:"error_rate"`
}

// ConfigLoader loads topology from JSON configuration file
type ConfigLoader struct {
	logger     *logrus.Logger
	repository *MemoryTopologyRepository
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(repository *MemoryTopologyRepository, logger *logrus.Logger) *ConfigLoader {
	return &ConfigLoader{
		logger:     logger,
		repository: repository,
	}
}

// LoadFromFile loads topology from a JSON file
func (l *ConfigLoader) LoadFromFile(configPath string) error {
	l.logger.WithField("config_path", configPath).Info("Loading topology configuration")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		l.logger.WithField("config_path", configPath).Warn("Topology config file not found, starting with empty topology")
		return nil // Not an error - just start with empty topology
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config TopologyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	l.logger.WithFields(logrus.Fields{
		"version": config.Version,
		"nodes":   len(config.Nodes),
		"edges":   len(config.Edges),
	}).Info("Parsed topology configuration")

	// Load nodes
	for _, nodeConfig := range config.Nodes {
		if err := l.loadNode(nodeConfig); err != nil {
			l.logger.WithError(err).WithField("node_id", nodeConfig.ID).Error("Failed to load node")
			continue
		}
	}

	// Load edges
	for _, edgeConfig := range config.Edges {
		if err := l.loadEdge(edgeConfig); err != nil {
			l.logger.WithError(err).WithField("edge_id", edgeConfig.ID).Error("Failed to load edge")
			continue
		}
	}

	l.logger.WithFields(logrus.Fields{
		"nodes_loaded": len(config.Nodes),
		"edges_loaded": len(config.Edges),
	}).Info("Successfully loaded topology configuration")

	return nil
}

// loadNode loads a single node into the repository
func (l *ConfigLoader) loadNode(config NodeConfig) error {
	// Create node with basic info
	node := entities.NewServiceNode(
		config.ID,
		config.Name,
		config.ServiceType,
		config.Name, // instance name (use name for now)
	)

	// Save node to repository
	ctx := context.Background()
	if err := l.repository.SaveNode(ctx, node); err != nil {
		return fmt.Errorf("failed to save node to repository: %w", err)
	}

	l.logger.WithFields(logrus.Fields{
		"node_id":      config.ID,
		"name":         config.Name,
		"service_type": config.ServiceType,
	}).Debug("Loaded node")

	return nil
}

// loadEdge loads a single edge into the repository
func (l *ConfigLoader) loadEdge(config EdgeConfig) error {
	// Parse connection type from protocol
	var connType entities.ConnectionType
	switch config.Protocol {
	case "gRPC":
		connType = entities.ConnectionTypeGRPC
	case "HTTP":
		connType = entities.ConnectionTypeHTTP
	default:
		connType = entities.ConnectionTypeGRPC // Default to gRPC
	}

	// Create edge
	edge := entities.NewServiceConnection(
		config.ID,
		config.SourceID,
		config.TargetID,
		connType,
	)

	// Save edge to repository
	ctx := context.Background()
	if err := l.repository.SaveConnection(ctx, edge); err != nil {
		return fmt.Errorf("failed to save edge to repository: %w", err)
	}

	l.logger.WithFields(logrus.Fields{
		"edge_id":      config.ID,
		"source_id":    config.SourceID,
		"target_id":    config.TargetID,
		"protocol":     config.Protocol,
		"relationship": config.Relationship,
	}).Debug("Loaded edge")

	return nil
}
