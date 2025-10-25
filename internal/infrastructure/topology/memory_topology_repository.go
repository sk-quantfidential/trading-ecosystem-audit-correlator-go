package topology

import (
	"context"
	"fmt"
	"sync"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// MemoryTopologyRepository implements TopologyRepository using in-memory storage
// This is suitable for development, testing, and small-scale deployments
type MemoryTopologyRepository struct {
	mu          sync.RWMutex
	nodes       map[string]*entities.ServiceNode
	connections map[string]*entities.ServiceConnection
	snapshotID  string
}

// NewMemoryTopologyRepository creates a new in-memory topology repository
func NewMemoryTopologyRepository() *MemoryTopologyRepository {
	return &MemoryTopologyRepository{
		nodes:       make(map[string]*entities.ServiceNode),
		connections: make(map[string]*entities.ServiceConnection),
		snapshotID:  "snapshot-0",
	}
}

// SaveNode saves a node to the repository
func (r *MemoryTopologyRepository) SaveNode(ctx context.Context, node *entities.ServiceNode) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodes[node.ID] = node
	return nil
}

// GetNode retrieves a node by ID
func (r *MemoryTopologyRepository) GetNode(ctx context.Context, nodeID string) (*entities.ServiceNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return node, nil
}

// GetNodes retrieves nodes with optional filtering
func (r *MemoryTopologyRepository) GetNodes(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*entities.ServiceNode, 0)
	for _, node := range r.nodes {
		if filters == nil || filters.MatchesNode(node) {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// DeleteNode removes a node from the repository
func (r *MemoryTopologyRepository) DeleteNode(ctx context.Context, nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	delete(r.nodes, nodeID)
	return nil
}

// SaveConnection saves a connection to the repository
func (r *MemoryTopologyRepository) SaveConnection(ctx context.Context, conn *entities.ServiceConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.connections[conn.ID] = conn
	return nil
}

// GetConnection retrieves a connection by ID
func (r *MemoryTopologyRepository) GetConnection(ctx context.Context, connID string) (*entities.ServiceConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.connections[connID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	return conn, nil
}

// GetConnections retrieves connections with optional filtering
func (r *MemoryTopologyRepository) GetConnections(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connections := make([]*entities.ServiceConnection, 0)
	for _, conn := range r.connections {
		if filters == nil || filters.MatchesConnection(conn) {
			connections = append(connections, conn)
		}
	}

	return connections, nil
}

// DeleteConnection removes a connection from the repository
func (r *MemoryTopologyRepository) DeleteConnection(ctx context.Context, connID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[connID]; !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	delete(r.connections, connID)
	return nil
}

// GetTopologySnapshot retrieves the current topology snapshot
func (r *MemoryTopologyRepository) GetTopologySnapshot(ctx context.Context) (*entities.NetworkTopology, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	topology := entities.NewNetworkTopology(r.snapshotID)

	// Copy nodes
	for id, node := range r.nodes {
		topology.Nodes[id] = node
	}

	// Copy connections
	for id, conn := range r.connections {
		topology.Connections[id] = conn
	}

	return topology, nil
}

// SaveTopologySnapshot saves the topology snapshot (updates snapshot ID)
func (r *MemoryTopologyRepository) SaveTopologySnapshot(ctx context.Context, topology *entities.NetworkTopology) error {
	if topology == nil {
		return fmt.Errorf("topology cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing data
	r.nodes = make(map[string]*entities.ServiceNode)
	r.connections = make(map[string]*entities.ServiceConnection)

	// Copy nodes
	for id, node := range topology.Nodes {
		r.nodes[id] = node
	}

	// Copy connections
	for id, conn := range topology.Connections {
		r.connections[id] = conn
	}

	r.snapshotID = topology.SnapshotID

	return nil
}

var _ ports.TopologyRepository = (*MemoryTopologyRepository)(nil)
