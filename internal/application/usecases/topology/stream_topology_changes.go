package topology

import (
	"context"
	"fmt"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// StreamTopologyChangesRequest represents the request to stream topology changes
type StreamTopologyChangesRequest struct {
	FromSnapshotID string
	ServiceTypes   []string
	MinInterval    time.Duration // Minimum interval between updates (e.g., 1 second)
}

// StreamTopologyChangesUseCase handles streaming topology changes to clients
type StreamTopologyChangesUseCase struct {
	changePublisher ports.TopologyChangePublisher
}

// NewStreamTopologyChangesUseCase creates a new StreamTopologyChangesUseCase
func NewStreamTopologyChangesUseCase(changePublisher ports.TopologyChangePublisher) *StreamTopologyChangesUseCase {
	return &StreamTopologyChangesUseCase{
		changePublisher: changePublisher,
	}
}

// Execute subscribes to topology changes and returns a channel of change events
func (uc *StreamTopologyChangesUseCase) Execute(ctx context.Context, req *StreamTopologyChangesRequest) (<-chan *ports.TopologyChangeEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate minimum interval
	if req.MinInterval < time.Second {
		req.MinInterval = time.Second // Enforce 1-second minimum
	}

	// Build filters
	filters := entities.NewTopologyFilters()
	for _, serviceType := range req.ServiceTypes {
		filters.AddServiceType(serviceType)
	}

	// Subscribe to changes
	changeChan, err := uc.changePublisher.Subscribe(ctx, req.FromSnapshotID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topology changes: %w", err)
	}

	// Apply rate limiting based on MinInterval
	throttledChan := make(chan *ports.TopologyChangeEvent)
	go func() {
		defer close(throttledChan)
		ticker := time.NewTicker(req.MinInterval)
		defer ticker.Stop()

		var pendingChange *ports.TopologyChangeEvent

		for {
			select {
			case <-ctx.Done():
				return
			case change, ok := <-changeChan:
				if !ok {
					// Channel closed, send any pending change and exit
					if pendingChange != nil {
						select {
						case throttledChan <- pendingChange:
						case <-ctx.Done():
						}
					}
					return
				}
				// Store the most recent change
				pendingChange = change
			case <-ticker.C:
				// Send pending change if available
				if pendingChange != nil {
					select {
					case throttledChan <- pendingChange:
						pendingChange = nil
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return throttledChan, nil
}
