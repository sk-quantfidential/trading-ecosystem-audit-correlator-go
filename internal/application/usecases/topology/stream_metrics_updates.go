package topology

import (
	"context"
	"fmt"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// StreamMetricsUpdatesRequest represents the request to stream metrics updates
type StreamMetricsUpdatesRequest struct {
	NodeIDs        []string
	EdgeIDs        []string
	UpdateInterval time.Duration // How often to collect metrics (minimum 1 second)
	RequestID      string
}

// StreamMetricsUpdatesUseCase handles streaming metrics updates to clients
type StreamMetricsUpdatesUseCase struct {
	metricsCollector ports.MetricsCollector
}

// NewStreamMetricsUpdatesUseCase creates a new StreamMetricsUpdatesUseCase
func NewStreamMetricsUpdatesUseCase(metricsCollector ports.MetricsCollector) *StreamMetricsUpdatesUseCase {
	return &StreamMetricsUpdatesUseCase{
		metricsCollector: metricsCollector,
	}
}

// Execute subscribes to metrics updates and returns a channel of metric updates
func (uc *StreamMetricsUpdatesUseCase) Execute(ctx context.Context, req *StreamMetricsUpdatesRequest) (<-chan *ports.MetricsUpdate, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate update interval (minimum 1 second)
	if req.UpdateInterval < time.Second {
		req.UpdateInterval = time.Second
	}

	// Stream metrics from collector
	metricsChan, err := uc.metricsCollector.StreamMetrics(ctx, req.NodeIDs, req.EdgeIDs, req.UpdateInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to stream metrics: %w", err)
	}

	return metricsChan, nil
}
