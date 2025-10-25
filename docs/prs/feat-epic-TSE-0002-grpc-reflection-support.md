# feat(epic-TSE-0002): Add gRPC Reflection Support for Command-Line Testing

## Summary

Adds gRPC reflection API support to enable command-line testing with `grpcurl` and other tools. This enhancement solves the "server does not support the reflection API" error encountered when testing the TopologyService from the command line.

This is a small but critical enhancement to the Epic TSE-0002 implementation that enables easier testing and debugging.

## What Changed

### Server Configuration (`internal/presentation/grpc/server.go`)

**Added gRPC Reflection Registration**:
- Import: `"google.golang.org/grpc/reflection"`
- Registration: `reflection.Register(server.server)` after all service registrations
- Log message: "gRPC server initialized with reflection support"

**Why This Matters**:
- Enables `grpcurl` to discover services without proto files
- Allows runtime introspection of available services and methods
- Standard practice for gRPC services in development/testing
- Zero impact on production performance

### Documentation (`docs/TOPOLOGY_GRPC_TESTING.md`)

**New Comprehensive Testing Guide** (458 lines):
- Prerequisites and setup instructions
- Examples for all 5 RPC methods:
  - `GetTopologyStructure` (unary)
  - `GetNodeMetadata` (unary)
  - `GetEdgeMetadata` (unary)
  - `StreamTopologyChanges` (server-streaming)
  - `StreamMetricsUpdates` (server-streaming)
- Common testing scenarios
- Health check examples
- Troubleshooting section
- Alternative method using proto files directly

## Testing

```bash
# Build verification
go build ./...
# ✅ Builds successfully

# Run all tests
go test ./... -short -count=1
# ✅ All 42 tests pass (no regression)

# Manual testing (after Docker rebuild):
# List available services
grpcurl -plaintext localhost:50052 list

# Expected output:
# audit.v1.TopologyService
# grpc.health.v1.Health

# Test GetTopologyStructure
grpcurl -plaintext \
  -d '{"request_id": "test-1"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure

# Test health check
grpcurl -plaintext \
  -d '{"service": "audit.v1.TopologyService"}' \
  localhost:50052 \
  grpc.health.v1.Health/Check
```

**Test Coverage**: No new tests needed - this is a presentation layer enhancement that doesn't affect business logic. All existing 42 tests continue to pass.

## Problem Solved

**User-Reported Error**:
```
Error invoking method "audit.v1.TopologyService/StreamTopologyChanges":
failed to query for service descriptor "audit.v1.TopologyService":
server does not support the reflection API
```

**Root Cause**: gRPC server lacked reflection support, preventing `grpcurl` from discovering services at runtime.

**Solution**: Added `reflection.Register(server.server)` to enable runtime service discovery.

## Architecture Impact

**Zero Changes to Architecture**:
- ✅ Clean Architecture layers unchanged
- ✅ Domain logic unaffected
- ✅ No new dependencies introduced (reflection is part of `google.golang.org/grpc`)
- ✅ All 42 tests still passing
- ✅ Build succeeds

**Enhancement Only**:
- Improves developer experience for testing
- Enables easier debugging with standard tools
- Standard practice for gRPC services

## Deployment Notes

### After PR Merge

1. **Rebuild Docker Container**:
   ```bash
   cd ../orchestrator-docker
   docker-compose build audit-correlator
   docker-compose up -d audit-correlator
   ```

2. **Verify Reflection Works**:
   ```bash
   # Should now list services successfully
   grpcurl -plaintext localhost:50052 list
   ```

3. **Test TopologyService**:
   ```bash
   # All commands from docs/TOPOLOGY_GRPC_TESTING.md should work
   grpcurl -plaintext localhost:50052 list audit.v1.TopologyService
   ```

### Before Container Rebuild

If you need to test before rebuilding, use the proto file method:
```bash
cd /path/to/protobuf-schemas
grpcurl -plaintext \
  -proto audit/v1/topology_service.proto \
  -import-path . \
  -d '{"request_id": "test-1"}' \
  localhost:50052 \
  audit.v1.TopologyService/GetTopologyStructure
```

## Related Work

- **Depends On**:
  - Epic TSE-0002 implementation (all 6 milestones) - ✅ merged to main
  - TopologyService gRPC server registration - ✅ merged to main

- **Enables**:
  - Command-line testing with `grpcurl` without proto files
  - Runtime service introspection for debugging
  - Easier manual testing during development
  - Standard tooling integration (grpcurl, grpc_cli, etc.)

- **Solves**:
  - ✅ "server does not support the reflection API" error
  - ✅ Need for proto files during manual testing
  - ✅ Lack of runtime service discovery

## Epic Context

**Epic**: TSE-0002 - Network Topology Visualization
**Type**: Enhancement (post-implementation improvement)
**Status**: ✅ Ready for validation and PR

This is a small enhancement to the completed Epic TSE-0002 implementation that improves testability without changing core functionality.

## Branch Information

- **Branch**: `feature/epic-TSE-0002-grpc-reflection-support`
- **Base**: `main`
- **Type**: `feature` (enhancement)
- **Epic**: TSE-0002
- **Scope**: Developer experience improvement

## Checklist

- [x] Code follows Clean Architecture principles
- [x] All tests pass (42 tests, no regression)
- [x] Build succeeds (`go build ./...`)
- [x] Reflection registration added to gRPC server
- [x] Comprehensive testing guide created (458 lines)
- [x] Troubleshooting documentation included
- [x] Zero impact on existing functionality
- [x] Standard gRPC best practice implemented
- [x] PR documentation complete
- [x] Branch name follows `feature/epic-XXX-description` format
- [x] Ready for validation suite and PR creation
