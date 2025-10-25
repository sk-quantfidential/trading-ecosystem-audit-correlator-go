package entities

import (
	"testing"
	"time"
)

func TestServiceNode_Creation(t *testing.T) {
	node := NewServiceNode("node-1", "risk-monitor", "risk-monitor-py", "risk-monitor-lh")

	if node.ID != "node-1" {
		t.Errorf("Expected ID 'node-1', got %s", node.ID)
	}
	if node.Name != "risk-monitor" {
		t.Errorf("Expected Name 'risk-monitor', got %s", node.Name)
	}
	if node.ServiceType != "risk-monitor-py" {
		t.Errorf("Expected ServiceType 'risk-monitor-py', got %s", node.ServiceType)
	}
	if node.InstanceName != "risk-monitor-lh" {
		t.Errorf("Expected InstanceName 'risk-monitor-lh', got %s", node.InstanceName)
	}
	if node.Status != NodeStatusLive {
		t.Errorf("Expected initial status NodeStatusLive, got %v", node.Status)
	}
	if node.Labels == nil {
		t.Error("Expected Labels map to be initialized")
	}
	if node.RegisteredAt.IsZero() {
		t.Error("Expected RegisteredAt to be set")
	}
	if node.LastSeenAt.IsZero() {
		t.Error("Expected LastSeenAt to be set")
	}
}

func TestServiceNode_UpdateStatus(t *testing.T) {
	node := NewServiceNode("node-1", "test", "test-type", "test-instance")
	initialLastSeen := node.LastSeenAt

	time.Sleep(10 * time.Millisecond)
	node.UpdateStatus(NodeStatusDegraded)

	if node.Status != NodeStatusDegraded {
		t.Errorf("Expected status NodeStatusDegraded, got %v", node.Status)
	}
	if !node.LastSeenAt.After(initialLastSeen) {
		t.Error("Expected LastSeenAt to be updated")
	}
}

func TestServiceNode_AddLabel(t *testing.T) {
	node := NewServiceNode("node-1", "test", "test-type", "test-instance")
	node.AddLabel("environment", "production")
	node.AddLabel("region", "us-east-1")

	if node.Labels["environment"] != "production" {
		t.Error("Expected environment label to be 'production'")
	}
	if node.Labels["region"] != "us-east-1" {
		t.Error("Expected region label to be 'us-east-1'")
	}
}

func TestServiceNode_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeStatus
		expected bool
	}{
		{"Live node is healthy", NodeStatusLive, true},
		{"Degraded node is not healthy", NodeStatusDegraded, false},
		{"Dead node is not healthy", NodeStatusDead, false},
		{"Unspecified node is not healthy", NodeStatusUnspecified, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewServiceNode("node-1", "test", "test-type", "test-instance")
			node.Status = tt.status

			if node.IsHealthy() != tt.expected {
				t.Errorf("Expected IsHealthy() to return %v for status %v", tt.expected, tt.status)
			}
		})
	}
}

func TestServiceConnection_Creation(t *testing.T) {
	conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)

	if conn.ID != "conn-1" {
		t.Errorf("Expected ID 'conn-1', got %s", conn.ID)
	}
	if conn.SourceID != "node-1" {
		t.Errorf("Expected SourceID 'node-1', got %s", conn.SourceID)
	}
	if conn.TargetID != "node-2" {
		t.Errorf("Expected TargetID 'node-2', got %s", conn.TargetID)
	}
	if conn.Type != ConnectionTypeGRPC {
		t.Errorf("Expected Type ConnectionTypeGRPC, got %v", conn.Type)
	}
	if conn.Status != EdgeStatusActive {
		t.Errorf("Expected initial status EdgeStatusActive, got %v", conn.Status)
	}
	if conn.IsCritical {
		t.Error("Expected IsCritical to be false initially")
	}
	if conn.Labels == nil {
		t.Error("Expected Labels map to be initialized")
	}
}

func TestServiceConnection_UpdateStatus(t *testing.T) {
	conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
	initialUpdated := conn.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	conn.UpdateStatus(EdgeStatusDegraded)

	if conn.Status != EdgeStatusDegraded {
		t.Errorf("Expected status EdgeStatusDegraded, got %v", conn.Status)
	}
	if !conn.UpdatedAt.After(initialUpdated) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestServiceConnection_MarkAsCritical(t *testing.T) {
	conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
	initialUpdated := conn.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	conn.MarkAsCritical()

	if !conn.IsCritical {
		t.Error("Expected IsCritical to be true")
	}
	if !conn.UpdatedAt.After(initialUpdated) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestServiceConnection_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		status   EdgeStatus
		expected bool
	}{
		{"Active connection is healthy", EdgeStatusActive, true},
		{"Degraded connection is not healthy", EdgeStatusDegraded, false},
		{"Failed connection is not healthy", EdgeStatusFailed, false},
		{"Unspecified connection is not healthy", EdgeStatusUnspecified, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
			conn.Status = tt.status

			if conn.IsHealthy() != tt.expected {
				t.Errorf("Expected IsHealthy() to return %v for status %v", tt.expected, tt.status)
			}
		})
	}
}

func TestNetworkTopology_Creation(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")

	if topology.SnapshotID != "snapshot-1" {
		t.Errorf("Expected SnapshotID 'snapshot-1', got %s", topology.SnapshotID)
	}
	if topology.Nodes == nil {
		t.Error("Expected Nodes map to be initialized")
	}
	if topology.Connections == nil {
		t.Error("Expected Connections map to be initialized")
	}
	if topology.SnapshotTime.IsZero() {
		t.Error("Expected SnapshotTime to be set")
	}
}

func TestNetworkTopology_NodeOperations(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")
	node1 := NewServiceNode("node-1", "test1", "type1", "instance1")
	node2 := NewServiceNode("node-2", "test2", "type2", "instance2")

	// Test AddNode
	topology.AddNode(node1)
	topology.AddNode(node2)

	if len(topology.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(topology.Nodes))
	}

	// Test GetNode
	retrieved, exists := topology.GetNode("node-1")
	if !exists {
		t.Error("Expected node-1 to exist")
	}
	if retrieved.ID != "node-1" {
		t.Errorf("Expected retrieved node ID 'node-1', got %s", retrieved.ID)
	}

	// Test RemoveNode
	topology.RemoveNode("node-1")
	if len(topology.Nodes) != 1 {
		t.Errorf("Expected 1 node after removal, got %d", len(topology.Nodes))
	}

	_, exists = topology.GetNode("node-1")
	if exists {
		t.Error("Expected node-1 to not exist after removal")
	}
}

func TestNetworkTopology_ConnectionOperations(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")
	conn1 := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
	conn2 := NewServiceConnection("conn-2", "node-2", "node-3", ConnectionTypeHTTP)

	// Test AddConnection
	topology.AddConnection(conn1)
	topology.AddConnection(conn2)

	if len(topology.Connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(topology.Connections))
	}

	// Test GetConnection
	retrieved, exists := topology.GetConnection("conn-1")
	if !exists {
		t.Error("Expected conn-1 to exist")
	}
	if retrieved.ID != "conn-1" {
		t.Errorf("Expected retrieved connection ID 'conn-1', got %s", retrieved.ID)
	}

	// Test RemoveConnection
	topology.RemoveConnection("conn-1")
	if len(topology.Connections) != 1 {
		t.Errorf("Expected 1 connection after removal, got %d", len(topology.Connections))
	}

	_, exists = topology.GetConnection("conn-1")
	if exists {
		t.Error("Expected conn-1 to not exist after removal")
	}
}

func TestNetworkTopology_GetNodesByServiceType(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")
	node1 := NewServiceNode("node-1", "risk1", "risk-monitor-py", "instance1")
	node2 := NewServiceNode("node-2", "risk2", "risk-monitor-py", "instance2")
	node3 := NewServiceNode("node-3", "exchange", "exchange-simulator-go", "instance3")

	topology.AddNode(node1)
	topology.AddNode(node2)
	topology.AddNode(node3)

	riskNodes := topology.GetNodesByServiceType("risk-monitor-py")
	if len(riskNodes) != 2 {
		t.Errorf("Expected 2 risk-monitor-py nodes, got %d", len(riskNodes))
	}

	exchangeNodes := topology.GetNodesByServiceType("exchange-simulator-go")
	if len(exchangeNodes) != 1 {
		t.Errorf("Expected 1 exchange-simulator-go node, got %d", len(exchangeNodes))
	}
}

func TestNetworkTopology_GetNodesByStatus(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")
	node1 := NewServiceNode("node-1", "test1", "type1", "instance1")
	node2 := NewServiceNode("node-2", "test2", "type2", "instance2")
	node3 := NewServiceNode("node-3", "test3", "type3", "instance3")

	node1.Status = NodeStatusLive
	node2.Status = NodeStatusDegraded
	node3.Status = NodeStatusLive

	topology.AddNode(node1)
	topology.AddNode(node2)
	topology.AddNode(node3)

	liveNodes := topology.GetNodesByStatus(NodeStatusLive)
	if len(liveNodes) != 2 {
		t.Errorf("Expected 2 live nodes, got %d", len(liveNodes))
	}

	degradedNodes := topology.GetNodesByStatus(NodeStatusDegraded)
	if len(degradedNodes) != 1 {
		t.Errorf("Expected 1 degraded node, got %d", len(degradedNodes))
	}
}

func TestNetworkTopology_GetConnectionsBySourceAndTarget(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")
	conn1 := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
	conn2 := NewServiceConnection("conn-2", "node-1", "node-3", ConnectionTypeHTTP)
	conn3 := NewServiceConnection("conn-3", "node-2", "node-3", ConnectionTypeGRPC)

	topology.AddConnection(conn1)
	topology.AddConnection(conn2)
	topology.AddConnection(conn3)

	// Test GetConnectionsBySource
	sourceConns := topology.GetConnectionsBySource("node-1")
	if len(sourceConns) != 2 {
		t.Errorf("Expected 2 connections from node-1, got %d", len(sourceConns))
	}

	// Test GetConnectionsByTarget
	targetConns := topology.GetConnectionsByTarget("node-3")
	if len(targetConns) != 2 {
		t.Errorf("Expected 2 connections to node-3, got %d", len(targetConns))
	}
}

func TestNetworkTopology_CountHealthyNodesAndConnections(t *testing.T) {
	topology := NewNetworkTopology("snapshot-1")

	// Add nodes with various statuses
	node1 := NewServiceNode("node-1", "test1", "type1", "instance1")
	node1.Status = NodeStatusLive
	node2 := NewServiceNode("node-2", "test2", "type2", "instance2")
	node2.Status = NodeStatusDegraded
	node3 := NewServiceNode("node-3", "test3", "type3", "instance3")
	node3.Status = NodeStatusLive

	topology.AddNode(node1)
	topology.AddNode(node2)
	topology.AddNode(node3)

	// Add connections with various statuses
	conn1 := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
	conn1.Status = EdgeStatusActive
	conn2 := NewServiceConnection("conn-2", "node-2", "node-3", ConnectionTypeHTTP)
	conn2.Status = EdgeStatusFailed

	topology.AddConnection(conn1)
	topology.AddConnection(conn2)

	// Test counts
	healthyNodes := topology.CountHealthyNodes()
	if healthyNodes != 2 {
		t.Errorf("Expected 2 healthy nodes, got %d", healthyNodes)
	}

	healthyConns := topology.CountHealthyConnections()
	if healthyConns != 1 {
		t.Errorf("Expected 1 healthy connection, got %d", healthyConns)
	}
}
