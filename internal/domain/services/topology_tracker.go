package services

import (
	"context"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

// TopologyTracker defines the domain service interface for topology management
// This is a pure domain interface with no infrastructure dependencies
type TopologyTracker interface {
	// Node operations
	RegisterNode(ctx context.Context, node *entities.ServiceNode) error
	DeregisterNode(ctx context.Context, nodeID string) error
	UpdateNodeStatus(ctx context.Context, nodeID string, status entities.NodeStatus) error
	GetNode(ctx context.Context, nodeID string) (*entities.ServiceNode, error)
	GetNodes(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceNode, error)

	// Connection operations
	RegisterConnection(ctx context.Context, conn *entities.ServiceConnection) error
	DeregisterConnection(ctx context.Context, connID string) error
	UpdateConnectionStatus(ctx context.Context, connID string, status entities.EdgeStatus) error
	GetConnection(ctx context.Context, connID string) (*entities.ServiceConnection, error)
	GetConnections(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceConnection, error)

	// Topology queries
	GetTopology(ctx context.Context, filters *entities.TopologyFilters) (*entities.NetworkTopology, error)
	GetTopologySnapshot(ctx context.Context) (*entities.NetworkTopology, error)

	// Metadata operations
	GetNodeMetadata(ctx context.Context, nodeID string) (*entities.NodeMetadata, error)
	GetEdgeMetadata(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error)
	UpdateNodeMetadata(ctx context.Context, metadata *entities.NodeMetadata) error
	UpdateEdgeMetadata(ctx context.Context, metadata *entities.EdgeMetadata) error
}
