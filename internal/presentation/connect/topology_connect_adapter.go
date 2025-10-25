package connectpresentation

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"

	auditv1 "github.com/quantfidential/trading-ecosystem/audit-correlator-go/gen/go/audit/v1"
	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/gen/go/audit/v1/auditv1connect"
	grpcservices "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/presentation/grpc/services"
)

// TopologyConnectAdapter adapts the gRPC TopologyService to Connect protocol
// Implements auditv1connect.TopologyServiceHandler interface
type TopologyConnectAdapter struct {
	grpcServer *grpcservices.TopologyServiceServer
}

// Ensure TopologyConnectAdapter implements TopologyServiceHandler
var _ auditv1connect.TopologyServiceHandler = (*TopologyConnectAdapter)(nil)

// NewTopologyConnectAdapter creates a new Connect-compatible topology adapter
func NewTopologyConnectAdapter(grpcServer *grpcservices.TopologyServiceServer) *TopologyConnectAdapter {
	return &TopologyConnectAdapter{
		grpcServer: grpcServer,
	}
}

// GetTopologyStructure implements the Connect handler for GetTopologyStructure
func (h *TopologyConnectAdapter) GetTopologyStructure(
	ctx context.Context,
	req *connect.Request[auditv1.GetTopologyStructureRequest],
) (*connect.Response[auditv1.TopologyStructureResponse], error) {
	resp, err := h.grpcServer.GetTopologyStructure(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetNodeMetadata implements the Connect handler for GetNodeMetadata
func (h *TopologyConnectAdapter) GetNodeMetadata(
	ctx context.Context,
	req *connect.Request[auditv1.GetNodeMetadataRequest],
) (*connect.Response[auditv1.GetNodeMetadataResponse], error) {
	resp, err := h.grpcServer.GetNodeMetadata(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetEdgeMetadata implements the Connect handler for GetEdgeMetadata
func (h *TopologyConnectAdapter) GetEdgeMetadata(
	ctx context.Context,
	req *connect.Request[auditv1.GetEdgeMetadataRequest],
) (*connect.Response[auditv1.GetEdgeMetadataResponse], error) {
	resp, err := h.grpcServer.GetEdgeMetadata(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// StreamTopologyChanges implements the Connect handler for StreamTopologyChanges
func (h *TopologyConnectAdapter) StreamTopologyChanges(
	ctx context.Context,
	req *connect.Request[auditv1.StreamTopologyChangesRequest],
	stream *connect.ServerStream[auditv1.TopologyChange],
) error {
	// Create a streaming adapter
	streamAdapter := &topologyChangeStreamAdapter{stream: stream, ctx: ctx}
	return h.grpcServer.StreamTopologyChanges(req.Msg, streamAdapter)
}

// StreamMetricsUpdates implements the Connect handler for StreamMetricsUpdates
func (h *TopologyConnectAdapter) StreamMetricsUpdates(
	ctx context.Context,
	req *connect.Request[auditv1.StreamMetricsUpdatesRequest],
	stream *connect.ServerStream[auditv1.MetricsUpdate],
) error {
	// Create a streaming adapter
	streamAdapter := &metricsUpdateStreamAdapter{stream: stream, ctx: ctx}
	return h.grpcServer.StreamMetricsUpdates(req.Msg, streamAdapter)
}

// topologyChangeStreamAdapter adapts Connect ServerStream to gRPC stream
type topologyChangeStreamAdapter struct {
	stream *connect.ServerStream[auditv1.TopologyChange]
	ctx    context.Context
}

func (s *topologyChangeStreamAdapter) Send(msg *auditv1.TopologyChange) error {
	return s.stream.Send(msg)
}

func (s *topologyChangeStreamAdapter) Context() context.Context {
	return s.ctx
}

// Implement required gRPC stream methods (unused but needed for interface)
func (s *topologyChangeStreamAdapter) SetHeader(md metadata.MD) error  { return nil }
func (s *topologyChangeStreamAdapter) SendHeader(md metadata.MD) error { return nil }
func (s *topologyChangeStreamAdapter) SetTrailer(md metadata.MD)       {}
func (s *topologyChangeStreamAdapter) SendMsg(m interface{}) error     { return nil }
func (s *topologyChangeStreamAdapter) RecvMsg(m interface{}) error     { return nil }

// metricsUpdateStreamAdapter adapts Connect ServerStream to gRPC stream
type metricsUpdateStreamAdapter struct {
	stream *connect.ServerStream[auditv1.MetricsUpdate]
	ctx    context.Context
}

func (s *metricsUpdateStreamAdapter) Send(msg *auditv1.MetricsUpdate) error {
	return s.stream.Send(msg)
}

func (s *metricsUpdateStreamAdapter) Context() context.Context {
	return s.ctx
}

// Implement required gRPC stream methods (unused but needed for interface)
func (s *metricsUpdateStreamAdapter) SetHeader(md metadata.MD) error  { return nil }
func (s *metricsUpdateStreamAdapter) SendHeader(md metadata.MD) error { return nil }
func (s *metricsUpdateStreamAdapter) SetTrailer(md metadata.MD)       {}
func (s *metricsUpdateStreamAdapter) SendMsg(m interface{}) error     { return nil }
func (s *metricsUpdateStreamAdapter) RecvMsg(m interface{}) error     { return nil }
