package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	auditv1 "github.com/quantfidential/trading-ecosystem/audit-correlator-go/gen/go/audit/v1"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/application/usecases/topology"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/ports"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
)

// TopologyServiceServer implements the gRPC TopologyService
type TopologyServiceServer struct {
	auditv1.UnimplementedTopologyServiceServer
	topologyService *services.TopologyService
	logger          *logrus.Logger
}

// NewTopologyServiceServer creates a new TopologyServiceServer
func NewTopologyServiceServer(topologyService *services.TopologyService, logger *logrus.Logger) *TopologyServiceServer {
	return &TopologyServiceServer{
		topologyService: topologyService,
		logger:          logger,
	}
}

// GetTopologyStructure returns lightweight topology structure for initial render
func (s *TopologyServiceServer) GetTopologyStructure(
	ctx context.Context,
	req *auditv1.GetTopologyStructureRequest,
) (*auditv1.TopologyStructureResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"request_id":    req.RequestId,
		"service_types": req.ServiceTypes,
		"statuses":      req.Statuses,
	}).Debug("GetTopologyStructure called")

	// Convert proto request to use case request
	ucReq := &topology.GetTopologyStructureRequest{
		ServiceTypes: req.ServiceTypes,
		Statuses:     convertProtoNodeStatuses(req.Statuses),
		RequestID:    req.RequestId,
	}

	// Execute use case
	ucResp, err := s.topologyService.GetTopologyStructureUseCase().Execute(ctx, ucReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get topology structure")
		return nil, fmt.Errorf("failed to get topology structure: %w", err)
	}

	// Convert use case response to proto response
	// Parse the SnapshotTime string back to time.Time
	snapshotTime, err := time.Parse(time.RFC3339, ucResp.SnapshotTime)
	if err != nil {
		snapshotTime = time.Now() // Fallback to current time if parsing fails
	}

	return &auditv1.TopologyStructureResponse{
		Nodes:        convertNodesToProto(ucResp.Nodes),
		Edges:        convertEdgesToProto(ucResp.Edges),
		SnapshotTime: timestamppb.New(snapshotTime),
		SnapshotId:   ucResp.SnapshotID,
		RequestId:    req.RequestId,
	}, nil
}

// GetNodeMetadata returns detailed metadata for specific nodes
func (s *TopologyServiceServer) GetNodeMetadata(
	ctx context.Context,
	req *auditv1.GetNodeMetadataRequest,
) (*auditv1.GetNodeMetadataResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"request_id": req.RequestId,
		"node_ids":   req.NodeIds,
	}).Debug("GetNodeMetadata called")

	// Convert proto request to use case request
	ucReq := &topology.GetNodeMetadataRequest{
		NodeIDs:   req.NodeIds,
		RequestID: req.RequestId,
	}

	// Execute use case
	ucResp, err := s.topologyService.GetNodeMetadataUseCase().Execute(ctx, ucReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get node metadata")
		return nil, fmt.Errorf("failed to get node metadata: %w", err)
	}

	// Convert use case response to proto response
	metadataMap := make(map[string]*auditv1.NodeMetadata)
	for _, metadata := range ucResp.Metadata {
		metadataMap[metadata.NodeID] = convertNodeMetadataToProto(metadata)
	}

	return &auditv1.GetNodeMetadataResponse{
		Metadata:  metadataMap,
		RequestId: req.RequestId,
	}, nil
}

// GetEdgeMetadata returns detailed metadata for specific edges
func (s *TopologyServiceServer) GetEdgeMetadata(
	ctx context.Context,
	req *auditv1.GetEdgeMetadataRequest,
) (*auditv1.GetEdgeMetadataResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"request_id": req.RequestId,
		"edge_ids":   req.EdgeIds,
	}).Debug("GetEdgeMetadata called")

	// Convert proto request to use case request
	ucReq := &topology.GetEdgeMetadataRequest{
		EdgeIDs:   req.EdgeIds,
		RequestID: req.RequestId,
	}

	// Execute use case
	ucResp, err := s.topologyService.GetEdgeMetadataUseCase().Execute(ctx, ucReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get edge metadata")
		return nil, fmt.Errorf("failed to get edge metadata: %w", err)
	}

	// Convert use case response to proto response
	metadataMap := make(map[string]*auditv1.EdgeMetadata)
	for _, metadata := range ucResp.Metadata {
		metadataMap[metadata.EdgeID] = convertEdgeMetadataToProto(metadata)
	}

	return &auditv1.GetEdgeMetadataResponse{
		Metadata:  metadataMap,
		RequestId: req.RequestId,
	}, nil
}

// StreamTopologyChanges streams structural topology changes
func (s *TopologyServiceServer) StreamTopologyChanges(
	req *auditv1.StreamTopologyChangesRequest,
	stream auditv1.TopologyService_StreamTopologyChangesServer,
) error {
	s.logger.WithFields(logrus.Fields{
		"from_snapshot_id": req.FromSnapshotId,
		"service_types":    req.ServiceTypes,
	}).Debug("StreamTopologyChanges called")

	// Convert proto request to use case request
	ucReq := &topology.StreamTopologyChangesRequest{
		FromSnapshotID: req.FromSnapshotId,
		ServiceTypes:   req.ServiceTypes,
	}

	// Execute use case (returns a channel)
	ctx := stream.Context()
	eventsChan, err := s.topologyService.StreamTopologyChangesUseCase().Execute(ctx, ucReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start topology changes stream")
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Stream events to client
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("StreamTopologyChanges context cancelled")
			return ctx.Err()
		case event, ok := <-eventsChan:
			if !ok {
				s.logger.Debug("StreamTopologyChanges channel closed")
				return nil
			}
			protoEvent := convertTopologyChangeToProto(event)
			if err := stream.Send(protoEvent); err != nil {
				s.logger.WithError(err).Error("Failed to send topology change")
				return err
			}
		}
	}
}

// StreamMetricsUpdates streams high-frequency metrics updates
func (s *TopologyServiceServer) StreamMetricsUpdates(
	req *auditv1.StreamMetricsUpdatesRequest,
	stream auditv1.TopologyService_StreamMetricsUpdatesServer,
) error {
	s.logger.WithFields(logrus.Fields{
		"request_id":    req.RequestId,
		"node_ids":      req.NodeIds,
		"edge_ids":      req.EdgeIds,
		"service_types": req.ServiceTypes,
	}).Debug("StreamMetricsUpdates called")

	// Convert proto request to use case request
	interval := time.Second
	if req.UpdateInterval != nil {
		interval = req.UpdateInterval.AsDuration()
	}

	ucReq := &topology.StreamMetricsUpdatesRequest{
		NodeIDs:        req.NodeIds,
		EdgeIDs:        req.EdgeIds,
		UpdateInterval: interval,
		RequestID:      req.RequestId,
	}

	// Execute use case (returns a channel)
	ctx := stream.Context()
	metricsChan, err := s.topologyService.StreamMetricsUpdatesUseCase().Execute(ctx, ucReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start metrics stream")
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Stream metrics to client
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("StreamMetricsUpdates context cancelled")
			return ctx.Err()
		case update, ok := <-metricsChan:
			if !ok {
				s.logger.Debug("StreamMetricsUpdates channel closed")
				return nil
			}
			protoUpdate := convertMetricsUpdateToProto(update)
			if err := stream.Send(protoUpdate); err != nil {
				s.logger.WithError(err).Error("Failed to send metrics update")
				return err
			}
		}
	}
}

// ========== Conversion Functions ==========

func convertProtoNodeStatuses(statuses []auditv1.NodeStatus) []entities.NodeStatus {
	result := make([]entities.NodeStatus, 0, len(statuses))
	for _, status := range statuses {
		switch status {
		case auditv1.NodeStatus_NODE_STATUS_LIVE:
			result = append(result, entities.NodeStatusLive)
		case auditv1.NodeStatus_NODE_STATUS_DEGRADED:
			result = append(result, entities.NodeStatusDegraded)
		case auditv1.NodeStatus_NODE_STATUS_DEAD:
			result = append(result, entities.NodeStatusDead)
		}
	}
	return result
}

func convertNodesToProto(nodes []*entities.ServiceNode) []*auditv1.NodeSummary {
	result := make([]*auditv1.NodeSummary, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, &auditv1.NodeSummary{
			Id:           node.ID,
			Name:         node.Name,
			ServiceType:  node.ServiceType,
			InstanceName: node.InstanceName,
			Status:       convertNodeStatusToProto(node.Status),
			Labels:       node.Labels,
		})
	}
	return result
}

func convertNodeStatusToProto(status entities.NodeStatus) auditv1.NodeStatus {
	switch status {
	case entities.NodeStatusLive:
		return auditv1.NodeStatus_NODE_STATUS_LIVE
	case entities.NodeStatusDegraded:
		return auditv1.NodeStatus_NODE_STATUS_DEGRADED
	case entities.NodeStatusDead:
		return auditv1.NodeStatus_NODE_STATUS_DEAD
	default:
		return auditv1.NodeStatus_NODE_STATUS_UNSPECIFIED
	}
}

func convertEdgesToProto(edges []*entities.ServiceConnection) []*auditv1.EdgeSummary {
	result := make([]*auditv1.EdgeSummary, 0, len(edges))
	for _, edge := range edges {
		result = append(result, &auditv1.EdgeSummary{
			Id:         edge.ID,
			SourceId:   edge.SourceID,
			TargetId:   edge.TargetID,
			Type:       convertConnectionTypeToProto(edge.Type),
			Status:     convertEdgeStatusToProto(edge.Status),
			IsCritical: false, // TODO: Add this field to domain entity
		})
	}
	return result
}

func convertConnectionTypeToProto(connType entities.ConnectionType) auditv1.ConnectionType {
	switch connType {
	case entities.ConnectionTypeGRPC:
		return auditv1.ConnectionType_CONNECTION_TYPE_GRPC
	case entities.ConnectionTypeHTTP:
		return auditv1.ConnectionType_CONNECTION_TYPE_HTTP
	case entities.ConnectionTypeDataFlow:
		return auditv1.ConnectionType_CONNECTION_TYPE_DATA_FLOW
	default:
		return auditv1.ConnectionType_CONNECTION_TYPE_UNSPECIFIED
	}
}

func convertEdgeStatusToProto(status entities.EdgeStatus) auditv1.EdgeStatus {
	switch status {
	case entities.EdgeStatusActive:
		return auditv1.EdgeStatus_EDGE_STATUS_ACTIVE
	case entities.EdgeStatusDegraded:
		return auditv1.EdgeStatus_EDGE_STATUS_DEGRADED
	case entities.EdgeStatusFailed:
		return auditv1.EdgeStatus_EDGE_STATUS_FAILED
	default:
		return auditv1.EdgeStatus_EDGE_STATUS_UNSPECIFIED
	}
}

func convertNodeMetadataToProto(metadata *entities.NodeMetadata) *auditv1.NodeMetadata {
	return &auditv1.NodeMetadata{
		BasicInfo: &auditv1.BasicInfo{
			Version:     "1.0.0",         // TODO: Add this field to domain entity
			StartedAt:   timestamppb.New(time.Now().Add(-metadata.Uptime)),
			LastSeen:    timestamppb.New(time.Now()),
			Environment: "docker",        // TODO: Add this field to domain entity
		},
		HealthMetrics: &auditv1.HealthMetrics{
			CpuPercent:    metadata.CPUUsage,
			MemoryMb:      metadata.MemoryUsage,
			TotalRequests: metadata.TotalReqs,
			TotalErrors:   int64(metadata.ErrorRate * float64(metadata.TotalReqs)),
			ErrorRate:     metadata.ErrorRate,
			MeasuredAt:    timestamppb.New(time.Now()),
		},
		Endpoints: &auditv1.EndpointInfo{
			GrpcEndpoints:    []string{}, // TODO: Add this field to domain entity
			HttpEndpoints:    []string{}, // TODO: Add this field to domain entity
			MetricsEndpoints: []string{}, // TODO: Add this field to domain entity
		},
		Configuration: make(map[string]string), // TODO: Add this field to domain entity
	}
}

func convertEdgeMetadataToProto(metadata *entities.EdgeMetadata) *auditv1.EdgeMetadata {
	return &auditv1.EdgeMetadata{
		Metrics: &auditv1.ConnectionMetrics{
			LatencyP50Ms:   float64(metadata.AvgLatency.Milliseconds()),
			LatencyP99Ms:   float64(metadata.P99Latency.Milliseconds()),
			ThroughputRps:  int64(metadata.Throughput),
			ErrorRate:      metadata.ErrorRate,
			TotalBytesSent: metadata.TotalMessages, // Approximation
			MeasuredAt:     timestamppb.New(time.Now()),
		},
		Details: &auditv1.ConnectionDetails{
			Protocol:      metadata.Protocol,
			Methods:       []string{}, // TODO: Add this field to domain entity
			EstablishedAt: timestamppb.New(metadata.LastActiveTime),
		},
	}
}

func convertTopologyChangeToProto(event *ports.TopologyChangeEvent) *auditv1.TopologyChange {
	change := &auditv1.TopologyChange{
		Timestamp:  timestamppb.New(event.Timestamp),
		SnapshotId: event.SnapshotID,
	}

	switch event.ChangeType {
	case ports.TopologyChangeTypeNodeAdded:
		if event.Node != nil {
			change.Change = &auditv1.TopologyChange_NodeAdded{
				NodeAdded: &auditv1.NodeAdded{
					Node: &auditv1.NodeSummary{
						Id:           event.Node.ID,
						Name:         event.Node.Name,
						ServiceType:  event.Node.ServiceType,
						InstanceName: event.Node.InstanceName,
						Status:       convertNodeStatusToProto(event.Node.Status),
						Labels:       event.Node.Labels,
					},
				},
			}
		}
	case ports.TopologyChangeTypeNodeRemoved:
		if event.Node != nil {
			change.Change = &auditv1.TopologyChange_NodeRemoved{
				NodeRemoved: &auditv1.NodeRemoved{
					NodeId: event.Node.ID,
					Reason: "removed",
				},
			}
		}
	case ports.TopologyChangeTypeEdgeAdded:
		if event.Connection != nil {
			change.Change = &auditv1.TopologyChange_EdgeAdded{
				EdgeAdded: &auditv1.EdgeAdded{
					Edge: &auditv1.EdgeSummary{
						Id:       event.Connection.ID,
						SourceId: event.Connection.SourceID,
						TargetId: event.Connection.TargetID,
						Type:     convertConnectionTypeToProto(event.Connection.Type),
						Status:   convertEdgeStatusToProto(event.Connection.Status),
					},
				},
			}
		}
	}

	return change
}

func convertMetricsUpdateToProto(update *ports.MetricsUpdate) *auditv1.MetricsUpdate {
	result := &auditv1.MetricsUpdate{
		Timestamp: timestamppb.New(update.Timestamp),
	}

	// Extract metrics from the generic map
	if update.NodeID != "" {
		result.Update = &auditv1.MetricsUpdate_NodeMetrics{
			NodeMetrics: &auditv1.NodeMetricsUpdate{
				NodeId: update.NodeID,
				Metrics: &auditv1.HealthMetrics{
					CpuPercent:    getFloat64FromMap(update.Metrics, "cpu_percent"),
					MemoryMb:      getFloat64FromMap(update.Metrics, "memory_mb"),
					TotalRequests: getInt64FromMap(update.Metrics, "total_requests"),
					TotalErrors:   getInt64FromMap(update.Metrics, "total_errors"),
					ErrorRate:     getFloat64FromMap(update.Metrics, "error_rate"),
					MeasuredAt:    timestamppb.New(update.Timestamp),
				},
			},
		}
	} else if update.EdgeID != "" {
		result.Update = &auditv1.MetricsUpdate_EdgeMetrics{
			EdgeMetrics: &auditv1.EdgeMetricsUpdate{
				EdgeId: update.EdgeID,
				Metrics: &auditv1.ConnectionMetrics{
					LatencyP50Ms:   getFloat64FromMap(update.Metrics, "latency_p50_ms"),
					LatencyP99Ms:   getFloat64FromMap(update.Metrics, "latency_p99_ms"),
					ThroughputRps:  getInt64FromMap(update.Metrics, "throughput_rps"),
					ErrorRate:      getFloat64FromMap(update.Metrics, "error_rate"),
					TotalBytesSent: getInt64FromMap(update.Metrics, "total_bytes_sent"),
					MeasuredAt:     timestamppb.New(update.Timestamp),
				},
			},
		}
	}

	return result
}

// Helper functions to extract typed values from generic map
func getFloat64FromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
		if i, ok := val.(int); ok {
			return int64(i)
		}
	}
	return 0
}
