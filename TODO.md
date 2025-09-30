# audit-correlator-go TODO

## epic-TSE-0001: Foundation Services & Infrastructure

### 🏗️ Milestone TSE-0001.1a: Go Services Bootstrapping
**Status**: ✅ COMPLETED
**Priority**: High

**Tasks**:
- [x] Create Go service directory structure following clean architecture
- [x] Implement health check endpoint (REST and gRPC)
- [x] Basic structured logging with levels
- [x] Error handling infrastructure
- [x] Dockerfile for service containerization
- [x] Load component-specific .claude configuration

**BDD Acceptance**: All Go services can start, respond to health checks, and shutdown gracefully

---

### 🔗 Milestone TSE-0001.3b: Go Services gRPC Integration
**Status**: ✅ **COMPLETED** - Ready for BDD acceptance verification
**Priority**: High

**Tasks**:
- [x] Implement gRPC server with health service
- [x] Service registration with Redis-based discovery
- [x] Configuration service client integration
- [x] Inter-service communication testing

**BDD Acceptance**: Go services can discover and communicate with each other via gRPC

**Dependencies**: TSE-0001.1a (Go Services Bootstrapping), TSE-0001.3a (Core Infrastructure)

---

### 🔍 Milestone TSE-0001.10: Audit Infrastructure (PRIMARY)
**Status**: ⚡ **IN PROGRESS** - Foundation established, integration needed
**Priority**: CRITICAL - Enables system validation and correlation

**Tasks**:
- [x] Fix JSON serialization issues (map[string]interface{} → json.RawMessage)
- [x] Create comprehensive test automation (Makefile with unit/integration targets)
- [x] Establish TDD Red-Green-Refactor pattern testing (4 test files, 15+ test scenarios)
- [ ] **CRITICAL**: Integrate audit-data-adapter-go component for all Redis/PostgreSQL operations
- [ ] Remove direct Redis/PostgreSQL dependencies (delegate to audit-data-adapter-go)
- [ ] OpenTelemetry trace collection from all services
- [ ] Basic event correlation (timeline reconstruction)
- [ ] Prometheus metrics aggregation
- [ ] Simple causation analysis (scenario event → system response)
- [ ] Timeline analysis engine
- [ ] Correlation reporting
- [ ] Validation assertion framework

**Current Test Status**:
- **Unit Tests**: 10 test scenarios (7 passing, 3 skipped - infrastructure dependencies)
- **Integration Tests**: 5 test scenarios (0 passing, 5 skipped - infrastructure dependencies)
- **Test Coverage**: Configuration, gRPC server, service discovery patterns established
- **Build Status**: ✅ Compiles successfully after JSON serialization fix

**BDD Acceptance**: Can correlate a chaos event with subsequent service behavior

**Dependencies**: TSE-0001.3b (Go Services gRPC Integration), TSE-0001.9 (Test Coordination Framework), **audit-data-adapter-go integration**

---

### 📈 Milestone TSE-0001.12c: Audit Integration
**Status**: Not Started
**Priority**: Medium

**Tasks**:
- [ ] All services emit telemetry to audit correlator
- [ ] Timeline reconstruction across all services
- [ ] Event correlation validation
- [ ] Audit trail generation and reporting

**BDD Acceptance**: Audit correlator successfully tracks and correlates events across all system components

**Dependencies**: TSE-0001.10 (Audit Infrastructure), TSE-0001.12b (Trading Flow Integration)

---

## Implementation Notes

- **Data Ingestion**: Collect from ALL system signals including risk monitor outputs
- **Event Correlation**: Link scenario injection → market changes → risk alerts
- **Timeline Analysis**: Complete causal chain reconstruction with timing
- **Independent Validation**: Objective validation of risk monitor effectiveness
- **Coverage Tracking**: Ensure all risk scenarios are tested and validated
- **Performance**: Real-time correlation analysis for large event volumes

---

---

## 🔄 Next Steps: Audit Data Adapter Integration

### TSE-0001.4 Integration Tasks (Phase-by-Phase Implementation)

#### ✅ **Task 0: Test Infrastructure Foundation** - COMPLETED
- [x] Fixed JSON serialization issues (map[string]interface{} → json.RawMessage)
- [x] Created comprehensive Makefile with unit/integration test targets
- [x] Established TDD Red-Green-Refactor pattern testing (4 test files, 15+ test scenarios)
- [x] Build status: ✅ Compiles successfully after JSON serialization fix
- [x] Test status: 10 unit tests (7 passing, 3 skipped), 5 integration tests (0 passing, 5 skipped - infrastructure dependencies)

#### ✅ **Task 1: Remove Direct Database Dependencies** - COMPLETED
**Goal**: Eliminate direct database imports and prepare for DataAdapter integration
**Files Modified**:
- `go.mod` - Redis/PostgreSQL dependencies kept as indirect via audit-data-adapter-go
- `internal/infrastructure/service_discovery.go` - No direct Redis imports (uses DataAdapter)
- `internal/services/audit.go` - Uses audit-data-adapter-go models

**Acceptance Criteria**:
- [x] No direct Redis client imports in any Go files
- [x] No direct PostgreSQL imports in any Go files (indirect via adapter)
- [x] Code compiles successfully
- [x] All database access points using DataAdapter integration

#### ✅ **Task 2: Refactor Infrastructure Layer** - COMPLETED
**Goal**: Replace direct database access with audit-data-adapter-go DataAdapter interfaces
**Files Modified**:
- `internal/infrastructure/service_discovery.go` → Uses `DataAdapter.ServiceDiscoveryRepository` ✅
- `internal/infrastructure/configuration_client.go` → Uses `DataAdapter.CacheRepository` ✅
- `internal/config/config.go` → DataAdapter initialization with `InitializeDataAdapter()` ✅
- `cmd/server/main.go` → Proper lifecycle management (Connect/Disconnect) ✅

**Implementation Completed**:
1. ✅ DataAdapter initialization in config with environment fallback
2. ✅ Service discovery using ServiceDiscoveryRepository interface
3. ✅ Configuration caching using CacheRepository interface (Set/Get/DeleteByPattern/GetKeysByPattern)
4. ✅ Connection lifecycle management in main.go with proper cleanup

**Acceptance Criteria**:
- [x] Service discovery uses only DataAdapter.ServiceDiscoveryRepository interface
- [x] Configuration caching uses only DataAdapter.CacheRepository interface
- [x] DataAdapter properly initialized with orchestrator credentials (from environment)
- [x] Connection lifecycle (Connect/Disconnect) working through adapter
- [x] Build compiles successfully
- [x] Tests: 7 unit tests passing, 3 passing integration tests

#### 🔧 **Task 3: Update Service Layer** - AFTER TASK 2
**Goal**: Integrate audit event operations with repository patterns
**Files to Modify**:
- `internal/services/audit.go` → Use `DataAdapter.AuditEventRepository`
- `internal/handlers/*` → Update to use repository patterns
- `internal/presentation/grpc/*` → Ensure proper model usage

**Implementation Steps**:
1. Update audit event creation to use AuditEventRepository.Create()
2. Update audit event queries to use AuditEventRepository.Query()
3. Update correlation logic to use AuditEventRepository.GetCorrelatedEvents()
4. Ensure all models align with audit-data-adapter-go patterns

**Acceptance Criteria**:
- [ ] All audit events created through AuditEventRepository interface
- [ ] Event correlation working through repository queries
- [ ] All models consistent with audit-data-adapter-go standards
- [ ] No direct database access in service layer

#### 🧪 **Task 4: Test Integration** - AFTER TASK 3
**Goal**: Enable tests to use shared orchestrator services and validate integration
**Files to Modify**:
- `internal/*_test.go` → Update to use audit-data-adapter-go test utilities
- Test configuration → Use shared PostgreSQL/Redis instances
- Integration tests → Validate cross-component functionality

**Implementation Steps**:
1. Configure tests to use audit-data-adapter-go test environment setup
2. Update test mocking to use repository interfaces
3. Enable integration tests with shared orchestrator PostgreSQL/Redis
4. Add cross-component audit validation tests

**Acceptance Criteria**:
- [ ] Unit tests: 10/10 passing (currently 7/10 due to infrastructure gaps)
- [ ] Integration tests: 5/5 passing (currently 0/5 due to infrastructure gaps)
- [ ] Tests use shared orchestrator database/Redis instances
- [ ] Cross-component audit functionality validated

#### ⚙️ **Task 5: Configuration Integration** - AFTER TASK 4
**Goal**: Complete environment alignment and lifecycle management
**Files to Create/Modify**:
- `.env.example` → Following audit-data-adapter-go pattern
- Docker configuration → Use shared environment variables
- Documentation → Integration patterns and usage

**Implementation Steps**:
1. Create .env.example with orchestrator-compatible configuration
2. Update Docker setup to use shared environment
3. Implement proper DataAdapter lifecycle management
4. Document integration patterns for replication

**Acceptance Criteria**:
- [ ] Environment configuration aligned with audit-data-adapter-go patterns
- [ ] Docker integration with orchestrator services working
- [ ] Proper connection lifecycle management implemented
- [ ] Integration pattern documented for replication to other Go services

### 🎯 **Success Metrics After Integration**:
- ✅ **Shared Infrastructure**: Service uses audit-data-adapter-go for all database operations
- ✅ **Test Success**: 100% test pass rate (15/15 scenarios passing)
- ✅ **Repository Pattern**: All database access through clean interfaces
- ✅ **Orchestrator Integration**: Seamless operation with shared services
- ✅ **Replication Ready**: Pattern established for custodian-simulator-go, exchange-simulator-go, market-data-simulator-go

### 📋 **Testing Commands During Integration**:
```bash
# After each task, validate progress
make test-unit              # Should improve pass rate as dependencies are resolved
make test-integration       # Should enable orchestrator connectivity
make status                 # Check overall integration health
make build                  # Ensure compilation throughout process
```

---

**Last Updated**: 2025-09-29 (Test status updated, audit-data-adapter-go integration tasks added)