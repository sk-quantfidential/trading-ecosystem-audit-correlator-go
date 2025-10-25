# Topology Service Integration Guide

## Overview

The topology service is now fully implemented following Clean Architecture with all layers complete:

- ✅ **Domain Layer** (TSE-0002.2): Entities, value objects, domain services
- ✅ **Application Layer** (TSE-0002.3): Use cases and ports
- ✅ **Infrastructure Layer** (TSE-0002.4): In-memory adapters
- ✅ **Service Layer** (TSE-0002.5): Dependency wiring

## Current Status

**Completed Components**:
- `TopologyService` - Wires all dependencies following Clean Architecture
- In-memory implementations for development/testing
- End-to-end tests validating full stack

**Pending Integration**:
- gRPC/Connect proto generation
- gRPC presentation layer (TopologyServiceServer)
- Registration in main.go

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Presentation Layer (gRPC/Connect) - TODO                │
│ ├─ TopologyServiceServer (from proto generation)        │
│ └─ Connect-RPC handlers for browser support             │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Service Layer (Dependency Wiring) - ✅ COMPLETE         │
│ └─ TopologyService (services/topology_service.go)       │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Application Layer (Use Cases) - ✅ COMPLETE             │
│ ├─ GetTopologyStructureUseCase                          │
│ ├─ GetNodeMetadataUseCase                               │
│ ├─ GetEdgeMetadataUseCase                               │
│ ├─ StreamTopologyChangesUseCase                         │
│ └─ StreamMetricsUpdatesUseCase                          │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Domain Layer (Business Logic) - ✅ COMPLETE             │
│ ├─ Entities: ServiceNode, ServiceConnection             │
│ ├─ Value Objects: NodeMetadata, TopologyFilters         │
│ └─ Ports: Repositories, Publishers, Collectors          │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Infrastructure Layer (Adapters) - ✅ COMPLETE           │
│ ├─ MemoryTopologyRepository                             │
│ ├─ MemoryMetadataRepository                             │
│ ├─ ChannelChangePublisher                               │
│ └─ MockMetricsCollector                                 │
└─────────────────────────────────────────────────────────┘
```

## Usage Example

```go
package main

import (
    "context"
    "github.com/sirupsen/logrus"
    "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
    "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/application/usecases/topology"
    "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

func main() {
    logger := logrus.New()

    // Initialize topology service (all layers wired automatically)
    topologyService := services.NewTopologyService(logger)

    ctx := context.Background()

    // Add topology data
    node := entities.NewServiceNode("node-1", "risk-monitor", "risk-monitor-py", "instance-1")
    topologyService.TopologyRepository().SaveNode(ctx, node)

    // Query through use cases
    req := &topology.GetTopologyStructureRequest{
        ServiceTypes: []string{"risk-monitor-py"},
        RequestID: "req-1",
    }

    resp, err := topologyService.GetTopologyStructureUseCase().Execute(ctx, req)
    if err != nil {
        logger.WithError(err).Error("Failed to get topology")
        return
    }

    logger.WithField("nodes", len(resp.Nodes)).Info("Topology retrieved")
}
```

## Next Steps: Proto Generation & gRPC Service

### Step 1: Generate Proto Code

The topology service proto schema was created in `simulator-ui-js` (TSE-0002.1). To generate Go code:

1. **Copy proto schema** to `proto/api/audit/v1/topology_service.proto`
2. **Run proto generation**:
   ```bash
   cd proto
   make generate
   ```
3. **Generated files** will be in `gen/go/audit/v1/`

### Step 2: Implement gRPC Service

Create `internal/presentation/grpc/services/topology_service_server.go`:

```go
package services

import (
    "context"
    "github.com/sirupsen/logrus"

    // Generated proto imports
    auditv1 "github.com/quantfidential/trading-ecosystem/audit-correlator-go/gen/go/audit/v1"

    "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/services"
    "github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/application/usecases/topology"
)

type TopologyServiceServer struct {
    auditv1.UnimplementedTopologyServiceServer
    topologyService *services.TopologyService
    logger          *logrus.Logger
}

func NewTopologyServiceServer(topologyService *services.TopologyService, logger *logrus.Logger) *TopologyServiceServer {
    return &TopologyServiceServer{
        topologyService: topologyService,
        logger:          logger,
    }
}

func (s *TopologyServiceServer) GetTopologyStructure(
    ctx context.Context,
    req *auditv1.GetTopologyStructureRequest,
) (*auditv1.GetTopologyStructureResponse, error) {
    // Convert proto request to use case request
    ucReq := &topology.GetTopologyStructureRequest{
        ServiceTypes: req.ServiceTypes,
        Statuses:     convertNodeStatuses(req.Statuses),
        RequestID:    req.RequestId,
    }

    // Execute use case
    ucResp, err := s.topologyService.GetTopologyStructureUseCase().Execute(ctx, ucReq)
    if err != nil {
        return nil, err
    }

    // Convert use case response to proto response
    return &auditv1.GetTopologyStructureResponse{
        Nodes:        convertNodesToProto(ucResp.Nodes),
        Edges:        convertEdgesToProto(ucResp.Edges),
        SnapshotId:   ucResp.SnapshotID,
        SnapshotTime: timestampProto(ucResp.SnapshotTime),
    }, nil
}

// Implement other 4 RPCs similarly...
```

### Step 3: Add Connect/gRPC-Web Support

For browser clients, add Connect-RPC support in main.go:

```go
import (
    "connectrpc.com/connect"
    "connectrpc.com/grpchealth"
    "connectrpc.com/grpcreflect"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"
)

// In main()
topologyService := services.NewTopologyService(logger)
topologyServer := grpcservices.NewTopologyServiceServer(topologyService, logger)

// Create Connect-RPC mux
mux := http.NewServeMux()

// Register topology service with Connect protocol (browser-compatible)
path, handler := auditv1connect.NewTopologyServiceHandler(topologyServer)
mux.Handle(path, handler)

// Enable CORS for browser clients
corsHandler := cors.New(cors.Options{
    AllowedOrigins: []string{"http://localhost:3002"},
    AllowedMethods: []string{"GET", "POST"},
    AllowedHeaders: []string{"*"},
}).Handler(mux)

// Start HTTP server with h2c (HTTP/2 Cleartext) for Connect
httpServer := &http.Server{
    Addr:    ":50051",
    Handler: h2c.NewHandler(corsHandler, &http2.Server{}),
}
```

### Step 4: Register in main.go

```go
// In main.go
topologyService := services.NewTopologyService(logger)

// gRPC registration
grpcServer := grpc.NewServer()
topologyServer := grpcservices.NewTopologyServiceServer(topologyService, logger)
auditv1.RegisterTopologyServiceServer(grpcServer, topologyServer)

// Start gRPC server
go func() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        logger.WithError(err).Fatal("Failed to listen on gRPC port")
    }
    if err := grpcServer.Serve(lis); err != nil {
        logger.WithError(err).Fatal("Failed to start gRPC server")
    }
}()
```

## Testing

### Unit Tests
```bash
# All layers already tested
go test ./internal/domain/entities/...        # 27 tests
go test ./internal/application/usecases/...   # 8 tests
go test ./internal/infrastructure/topology/... # 5 tests
go test ./internal/services/...               # 2 tests
```

### Integration Tests (after gRPC implementation)
```bash
# Test with grpcurl
grpcurl -plaintext localhost:50051 audit.v1.TopologyService/GetTopologyStructure

# Test from simulator-ui-js browser
# Navigate to http://localhost:3002/topology
```

## Benefits of Current Implementation

1. **Clean Architecture**: Strict dependency rules enforced
2. **Testable**: Each layer tested in isolation (42 tests total)
3. **Flexible**: Easy to swap in-memory adapters for PostgreSQL, Redis, etc.
4. **Ready**: Just needs proto generation and presentation layer

## Files Reference

**Service Layer**:
- `internal/services/topology_service.go` - Main service with dependency wiring
- `internal/services/topology_service_test.go` - End-to-end service tests

**Application Layer**:
- `internal/application/usecases/topology/*.go` - 5 use cases + tests

**Domain Layer**:
- `internal/domain/entities/*.go` - Entities, value objects + tests
- `internal/domain/ports/topology.go` - Port interfaces
- `internal/domain/services/topology_tracker.go` - Domain service interface

**Infrastructure Layer**:
- `internal/infrastructure/topology/*.go` - 4 adapters + tests
