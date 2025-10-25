package topology

import (
	"context"
	"testing"
	"time"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

// Mock MetadataRepository
type mockMetadataRepository struct {
	nodesMetadata map[string]*entities.NodeMetadata
	edgesMetadata map[string]*entities.EdgeMetadata
	err           error
}

func (m *mockMetadataRepository) SaveNodeMetadata(ctx context.Context, metadata *entities.NodeMetadata) error {
	return m.err
}

func (m *mockMetadataRepository) GetNodeMetadata(ctx context.Context, nodeID string) (*entities.NodeMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	if metadata, exists := m.nodesMetadata[nodeID]; exists {
		return metadata, nil
	}
	return nil, nil
}

func (m *mockMetadataRepository) GetNodesMetadata(ctx context.Context, nodeIDs []string) ([]*entities.NodeMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*entities.NodeMetadata, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if metadata, exists := m.nodesMetadata[nodeID]; exists {
			result = append(result, metadata)
		}
	}
	return result, nil
}

func (m *mockMetadataRepository) SaveEdgeMetadata(ctx context.Context, metadata *entities.EdgeMetadata) error {
	return m.err
}

func (m *mockMetadataRepository) GetEdgeMetadata(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	if metadata, exists := m.edgesMetadata[edgeID]; exists {
		return metadata, nil
	}
	return nil, nil
}

func (m *mockMetadataRepository) GetEdgesMetadata(ctx context.Context, edgeIDs []string) ([]*entities.EdgeMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*entities.EdgeMetadata, 0, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		if metadata, exists := m.edgesMetadata[edgeID]; exists {
			result = append(result, metadata)
		}
	}
	return result, nil
}

func TestGetNodeMetadataUseCase_Execute(t *testing.T) {
	// Create test metadata
	metadata1 := entities.NewNodeMetadata("node-1")
	metadata1.Uptime = 3600 * time.Second
	metadata1.HealthScore = 95.5

	metadata2 := entities.NewNodeMetadata("node-2")
	metadata2.Uptime = 1800 * time.Second
	metadata2.HealthScore = 82.3

	metadataRepo := &mockMetadataRepository{
		nodesMetadata: map[string]*entities.NodeMetadata{
			"node-1": metadata1,
			"node-2": metadata2,
		},
	}

	// Create topology for "get all" scenario
	topology := entities.NewNetworkTopology("snapshot-123")
	node1 := entities.NewServiceNode("node-1", "test1", "type1", "instance1")
	node2 := entities.NewServiceNode("node-2", "test2", "type2", "instance2")
	topology.AddNode(node1)
	topology.AddNode(node2)

	topologyRepo := &mockTopologyRepository{
		topology: topology,
	}

	tests := []struct {
		name           string
		request        *GetNodeMetadataRequest
		expectError    bool
		expectMetadata int
	}{
		{
			name: "get specific node metadata",
			request: &GetNodeMetadataRequest{
				NodeIDs:   []string{"node-1"},
				RequestID: "req-1",
			},
			expectError:    false,
			expectMetadata: 1,
		},
		{
			name: "get multiple nodes metadata",
			request: &GetNodeMetadataRequest{
				NodeIDs:   []string{"node-1", "node-2"},
				RequestID: "req-2",
			},
			expectError:    false,
			expectMetadata: 2,
		},
		{
			name: "get all nodes metadata (empty NodeIDs)",
			request: &GetNodeMetadataRequest{
				NodeIDs:   []string{},
				RequestID: "req-3",
			},
			expectError:    false,
			expectMetadata: 2,
		},
		{
			name:        "nil request returns error",
			request:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewGetNodeMetadataUseCase(metadataRepo, topologyRepo)
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

			if len(resp.Metadata) != tt.expectMetadata {
				t.Errorf("Expected %d metadata entries, got %d", tt.expectMetadata, len(resp.Metadata))
			}
		})
	}
}

func TestGetNodeMetadataUseCase_RepositoryError(t *testing.T) {
	metadataRepo := &mockMetadataRepository{
		err: context.DeadlineExceeded,
	}

	topologyRepo := &mockTopologyRepository{}

	useCase := NewGetNodeMetadataUseCase(metadataRepo, topologyRepo)
	_, err := useCase.Execute(context.Background(), &GetNodeMetadataRequest{
		NodeIDs:   []string{"node-1"},
		RequestID: "req-1",
	})

	if err == nil {
		t.Error("Expected error when repository fails")
	}
}
