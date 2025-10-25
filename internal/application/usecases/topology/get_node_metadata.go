package topology

import (
	"context"
	"fmt"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// GetNodeMetadataRequest represents the request to get node metadata
type GetNodeMetadataRequest struct {
	NodeIDs   []string
	RequestID string
}

// GetNodeMetadataResponse represents the node metadata response
type GetNodeMetadataResponse struct {
	Metadata []*entities.NodeMetadata
}

// GetNodeMetadataUseCase handles retrieving metadata for one or more nodes
type GetNodeMetadataUseCase struct {
	metadataRepo ports.MetadataRepository
	topologyRepo ports.TopologyRepository
}

// NewGetNodeMetadataUseCase creates a new GetNodeMetadataUseCase
func NewGetNodeMetadataUseCase(metadataRepo ports.MetadataRepository, topologyRepo ports.TopologyRepository) *GetNodeMetadataUseCase {
	return &GetNodeMetadataUseCase{
		metadataRepo: metadataRepo,
		topologyRepo: topologyRepo,
	}
}

// Execute retrieves metadata for specified nodes
func (uc *GetNodeMetadataUseCase) Execute(ctx context.Context, req *GetNodeMetadataRequest) (*GetNodeMetadataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var metadata []*entities.NodeMetadata

	// If no node IDs specified, get all nodes
	if len(req.NodeIDs) == 0 {
		topology, err := uc.topologyRepo.GetTopologySnapshot(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get topology snapshot: %w", err)
		}

		// Extract node IDs from topology
		nodeIDs := make([]string, 0, len(topology.Nodes))
		for nodeID := range topology.Nodes {
			nodeIDs = append(nodeIDs, nodeID)
		}

		metadata, err = uc.metadataRepo.GetNodesMetadata(ctx, nodeIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get nodes metadata: %w", err)
		}
	} else {
		// Get metadata for specified nodes
		var err error
		metadata, err = uc.metadataRepo.GetNodesMetadata(ctx, req.NodeIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get nodes metadata: %w", err)
		}
	}

	return &GetNodeMetadataResponse{
		Metadata: metadata,
	}, nil
}
