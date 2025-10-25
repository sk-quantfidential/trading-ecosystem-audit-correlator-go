package services

import (
	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/application/usecases/topology"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
	infratopology "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/infrastructure/topology"
)

// TopologyService provides access to topology-related functionality
// This service wires together all the topology components following Clean Architecture
type TopologyService struct {
	// Repositories (Infrastructure layer)
	topologyRepo ports.TopologyRepository
	metadataRepo ports.MetadataRepository

	// Publishers and Collectors (Infrastructure layer)
	changePublisher  ports.TopologyChangePublisher
	metricsCollector ports.MetricsCollector

	// Use Cases (Application layer)
	getTopologyStructure *topology.GetTopologyStructureUseCase
	getNodeMetadata      *topology.GetNodeMetadataUseCase
	getEdgeMetadata      *topology.GetEdgeMetadataUseCase
	streamChanges        *topology.StreamTopologyChangesUseCase
	streamMetrics        *topology.StreamMetricsUpdatesUseCase

	logger *logrus.Logger
}

// NewTopologyService creates a new topology service with all dependencies wired
func NewTopologyService(logger *logrus.Logger) *TopologyService {
	// Initialize infrastructure layer (adapters)
	topologyRepo := infratopology.NewMemoryTopologyRepository()
	metadataRepo := infratopology.NewMemoryMetadataRepository()
	changePublisher := infratopology.NewChannelChangePublisher()
	metricsCollector := infratopology.NewMockMetricsCollector()

	// Initialize application layer (use cases)
	getTopologyStructure := topology.NewGetTopologyStructureUseCase(topologyRepo)
	getNodeMetadata := topology.NewGetNodeMetadataUseCase(metadataRepo, topologyRepo)
	getEdgeMetadata := topology.NewGetEdgeMetadataUseCase(metadataRepo, topologyRepo)
	streamChanges := topology.NewStreamTopologyChangesUseCase(changePublisher)
	streamMetrics := topology.NewStreamMetricsUpdatesUseCase(metricsCollector)

	return &TopologyService{
		topologyRepo:         topologyRepo,
		metadataRepo:         metadataRepo,
		changePublisher:      changePublisher,
		metricsCollector:     metricsCollector,
		getTopologyStructure: getTopologyStructure,
		getNodeMetadata:      getNodeMetadata,
		getEdgeMetadata:      getEdgeMetadata,
		streamChanges:        streamChanges,
		streamMetrics:        streamMetrics,
		logger:               logger,
	}
}

// GetTopologyStructureUseCase returns the use case for getting topology structure
func (s *TopologyService) GetTopologyStructureUseCase() *topology.GetTopologyStructureUseCase {
	return s.getTopologyStructure
}

// GetNodeMetadataUseCase returns the use case for getting node metadata
func (s *TopologyService) GetNodeMetadataUseCase() *topology.GetNodeMetadataUseCase {
	return s.getNodeMetadata
}

// GetEdgeMetadataUseCase returns the use case for getting edge metadata
func (s *TopologyService) GetEdgeMetadataUseCase() *topology.GetEdgeMetadataUseCase {
	return s.getEdgeMetadata
}

// StreamTopologyChangesUseCase returns the use case for streaming topology changes
func (s *TopologyService) StreamTopologyChangesUseCase() *topology.StreamTopologyChangesUseCase {
	return s.streamChanges
}

// StreamMetricsUpdatesUseCase returns the use case for streaming metrics updates
func (s *TopologyService) StreamMetricsUpdatesUseCase() *topology.StreamMetricsUpdatesUseCase {
	return s.streamMetrics
}

// TopologyRepository returns the topology repository for direct access (if needed)
func (s *TopologyService) TopologyRepository() ports.TopologyRepository {
	return s.topologyRepo
}

// MetadataRepository returns the metadata repository for direct access (if needed)
func (s *TopologyService) MetadataRepository() ports.MetadataRepository {
	return s.metadataRepo
}

// ChangePublisher returns the change publisher for direct access (if needed)
func (s *TopologyService) ChangePublisher() ports.TopologyChangePublisher {
	return s.changePublisher
}

// MetricsCollector returns the metrics collector for direct access (if needed)
func (s *TopologyService) MetricsCollector() ports.MetricsCollector {
	return s.metricsCollector
}
