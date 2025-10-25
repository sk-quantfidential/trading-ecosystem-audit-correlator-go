# TopologyService gRPC Testing Guide

## Prerequisites

Before testing, you need to:

1. **Merge the PR** (or checkout the feature branch)
2. **Rebuild the Docker container** (includes gRPC reflection support)
3. **Restart the service**

## Important Note

The gRPC server now includes **reflection support** (added in commit 79e6e4e), which allows `grpcurl` to discover services without proto files. After rebuilding the Docker container, all commands below will work.

## Quick Start - After Deployment

### Step 1: Check Container Status

```bash
# Check if audit-correlator is running
docker ps | grep audit-correlator

# Expected output:
# trading-ecosystem-audit-correlator   Up X minutes   0.0.0.0:50052->50051/tcp
```

### Step 2: List Available Services

```bash
# List all gRPC services on audit-correlator
grpcurl -plaintext localhost:50052 list

# Expected output:
# audit.v1.TopologyService
# grpc.health.v1.Health
```

### Step 3: List Service Methods

```bash
# List all methods on TopologyService
grpcurl -plaintext localhost:50052 list audit.v1.TopologyService

# Expected output:
# audit.v1.TopologyService.GetEdgeMetadata
# audit.v1.TopologyService.GetNodeMetadata
# audit.v1.TopologyService.GetTopologyStructure
# audit.v1.TopologyService.StreamMetricsUpdates
# audit.v1.TopologyService.StreamTopologyChanges
```

## Testing Individual RPCs

### 1. GetTopologyStructure (Unary RPC)

Get the full network topology structure:

```bash
# Get all topology (no filters)
grpcurl -plaintext \
  -d '{"request_id": "test-1"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure

# Expected response:
# {
#   "nodes": [],
#   "edges": [],
#   "snapshotTime": "2025-10-25T...",
#   "snapshotId": "snapshot-...",
#   "requestId": "test-1"
# }
```

**Note**: Initially empty because the in-memory repository has no data yet.

#### Filter by Service Types

```bash
grpcurl -plaintext \
  -d '{
    "service_types": ["risk-monitor-py", "trading-system-engine-py"],
    "request_id": "test-2"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure
```

#### Filter by Node Status

```bash
grpcurl -plaintext \
  -d '{
    "statuses": ["NODE_STATUS_LIVE", "NODE_STATUS_DEGRADED"],
    "request_id": "test-3"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure
```

### 2. GetNodeMetadata (Unary RPC)

Get detailed metadata for specific nodes:

```bash
# Get metadata for specific nodes
grpcurl -plaintext \
  -d '{
    "node_ids": ["node-1", "node-2"],
    "request_id": "test-4"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetNodeMetadata

# Expected response:
# {
#   "metadata": {
#     "node-1": {
#       "basicInfo": {...},
#       "healthMetrics": {...},
#       "endpoints": {...}
#     }
#   },
#   "requestId": "test-4"
# }
```

#### Get metadata for all nodes (empty node_ids)

```bash
grpcurl -plaintext \
  -d '{
    "node_ids": [],
    "request_id": "test-5"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetNodeMetadata
```

### 3. GetEdgeMetadata (Unary RPC)

Get detailed metadata for specific edges:

```bash
grpcurl -plaintext \
  -d '{
    "edge_ids": ["edge-1", "edge-2"],
    "request_id": "test-6"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetEdgeMetadata

# Expected response:
# {
#   "metadata": {
#     "edge-1": {
#       "metrics": {
#         "latencyP50Ms": 12.5,
#         "latencyP99Ms": 45.2,
#         "throughputRps": 1500,
#         "errorRate": 0.001
#       },
#       "details": {
#         "protocol": "gRPC",
#         "establishedAt": "2025-10-25T..."
#       }
#     }
#   },
#   "requestId": "test-6"
# }
```

### 4. StreamTopologyChanges (Server Streaming)

Subscribe to real-time topology changes:

```bash
# Stream topology changes (Ctrl+C to stop)
grpcurl -plaintext \
  -d '{
    "service_types": ["risk-monitor-py"]
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamTopologyChanges

# Expected output (streaming):
# {
#   "timestamp": "2025-10-25T...",
#   "snapshotId": "snapshot-123",
#   "nodeAdded": {
#     "node": {
#       "id": "node-3",
#       "name": "Risk Monitor LH",
#       "serviceType": "risk-monitor-py",
#       "status": "NODE_STATUS_LIVE"
#     }
#   }
# }
# {
#   "timestamp": "2025-10-25T...",
#   "snapshotId": "snapshot-124",
#   "nodeStatusChanged": {
#     "nodeId": "node-1",
#     "oldStatus": "NODE_STATUS_LIVE",
#     "newStatus": "NODE_STATUS_DEGRADED"
#   }
# }
```

#### Resume from snapshot ID

```bash
grpcurl -plaintext \
  -d '{
    "from_snapshot_id": "snapshot-100",
    "service_types": []
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamTopologyChanges
```

### 5. StreamMetricsUpdates (Server Streaming)

Subscribe to real-time metrics updates:

```bash
# Stream metrics for all nodes/edges (1 second interval)
grpcurl -plaintext \
  -d '{
    "node_ids": [],
    "edge_ids": [],
    "update_interval": "1s",
    "request_id": "test-7"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamMetricsUpdates

# Expected output (streaming every 1 second):
# {
#   "timestamp": "2025-10-25T...",
#   "nodeMetrics": {
#     "nodeId": "node-1",
#     "metrics": {
#       "cpuPercent": 45.2,
#       "memoryMb": 512.8,
#       "totalRequests": 10523,
#       "totalErrors": 12,
#       "errorRate": 0.0011,
#       "measuredAt": "2025-10-25T..."
#     }
#   }
# }
```

#### Stream specific nodes only

```bash
grpcurl -plaintext \
  -d '{
    "node_ids": ["node-1", "node-2"],
    "update_interval": "2s"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamMetricsUpdates
```

#### Stream edge metrics

```bash
grpcurl -plaintext \
  -d '{
    "edge_ids": ["edge-1", "edge-2"],
    "update_interval": "1s"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamMetricsUpdates
```

## Health Check

Verify the TopologyService is registered and healthy:

```bash
# Check health of TopologyService
grpcurl -plaintext \
  -d '{"service": "audit.v1.TopologyService"}' \
  localhost:50052 \
  grpc.health.v1.Health/Check

# Expected response:
# {
#   "status": "SERVING"
# }
```

## Testing with Reflection (Describe Methods)

Get detailed information about any method:

```bash
# Describe GetTopologyStructure method
grpcurl -plaintext \
  localhost:50052 \
  describe audit.v1.TopologyService.GetTopologyStructure

# Expected output:
# audit.v1.TopologyService.GetTopologyStructure is a method:
# rpc GetTopologyStructure ( .audit.v1.GetTopologyStructureRequest ) returns ( .audit.v1.TopologyStructureResponse );
```

```bash
# Describe request message
grpcurl -plaintext \
  localhost:50052 \
  describe audit.v1.GetTopologyStructureRequest

# Shows all fields in the request message
```

## Common Testing Scenarios

### Scenario 1: Initial Topology Load

```bash
# 1. Get current topology
grpcurl -plaintext \
  -d '{"request_id": "initial-load"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure

# 2. If nodes found, get their metadata
grpcurl -plaintext \
  -d '{"node_ids": [], "request_id": "metadata-load"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetNodeMetadata
```

### Scenario 2: Real-Time Monitoring

```bash
# Terminal 1: Stream topology changes
grpcurl -plaintext \
  -d '{}' \
  localhost:50052 \
  audit.v1.TopologyService/StreamTopologyChanges

# Terminal 2: Stream metrics updates
grpcurl -plaintext \
  -d '{"update_interval": "1s"}' \
  localhost:50052 \
  audit.v1.TopologyService/StreamMetricsUpdates
```

### Scenario 3: Filter by Service Type

```bash
# Get only risk monitoring services
grpcurl -plaintext \
  -d '{
    "service_types": ["risk-monitor-py"],
    "request_id": "risk-only"
  }' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure

# Stream changes for trading services only
grpcurl -plaintext \
  -d '{
    "service_types": ["trading-system-engine-py"]
  }' \
  localhost:50052 \
  audit.v1.TopologyService/StreamTopologyChanges
```

## Alternative: Testing with Proto Files (No Reflection Needed)

If you need to test before the Docker container is rebuilt, you can use the proto files directly:

```bash
# From the protobuf-schemas directory
cd /path/to/protobuf-schemas

# Test using proto file
grpcurl -plaintext \
  -proto audit/v1/topology_service.proto \
  -import-path . \
  -d '{"request_id": "test-1"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure
```

This works with or without reflection support, but requires access to the proto files.

## Troubleshooting

### Error: "server does not support the reflection API"

```
Error:
  failed to query for service descriptor "audit.v1.TopologyService":
  server does not support the reflection API
```

**Solutions**:
1. **Rebuild Docker container** with latest code (includes reflection support as of commit 79e6e4e)
2. **Use proto files directly** (see "Alternative: Testing with Proto Files" above)
3. **Verify the branch** has the reflection commit: `git log --oneline | grep reflection`

### Error: "Failed to dial target host"

```
Error:
  Failed to dial target host "localhost:50052": context deadline exceeded
```

**Solutions**:
1. Check if Docker container is running: `docker ps | grep audit-correlator`
2. Check if port is mapped correctly: `docker port trading-ecosystem-audit-correlator`
3. Check container logs: `docker logs trading-ecosystem-audit-correlator`

### Error: "Service not found"

```
Error:
  Code: Unimplemented
  Message: unknown service audit.v1.TopologyService
```

**Solutions**:
1. Verify the feature branch is merged or checked out
2. Rebuild Docker container: `docker-compose build audit-correlator`
3. Restart container: `docker-compose up -d audit-correlator`
4. Check logs for startup errors: `docker logs trading-ecosystem-audit-correlator --tail 50`

### Empty Response (No Nodes/Edges)

```json
{
  "nodes": [],
  "edges": [],
  "snapshotTime": "2025-10-25T...",
  "snapshotId": "snapshot-...",
  "requestId": "test-1"
}
```

**Explanation**: The in-memory repository starts empty. In a real deployment, the TopologyService would be populated by:
1. Service discovery watching Redis for registered services
2. Prometheus metrics collector scraping node health
3. Manual registration via internal APIs (future work)

**For testing purposes**, this is expected behavior until milestone TSE-0002.4 (production service discovery adapter) is implemented.

## Advanced: JSON Input from File

Create a file `request.json`:

```json
{
  "service_types": ["risk-monitor-py", "trading-system-engine-py"],
  "statuses": ["NODE_STATUS_LIVE"],
  "request_id": "from-file"
}
```

Use it with grpcurl:

```bash
grpcurl -plaintext \
  -d @ \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure \
  < request.json
```

## Next Steps

1. **After PR merge**: Rebuild container and test all methods
2. **Verify browser connectivity**: http://localhost:3002/topology should work
3. **Monitor logs**: Watch for TopologyService initialization messages
4. **Future work**: Implement production service discovery adapter to auto-populate topology

## Proto Schema Reference

Full proto schema available at:
- `protobuf-schemas/audit/v1/topology_service.proto`
- Generated Go code: `audit-correlator-go/gen/go/audit/v1/`

For detailed message definitions, use:
```bash
grpcurl -plaintext localhost:50052 describe audit.v1.NodeSummary
grpcurl -plaintext localhost:50052 describe audit.v1.EdgeSummary
grpcurl -plaintext localhost:50052 describe audit.v1.NodeMetadata
```
