package topology

import (
	"context"
	"fmt"
	"sync"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// subscription represents an active subscription to topology changes
type subscription struct {
	id       string
	channel  chan *ports.TopologyChangeEvent
	filters  *entities.TopologyFilters
	fromSnap string
}

// ChannelChangePublisher implements TopologyChangePublisher using Go channels
// This is suitable for single-instance deployments
type ChannelChangePublisher struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	events        []*ports.TopologyChangeEvent // Store recent events for catch-up
	maxEvents     int
}

// NewChannelChangePublisher creates a new channel-based change publisher
func NewChannelChangePublisher() *ChannelChangePublisher {
	return &ChannelChangePublisher{
		subscriptions: make(map[string]*subscription),
		events:        make([]*ports.TopologyChangeEvent, 0),
		maxEvents:     1000, // Keep last 1000 events
	}
}

// PublishChange publishes a topology change event to all subscribers
func (p *ChannelChangePublisher) PublishChange(ctx context.Context, event *ports.TopologyChangeEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Store event for catch-up
	p.events = append(p.events, event)
	if len(p.events) > p.maxEvents {
		// Remove oldest events to maintain max size
		p.events = p.events[len(p.events)-p.maxEvents:]
	}

	// Publish to all matching subscribers
	for _, sub := range p.subscriptions {
		// Check if event matches subscriber's filters
		if p.eventMatchesFilters(event, sub.filters) {
			select {
			case sub.channel <- event:
				// Event sent successfully
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel full, skip (non-blocking)
			}
		}
	}

	return nil
}

// Subscribe subscribes to topology changes from a specific snapshot
func (p *ChannelChangePublisher) Subscribe(ctx context.Context, fromSnapshotID string, filters *entities.TopologyFilters) (<-chan *ports.TopologyChangeEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create new subscription
	subID := fmt.Sprintf("sub-%d", len(p.subscriptions))
	eventChan := make(chan *ports.TopologyChangeEvent, 100) // Buffered channel

	sub := &subscription{
		id:       subID,
		channel:  eventChan,
		filters:  filters,
		fromSnap: fromSnapshotID,
	}

	p.subscriptions[subID] = sub

	// Send historical events if requested (catch-up from snapshot)
	if fromSnapshotID != "" {
		go p.sendHistoricalEvents(ctx, sub)
	}

	return eventChan, nil
}

// Unsubscribe removes a subscription
func (p *ChannelChangePublisher) Unsubscribe(ctx context.Context, subscriptionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub, exists := p.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	close(sub.channel)
	delete(p.subscriptions, subscriptionID)

	return nil
}

// sendHistoricalEvents sends historical events to a new subscriber (catch-up)
func (p *ChannelChangePublisher) sendHistoricalEvents(ctx context.Context, sub *subscription) {
	p.mu.RLock()
	events := p.events // Get a reference to events
	p.mu.RUnlock()

	for _, event := range events {
		// Check if event is after the requested snapshot
		if event.SnapshotID > sub.fromSnap {
			if p.eventMatchesFilters(event, sub.filters) {
				select {
				case sub.channel <- event:
					// Event sent
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// eventMatchesFilters checks if an event matches the subscriber's filters
func (p *ChannelChangePublisher) eventMatchesFilters(event *ports.TopologyChangeEvent, filters *entities.TopologyFilters) bool {
	if filters == nil {
		return true // No filters, match all
	}

	// Check node filters
	if event.Node != nil {
		return filters.MatchesNode(event.Node)
	}

	// Check connection filters
	if event.Connection != nil {
		return filters.MatchesConnection(event.Connection)
	}

	return true
}

var _ ports.TopologyChangePublisher = (*ChannelChangePublisher)(nil)
