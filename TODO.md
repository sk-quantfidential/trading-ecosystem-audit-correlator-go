# audit-correlator-go TODO

> **Note**: Completed milestones have been moved to [TODO-HISTORY.md](./TODO-HISTORY.md). This file tracks only active and future work.

---

## epic-TSE-0001: Foundation Services & Infrastructure

### 🔍 Milestone TSE-0001.10: Audit Infrastructure (PRIMARY)
**Status**: ⚡ **IN PROGRESS** - Foundation established, integration needed
**Priority**: CRITICAL - Enables system validation and correlation

**Tasks**:
- [x] Fix JSON serialization issues (map[string]interface{} → json.RawMessage)
- [x] Create comprehensive test automation (Makefile with unit/integration targets)
- [x] Establish TDD Red-Green-Refactor pattern testing (4 test files, 15+ test scenarios)
- [x] Integrate audit-data-adapter-go component for all Redis/PostgreSQL operations
- [x] Remove direct Redis/PostgreSQL dependencies (delegate to audit-data-adapter-go)
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

**Dependencies**: TSE-0001.3b (Go Services gRPC Integration), TSE-0001.9 (Test Coordination Framework), audit-data-adapter-go integration ✅

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

**Last Updated**: 2025-10-29 (Completed milestones archived to TODO-HISTORY.md)
