package topology

import (
	"context"
	"testing"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

// Mock TopologyRepository
type mockTopologyRepository struct {
	topology *entities.NetworkTopology
	err      error
}

func (m *mockTopologyRepository) SaveNode(ctx context.Context, node *entities.ServiceNode) error {
	return m.err
}

func (m *mockTopologyRepository) GetNode(ctx context.Context, nodeID string) (*entities.ServiceNode, error) {
	if m.topology != nil {
		if node, exists := m.topology.Nodes[nodeID]; exists {
			return node, nil
		}
	}
	return nil, m.err
}

func (m *mockTopologyRepository) GetNodes(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	nodes := make([]*entities.ServiceNode, 0)
	if m.topology != nil {
		for _, node := range m.topology.Nodes {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (m *mockTopologyRepository) DeleteNode(ctx context.Context, nodeID string) error {
	return m.err
}

func (m *mockTopologyRepository) SaveConnection(ctx context.Context, conn *entities.ServiceConnection) error {
	return m.err
}

func (m *mockTopologyRepository) GetConnection(ctx context.Context, connID string) (*entities.ServiceConnection, error) {
	if m.topology != nil {
		if conn, exists := m.topology.Connections[connID]; exists {
			return conn, nil
		}
	}
	return nil, m.err
}

func (m *mockTopologyRepository) GetConnections(ctx context.Context, filters *entities.TopologyFilters) ([]*entities.ServiceConnection, error) {
	if m.err != nil {
		return nil, m.err
	}
	connections := make([]*entities.ServiceConnection, 0)
	if m.topology != nil {
		for _, conn := range m.topology.Connections {
			connections = append(connections, conn)
		}
	}
	return connections, nil
}

func (m *mockTopologyRepository) DeleteConnection(ctx context.Context, connID string) error {
	return m.err
}

func (m *mockTopologyRepository) GetTopologySnapshot(ctx context.Context) (*entities.NetworkTopology, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.topology, nil
}

func (m *mockTopologyRepository) SaveTopologySnapshot(ctx context.Context, topology *entities.NetworkTopology) error {
	return m.err
}

func TestGetTopologyStructureUseCase_Execute(t *testing.T) {
	// Create test topology
	topology := entities.NewNetworkTopology("snapshot-123")

	node1 := entities.NewServiceNode("node-1", "risk-monitor", "risk-monitor-py", "risk-monitor-lh")
	node1.Status = entities.NodeStatusLive
	topology.AddNode(node1)

	node2 := entities.NewServiceNode("node-2", "exchange", "exchange-simulator-go", "exchange-okx")
	node2.Status = entities.NodeStatusDegraded
	topology.AddNode(node2)

	conn1 := entities.NewServiceConnection("conn-1", "node-1", "node-2", entities.ConnectionTypeGRPC)
	topology.AddConnection(conn1)

	tests := []struct {
		name        string
		repo        *mockTopologyRepository
		request     *GetTopologyStructureRequest
		expectError bool
		expectNodes int
		expectEdges int
	}{
		{
			name: "successful topology retrieval",
			repo: &mockTopologyRepository{
				topology: topology,
			},
			request: &GetTopologyStructureRequest{
				RequestID: "req-1",
			},
			expectError: false,
			expectNodes: 2,
			expectEdges: 1,
		},
		{
			name: "filter by service type",
			repo: &mockTopologyRepository{
				topology: topology,
			},
			request: &GetTopologyStructureRequest{
				ServiceTypes: []string{"risk-monitor-py"},
				RequestID:    "req-2",
			},
			expectError: false,
			expectNodes: 1,
			expectEdges: 1,
		},
		{
			name: "filter by status",
			repo: &mockTopologyRepository{
				topology: topology,
			},
			request: &GetTopologyStructureRequest{
				Statuses:  []entities.NodeStatus{entities.NodeStatusLive},
				RequestID: "req-3",
			},
			expectError: false,
			expectNodes: 1,
			expectEdges: 1,
		},
		{
			name: "nil request returns error",
			repo: &mockTopologyRepository{
				topology: topology,
			},
			request:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewGetTopologyStructureUseCase(tt.repo)
			resp, err := useCase.Execute(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Fatal("Expected response but got nil")
			}

			if len(resp.Nodes) != tt.expectNodes {
				t.Errorf("Expected %d nodes, got %d", tt.expectNodes, len(resp.Nodes))
			}

			if len(resp.Edges) != tt.expectEdges {
				t.Errorf("Expected %d edges, got %d", tt.expectEdges, len(resp.Edges))
			}

			if resp.SnapshotID != "snapshot-123" {
				t.Errorf("Expected snapshot ID 'snapshot-123', got %s", resp.SnapshotID)
			}
		})
	}
}

func TestGetTopologyStructureUseCase_RepositoryError(t *testing.T) {
	repo := &mockTopologyRepository{
		err: context.DeadlineExceeded,
	}

	useCase := NewGetTopologyStructureUseCase(repo)
	_, err := useCase.Execute(context.Background(), &GetTopologyStructureRequest{
		RequestID: "req-1",
	})

	if err == nil {
		t.Error("Expected error when repository fails")
	}
}
