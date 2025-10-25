# feat(epic-TSE-0002): Add topology configuration loader for initial data

## Summary

Implements topology configuration loading from JSON file to populate the TopologyService with initial network structure on startup. This solves the "No topology data available" issue in the UI by loading service nodes and connections from a statically generated configuration file.

**Problem**: The TopologyService started with empty in-memory repositories, so the UI showed "No topology data available" even though the Connect protocol was working correctly.

**Solution**: Created a configuration loader that reads topology structure from a JSON file mounted via Docker volume, populating nodes and edges on service startup.

## What Changed

### Infrastructure Layer (`internal/infrastructure/topology/`)

- **`config_loader.go`** (193 lines) - New file:
  - `ConfigLoader` struct for loading topology from JSON
  - `TopologyConfig` structs matching JSON format
  - `LoadFromFile()` - Main loading function
  - `loadNode()` - Converts JSON node config to domain entity
  - `loadEdge()` - Converts JSON edge config to domain entity
  - Context-aware repository operations
  - Graceful handling of missing config files

### Service Layer (`internal/services/`)

- **`topology_service.go`** - Enhanced:
  - Added `GetTopologyRepository()` method
  - Added `LoadConfigFromFile()` method
  - Type assertion for memory repository access
  - Integration with ConfigLoader

### Presentation Layer (`cmd/server/`)

- **`main.go`** - Enhanced:
  - Topology config loading in `registerConnectHandlers()`
  - Config path: `/app/config/topology.json`
  - Graceful fallback if config missing
  - Logging of load success/failure

### Documentation

- **`docs/CONNECT_PROTOCOL_SETUP.md`** (New):
  - Complete setup guide for Connect protocol
  - Port mapping reference
  - Troubleshooting steps
  - Testing instructions

## Configuration File Format

The loader expects JSON in this format:

```json
{
  "version": "1.0",
  "generated_at": "startup",
  "nodes": [
    {
      "id": "node-audit-correlator",
      "name": "Audit Correlator",
      "service_type": "audit-correlator-go",
      "category": "monitoring",
      "status": "LIVE",
      "version": "1.0.0",
      "endpoints": {
        "grpc": "localhost:50052",
        "http": "localhost:8082",
        "internal_ip": "172.20.0.80"
      }
    }
  ],
  "edges": [
    {
      "id": "edge-risk-monitor-lh-to-trading-engine-lh",
      "source_id": "node-risk-monitor-lh",
      "target_id": "node-trading-engine-lh",
      "protocol": "gRPC",
      "relationship": "monitors",
      "status": "ACTIVE"
    }
  ]
}
```

## Testing

```bash
# Build verification
go build ./...
# ✅ Builds successfully

# Run all tests
go test ./... -short -count=1
# ✅ All 42 tests pass (no regression)

# Docker deployment test:
# 1. Generate config (in orchestrator-docker):
#    python3 scripts/generate-topology-config.py
#
# 2. Rebuild container:
#    docker-compose build --no-cache audit-correlator
#
# 3. Restart service:
#    docker-compose up -d audit-correlator
#
# 4. Check logs:
#    docker logs trading-ecosystem-audit-correlator | grep topology
#
# Expected output:
# {"level":"info","msg":"Loading topology configuration","config_path":"/app/config/topology.json"}
# {"level":"info","msg":"Parsed topology configuration","nodes":7,"edges":11}
# {"level":"info","msg":"Successfully loaded topology configuration","nodes_loaded":7,"edges_loaded":11}
```

## Integration with orchestrator-docker

This feature requires:
1. **Config generation script** (in orchestrator-docker repo)
2. **Volume mount** in docker-compose.yml
3. **topology.json** file generation before container start

See orchestrator-docker PR for the companion changes.

## Architecture Impact

✅ **Clean Architecture Maintained**:
- Config loader in infrastructure layer (adapter pattern)
- No changes to domain or application layers
- Service layer orchestrates loading at startup
- Presentation layer triggers load during initialization

✅ **Graceful Degradation**:
- Missing config file → Warning logged, empty topology
- Invalid JSON → Error logged, empty topology
- Invalid nodes/edges → Skipped with error log
- Service continues to run in all cases

✅ **Testability**:
- Config loader can be unit tested independently
- Repository injection enables mocking
- No side effects on existing tests

## Deployment Flow

### Production Startup Sequence

1. **Container starts** with mounted `/app/config` volume
2. **HTTP server initializes**
3. **registerConnectHandlers()** called
4. **TopologyService created**
5. **LoadConfigFromFile()** executes:
   - Reads `/app/config/topology.json`
   - Parses nodes and edges
   - Populates in-memory repositories
6. **Connect handlers registered** with populated data
7. **Service ready** - UI can fetch topology

### Configuration Update Flow

To update topology without code changes:
1. Regenerate `topology.json`
2. Restart audit-correlator container
3. New topology loaded automatically

## Problem Solved

**Original Issue**: "No topology data available" in UI

**Root Cause**: TopologyService in-memory repositories started empty with no initial data.

**Solution Verification**:
```bash
# After deployment:
curl -X POST http://localhost:8082/audit.v1.TopologyService/GetTopologyStructure \
  -H "Content-Type: application/json" \
  -d '{"request_id": "test"}'

# Before fix:
# {"nodes":[],"edges":[],"snapshotTime":"...","snapshotId":"..."}

# After fix:
# {"nodes":[{...7 nodes...}],"edges":[{...11 edges...}],"snapshotTime":"..."}
```

## Related Work

- **Depends On**:
  - Epic TSE-0002 TopologyService implementation - ✅ merged
  - Connect protocol support - ✅ merged

- **Requires (orchestrator-docker)**:
  - Config generation script (`scripts/generate-topology-config.py`)
  - Docker volume mount (`./config:/app/config:ro`)
  - Generated `config/topology.json` file

- **Enables**:
  - ✅ **Immediate**: UI shows actual network topology
  - ✅ **Immediate**: D3.js visualization with 7 nodes, 11 edges
  - Future: Dynamic topology updates from service discovery
  - Future: Topology persistence to database

## Future Enhancements

This static config loader is **Phase 1** of topology population:

- **Phase 1** (This PR): Static config file loading ✅
- **Phase 2** (Future): Service discovery integration
- **Phase 3** (Future): Real-time metrics collection
- **Phase 4** (Future): Topology persistence to PostgreSQL

The config loader provides immediate value while dynamic discovery is implemented.

## Epic Context

**Epic**: TSE-0002 - Network Topology Visualization
**Type**: Feature (configuration loading)
**Status**: ✅ Ready for testing with orchestrator-docker changes

This completes the data layer needed for full topology visualization in the UI.

## Branch Information

- **Branch**: `feature/epic-TSE-0002-topology-config-loader`
- **Base**: `main`
- **Type**: `feature` (new functionality)
- **Epic**: TSE-0002
- **Milestone**: Infrastructure enhancement

## Checklist

- [x] Code follows Clean Architecture principles
- [x] All tests pass (42 tests, no regression)
- [x] Build succeeds (`go build ./...`)
- [x] Config loader gracefully handles errors
- [x] Missing config file doesn't crash service
- [x] Logging at appropriate levels (info/warn/error)
- [x] JSON parsing with proper error handling
- [x] Domain entity conversion implemented
- [x] Integration with service layer complete
- [x] PR documentation complete
- [x] Branch name follows `feature/epic-XXX-description` format
- [x] Ready for Docker testing with orchestrator-docker
