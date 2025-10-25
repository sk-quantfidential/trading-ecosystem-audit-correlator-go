package topology

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// MockMetricsCollector implements MetricsCollector with generated sample data
// This is suitable for development and testing without real infrastructure
type MockMetricsCollector struct {
	rand *rand.Rand
}

// NewMockMetricsCollector creates a new mock metrics collector
func NewMockMetricsCollector() *MockMetricsCollector {
	return &MockMetricsCollector{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CollectNodeMetrics collects (generates) metrics for a node
func (c *MockMetricsCollector) CollectNodeMetrics(ctx context.Context, nodeID string) (*entities.NodeMetadata, error) {
	metadata := entities.NewNodeMetadata(nodeID)

	// Generate realistic sample metrics
	metadata.Uptime = time.Duration(c.rand.Int63n(86400)) * time.Second // 0-24 hours
	metadata.HealthScore = 70.0 + c.rand.Float64()*30.0                 // 70-100
	metadata.RequestRate = c.rand.Float64() * 1000.0                    // 0-1000 req/s
	metadata.ErrorRate = c.rand.Float64() * 10.0                        // 0-10 errors/s
	metadata.Latency = time.Duration(c.rand.Int63n(1000)) * time.Millisecond
	metadata.CPUUsage = 20.0 + c.rand.Float64()*60.0    // 20-80%
	metadata.MemoryUsage = 30.0 + c.rand.Float64()*50.0 // 30-80%
	metadata.ActiveConns = c.rand.Int63n(100)           // 0-100 connections
	metadata.TotalReqs = c.rand.Int63n(1000000)         // 0-1M requests

	return metadata, nil
}

// CollectEdgeMetrics collects (generates) metrics for an edge
func (c *MockMetricsCollector) CollectEdgeMetrics(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error) {
	metadata := entities.NewEdgeMetadata(edgeID)

	// Generate realistic sample metrics
	metadata.Protocol = "gRPC"
	metadata.Throughput = c.rand.Float64() * 500.0              // 0-500 msg/s
	metadata.ErrorRate = c.rand.Float64() * 5.0                 // 0-5 errors/s
	metadata.AvgLatency = time.Duration(c.rand.Int63n(100)) * time.Millisecond
	metadata.P99Latency = time.Duration(c.rand.Int63n(500)) * time.Millisecond
	metadata.TotalMessages = c.rand.Int63n(10000000) // 0-10M messages
	metadata.TotalErrors = c.rand.Int63n(1000)       // 0-1000 errors
	metadata.LastActiveTime = time.Now().Add(-time.Duration(c.rand.Int63n(3600)) * time.Second)

	return metadata, nil
}

// StreamMetrics streams periodic metrics updates
func (c *MockMetricsCollector) StreamMetrics(ctx context.Context, nodeIDs []string, edgeIDs []string, interval time.Duration) (<-chan *ports.MetricsUpdate, error) {
	if interval < time.Second {
		interval = time.Second
	}

	updateChan := make(chan *ports.MetricsUpdate, 10)

	go func() {
		defer close(updateChan)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		updateID := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Generate updates for nodes
				for _, nodeID := range nodeIDs {
					metadata, err := c.CollectNodeMetrics(ctx, nodeID)
					if err != nil {
						continue
					}

					update := &ports.MetricsUpdate{
						UpdateID:  fmt.Sprintf("update-%d", updateID),
						Timestamp: time.Now(),
						NodeID:    nodeID,
						Metrics: map[string]interface{}{
							"health_score":   metadata.HealthScore,
							"request_rate":   metadata.RequestRate,
							"error_rate":     metadata.ErrorRate,
							"latency_ms":     metadata.Latency.Milliseconds(),
							"cpu_usage":      metadata.CPUUsage,
							"memory_usage":   metadata.MemoryUsage,
							"active_conns":   metadata.ActiveConns,
							"total_requests": metadata.TotalReqs,
						},
					}

					select {
					case updateChan <- update:
						updateID++
					case <-ctx.Done():
						return
					}
				}

				// Generate updates for edges
				for _, edgeID := range edgeIDs {
					metadata, err := c.CollectEdgeMetrics(ctx, edgeID)
					if err != nil {
						continue
					}

					update := &ports.MetricsUpdate{
						UpdateID:  fmt.Sprintf("update-%d", updateID),
						Timestamp: time.Now(),
						EdgeID:    edgeID,
						Metrics: map[string]interface{}{
							"throughput":      metadata.Throughput,
							"error_rate":      metadata.ErrorRate,
							"avg_latency_ms":  metadata.AvgLatency.Milliseconds(),
							"p99_latency_ms":  metadata.P99Latency.Milliseconds(),
							"total_messages":  metadata.TotalMessages,
							"total_errors":    metadata.TotalErrors,
							"last_active":     metadata.LastActiveTime.Unix(),
						},
					}

					select {
					case updateChan <- update:
						updateID++
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return updateChan, nil
}

var _ ports.MetricsCollector = (*MockMetricsCollector)(nil)
