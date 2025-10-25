package entities

import (
	"testing"
	"time"
)

func TestNodeMetadata_Creation(t *testing.T) {
	metadata := NewNodeMetadata("node-1")

	if metadata.NodeID != "node-1" {
		t.Errorf("Expected NodeID 'node-1', got %s", metadata.NodeID)
	}
	if metadata.HealthScore != 100.0 {
		t.Errorf("Expected initial HealthScore 100.0, got %f", metadata.HealthScore)
	}
	if metadata.CustomFields == nil {
		t.Error("Expected CustomFields map to be initialized")
	}
}

func TestNodeMetadata_UpdateHealthScore(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Valid score", 75.5, 75.5},
		{"Negative score clamped to 0", -10.0, 0.0},
		{"Score over 100 clamped to 100", 150.0, 100.0},
		{"Zero score", 0.0, 0.0},
		{"Max score", 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewNodeMetadata("node-1")
			metadata.UpdateHealthScore(tt.input)

			if metadata.HealthScore != tt.expected {
				t.Errorf("Expected HealthScore %f, got %f", tt.expected, metadata.HealthScore)
			}
		})
	}
}

func TestNodeMetadata_AddCustomField(t *testing.T) {
	metadata := NewNodeMetadata("node-1")
	metadata.AddCustomField("version", "1.0.0")
	metadata.AddCustomField("datacenter", "us-east-1")
	metadata.AddCustomField("replicas", 3)

	if metadata.CustomFields["version"] != "1.0.0" {
		t.Error("Expected version custom field to be '1.0.0'")
	}
	if metadata.CustomFields["datacenter"] != "us-east-1" {
		t.Error("Expected datacenter custom field to be 'us-east-1'")
	}
	if metadata.CustomFields["replicas"] != 3 {
		t.Error("Expected replicas custom field to be 3")
	}
}

func TestEdgeMetadata_Creation(t *testing.T) {
	metadata := NewEdgeMetadata("edge-1")

	if metadata.EdgeID != "edge-1" {
		t.Errorf("Expected EdgeID 'edge-1', got %s", metadata.EdgeID)
	}
	if metadata.LastActiveTime.IsZero() {
		t.Error("Expected LastActiveTime to be set")
	}
	if metadata.CustomFields == nil {
		t.Error("Expected CustomFields map to be initialized")
	}
}

func TestEdgeMetadata_UpdateThroughput(t *testing.T) {
	metadata := NewEdgeMetadata("edge-1")
	initialTime := metadata.LastActiveTime

	time.Sleep(10 * time.Millisecond)
	metadata.UpdateThroughput(1234.56)

	if metadata.Throughput != 1234.56 {
		t.Errorf("Expected Throughput 1234.56, got %f", metadata.Throughput)
	}
	if !metadata.LastActiveTime.After(initialTime) {
		t.Error("Expected LastActiveTime to be updated")
	}
}

func TestTopologyFilters_Creation(t *testing.T) {
	filters := NewTopologyFilters()

	if filters.ServiceTypes == nil {
		t.Error("Expected ServiceTypes to be initialized")
	}
	if filters.NodeStatuses == nil {
		t.Error("Expected NodeStatuses to be initialized")
	}
	if filters.EdgeStatuses == nil {
		t.Error("Expected EdgeStatuses to be initialized")
	}
	if filters.LabelFilters == nil {
		t.Error("Expected LabelFilters to be initialized")
	}
	if !filters.IncludeNodes {
		t.Error("Expected IncludeNodes to be true by default")
	}
	if !filters.IncludeEdges {
		t.Error("Expected IncludeEdges to be true by default")
	}
}

func TestTopologyFilters_AddFilters(t *testing.T) {
	filters := NewTopologyFilters()

	filters.AddServiceType("risk-monitor-py")
	filters.AddServiceType("exchange-simulator-go")
	filters.AddNodeStatus(NodeStatusLive)
	filters.AddEdgeStatus(EdgeStatusActive)
	filters.AddLabelFilter("environment", "production")

	if len(filters.ServiceTypes) != 2 {
		t.Errorf("Expected 2 service types, got %d", len(filters.ServiceTypes))
	}
	if len(filters.NodeStatuses) != 1 {
		t.Errorf("Expected 1 node status, got %d", len(filters.NodeStatuses))
	}
	if len(filters.EdgeStatuses) != 1 {
		t.Errorf("Expected 1 edge status, got %d", len(filters.EdgeStatuses))
	}
	if filters.LabelFilters["environment"] != "production" {
		t.Error("Expected environment label filter to be 'production'")
	}
}

func TestTopologyFilters_MatchesNode(t *testing.T) {
	tests := []struct {
		name         string
		setupNode    func() *ServiceNode
		setupFilters func() *TopologyFilters
		shouldMatch  bool
	}{
		{
			name: "Empty filters match any node",
			setupNode: func() *ServiceNode {
				return NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
			},
			setupFilters: func() *TopologyFilters {
				return NewTopologyFilters()
			},
			shouldMatch: true,
		},
		{
			name: "Service type filter matches",
			setupNode: func() *ServiceNode {
				return NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddServiceType("risk-monitor-py")
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Service type filter does not match",
			setupNode: func() *ServiceNode {
				return NewServiceNode("node-1", "test", "exchange-simulator-go", "instance1")
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddServiceType("risk-monitor-py")
				return f
			},
			shouldMatch: false,
		},
		{
			name: "Node status filter matches",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.Status = NodeStatusLive
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddNodeStatus(NodeStatusLive)
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Node status filter does not match",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.Status = NodeStatusDegraded
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddNodeStatus(NodeStatusLive)
				return f
			},
			shouldMatch: false,
		},
		{
			name: "Label filter matches",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.AddLabel("environment", "production")
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddLabelFilter("environment", "production")
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Label filter does not match",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.AddLabel("environment", "staging")
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddLabelFilter("environment", "production")
				return f
			},
			shouldMatch: false,
		},
		{
			name: "Multiple filters all match",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.Status = NodeStatusLive
				node.AddLabel("environment", "production")
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddServiceType("risk-monitor-py")
				f.AddNodeStatus(NodeStatusLive)
				f.AddLabelFilter("environment", "production")
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Multiple filters one does not match",
			setupNode: func() *ServiceNode {
				node := NewServiceNode("node-1", "test", "risk-monitor-py", "instance1")
				node.Status = NodeStatusDegraded
				node.AddLabel("environment", "production")
				return node
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddServiceType("risk-monitor-py")
				f.AddNodeStatus(NodeStatusLive)
				f.AddLabelFilter("environment", "production")
				return f
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setupNode()
			filters := tt.setupFilters()

			matches := filters.MatchesNode(node)
			if matches != tt.shouldMatch {
				t.Errorf("Expected MatchesNode() to return %v, got %v", tt.shouldMatch, matches)
			}
		})
	}
}

func TestTopologyFilters_MatchesConnection(t *testing.T) {
	tests := []struct {
		name            string
		setupConnection func() *ServiceConnection
		setupFilters    func() *TopologyFilters
		shouldMatch     bool
	}{
		{
			name: "Empty filters match any connection",
			setupConnection: func() *ServiceConnection {
				return NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
			},
			setupFilters: func() *TopologyFilters {
				return NewTopologyFilters()
			},
			shouldMatch: true,
		},
		{
			name: "Edge status filter matches",
			setupConnection: func() *ServiceConnection {
				conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
				conn.Status = EdgeStatusActive
				return conn
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddEdgeStatus(EdgeStatusActive)
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Edge status filter does not match",
			setupConnection: func() *ServiceConnection {
				conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
				conn.Status = EdgeStatusFailed
				return conn
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddEdgeStatus(EdgeStatusActive)
				return f
			},
			shouldMatch: false,
		},
		{
			name: "Label filter matches",
			setupConnection: func() *ServiceConnection {
				conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
				conn.AddLabel("protocol", "grpc")
				return conn
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddLabelFilter("protocol", "grpc")
				return f
			},
			shouldMatch: true,
		},
		{
			name: "Label filter does not match",
			setupConnection: func() *ServiceConnection {
				conn := NewServiceConnection("conn-1", "node-1", "node-2", ConnectionTypeGRPC)
				conn.AddLabel("protocol", "http")
				return conn
			},
			setupFilters: func() *TopologyFilters {
				f := NewTopologyFilters()
				f.AddLabelFilter("protocol", "grpc")
				return f
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.setupConnection()
			filters := tt.setupFilters()

			matches := filters.MatchesConnection(conn)
			if matches != tt.shouldMatch {
				t.Errorf("Expected MatchesConnection() to return %v, got %v", tt.shouldMatch, matches)
			}
		})
	}
}
