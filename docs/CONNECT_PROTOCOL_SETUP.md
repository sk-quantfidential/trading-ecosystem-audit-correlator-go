# Connect Protocol Setup - Browser Integration Fix

## Issue Resolved

**Original Error**:
```
[gRPC Error] audit.v1.TopologyService.GetTopologyStructure {}
NetworkError when attempting to fetch resource
```

**Root Cause**:
- Browsers cannot use native gRPC protocol directly
- UI was pointing to wrong port (gRPC server instead of HTTP server with Connect support)

## Solution

### 1. Backend Changes ✅ (Already Deployed)

Added Connect protocol support to audit-correlator:
- Connect handlers on HTTP server (port 8080 in container)
- CORS middleware for browser requests
- HTTP/2 support for streaming
- All 5 TopologyService RPCs available via Connect protocol

**Container Rebuilt**: audit-correlator now includes Connect support

### 2. UI Configuration Change ✅ (Just Updated)

Updated `simulator-ui-js/.env.local`:

**Before**:
```bash
NEXT_PUBLIC_AUDIT_CORRELATOR_URL=http://localhost:50052
```

**After**:
```bash
NEXT_PUBLIC_AUDIT_CORRELATOR_URL=http://localhost:8082
```

### 3. Required Action: Restart UI Development Server

The Next.js dev server must be restarted to load the new environment variable:

```bash
# In the simulator-ui-js terminal:
# 1. Stop the dev server (Ctrl+C)

# 2. Restart it
npm run dev
```

### 4. Verification

After restarting the UI:

1. **Open Browser**: http://localhost:3002/topology
2. **Check Console**: Should see NO errors
3. **Verify Logs**: Should see successful Connect requests

**Expected Console Output**:
```
[gRPC Request] audit.v1.TopologyService.GetTopologyStructure
[gRPC Response] audit.v1.TopologyService.GetTopologyStructure { status: 'success' }
```

## Port Mapping Reference

| Port | Service | Protocol | Purpose |
|------|---------|----------|---------|
| 8082 | HTTP Server | Connect/REST | ✅ Browser gRPC via Connect |
| 50052 | gRPC Server | Native gRPC | Native clients only |

**Key Point**: Browsers MUST use port 8082 (HTTP with Connect), not port 50052 (native gRPC).

## Testing the Connect Endpoint

You can test the Connect endpoint directly with curl:

```bash
curl -X POST http://localhost:8082/audit.v1.TopologyService/GetTopologyStructure \
  -H "Content-Type: application/json" \
  -d '{"request_id": "test-1"}'

# Expected response:
# {"snapshotTime":"...", "snapshotId":"snapshot-0", "requestId":"test-1"}
```

## Technical Details

### Why Two Ports?

**Port 8082 (HTTP Server)**:
- REST API endpoints (`/api/v1/*`)
- Prometheus metrics (`/metrics`)
- **Connect protocol** (`/audit.v1.TopologyService/*`) ← Browsers use this
- Supports: HTTP/1.1, HTTP/2, JSON, Binary Protobuf
- CORS-enabled for browser requests

**Port 50052 (gRPC Server)**:
- Native gRPC protocol only
- Requires HTTP/2 with trailers
- Browsers cannot use this directly
- For: Native gRPC clients (Go, Python, etc.)

### Protocol Compatibility

The Connect protocol implementation supports **three protocols**:
1. **Connect** - Preferred for browsers (used by simulator-ui-js)
2. **gRPC** - Native gRPC protocol (server-to-server)
3. **gRPC-Web** - Alternative browser protocol

All three protocols work on port 8082, making it the universal endpoint.

## Troubleshooting

### Still Getting Errors?

1. **Check UI is using port 8082**:
   ```bash
   cd simulator-ui-js
   cat .env.local | grep AUDIT_CORRELATOR
   # Should show: http://localhost:8082
   ```

2. **Verify dev server restarted**:
   - Environment variables only load when dev server starts
   - Must stop (Ctrl+C) and restart (`npm run dev`)

3. **Check browser console**:
   - Should see requests to `http://localhost:8082/audit.v1.TopologyService/*`
   - NOT `http://localhost:50052/*`

4. **Test Connect endpoint**:
   ```bash
   curl -X POST http://localhost:8082/audit.v1.TopologyService/GetTopologyStructure \
     -H "Content-Type: application/json" \
     -d '{"request_id": "test"}'
   ```
   Should return JSON response with `snapshotTime`, `snapshotId`, `requestId`.

## Summary

✅ **Backend**: Connect protocol support added and deployed
✅ **UI Config**: Updated to use correct port (8082)
⏳ **Next Step**: Restart Next.js dev server to apply config change

Once the dev server restarts, the UI should successfully connect to the TopologyService via Connect protocol!
