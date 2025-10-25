package services

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/application/usecases/topology"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

func TestNewTopologyService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	service := NewTopologyService(logger)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.GetTopologyStructureUseCase() == nil {
		t.Error("Expected GetTopologyStructureUseCase to be initialized")
	}

	if service.GetNodeMetadataUseCase() == nil {
		t.Error("Expected GetNodeMetadataUseCase to be initialized")
	}

	if service.GetEdgeMetadataUseCase() == nil {
		t.Error("Expected GetEdgeMetadataUseCase to be initialized")
	}

	if service.StreamTopologyChangesUseCase() == nil {
		t.Error("Expected StreamTopologyChangesUseCase to be initialized")
	}

	if service.StreamMetricsUpdatesUseCase() == nil {
		t.Error("Expected StreamMetricsUpdatesUseCase to be initialized")
	}

	if service.TopologyRepository() == nil {
		t.Error("Expected TopologyRepository to be initialized")
	}

	if service.MetadataRepository() == nil {
		t.Error("Expected MetadataRepository to be initialized")
	}

	if service.ChangePublisher() == nil {
		t.Error("Expected ChangePublisher to be initialized")
	}

	if service.MetricsCollector() == nil {
		t.Error("Expected MetricsCollector to be initialized")
	}
}

func TestTopologyService_EndToEnd(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	service := NewTopologyService(logger)
	ctx := context.Background()

	// Add a node to the topology repository
	node := entities.NewServiceNode("node-1", "risk-monitor", "risk-monitor-py", "risk-monitor-lh")
	err := service.TopologyRepository().SaveNode(ctx, node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	// Add a connection
	conn := entities.NewServiceConnection("conn-1", "node-1", "node-2", entities.ConnectionTypeGRPC)
	err = service.TopologyRepository().SaveConnection(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to save connection: %v", err)
	}

	// Query topology through use case
	req := &topology.GetTopologyStructureRequest{
		RequestID: "test-req-1",
	}

	resp, err := service.GetTopologyStructureUseCase().Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to get topology structure: %v", err)
	}

	if len(resp.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(resp.Nodes))
	}

	if len(resp.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(resp.Edges))
	}

	// Add metadata
	metadata := entities.NewNodeMetadata("node-1")
	metadata.HealthScore = 95.5
	err = service.MetadataRepository().SaveNodeMetadata(ctx, metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Query metadata through use case
	metadataReq := &topology.GetNodeMetadataRequest{
		NodeIDs:   []string{"node-1"},
		RequestID: "test-req-2",
	}

	metadataResp, err := service.GetNodeMetadataUseCase().Execute(ctx, metadataReq)
	if err != nil {
		t.Fatalf("Failed to get node metadata: %v", err)
	}

	if len(metadataResp.Metadata) != 1 {
		t.Errorf("Expected 1 metadata entry, got %d", len(metadataResp.Metadata))
	}

	if metadataResp.Metadata[0].HealthScore != 95.5 {
		t.Errorf("Expected health score 95.5, got %f", metadataResp.Metadata[0].HealthScore)
	}
}
