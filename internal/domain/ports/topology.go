package ports

import (
	"context"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

// TopologyRepository defines the port for persisting and querying topology data
// This is an interface that will be implemented by infrastructure adapters
type TopologyRepository interface {
	// Node operations
	SaveNode(ctx context.Context, node *entities.ServiceNode) error
	GetNode(ctx context.Context, nodeID string) (*entities.ServiceNode, error)
	GetNodes(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceNode, error)
	DeleteNode(ctx context.Context, nodeID string) error

	// Connection operations
	SaveConnection(ctx context.Context, conn *entities.ServiceConnection) error
	GetConnection(ctx context.Context, connID string) (*entities.ServiceConnection, error)
	GetConnections(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceConnection, error)
	DeleteConnection(ctx context.Context, connID string) error

	// Snapshot operations
	GetTopologySnapshot(ctx context.Context) (*entities.NetworkTopology, error)
	SaveTopologySnapshot(ctx context.Context, topology *entities.NetworkTopology) error
}

// MetadataRepository defines the port for persisting and querying metadata
type MetadataRepository interface {
	// Node metadata operations
	SaveNodeMetadata(ctx context.Context, metadata *entities.NodeMetadata) error
	GetNodeMetadata(ctx context.Context, nodeID string) (*entities.NodeMetadata, error)
	GetNodesMetadata(ctx context.Context, nodeIDs []string) ([]*entities.NodeMetadata, error)

	// Edge metadata operations
	SaveEdgeMetadata(ctx context.Context, metadata *entities.EdgeMetadata) error
	GetEdgeMetadata(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error)
	GetEdgesMetadata(ctx context.Context, edgeIDs []string) ([]*entities.EdgeMetadata, error)
}

// TopologyChangeType represents the type of topology change
type TopologyChangeType int

const (
	TopologyChangeTypeUnspecified TopologyChangeType = iota
	TopologyChangeTypeNodeAdded
	TopologyChangeTypeNodeUpdated
	TopologyChangeTypeNodeRemoved
	TopologyChangeTypeEdgeAdded
	TopologyChangeTypeEdgeUpdated
	TopologyChangeTypeEdgeRemoved
)

// TopologyChangeEvent represents a change in the topology
type TopologyChangeEvent struct {
	ChangeID   string
	ChangeType TopologyChangeType
	Timestamp  time.Time
	SnapshotID string
	Node       *entities.ServiceNode
	Connection *entities.ServiceConnection
}

// TopologyChangePublisher defines the port for publishing topology changes
type TopologyChangePublisher interface {
	// Publish a topology change event
	PublishChange(ctx context.Context, event *TopologyChangeEvent) error

	// Subscribe to topology changes (returns a channel for streaming)
	Subscribe(ctx context.Context, fromSnapshotID string, filters *entities.TopologyFilters) (<-chan *TopologyChangeEvent, error)

	// Unsubscribe from topology changes
	Unsubscribe(ctx context.Context, subscriptionID string) error
}

// MetricsUpdate represents a metrics update for a node or edge
type MetricsUpdate struct {
	UpdateID  string
	Timestamp time.Time
	NodeID    string
	EdgeID    string
	Metrics   map[string]interface{}
}

// MetricsCollector defines the port for collecting metrics from infrastructure
type MetricsCollector interface {
	// Collect node metrics (from Prometheus, CloudWatch, etc.)
	CollectNodeMetrics(ctx context.Context, nodeID string) (*entities.NodeMetadata, error)

	// Collect edge metrics
	CollectEdgeMetrics(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error)

	// Stream metrics updates (returns a channel for streaming)
	StreamMetrics(ctx context.Context, nodeIDs []string, edgeIDs []string, interval time.Duration) (<-chan *MetricsUpdate, error)
}
