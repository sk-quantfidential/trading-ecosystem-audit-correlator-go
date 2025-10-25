package topology

import (
	"context"
	"fmt"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// GetTopologyStructureRequest represents the request to get topology structure
type GetTopologyStructureRequest struct {
	ServiceTypes []string
	Statuses     []entities.NodeStatus
	RequestID    string
}

// GetTopologyStructureResponse represents the topology structure response
type GetTopologyStructureResponse struct {
	Nodes        []*entities.ServiceNode
	Edges        []*entities.ServiceConnection
	SnapshotID   string
	SnapshotTime string
}

// GetTopologyStructureUseCase handles retrieving the current topology structure
type GetTopologyStructureUseCase struct {
	topologyRepo ports.TopologyRepository
}

// NewGetTopologyStructureUseCase creates a new GetTopologyStructureUseCase
func NewGetTopologyStructureUseCase(topologyRepo ports.TopologyRepository) *GetTopologyStructureUseCase {
	return &GetTopologyStructureUseCase{
		topologyRepo: topologyRepo,
	}
}

// Execute retrieves the topology structure with optional filtering
func (uc *GetTopologyStructureUseCase) Execute(ctx context.Context, req *GetTopologyStructureRequest) (*GetTopologyStructureResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Build filters
	filters := entities.NewTopologyFilters()
	for _, serviceType := range req.ServiceTypes {
		filters.AddServiceType(serviceType)
	}
	for _, status := range req.Statuses {
		filters.AddNodeStatus(status)
	}

	// Get topology snapshot
	topology, err := uc.topologyRepo.GetTopologySnapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get topology snapshot: %w", err)
	}

	// Apply filters to nodes
	var filteredNodes []*entities.ServiceNode
	for _, node := range topology.Nodes {
		if filters.MatchesNode(node) {
			filteredNodes = append(filteredNodes, node)
		}
	}

	// Get all connections (no filtering for now, but could add edge filters)
	var connections []*entities.ServiceConnection
	for _, conn := range topology.Connections {
		connections = append(connections, conn)
	}

	return &GetTopologyStructureResponse{
		Nodes:        filteredNodes,
		Edges:        connections,
		SnapshotID:   topology.SnapshotID,
		SnapshotTime: topology.SnapshotTime.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
