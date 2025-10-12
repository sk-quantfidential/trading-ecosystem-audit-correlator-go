# Pull Request: TSE-0001.12.0b - Prometheus Metrics with Clean Architecture

**Epic:** TSE-0001 - Foundation Services & Infrastructure
**Milestone:** TSE-0001.12.0b - Prometheus Metrics (Clean Architecture)
**Branch:** `feature/TSE-0001.12.0-prometheus-metric-client`
**Status:** ✅ Ready for Review

## Summary

This PR implements Prometheus metrics collection using **Clean Architecture principles**, ensuring the domain layer never depends on infrastructure concerns. The implementation follows the port/adapter pattern, enabling future migration to OpenTelemetry without changing domain logic.

**Key Achievements:**
1. ✅ **Clean Architecture**: MetricsPort interface separates domain from infrastructure
2. ✅ **RED Pattern**: Rate, Errors, Duration metrics for all HTTP requests
3. ✅ **Low Cardinality**: Constant labels (service, instance, version) + request labels (method, route, code)
4. ✅ **Future-Proof**: Can swap Prometheus for OpenTelemetry by changing adapter
5. ✅ **Testable**: Mock MetricsPort for unit tests
6. ✅ **Comprehensive Tests**: 8 BDD test scenarios covering all functionality

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Presentation Layer                     │
│  ┌────────────────┐  ┌─────────────────────────────────┐  │
│  │  HTTP Handler  │  │   RED Metrics Middleware        │  │
│  │  /metrics      │  │   (instruments all requests)    │  │
│  └────────┬───────┘  └──────────────┬──────────────────┘  │
│           │                          │                      │
└───────────┼──────────────────────────┼──────────────────────┘
            │                          │
            │  depends on interface    │
            ▼                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer (Port)                    │
│  ┌───────────────────────────────────────────────────────┐ │
│  │           MetricsPort (interface)                     │ │
│  │  - IncCounter(name, labels)                           │ │
│  │  - ObserveHistogram(name, value, labels)              │ │
│  │  - SetGauge(name, value, labels)                      │ │
│  │  - GetHTTPHandler() http.Handler                      │ │
│  └───────────────────────────────────────────────────────┘ │
└───────────────────────────┬─────────────────────────────────┘
                            │  implemented by adapter
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer (Adapter)            │
│  ┌───────────────────────────────────────────────────────┐ │
│  │       PrometheusMetricsAdapter                        │ │
│  │  implements MetricsPort                               │ │
│  │                                                        │ │
│  │  - Uses prometheus/client_golang                      │ │
│  │  - Thread-safe lazy initialization                    │ │
│  │  - Registers Go runtime metrics                       │ │
│  │  - Applies constant labels (service, instance, ver)   │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

Future: Swap for OtelMetricsAdapter without changing domain/presentation
```

## Changes

### 1. Domain Layer - MetricsPort Interface

**File:** `internal/domain/ports/metrics.go` (NEW)

**Purpose:** Define the contract for metrics collection, independent of implementation

**Interface Methods:**
```go
type MetricsPort interface {
    // RED Pattern methods
    IncCounter(name string, labels map[string]string)
    ObserveHistogram(name string, value float64, labels map[string]string)
    SetGauge(name string, value float64, labels map[string]string)

    // HTTP serving
    GetHTTPHandler() http.Handler
}
```

**MetricsLabels Helper:**
- `ToMap()`: Converts labels struct to map
- `ConstantLabels()`: Returns only service, instance, version
- Ensures low cardinality by design

**Clean Architecture Benefits:**
- Domain never imports Prometheus packages
- Interface can be mocked for testing
- Future implementations (OpenTelemetry) implement same interface

### 2. Infrastructure Layer - PrometheusMetricsAdapter

**File:** `internal/infrastructure/observability/prometheus_adapter.go` (NEW)

**Purpose:** Implement MetricsPort using Prometheus client library

**Features:**
- **Thread-safe lazy initialization**: Metrics created on first use
- **Constant labels**: Applied to all metrics (service, instance, version)
- **Separate registry**: Isolated from default Prometheus registry
- **Go runtime metrics**: Automatic collection (goroutines, memory, GC, etc.)
- **Sensible histogram buckets**: 5ms to 10s for request duration

**Implementation Details:**
```go
type PrometheusMetricsAdapter struct {
    registry       *prometheus.Registry
    counters       map[string]*prometheus.CounterVec
    histograms     map[string]*prometheus.HistogramVec
    gauges         map[string]*prometheus.GaugeVec
    mu             sync.RWMutex
    constantLabels map[string]string
}
```

**Lazy Initialization Pattern:**
1. Fast path: Read lock check
2. Slow path: Write lock + double-check + create
3. Thread-safe for concurrent requests

**Histogram Buckets:**
```
5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
```
Chosen for typical HTTP API response times.

### 3. RED Metrics Middleware

**File:** `internal/infrastructure/observability/middleware.go` (NEW)

**Purpose:** Instrument all HTTP requests with RED pattern metrics

**RED Pattern Metrics:**
1. **Rate**: `http_requests_total` (counter)
   - Labels: method, route, code
   - Incremented for every request

2. **Errors**: `http_request_errors_total` (counter)
   - Labels: method, route, code
   - Incremented only for 4xx/5xx responses

3. **Duration**: `http_request_duration_seconds` (histogram)
   - Labels: method, route, code
   - Observes request latency in seconds

**Low Cardinality Enforcement:**
- **Route**: Uses `c.FullPath()` (pattern `/api/v1/events/:id`) NOT full path (`/api/v1/events/123`)
- **Unknown routes**: Labeled as `"unknown"` to avoid metric explosion
- **Method**: HTTP method (GET, POST, etc.) - naturally low cardinality
- **Code**: HTTP status code (200, 404, 500) - naturally low cardinality

**Middleware Usage:**
```go
router.Use(observability.REDMetricsMiddleware(metricsPort))
```

### 4. Updated Metrics Handler

**File:** `internal/handlers/metrics.go` (MODIFIED)

**Before (Coupled to Prometheus):**
```go
type MetricsHandler struct {
    handler gin.HandlerFunc  // Direct Prometheus dependency
}

func NewMetricsHandler() *MetricsHandler {
    promHandler := promhttp.Handler()  // Coupled!
    // ...
}
```

**After (Clean Architecture):**
```go
type MetricsHandler struct {
    metricsPort ports.MetricsPort  // Interface dependency
}

func NewMetricsHandler(metricsPort ports.MetricsPort) *MetricsHandler {
    return &MetricsHandler{
        metricsPort: metricsPort,
    }
}

func (h *MetricsHandler) Metrics(c *gin.Context) {
    handler := h.metricsPort.GetHTTPHandler()
    handler.ServeHTTP(c.Writer, c.Request)
}
```

**Benefits:**
- ✅ Depends on interface, not concrete implementation
- ✅ Can be tested with mock MetricsPort
- ✅ Future OpenTelemetry: just pass OtelMetricsAdapter

### 5. Main Server Integration

**File:** `cmd/server/main.go` (MODIFIED)

**Setup Observability:**
```go
// Initialize observability (Clean Architecture: port + adapter)
constantLabels := map[string]string{
    "service":  cfg.ServiceName,         // "audit-correlator"
    "instance": cfg.ServiceInstanceName, // "audit-correlator"
    "version":  cfg.Version,             // "1.0.0"
}
metricsPort := observability.NewPrometheusMetricsAdapter(constantLabels)

// Add RED metrics middleware (Rate, Errors, Duration)
router.Use(observability.REDMetricsMiddleware(metricsPort))

// Initialize handlers
metricsHandler := handlers.NewMetricsHandler(metricsPort)
router.GET("/metrics", metricsHandler.Metrics)
```

**Dependency Injection:**
- MetricsPort interface passed to middleware and handler
- Concrete PrometheusMetricsAdapter created once at startup
- All components depend on interface, not implementation

### 6. Comprehensive Tests

**File:** `internal/handlers/metrics_test.go` (UPDATED)

**Test Scenarios:**
1. ✅ `exposes_prometheus_metrics_through_port`: Verifies /metrics returns Prometheus format
2. ✅ `returns_text_plain_content_type`: Verifies Content-Type header
3. ✅ `includes_standard_go_runtime_metrics`: Verifies Go runtime metrics present
4. ✅ `metrics_are_parseable_by_prometheus`: Verifies Prometheus text format (HELP, TYPE, metric lines)

**File:** `internal/infrastructure/observability/middleware_test.go` (NEW)

**Test Scenarios:**
1. ✅ `instruments_successful_requests_with_RED_metrics`: Verifies all RED metrics recorded
2. ✅ `instruments_error_requests_with_error_counter`: Verifies error counter for 4xx/5xx
3. ✅ `uses_route_pattern_not_full_path_for_low_cardinality`: Verifies `/api/v1/events/:id` not `/api/v1/events/123`
4. ✅ `handles_unknown_routes_gracefully`: Verifies unknown routes labeled as `"unknown"`

**All tests follow BDD Given/When/Then pattern:**
```go
// Given: A Prometheus metrics adapter
constantLabels := map[string]string{...}
metricsPort := observability.NewPrometheusMetricsAdapter(constantLabels)

// When: A request is made
req := httptest.NewRequest(...)
router.ServeHTTP(w, req)

// Then: Metrics should be recorded
if !strings.Contains(metricsOutput, "http_requests_total") {
    t.Error("Expected metric to be present")
}
```

## Metrics Exposed

### Standard Go Runtime Metrics

Automatically collected by Prometheus client:
- `go_goroutines`: Number of goroutines
- `go_threads`: Number of OS threads
- `go_gc_duration_seconds`: GC pause duration
- `go_memstats_alloc_bytes`: Heap memory allocated
- `process_cpu_seconds_total`: CPU time consumed
- `process_resident_memory_bytes`: Resident memory size

### RED Pattern Metrics

**1. http_requests_total** (counter)
```promql
http_requests_total{
  service="audit-correlator",
  instance="audit-correlator",
  version="1.0.0",
  method="GET",
  route="/api/v1/health",
  code="200"
}
```

**2. http_request_duration_seconds** (histogram)
```promql
http_request_duration_seconds_bucket{
  service="audit-correlator",
  instance="audit-correlator",
  version="1.0.0",
  method="GET",
  route="/api/v1/health",
  code="200",
  le="0.1"
} 42
```

**3. http_request_errors_total** (counter)
```promql
http_request_errors_total{
  service="audit-correlator",
  instance="audit-correlator",
  version="1.0.0",
  method="GET",
  route="/api/v1/nonexistent",
  code="404"
}
```

## Example Prometheus Queries

### Request Rate (Requests per second)
```promql
rate(http_requests_total{service="audit-correlator"}[5m])
```

### Request Rate by Route
```promql
sum by (route) (rate(http_requests_total{service="audit-correlator"}[5m]))
```

### Request Duration (95th percentile)
```promql
histogram_quantile(0.95,
  sum by (le) (rate(http_request_duration_seconds_bucket{service="audit-correlator"}[5m]))
)
```

### Error Rate
```promql
rate(http_request_errors_total{service="audit-correlator"}[5m])
```

### Error Percentage
```promql
(
  rate(http_request_errors_total{service="audit-correlator"}[5m])
  /
  rate(http_requests_total{service="audit-correlator"}[5m])
) * 100
```

## Testing Instructions

### 1. Run Unit Tests

```bash
cd /home/skingham/Projects/Quantfidential/trading-ecosystem/audit-correlator-go

# Run all metrics-related tests
go test -v -tags=unit ./internal/handlers/metrics_test.go ./internal/handlers/metrics.go
go test -v -tags=unit ./internal/infrastructure/observability/...

# Run with coverage
go test -cover -tags=unit ./internal/handlers/metrics_test.go ./internal/handlers/metrics.go
go test -cover -tags=unit ./internal/infrastructure/observability/...
```

**Expected:** All tests pass ✅

### 2. Build and Run Service

```bash
# Rebuild service
cd /home/skingham/Projects/Quantfidential/trading-ecosystem/orchestrator-docker
docker-compose build audit-correlator

# Start service
docker-compose up -d audit-correlator

# Wait for startup
sleep 10
```

### 3. Verify Metrics Endpoint

```bash
# Check metrics endpoint
curl http://localhost:8083/metrics

# Should see:
# - # HELP go_goroutines ...
# - # TYPE go_goroutines gauge
# - go_goroutines 13
# - (many more Go runtime metrics)
```

### 4. Generate Traffic and Verify RED Metrics

```bash
# Make some requests
for i in {1..10}; do
  curl http://localhost:8083/api/v1/health
done

# Make an error request
curl http://localhost:8083/nonexistent

# Check RED metrics
curl http://localhost:8083/metrics | grep -E "http_requests_total|http_request_duration|http_request_errors"
```

**Expected Output:**
```
http_requests_total{code="200",method="GET",route="/api/v1/health",...} 10
http_requests_total{code="404",method="GET",route="unknown",...} 1
http_request_duration_seconds_bucket{code="200",...,le="0.005"} 8
http_request_duration_seconds_bucket{code="200",...,le="0.01"} 10
http_request_errors_total{code="404",method="GET",route="unknown",...} 1
```

### 5. Verify Prometheus Scraping

```bash
# Check Prometheus targets
open http://localhost:9090/targets

# Should see:
# audit-correlator (172.20.0.80:8083) - UP
```

**Query in Prometheus:**
```promql
up{job="audit-correlator"}
# Should return: 1
```

### 6. Verify Constant Labels

```bash
curl http://localhost:8083/metrics | grep -E "service=|instance=|version="
```

**Expected:**
```
http_requests_total{...,service="audit-correlator",instance="audit-correlator",version="1.0.0",...}
```

## Migration Path to OpenTelemetry (Phase 2)

### Current Implementation (Phase 1)
```go
// Prometheus adapter
metricsPort := observability.NewPrometheusMetricsAdapter(constantLabels)
```

### Future Implementation (Phase 2 - No Domain Changes!)
```go
// OpenTelemetry adapter (same interface!)
metricsPort := observability.NewOtelMetricsAdapter(constantLabels)
```

**Steps for OpenTelemetry Migration:**
1. Create `OtelMetricsAdapter` implementing `MetricsPort`
2. Use OpenTelemetry SDK meters instead of Prometheus client
3. Add OpenTelemetry Prometheus bridge for `/metrics` endpoint
4. Swap adapter in `main.go`
5. **Zero changes to handlers, middleware, or domain logic** ✅

**Metric Names Remain the Same:**
- `http_requests_total`
- `http_request_duration_seconds`
- `http_request_errors_total`

**Dashboards Remain the Same:** No Grafana dashboard changes needed!

## Repeatable Pattern for Other Services

This implementation provides a **repeatable pattern** for all other services:

### 1. Copy Port Interface
```bash
cp internal/domain/ports/metrics.go ../other-service-go/internal/domain/ports/
```

### 2. Copy Prometheus Adapter
```bash
cp -r internal/infrastructure/observability ../other-service-go/internal/infrastructure/
```

### 3. Update main.go
```go
// Add to main.go setupHTTPServer
constantLabels := map[string]string{
    "service":  cfg.ServiceName,
    "instance": cfg.ServiceInstanceName,
    "version":  cfg.Version,
}
metricsPort := observability.NewPrometheusMetricsAdapter(constantLabels)
router.Use(observability.REDMetricsMiddleware(metricsPort))

metricsHandler := handlers.NewMetricsHandler(metricsPort)
router.GET("/metrics", metricsHandler.Metrics)
```

### 4. Update Prometheus Config
```yaml
# Add scrape target in prometheus.yml
- job_name: 'other-service'
  static_configs:
    - targets: ['other-service:8084']
      labels:
        service: 'other-service'
        instance_name: 'other-service'
        service_type: 'singleton'
```

### 5. Copy Tests
```bash
cp internal/handlers/metrics_test.go ../other-service-go/internal/handlers/
cp internal/infrastructure/observability/middleware_test.go ../other-service-go/internal/infrastructure/observability/
```

**Result:** Consistent metrics across all services with minimal effort!

## Dependencies

**New Dependencies Added:**
- `github.com/prometheus/client_golang v1.23.2`
- `github.com/prometheus/client_model v0.6.2`
- `github.com/prometheus/common v0.66.1`
- `github.com/prometheus/procfs v0.16.1`

**go.mod Updated:** Yes (go get ran successfully)

## Files Changed

**New Files:**
- `internal/domain/ports/metrics.go` (113 lines)
- `internal/infrastructure/observability/prometheus_adapter.go` (180 lines)
- `internal/infrastructure/observability/middleware.go` (61 lines)
- `internal/handlers/metrics_test.go` (220 lines)
- `internal/infrastructure/observability/middleware_test.go` (200 lines)
- `docs/prs/feature-TSE-0001.12.0-prometheus-metric-client.md` (THIS FILE)

**Modified Files:**
- `internal/handlers/metrics.go` (simplified to use MetricsPort)
- `cmd/server/main.go` (added observability setup)
- `TODO.md` (added milestone TSE-0001.12.0b)
- `go.mod` (added Prometheus client dependencies)
- `go.sum` (dependency checksums)

## Integration Testing (Added 2025-10-10)

**New File:** `tests/data_adapter_smoke_test.go` (114 lines)

**Purpose:** Minimal smoke tests to verify data adapter integration with infrastructure

**Test Coverage:**
1. ✅ **adapter_initialization**: Verifies adapter creation and connection
2. ✅ **cache_repository_smoke**: Tests basic cache Set/Get/Delete operations
3. ⏭️ **audit_event_repository_basic_crud**: Deferred (requires UUID generation)
4. ⏭️ **service_discovery_smoke**: Deferred (requires Redis ACL enhancement)

**Features:**
- Tagged with `//go:build integration` for selective execution
- Graceful skips when infrastructure unavailable
- Uses orchestrator credentials for integration testing
- Time-based unique keys to prevent test collisions

**Running Tests:**
```bash
# Run integration tests
make test-integration

# Run all tests
make test-all
```

This brings audit-correlator-go to parity with exchange-simulator-go for minimal integration testing.

## Merge Checklist

- [x] Clean Architecture port/adapter pattern implemented
- [x] MetricsPort interface defined in domain layer
- [x] PrometheusMetricsAdapter implements MetricsPort
- [x] RED metrics middleware created
- [x] /metrics endpoint updated to use port
- [x] Constant labels applied (service, instance, version)
- [x] Low-cardinality request labels (method, route, code)
- [x] All unit tests passing (8 test scenarios)
- [x] BDD Given/When/Then test pattern followed
- [x] Integration with main.go complete
- [x] Dependencies added to go.mod
- [x] TODO.md updated with milestone
- [x] PR documentation complete
- [x] Repeatable pattern documented for other services
- [x] OpenTelemetry migration path documented

## Approval

**Ready for Merge**: ✅ Yes

All requirements satisfied:
- ✅ Clean Architecture principles followed
- ✅ Domain layer independent of infrastructure
- ✅ Future-proof for OpenTelemetry migration
- ✅ RED pattern metrics implemented
- ✅ Low-cardinality labels enforced
- ✅ Comprehensive test coverage
- ✅ Repeatable pattern for other services
- ✅ Documentation complete

---

**Epic:** TSE-0001.12.0b
**Branch:** feature/TSE-0001.12.0-prometheus-metric-client
**Test Results:** 8/8 tests passing
**Build Status:** ✅ Successful
**Runtime Status:** ✅ Verified with Prometheus scraping

🎯 **Achievement:** Prometheus metrics with Clean Architecture - ready to replicate across all services!

🤖 Generated with [Claude Code](https://claude.com/claude-code)
