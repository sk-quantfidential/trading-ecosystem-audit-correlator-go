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

### Priority Tasks for audit-data-adapter-go Integration:

1. **Remove Direct Database Dependencies**:
   - Remove direct Redis client imports (`github.com/redis/go-redis/v9`)
   - Remove direct PostgreSQL imports (`github.com/lib/pq`)
   - Update go.mod to only depend on audit-data-adapter-go for data operations

2. **Refactor Infrastructure Layer**:
   - Replace `internal/infrastructure` Redis clients with audit-data-adapter-go DataAdapter
   - Update service discovery to use DataAdapter.ServiceDiscoveryRepository interface
   - Update configuration caching to use DataAdapter.CacheRepository interface

3. **Update Service Layer**:
   - Modify `internal/services/audit.go` to use DataAdapter.AuditEventRepository
   - Update audit event creation and correlation logic to use repository patterns
   - Ensure all audit events use proper audit-data-adapter-go models

4. **Test Integration**:
   - Update test dependencies to use audit-data-adapter-go test utilities
   - Configure tests to use shared orchestrator database/Redis instances
   - Validate that all tests pass with delegated data operations

5. **Configuration Integration**:
   - Update configuration to initialize audit-data-adapter-go DataAdapter
   - Ensure environment variables align with audit-data-adapter-go patterns
   - Implement proper connection lifecycle management

---

**Last Updated**: 2025-09-29 (Test status updated, audit-data-adapter-go integration tasks added)