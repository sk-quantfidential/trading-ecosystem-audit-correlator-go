package topology

import (
	"context"
	"fmt"
	"sync"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
)

// MemoryMetadataRepository implements MetadataRepository using in-memory storage
type MemoryMetadataRepository struct {
	mu            sync.RWMutex
	nodesMetadata map[string]*entities.NodeMetadata
	edgesMetadata map[string]*entities.EdgeMetadata
}

// NewMemoryMetadataRepository creates a new in-memory metadata repository
func NewMemoryMetadataRepository() *MemoryMetadataRepository {
	return &MemoryMetadataRepository{
		nodesMetadata: make(map[string]*entities.NodeMetadata),
		edgesMetadata: make(map[string]*entities.EdgeMetadata),
	}
}

// SaveNodeMetadata saves node metadata
func (r *MemoryMetadataRepository) SaveNodeMetadata(ctx context.Context, metadata *entities.NodeMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodesMetadata[metadata.NodeID] = metadata
	return nil
}

// GetNodeMetadata retrieves node metadata by ID
func (r *MemoryMetadataRepository) GetNodeMetadata(ctx context.Context, nodeID string) (*entities.NodeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.nodesMetadata[nodeID]
	if !exists {
		return nil, fmt.Errorf("node metadata not found: %s", nodeID)
	}

	return metadata, nil
}

// GetNodesMetadata retrieves metadata for multiple nodes
func (r *MemoryMetadataRepository) GetNodesMetadata(ctx context.Context, nodeIDs []string) ([]*entities.NodeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entities.NodeMetadata, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if metadata, exists := r.nodesMetadata[nodeID]; exists {
			result = append(result, metadata)
		}
	}

	return result, nil
}

// SaveEdgeMetadata saves edge metadata
func (r *MemoryMetadataRepository) SaveEdgeMetadata(ctx context.Context, metadata *entities.EdgeMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.edgesMetadata[metadata.EdgeID] = metadata
	return nil
}

// GetEdgeMetadata retrieves edge metadata by ID
func (r *MemoryMetadataRepository) GetEdgeMetadata(ctx context.Context, edgeID string) (*entities.EdgeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.edgesMetadata[edgeID]
	if !exists {
		return nil, fmt.Errorf("edge metadata not found: %s", edgeID)
	}

	return metadata, nil
}

// GetEdgesMetadata retrieves metadata for multiple edges
func (r *MemoryMetadataRepository) GetEdgesMetadata(ctx context.Context, edgeIDs []string) ([]*entities.EdgeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entities.EdgeMetadata, 0, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		if metadata, exists := r.edgesMetadata[edgeID]; exists {
			result = append(result, metadata)
		}
	}

	return result, nil
}

var _ ports.MetadataRepository = (*MemoryMetadataRepository)(nil)
