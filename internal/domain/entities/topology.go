package entities

import (
	"time"
)

// NodeStatus represents the health status of a service node
type NodeStatus int

const (
	NodeStatusUnspecified NodeStatus = iota
	NodeStatusLive                   // Service is healthy and responding
	NodeStatusDegraded               // Service is responding but with issues
	NodeStatusDead                   // Service is not responding
)

// EdgeStatus represents the status of a connection between services
type EdgeStatus int

const (
	EdgeStatusUnspecified EdgeStatus = iota
	EdgeStatusActive                 // Connection is healthy
	EdgeStatusDegraded               // Connection is experiencing issues
	EdgeStatusFailed                 // Connection has failed
)

// ConnectionType represents the type of connection between services
type ConnectionType int

const (
	ConnectionTypeUnspecified ConnectionType = iota
	ConnectionTypeGRPC                       // gRPC connection
	ConnectionTypeHTTP                       // HTTP/REST connection
	ConnectionTypeDataFlow                   // Data flow connection
)

// ServiceNode represents a service instance in the network topology
type ServiceNode struct {
	ID           string
	Name         string
	ServiceType  string
	InstanceName string
	Status       NodeStatus
	Labels       map[string]string
	RegisteredAt time.Time
	LastSeenAt   time.Time
}

// NewServiceNode creates a new service node
func NewServiceNode(id, name, serviceType, instanceName string) *ServiceNode {
	now := time.Now()
	return &ServiceNode{
		ID:           id,
		Name:         name,
		ServiceType:  serviceType,
		InstanceName: instanceName,
		Status:       NodeStatusLive,
		Labels:       make(map[string]string),
		RegisteredAt: now,
		LastSeenAt:   now,
	}
}

// UpdateStatus updates the node status and last seen time
func (n *ServiceNode) UpdateStatus(status NodeStatus) {
	n.Status = status
	n.LastSeenAt = time.Now()
}

// UpdateLastSeen updates the last seen timestamp
func (n *ServiceNode) UpdateLastSeen() {
	n.LastSeenAt = time.Now()
}

// AddLabel adds a label to the node
func (n *ServiceNode) AddLabel(key, value string) {
	if n.Labels == nil {
		n.Labels = make(map[string]string)
	}
	n.Labels[key] = value
}

// IsHealthy returns true if the node is in a healthy state
func (n *ServiceNode) IsHealthy() bool {
	return n.Status == NodeStatusLive
}

// ServiceConnection represents a connection between two service nodes
type ServiceConnection struct {
	ID         string
	SourceID   string
	TargetID   string
	Type       ConnectionType
	Status     EdgeStatus
	IsCritical bool
	Labels     map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewServiceConnection creates a new service connection
func NewServiceConnection(id, sourceID, targetID string, connType ConnectionType) *ServiceConnection {
	now := time.Now()
	return &ServiceConnection{
		ID:         id,
		SourceID:   sourceID,
		TargetID:   targetID,
		Type:       connType,
		Status:     EdgeStatusActive,
		IsCritical: false,
		Labels:     make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// UpdateStatus updates the connection status and updated time
func (c *ServiceConnection) UpdateStatus(status EdgeStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}

// MarkAsCritical marks the connection as critical
func (c *ServiceConnection) MarkAsCritical() {
	c.IsCritical = true
	c.UpdatedAt = time.Now()
}

// AddLabel adds a label to the connection
func (c *ServiceConnection) AddLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}

// IsHealthy returns true if the connection is in a healthy state
func (c *ServiceConnection) IsHealthy() bool {
	return c.Status == EdgeStatusActive
}

// NetworkTopology represents the complete network topology
type NetworkTopology struct {
	SnapshotID   string
	Nodes        map[string]*ServiceNode
	Connections  map[string]*ServiceConnection
	SnapshotTime time.Time
}

// NewNetworkTopology creates a new network topology
func NewNetworkTopology(snapshotID string) *NetworkTopology {
	return &NetworkTopology{
		SnapshotID:   snapshotID,
		Nodes:        make(map[string]*ServiceNode),
		Connections:  make(map[string]*ServiceConnection),
		SnapshotTime: time.Now(),
	}
}

// AddNode adds a node to the topology
func (t *NetworkTopology) AddNode(node *ServiceNode) {
	t.Nodes[node.ID] = node
}

// RemoveNode removes a node from the topology
func (t *NetworkTopology) RemoveNode(nodeID string) {
	delete(t.Nodes, nodeID)
}

// GetNode retrieves a node by ID
func (t *NetworkTopology) GetNode(nodeID string) (*ServiceNode, bool) {
	node, exists := t.Nodes[nodeID]
	return node, exists
}

// AddConnection adds a connection to the topology
func (t *NetworkTopology) AddConnection(conn *ServiceConnection) {
	t.Connections[conn.ID] = conn
}

// RemoveConnection removes a connection from the topology
func (t *NetworkTopology) RemoveConnection(connID string) {
	delete(t.Connections, connID)
}

// GetConnection retrieves a connection by ID
func (t *NetworkTopology) GetConnection(connID string) (*ServiceConnection, bool) {
	conn, exists := t.Connections[connID]
	return conn, exists
}

// GetNodesByServiceType returns all nodes of a specific service type
func (t *NetworkTopology) GetNodesByServiceType(serviceType string) []*ServiceNode {
	var nodes []*ServiceNode
	for _, node := range t.Nodes {
		if node.ServiceType == serviceType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodesByStatus returns all nodes with a specific status
func (t *NetworkTopology) GetNodesByStatus(status NodeStatus) []*ServiceNode {
	var nodes []*ServiceNode
	for _, node := range t.Nodes {
		if node.Status == status {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetConnectionsBySource returns all connections from a source node
func (t *NetworkTopology) GetConnectionsBySource(sourceID string) []*ServiceConnection {
	var connections []*ServiceConnection
	for _, conn := range t.Connections {
		if conn.SourceID == sourceID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// GetConnectionsByTarget returns all connections to a target node
func (t *NetworkTopology) GetConnectionsByTarget(targetID string) []*ServiceConnection {
	var connections []*ServiceConnection
	for _, conn := range t.Connections {
		if conn.TargetID == targetID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// CountHealthyNodes returns the number of healthy nodes
func (t *NetworkTopology) CountHealthyNodes() int {
	count := 0
	for _, node := range t.Nodes {
		if node.IsHealthy() {
			count++
		}
	}
	return count
}

// CountHealthyConnections returns the number of healthy connections
func (t *NetworkTopology) CountHealthyConnections() int {
	count := 0
	for _, conn := range t.Connections {
		if conn.IsHealthy() {
			count++
		}
	}
	return count
}
