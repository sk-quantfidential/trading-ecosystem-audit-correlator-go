package topology

import (
	"context"
	"fmt"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// GetEdgeMetadataRequest represents the request to get edge metadata
type GetEdgeMetadataRequest struct {
	EdgeIDs   []string
	RequestID string
}

// GetEdgeMetadataResponse represents the edge metadata response
type GetEdgeMetadataResponse struct {
	Metadata []*entities.EdgeMetadata
}

// GetEdgeMetadataUseCase handles retrieving metadata for one or more edges
type GetEdgeMetadataUseCase struct {
	metadataRepo ports.MetadataRepository
	topologyRepo ports.TopologyRepository
}

// NewGetEdgeMetadataUseCase creates a new GetEdgeMetadataUseCase
func NewGetEdgeMetadataUseCase(metadataRepo ports.MetadataRepository, topologyRepo ports.TopologyRepository) *GetEdgeMetadataUseCase {
	return &GetEdgeMetadataUseCase{
		metadataRepo: metadataRepo,
		topologyRepo: topologyRepo,
	}
}

// Execute retrieves metadata for specified edges
func (uc *GetEdgeMetadataUseCase) Execute(ctx context.Context, req *GetEdgeMetadataRequest) (*GetEdgeMetadataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var metadata []*entities.EdgeMetadata

	// If no edge IDs specified, get all edges
	if len(req.EdgeIDs) == 0 {
		topology, err := uc.topologyRepo.GetTopologySnapshot(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get topology snapshot: %w", err)
		}

		// Extract edge IDs from topology
		edgeIDs := make([]string, 0, len(topology.Connections))
		for edgeID := range topology.Connections {
			edgeIDs = append(edgeIDs, edgeID)
		}

		metadata, err = uc.metadataRepo.GetEdgesMetadata(ctx, edgeIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get edges metadata: %w", err)
		}
	} else {
		// Get metadata for specified edges
		var err error
		metadata, err = uc.metadataRepo.GetEdgesMetadata(ctx, req.EdgeIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get edges metadata: %w", err)
		}
	}

	return &GetEdgeMetadataResponse{
		Metadata: metadata,
	}, nil
}
