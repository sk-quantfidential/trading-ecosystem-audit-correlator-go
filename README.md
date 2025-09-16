# Audit Correlator

A high-performance event correlation engine built in Go that provides comprehensive system-wide observability by ingesting all telemetry streams, reconstructing causal event chains, and validating system behavior against chaos engineering scenarios.

## 🎯 Overview

The Audit Correlator serves as the "omniscient observer" of the trading ecosystem, with complete visibility into all service interactions, scenario events, and risk system responses. Unlike the production-constrained Risk Monitor, the Audit Correlator has access to the full system state, enabling it to validate that chaos scenarios execute correctly and that the risk system responds appropriately within defined SLAs.

### Key Features
- **Complete System Visibility**: Ingests telemetry from all services including risk monitor compliance signals
- **Real-Time Event Correlation**: Sub-second correlation of related events across multiple services
- **Causal Chain Reconstruction**: Builds complete cause-and-effect timelines from scenario injection to system response
- **Scenario Validation**: Automated verification that chaos scenarios execute as intended
- **Risk System Validation**: Proves risk monitor detects scenarios within SLA timeframes
- **Timeline Analytics**: Statistical analysis of event propagation delays and system behavior
- **Audit Trail Generation**: Regulatory-grade audit trails with complete event lineage
- **Performance Impact Analysis**: Quantifies system performance during chaos scenarios

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 Audit Correlator                        │
├─────────────────────────────────────────────────────────┤
│  Data Ingestion Layer                                   │
│  ├─OpenTelemetry Collector (Distributed traces)        │
│  ├─Prometheus Scraper (Metrics from all services)      │
│  ├─Log Stream Processor (Structured logs ingestion)    │
│  └─Risk Monitor Signal Receiver (Compliance signals)   │
├─────────────────────────────────────────────────────────┤
│  Event Processing Engine                                │
│  ├─Event Normalizer (Unified event format)             │
│  ├─Timestamp Synchronizer (Cross-service time alignment)│
│  ├─Correlation ID Tracker (Request flow reconstruction) │
│  └─Event Enricher (Context and metadata augmentation)  │
├─────────────────────────────────────────────────────────┤
│  Correlation Engine                                     │
│  ├─Causal Chain Builder (Event relationship detection) │
│  ├─Scenario Validator (Expected behavior verification)  │
│  ├─Timeline Reconstructor (Temporal sequence analysis)  │
│  ├─Confidence Calculator (Correlation strength scoring) │
│  └─Anomaly Detector (Unexpected behavior identification)│
├─────────────────────────────────────────────────────────┤
│  Analytics & Validation                                 │
│  ├─SLA Validator (Response time verification)           │
│  ├─Coverage Analyzer (Scenario completeness tracking)   │
│  ├─Performance Analyzer (System impact quantification) │
│  └─Audit Trail Generator (Regulatory compliance logs)  │
├─────────────────────────────────────────────────────────┤
│  Storage & Retrieval                                    │
│  ├─Event Store (High-performance event storage)        │
│  ├─Correlation Cache (Fast correlation lookup)         │
│  ├─Timeline Database (Temporal event sequences)        │
│  └─Audit Archive (Long-term compliance storage)        │
└─────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker and Docker Compose
- Access to all ecosystem service metrics and logs
- OpenTelemetry collector configuration

### Development Setup
```bash
# Clone the repository
git clone <repo-url>
cd audit-correlator

# Install dependencies
go mod download

# Build the application
make build

# Run tests
make test

# Start development server with mock data
make run-dev
```

### Docker Deployment
```bash
# Build container
docker build -t audit-correlator .

# Run with docker-compose (recommended)
docker-compose up audit-correlator

# Verify health and data ingestion
curl http://localhost:8085/health
curl http://localhost:8085/api/v1/ingestion/status
```

## 📡 API Reference

### gRPC Services

#### Correlation Service
```protobuf
service CorrelationService {
  rpc GetCorrelations(CorrelationRequest) returns (CorrelationResponse);
  rpc GetCausalChain(CausalChainRequest) returns (CausalChainResponse);
  rpc GetScenarioValidation(ScenarioValidationRequest) returns (ValidationResponse);
  rpc GetTimelineReconstruction(TimelineRequest) returns (TimelineResponse);
}
```

#### Analytics Service
```protobuf
service AnalyticsService {
  rpc GetSystemPerformance(PerformanceRequest) returns (PerformanceResponse);
  rpc GetCoverageAnalysis(CoverageRequest) returns (CoverageResponse);
  rpc GetAnomalyDetection(AnomalyRequest) returns (AnomalyResponse);
  rpc GenerateAuditReport(AuditReportRequest) returns (AuditReportResponse);
}
```

### REST Endpoints

#### Data Ingestion APIs
```
POST   /api/v1/events/ingest
GET    /api/v1/ingestion/status
GET    /api/v1/ingestion/metrics
POST   /api/v1/ingestion/configure
```

#### Correlation APIs
```
GET    /api/v1/correlations/scenario/{scenario_id}
GET    /api/v1/correlations/timeline/{start}/{end}
GET    /api/v1/correlations/causal-chain/{event_id}
GET    /api/v1/correlations/confidence/{correlation_id}
```

#### Validation APIs  
```
GET    /api/v1/validation/scenario/{scenario_id}
GET    /api/v1/validation/risk-sla/{timeframe}
GET    /api/v1/validation/coverage/summary
POST   /api/v1/validation/assert
```

#### Analytics APIs
```
GET    /api/v1/analytics/performance-impact/{scenario_id}
GET    /api/v1/analytics/system-behavior/{service}/{timeframe}
GET    /api/v1/analytics/anomalies/{timeframe}
GET    /api/v1/analytics/audit-trail/{start}/{end}
```

#### Real-Time Monitoring
```
WebSocket /ws/correlations (Live correlation updates)
WebSocket /ws/scenario-validation (Real-time scenario validation)
GET    /debug/correlation-engine/state
GET    /debug/event-streams/status
```

## 📊 Event Ingestion & Processing

### Multi-Source Data Ingestion
```go
type EventIngestionManager struct {
    sources map[string]DataSource
    processor *EventProcessor
    correlator *EventCorrelator
}

type DataSource interface {
    StartIngestion() error
    StopIngestion() error
    GetEventStream() <-chan RawEvent
    GetSourceInfo() SourceInfo
}

// OpenTelemetry Trace Ingestion
type OTelTraceSource struct {
    collector *otlp.Collector
    traceProcessor *TraceProcessor
}

func (ots *OTelTraceSource) StartIngestion() error {
    return ots.collector.Start(func(span *trace.Span) {
        event := ots.convertSpanToEvent(span)
        ots.processor.ProcessEvent(event)
    })
}

// Prometheus Metrics Ingestion  
type PrometheusSource struct {
    client api.Client
    queries []MetricQuery
    scrapeInterval time.Duration
}

func (ps *PrometheusSource) StartIngestion() error {
    ticker := time.NewTicker(ps.scrapeInterval)
    go func() {
        for range ticker.C {
            for _, query := range ps.queries {
                metrics, err := ps.client.Query(context.Background(), query.PromQL, time.Now())
                if err != nil {
                    continue
                }
                events := ps.convertMetricsToEvents(metrics, query)
                for _, event := range events {
                    ps.processor.ProcessEvent(event)
                }
            }
        }
    }()
    return nil
}

// Structured Log Ingestion
type LogStreamSource struct {
    logTail *tail.Tail
    parser *StructuredLogParser
}

func (lss *LogStreamSource) StartIngestion() error {
    go func() {
        for line := range lss.logTail.Lines {
            if event, err := lss.parser.ParseLogLine(line.Text); err == nil {
                lss.processor.ProcessEvent(event)
            }
        }
    }()
    return nil
}
```

### Event Normalization & Enrichment
```go
type EventProcessor struct {
    normalizer *EventNormalizer
    enricher   *EventEnricher
    validator  *EventValidator
}

type UnifiedEvent struct {
    ID            string                 `json:"id"`
    Timestamp     time.Time             `json:"timestamp"`
    Source        string                `json:"source"`  // service name
    EventType     string                `json:"event_type"`
    CorrelationID string                `json:"correlation_id,omitempty"`
    TraceID       string                `json:"trace_id,omitempty"`
    SpanID        string                `json:"span_id,omitempty"`
    Metadata      map[string]interface{} `json:"metadata"`
    Payload       map[string]interface{} `json:"payload"`
    
    // Audit-specific fields
    ScenarioID    string                `json:"scenario_id,omitempty"`
    ChaosType     string                `json:"chaos_type,omitempty"`
    ValidationTag string                `json:"validation_tag,omitempty"`
}

func (ep *EventProcessor) ProcessEvent(rawEvent RawEvent) error {
    // Normalize to unified format
    unifiedEvent, err := ep.normalizer.Normalize(rawEvent)
    if err != nil {
        return fmt.Errorf("normalization failed: %w", err)
    }
    
    // Enrich with context and metadata
    enrichedEvent, err := ep.enricher.Enrich(unifiedEvent)
    if err != nil {
        return fmt.Errorf("enrichment failed: %w", err)
    }
    
    // Validate event structure
    if err := ep.validator.Validate(enrichedEvent); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Send to correlation engine
    ep.correlationEngine.ProcessEvent(enrichedEvent)
    
    return nil
}
```

## 🔗 Correlation Engine

### Causal Chain Detection
```go
type CorrelationEngine struct {
    correlationRules []CorrelationRule
    eventStore       *EventStore
    timeWindow       time.Duration
    confidenceCalc   *ConfidenceCalculator
}

type CorrelationRule struct {
    Name           string        `yaml:"name"`
    TriggerPattern EventPattern  `yaml:"trigger"`
    EffectPattern  EventPattern  `yaml:"effect"`
    MaxDelay       time.Duration `yaml:"max_delay"`
    MinConfidence  float64       `yaml:"min_confidence"`
}

type EventPattern struct {
    Source    string            `yaml:"source"`
    EventType string            `yaml:"event_type"`
    Metadata  map[string]string `yaml:"metadata"`
}

// Example correlation rules
var DefaultCorrelationRules = []CorrelationRule{
    {
        Name: "chaos_injection_to_risk_alert",
        TriggerPattern: EventPattern{
            Source:    "test-coordinator", 
            EventType: "chaos_injection_started",
        },
        EffectPattern: EventPattern{
            Source:    "risk-monitor",
            EventType: "risk_alert_generated", 
        },
        MaxDelay:      2 * time.Minute,
        MinConfidence: 0.8,
    },
    {
        Name: "market_crash_to_trading_halt",
        TriggerPattern: EventPattern{
            Source:    "market-data-simulator",
            EventType: "price_shock_injected",
        },
        EffectPattern: EventPattern{
            Source:    "trading-engine", 
            EventType: "strategy_halted",
        },
        MaxDelay:      30 * time.Second,
        MinConfidence: 0.9,
    },
    {
        Name: "settlement_delay_to_liquidity_alert",
        TriggerPattern: EventPattern{
            Source:    "custodian-simulator",
            EventType: "settlement_delayed",
        },
        EffectPattern: EventPattern{
            Source:    "risk-monitor",
            EventType: "liquidity_risk_alert",
        },
        MaxDelay:      5 * time.Minute,
        MinConfidence: 0.7,
    },
}

func (ce *CorrelationEngine) ProcessEvent(event UnifiedEvent) {
    // Find potential correlations
    correlations := ce.findPotentialCorrelations(event)
    
    for _, correlation := range correlations {
        // Calculate confidence score
        confidence := ce.confidenceCalc.Calculate(correlation)
        
        if confidence >= correlation.Rule.MinConfidence {
            // Store validated correlation
            ce.storeCorrelation(correlation, confidence)
            
            // Update scenario validation if applicable
            if correlation.ScenarioID != "" {
                ce.updateScenarioValidation(correlation.ScenarioID, correlation)
            }
            
            // Emit correlation event
            ce.emitCorrelationEvent(correlation, confidence)
        }
    }
}

func (ce *CorrelationEngine) findPotentialCorrelations(event UnifiedEvent) []PotentialCorrelation {
    var correlations []PotentialCorrelation
    
    // Look for matching trigger patterns
    for _, rule := range ce.correlationRules {
        if ce.matchesPattern(event, rule.TriggerPattern) {
            // Search for corresponding effect events within time window
            effectEvents := ce.eventStore.FindEvents(EventQuery{
                Pattern:   rule.EffectPattern,
                StartTime: event.Timestamp,
                EndTime:   event.Timestamp.Add(rule.MaxDelay),
                CorrelationID: event.CorrelationID, // Use correlation ID if available
            })
            
            for _, effectEvent := range effectEvents {
                correlations = append(correlations, PotentialCorrelation{
                    TriggerEvent: event,
                    EffectEvent:  effectEvent,
                    Rule:         rule,
                    Delay:        effectEvent.Timestamp.Sub(event.Timestamp),
                })
            }
        }
        
        // Also check if this event is an effect of previous triggers
        if ce.matchesPattern(event, rule.EffectPattern) {
            triggerEvents := ce.eventStore.FindEvents(EventQuery{
                Pattern:   rule.TriggerPattern,
                StartTime: event.Timestamp.Add(-rule.MaxDelay),
                EndTime:   event.Timestamp,
                CorrelationID: event.CorrelationID,
            })
            
            for _, triggerEvent := range triggerEvents {
                correlations = append(correlations, PotentialCorrelation{
                    TriggerEvent: triggerEvent,
                    EffectEvent:  event,
                    Rule:         rule,
                    Delay:        event.Timestamp.Sub(triggerEvent.Timestamp),
                })
            }
        }
    }
    
    return correlations
}
```

### Timeline Reconstruction
```go
type TimelineReconstructor struct {
    eventStore     *EventStore
    correlationStore *CorrelationStore
    timelineCache  *TimelineCache
}

type Timeline struct {
    ID             string           `json:"id"`
    ScenarioID     string           `json:"scenario_id,omitempty"`
    StartTime      time.Time        `json:"start_time"`
    EndTime        time.Time        `json:"end_time"`
    Events         []TimelineEvent  `json:"events"`
    Correlations   []Correlation    `json:"correlations"`
    CausalChains   []CausalChain    `json:"causal_chains"`
    ValidationResults []ValidationResult `json:"validation_results"`
}

type TimelineEvent struct {
    Event          UnifiedEvent      `json:"event"`
    RelatedEvents  []string         `json:"related_events"`  // Event IDs
    CausalDepth    int              `json:"causal_depth"`    // 0 = root cause, 1 = direct effect, etc.
    ValidationTags []string         `json:"validation_tags"`
}

type CausalChain struct {
    ID           string              `json:"id"`
    RootCause    string              `json:"root_cause"`      // Event ID
    Events       []CausalChainEvent  `json:"events"`
    TotalDelay   time.Duration       `json:"total_delay"`
    Confidence   float64             `json:"confidence"`
    Validated    bool                `json:"validated"`
}

type CausalChainEvent struct {
    EventID      string        `json:"event_id"`
    CausalDepth  int           `json:"causal_depth"`
    DelayFromRoot time.Duration `json:"delay_from_root"`
    Confidence   float64       `json:"confidence"`
}

func (tr *TimelineReconstructor) ReconstructTimeline(query TimelineQuery) (*Timeline, error) {
    // Get all events in time range
    events, err := tr.eventStore.FindEvents(EventQuery{
        StartTime: query.StartTime,
        EndTime:   query.EndTime,
        ScenarioID: query.ScenarioID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get events: %w", err)
    }
    
    // Get all correlations for these events
    correlations, err := tr.correlationStore.FindCorrelations(CorrelationQuery{
        EventIDs:  extractEventIDs(events),
        StartTime: query.StartTime,
        EndTime:   query.EndTime,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get correlations: %w", err)
    }
    
    // Build causal chains
    causalChains := tr.buildCausalChains(events, correlations)
    
    // Create timeline events with causal depth
    timelineEvents := tr.createTimelineEvents(events, correlations, causalChains)
    
    // Sort chronologically
    sort.Slice(timelineEvents, func(i, j int) bool {
        return timelineEvents[i].Event.Timestamp.Before(timelineEvents[j].Event.Timestamp)
    })
    
    timeline := &Timeline{
        ID:            tr.generateTimelineID(),
        ScenarioID:    query.ScenarioID,
        StartTime:     query.StartTime,
        EndTime:       query.EndTime,
        Events:        timelineEvents,
        Correlations:  correlations,
        CausalChains:  causalChains,
    }
    
    return timeline, nil
}

func (tr *TimelineReconstructor) buildCausalChains(events []UnifiedEvent, correlations []Correlation) []CausalChain {
    // Group correlations by root cause
    rootCauses := make(map[string][]Correlation)
    
    for _, correlation := range correlations {
        rootID := tr.findRootCause(correlation.TriggerEvent.ID, correlations)
        rootCauses[rootID] = append(rootCauses[rootID], correlation)
    }
    
    var chains []CausalChain
    
    for rootID, relatedCorrelations := range rootCauses {
        chain := tr.buildChainFromRoot(rootID, relatedCorrelations, events)
        chains = append(chains, chain)
    }
    
    return chains
}
```

## ✅ Scenario Validation Framework

### Scenario Execution Validation
```go
type ScenarioValidator struct {
    expectedBehaviors map[string]ExpectedBehavior
    validationRules   []ValidationRule
    slaChecker       *SLAChecker
}

type ExpectedBehavior struct {
    ScenarioID    string                 `yaml:"scenario_id"`
    Phase         string                 `yaml:"phase"`
    TriggerEvent  EventPattern           `yaml:"trigger_event"`
    ExpectedEvents []ExpectedEventRule   `yaml:"expected_events"`
    SLARequirements []SLARequirement     `yaml:"sla_requirements"`
}

type ExpectedEventRule struct {
    Pattern       EventPattern  `yaml:"pattern"`
    WithinDelay   time.Duration `yaml:"within_delay"`
    MinConfidence float64       `yaml:"min_confidence"`
    Required      bool          `yaml:"required"`
}

type SLARequirement struct {
    Name        string        `yaml:"name"`
    MaxDelay    time.Duration `yaml:"max_delay"`
    Description string        `yaml:"description"`
}

// Example expected behaviors for scenarios
var ScenarioExpectedBehaviors = map[string]ExpectedBehavior{
    "stablecoin-depeg": {
        ScenarioID: "stablecoin-depeg",
        TriggerEvent: EventPattern{
            Source:    "market-data-simulator",
            EventType: "stablecoin_depeg_started",
        },
        ExpectedEvents: []ExpectedEventRule{
            {
                Pattern: EventPattern{
                    Source:    "risk-monitor",
                    EventType: "correlation_risk_alert",
                },
                WithinDelay:   4 * time.Hour,  // Should detect within 4 hours of depeg start
                MinConfidence: 0.8,
                Required:      true,
            },
            {
                Pattern: EventPattern{
                    Source:    "trading-engine", 
                    EventType: "arbitrage_opportunity_detected",
                },
                WithinDelay:   2 * time.Hour,  // Should detect arbitrage opportunities
                MinConfidence: 0.9,
                Required:      true,
            },
        },
        SLARequirements: []SLARequirement{
            {
                Name:        "risk_detection_sla",
                MaxDelay:    15 * time.Minute,
                Description: "Risk monitor must detect depeg within 15 minutes",
            },
            {
                Name:        "alert_escalation_sla", 
                MaxDelay:    5 * time.Minute,
                Description: "High severity alerts must escalate within 5 minutes",
            },
        },
    },
    "market-crash": {
        ScenarioID: "market-crash",
        TriggerEvent: EventPattern{
            Source:    "market-data-simulator",
            EventType: "price_shock_injected",
        },
        ExpectedEvents: []ExpectedEventRule{
            {
                Pattern: EventPattern{
                    Source:    "risk-monitor",
                    EventType: "drawdown_limit_breach",
                },
                WithinDelay:   2 * time.Minute,
                MinConfidence: 0.95,
                Required:      true,
            },
            {
                Pattern: EventPattern{
                    Source:    "trading-engine",
                    EventType: "emergency_position_reduction",
                },
                WithinDelay:   5 * time.Minute,
                MinConfidence: 0.85,
                Required:      true,
            },
        },
    },
}

func (sv *ScenarioValidator) ValidateScenarioExecution(scenarioID string, timeline *Timeline) ScenarioValidationResult {
    expectedBehavior, exists := sv.expectedBehaviors[scenarioID]
    if !exists {
        return ScenarioValidationResult{
            ScenarioID: scenarioID,
            Status:     ValidationStatusSkipped,
            Message:    "No expected behavior defined for scenario",
        }
    }
    
    result := ScenarioValidationResult{
        ScenarioID:     scenarioID,
        Status:         ValidationStatusPassed,
        EventResults:   make([]EventValidationResult, 0),
        SLAResults:     make([]SLAValidationResult, 0),
        Timestamp:      time.Now(),
    }
    
    // Find trigger event in timeline
    var triggerEvent *TimelineEvent
    for _, event := range timeline.Events {
        if sv.matchesPattern(event.Event, expectedBehavior.TriggerEvent) {
            triggerEvent = &event
            break
        }
    }
    
    if triggerEvent == nil {
        result.Status = ValidationStatusFailed
        result.Message = "Trigger event not found in timeline"
        return result
    }
    
    // Validate expected events
    for _, expectedEvent := range expectedBehavior.ExpectedEvents {
        eventResult := sv.validateExpectedEvent(expectedEvent, triggerEvent, timeline)
        result.EventResults = append(result.EventResults, eventResult)
        
        if expectedEvent.Required && !eventResult.Found {
            result.Status = ValidationStatusFailed
        }
    }
    
    // Validate SLA requirements
    for _, slaReq := range expectedBehavior.SLARequirements {
        slaResult := sv.validateSLARequirement(slaReq, triggerEvent, timeline)
        result.SLAResults = append(result.SLAResults, slaResult)
        
        if !slaResult.Met {
            result.Status = ValidationStatusFailed
        }
    }
    
    return result
}

func (sv *ScenarioValidator) validateExpectedEvent(expected ExpectedEventRule, trigger *TimelineEvent, timeline *Timeline) EventValidationResult {
    deadline := trigger.Event.Timestamp.Add(expected.WithinDelay)
    
    for _, event := range timeline.Events {
        if event.Event.Timestamp.After(deadline) {
            break  // Timeline is sorted, no more events within deadline
        }
        
        if event.Event.Timestamp.After(trigger.Event.Timestamp) && sv.matchesPattern(event.Event, expected.Pattern) {
            // Found matching event within deadline
            delay := event.Event.Timestamp.Sub(trigger.Event.Timestamp)
            
            return EventValidationResult{
                ExpectedEvent: expected,
                Found:         true,
                ActualEvent:   &event.Event,
                Delay:         delay,
                WithinSLA:     delay <= expected.WithinDelay,
            }
        }
    }
    
    return EventValidationResult{
        ExpectedEvent: expected,
        Found:         false,
        WithinSLA:     false,
    }
}
```

## 📊 Monitoring & Observability

### Prometheus Metrics (Audit Layer)
```
# Event ingestion metrics
audit_events_ingested_total{source, event_type}
audit_events_processed_total{source, status="success|failed"}
audit_ingestion_lag_seconds{source}
audit_ingestion_rate_events_per_second{source}

# Correlation metrics
audit_correlations_found_total{rule_name, confidence_bucket}
audit_correlation_detection_time_seconds{rule_name}
audit_causal_chain_length{scenario_id}
audit_causal_chain_confidence{scenario_id}

# Scenario validation metrics  
audit_scenario_validation_results_total{scenario_id, status="passed|failed|skipped"}
audit_scenario_sla_compliance{scenario_id, sla_name, status="met|violated"}
audit_expected_behavior_verification_rate{scenario_id}

# Risk system validation metrics
audit_risk_alert_detection_delay_seconds{alert_type, scenario_id}
audit_risk_sla_compliance_rate{sla_type}
audit_false_positive_rate{detection_type}
audit_false_negative_rate{detection_type}

# System performance impact metrics
audit_system_performance_degradation{service, metric_type, scenario_id}
audit_recovery_time_seconds{service, scenario_id}
audit_availability_during_chaos{service, scenario_id}

# Audit trail metrics
audit_trail_completeness_score{scenario_id}
audit_timeline_reconstruction_accuracy{scenario_id}
audit_regulatory_compliance_score{requirement_type}
```

### Event Stream Monitoring
```json
{
  "timestamp": "2025-09-16T14:23:45.123Z",
  "level": "info",
  "service": "audit-correlator",
  "correlation_id": "scenario-depeg-001",
  "event": "causal_chain_detected",
  "scenario_id": "stablecoin-depeg",
  "causal_chain_id": "chain-abc123",
  "root_cause": {
    "event_id": "event-xyz789",
    "source": "market-data-simulator",
    "event_type": "stablecoin_depeg_started",
    "timestamp": "2025-09-16T14:20:00.000Z"
  },
  "effects_detected": 5,
  "total_propagation_delay_ms": 900000,
  "confidence_score": 0.92,
  "sla_compliance": {
    "risk_detection_sla": "met",
    "alert_escalation_sla": "met"
  },
  "validation_status": "passed"
}
```

## 🧪 Testing

### Correlation Engine Testing
```go
func TestCausalChainDetection(t *testing.T) {
    correlator := NewTestCorrelationEngine()
    
    // Inject scenario events in sequence
    triggerEvent := UnifiedEvent{
        ID:        "trigger-1",
        Timestamp: time.Now(),
        Source:    "market-data-simulator",
        EventType: "price_shock_injected",
        ScenarioID: "market-crash-test",
    }
    
    correlator.ProcessEvent(triggerEvent)
    
    // Simulate risk alert 30 seconds later
    alertEvent := UnifiedEvent{
        ID:        "alert-1",
        Timestamp: triggerEvent.Timestamp.Add(30 * time.Second),
        Source:    "risk-monitor", 
        EventType: "drawdown_limit_breach",
        ScenarioID: "market-crash-test",
    }
    
    correlator.ProcessEvent(alertEvent)
    
    // Verify correlation detected
    correlations := correlator.GetCorrelations("market-crash-test")
    require.Len(t, correlations, 1)
    
    correlation := correlations[0]
    assert.Equal(t, triggerEvent.ID, correlation.TriggerEvent.ID)
    assert.Equal(t, alertEvent.ID, correlation.EffectEvent.ID)
    assert.Equal(t, 30*time.Second, correlation.Delay)
    assert.Greater(t, correlation.Confidence, 0.8)
}

func TestScenarioValidation(t *testing.T) {
    validator := NewTestScenarioValidator()
    
    // Create test timeline with expected events
    timeline := createTestTimeline("stablecoin-depeg", []TestEvent{
        {Source: "market-data-simulator", Type: "stablecoin_depeg_started", Offset: 0},
        {Source: "risk-monitor", Type: "correlation_risk_alert", Offset: 10 * time.Minute},
        {Source: "trading-engine", Type: "arbitrage_opportunity_detected", Offset: 30 * time.Minute},
    })
    
    result := validator.ValidateScenarioExecution("stablecoin-depeg", timeline)
    
    assert.Equal(t, ValidationStatusPassed, result.Status)
    assert.Len(t, result.EventResults, 2)
    assert.All(t, result.EventResults, func(r EventValidationResult) bool {
        return r.Found && r.WithinSLA
    })
}

func TestSLACompliance(t *testing.T) {
    slaChecker := NewTestSLAChecker()
    
    // Test risk detection SLA compliance
    triggerTime := time.Now()
    alertTime := triggerTime.Add(30 * time.Second)  // Within 15 minute SLA
    
    compliance := slaChecker.CheckSLA("risk_detection_sla", triggerTime, alertTime)
    assert.True(t, compliance.Met)
    assert.Equal(t, 30*time.Second, compliance.ActualDelay)
    assert.Less(t, compliance.ActualDelay, 15*time.Minute)
}
```

### Integration Testing
```bash
# Test complete correlation pipeline
make test-correlation-pipeline

# Test scenario validation with real telemetry
make test-scenario-validation

# Test performance under high event load
make test-high-load-correlation

# Test audit trail generation
make test-audit-trail-generation
```

### Performance Testing
```go
func BenchmarkEventCorrelation(b *testing.B) {
    correlator := NewCorrelationEngine()
    events := generateTestEvents(10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        correlator.ProcessEvent(events[i%len(events)])
    }
}

func BenchmarkTimelineReconstruction(b *testing.B) {
    reconstructor := NewTimelineReconstructor()
    
    // Pre-populate with test data
    events := generateTestEvents(50000)
    correlations := generateTestCorrelations(5000)
    
    query := TimelineQuery{
        StartTime: time.Now().Add(-1 * time.Hour),
        EndTime:   time.Now(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := reconstructor.ReconstructTimeline(query)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## ⚙️ Configuration

### Environment Variables
```bash
# Core settings
AUDIT_CORRELATOR_PORT=8085
AUDIT_CORRELATOR_GRPC_PORT=50057
AUDIT_CORRELATOR_LOG_LEVEL=info

# Data source endpoints
PROMETHEUS_URL=http://prometheus:9090
OTEL_COLLECTOR_ENDPOINT=http://otel-collector:4317
LOG_AGGREGATOR_URL=http://loki:3100

# Service endpoints (for direct API access)
RISK_MONITOR_URL=http://risk-monitor:8080
EXCHANGE_SIMULATOR_URL=http://exchange-simulator:8080
CUSTODIAN_SIMULATOR_URL=http://custodian-simulator:8081
MARKET_DATA_SIMULATOR_URL=http://market-data-simulator:8082
TRADING_ENGINE_URL=http://trading-engine:8083

# Correlation settings
CORRELATION_TIME_WINDOW=10m
MAX_CORRELATION_DELAY=5m
MIN_CONFIDENCE_THRESHOLD=0.7
ENABLE_REAL_TIME_CORRELATION=true

# Storage settings
EVENT_RETENTION_DAYS=90
CORRELATION_RETENTION_DAYS=365
AUDIT_TRAIL_RETENTION_DAYS=2555  # 7 years for regulatory compliance
TIMELINE_CACHE_SIZE=1000

# Performance settings
MAX_CONCURRENT_CORRELATIONS=100
EVENT_BUFFER_SIZE=10000
CORRELATION_BATCH_SIZE=100
```

### Configuration File
```yaml
# config.yaml
audit_correlator:
  name: "audit-correlator"
  version: "1.0.0"
  
data_sources:
  prometheus:
    url: "http://prometheus:9090"
    scrape_interval: "5s"
    queries:
      - name: "risk_monitor_signals"
        promql: "risk_alert_generated"
        labels: ["severity", "breach_type", "symbol"]
      - name: "trading_engine_metrics"
        promql: "trading_strategy_pnl"
        labels: ["strategy_name"]
      - name: "system_health"
        promql: "up"
        labels: ["job", "instance"]
  
  opentelemetry:
    endpoint: "http://otel-collector:4317"
    insecure: true
    timeout: "30s"
    
  structured_logs:
    sources:
      - path: "/var/log/trading-ecosystem/*.log"
        format: "json"
        tail: true
      - path: "/var/log/risk-monitor/*.log" 
        format: "json"
        tail: true

correlation_rules:
  - name: "chaos_to_risk_alert"
    trigger:
      source: "test-coordinator"
      event_type: "chaos_injection_started"
    effect:
      source: "risk-monitor"
      event_type: "risk_alert_generated"
    max_delay: "2m"
    min_confidence: 0.8
    
  - name: "price_shock_to_strategy_halt"
    trigger:
      source: "market-data-simulator"
      event_type: "price_shock_injected"
    effect:
      source: "trading-engine"
      event_type: "strategy_halted"
    max_delay: "30s"
    min_confidence: 0.9
    
  - name: "settlement_delay_to_liquidity_alert"
    trigger:
      source: "custodian-simulator" 
      event_type: "settlement_delayed"
    effect:
      source: "risk-monitor"
      event_type: "liquidity_risk_alert"
    max_delay: "5m"
    min_confidence: 0.7

scenario_validation:
  expected_behaviors:
    stablecoin-depeg:
      trigger_event:
        source: "market-data-simulator"
        event_type: "stablecoin_depeg_started"
      expected_events:
        - pattern:
            source: "risk-monitor"
            event_type: "correlation_risk_alert"
          within_delay: "4h"
          min_confidence: 0.8
          required: true
        - pattern:
            source: "trading-engine"
            event_type: "arbitrage_opportunity_detected"
          within_delay: "2h"
          min_confidence: 0.9
          required: true
      sla_requirements:
        - name: "risk_detection_sla"
          max_delay: "15m"
          description: "Risk monitor must detect depeg within 15 minutes"

storage:
  event_store:
    type: "clickhouse"  # or "postgresql", "mongodb"
    connection_string: "clickhouse://clickhouse:9000/audit"
    retention_days: 90
    
  correlation_store:
    type: "postgresql"
    connection_string: "postgres://user:pass@postgres:5432/correlations"
    retention_days: 365
    
  timeline_cache:
    type: "redis"
    connection_string: "redis://redis:6379/1"
    cache_size: 1000
    ttl: "24h"

performance:
  max_concurrent_correlations: 100
  event_buffer_size: 10000
  correlation_batch_size: 100
  timeline_reconstruction_timeout: "30s"
  
audit_trail:
  enabled: true
  output_format: "json"
  include_raw_events: true
  include_correlation_details: true
  regulatory_compliance_mode: true
  retention_years: 7
```

## 🐳 Docker Configuration

### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o audit-correlator cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl jq
WORKDIR /root/

COPY --from=builder /app/audit-correlator .
COPY --from=builder /app/config ./config
COPY --from=builder /app/correlation-rules ./correlation-rules

# Create directories for data and logs
RUN mkdir -p /data/events /data/correlations /data/timelines /var/log/audit-correlator

EXPOSE 8085 50057
CMD ["./audit-correlator", "--config=config/config.yaml"]
```

### Health Checks
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8085/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 60s
```

### Docker Compose Integration
```yaml
# docker-compose.yaml
version: '3.8'
services:
  audit-correlator:
    build: .
    ports:
      - "8085:8085"
      - "50057:50057"
    environment:
      - AUDIT_CORRELATOR_LOG_LEVEL=info
      - PROMETHEUS_URL=http://prometheus:9090
      - OTEL_COLLECTOR_ENDPOINT=http://otel-collector:4317
    volumes:
      - ./correlation-rules:/app/correlation-rules
      - ./data/audit:/data
      - ./logs:/var/log/trading-ecosystem:ro  # Read-only access to system logs
    depends_on:
      - prometheus
      - otel-collector
      - clickhouse
      - redis
    networks:
      - trading-network
      - audit-network  # Separate network for audit data

  # Supporting infrastructure
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    ports:
      - "9000:9000"
    networks:
      - audit-network

  redis:
    image: redis:alpine
    volumes:
      - redis_data:/data
    networks:
      - audit-network

volumes:
  clickhouse_data:
  redis_data:

networks:
  trading-network:
    external: true
  audit-network:
    driver: bridge
```

## 🔒 Security & Compliance

### Data Access Controls
```go
type DataAccessController struct {
    permissions map[string][]Permission
    auditLogger *AuditLogger
}

type Permission struct {
    Resource string   `json:"resource"`
    Actions  []string `json:"actions"`
    Scope    string   `json:"scope"`
}

func (dac *DataAccessController) ValidateAccess(userID, resource, action string) error {
    permissions, exists := dac.permissions[userID]
    if !exists {
        dac.auditLogger.LogUnauthorizedAccess(userID, resource, action)
        return fmt.Errorf("no permissions found for user: %s", userID)
    }
    
    for _, permission := range permissions {
        if permission.Resource == resource || permission.Resource == "*" {
            for _, allowedAction := range permission.Actions {
                if allowedAction == action || allowedAction == "*" {
                    dac.auditLogger.LogAuthorizedAccess(userID, resource, action)
                    return nil
                }
            }
        }
    }
    
    dac.auditLogger.LogUnauthorizedAccess(userID, resource, action)
    return fmt.Errorf("insufficient permissions for %s on %s", action, resource)
}
```

### Regulatory Compliance Features
- **Immutable Audit Trail**: Write-only event storage with cryptographic integrity
- **Data Retention Policies**: Automated data lifecycle management per regulatory requirements
- **Access Logging**: Complete audit trail of all data access and correlation queries
- **Data Anonymization**: PII scrubbing for events containing sensitive information
- **Compliance Reporting**: Automated generation of regulatory compliance reports

### Security Controls
- **TLS Encryption**: All service communication encrypted in transit
- **API Authentication**: JWT-based authentication for API access
- **Role-Based Access**: Granular permissions for different types of audit data
- **Data Encryption**: Sensitive event data encrypted at rest

## 📋 CLI Interface

### Command Line Usage
```bash
# Event monitoring and correlation
./audit-correlator events --source risk-monitor --tail        # Tail events from specific source
./audit-correlator correlate --scenario-id scenario-123       # Show correlations for scenario
./audit-correlator timeline --start 2025-09-16T10:00:00Z --end 2025-09-16T12:00:00Z

# Scenario validation
./audit-correlator validate-scenario --scenario-id stablecoin-depeg
./audit-correlator sla-report --timeframe 24h
./audit-correlator coverage-analysis --scenario-category market

# System analysis
./audit-correlator performance-impact --scenario-id market-crash-001
./audit-correlator anomaly-detection --timeframe 1h --confidence 0.9
./audit-correlator causal-chain --event-id event-abc123

# Audit and compliance
./audit-correlator audit-trail --start 2025-09-01 --end 2025-09-16 --format json
./audit-correlator compliance-report --requirement sox --timeframe quarterly
./audit-correlator export-evidence --scenario-id scenario-123 --format regulatory

# Data management
./audit-correlator health-check                               # Check all data sources
./audit-correlator reprocess-events --start-time 2025-09-16T10:00:00Z
./audit-correlator cleanup-old-data --older-than 90d
```

### Example CLI Workflows
```bash
# Monitor real-time correlations during scenario execution
./audit-correlator events --tail --format json | jq '.correlation_id' | sort | uniq -c

# Validate that a chaos scenario executed correctly
SCENARIO_ID="stablecoin-depeg-$(date +%Y%m%d)"
./audit-correlator validate-scenario --scenario-id $SCENARIO_ID --output-format html

# Generate comprehensive audit report for regulatory review
./audit-correlator audit-trail \
  --start "2025-09-01T00:00:00Z" \
  --end "2025-09-16T23:59:59Z" \
  --include-correlations \
  --include-validations \
  --format regulatory-xml \
  --output quarterly-audit-report.xml

# Analyze system performance impact of chaos scenarios
./audit-correlator performance-impact \
  --scenario-category operational \
  --timeframe 30d \
  --services "risk-monitor,trading-engine" \
  --metrics "latency,throughput,availability"
```

## 📊 Reporting & Analytics

### Audit Report Generation
```go
type AuditReportGenerator struct {
    eventStore       *EventStore
    correlationStore *CorrelationStore
    validationStore  *ValidationStore
    templateEngine   *ReportTemplateEngine
}

type AuditReport struct {
    ReportID        string              `json:"report_id"`
    GeneratedAt     time.Time           `json:"generated_at"`
    TimeRange       TimeRange           `json:"time_range"`
    ExecutiveSummary ExecutiveSummary   `json:"executive_summary"`
    SystemOverview  SystemOverview      `json:"system_overview"`
    ScenarioResults []ScenarioResult    `json:"scenario_results"`
    ComplianceStatus ComplianceStatus   `json:"compliance_status"`
    Recommendations []Recommendation    `json:"recommendations"`
    TechnicalDetails TechnicalDetails   `json:"technical_details"`
}

type ExecutiveSummary struct {
    TotalScenarios      int     `json:"total_scenarios"`
    SuccessfulScenarios int     `json:"successful_scenarios"`
    FailedScenarios     int     `json:"failed_scenarios"`
    OverallSuccessRate  float64 `json:"overall_success_rate"`
    SLAComplianceRate   float64 `json:"sla_compliance_rate"`
    SystemAvailability  float64 `json:"system_availability"`
    KeyFindings         []string `json:"key_findings"`
}

type ScenarioResult struct {
    ScenarioID          string                    `json:"scenario_id"`
    ScenarioName        string                    `json:"scenario_name"`
    ExecutionTime       time.Time                 `json:"execution_time"`
    Duration            time.Duration             `json:"duration"`
    Status              ValidationStatus          `json:"status"`
    EventsCorrelated    int                       `json:"events_correlated"`
    CausalChains        []CausalChainSummary     `json:"causal_chains"`
    SLACompliance       []SLAComplianceResult    `json:"sla_compliance"`
    PerformanceImpact   PerformanceImpactSummary `json:"performance_impact"`
    AnomaliesDetected   []AnomalySummary         `json:"anomalies_detected"`
}

func (arg *AuditReportGenerator) GenerateComplianceReport(timeRange TimeRange) (*AuditReport, error) {
    // Gather all scenario executions in time range
    scenarios, err := arg.getScenarioExecutions(timeRange)
    if err != nil {
        return nil, fmt.Errorf("failed to get scenario executions: %w", err)
    }
    
    report := &AuditReport{
        ReportID:    arg.generateReportID(),
        GeneratedAt: time.Now(),
        TimeRange:   timeRange,
    }
    
    // Calculate executive summary
    report.ExecutiveSummary = arg.calculateExecutiveSummary(scenarios)
    
    // Generate system overview
    report.SystemOverview = arg.generateSystemOverview(timeRange)
    
    // Process each scenario
    for _, scenario := range scenarios {
        scenarioResult := arg.processScenarioResult(scenario)
        report.ScenarioResults = append(report.ScenarioResults, scenarioResult)
    }
    
    // Determine compliance status
    report.ComplianceStatus = arg.assessComplianceStatus(scenarios)
    
    // Generate recommendations
    report.Recommendations = arg.generateRecommendations(scenarios)
    
    return report, nil
}
```

### Performance Analytics Dashboard
```go
type PerformanceAnalyticsDashboard struct {
    metricsClient    prometheus.API
    eventStore       *EventStore
    dashboardServer  *DashboardServer
}

type SystemPerformanceMetrics struct {
    CorrelationLatency    time.Duration            `json:"correlation_latency"`
    EventProcessingRate   float64                  `json:"event_processing_rate"`
    TimelineAccuracy      float64                  `json:"timeline_accuracy"`
    ValidationSuccessRate float64                  `json:"validation_success_rate"`
    ServiceMetrics        map[string]ServiceMetric `json:"service_metrics"`
}

type ServiceMetric struct {
    Availability          float64       `json:"availability"`
    AverageResponseTime   time.Duration `json:"average_response_time"`
    ErrorRate            float64       `json:"error_rate"`
    ThroughputDegradation float64      `json:"throughput_degradation"`
}

func (pad *PerformanceAnalyticsDashboard) GetRealTimeMetrics() (*SystemPerformanceMetrics, error) {
    metrics := &SystemPerformanceMetrics{
        ServiceMetrics: make(map[string]ServiceMetric),
    }
    
    // Query correlation latency
    latencyQuery := "histogram_quantile(0.95, audit_correlation_detection_time_seconds)"
    latencyResult, err := pad.metricsClient.Query(context.Background(), latencyQuery, time.Now())
    if err == nil {
        metrics.CorrelationLatency = pad.extractDuration(latencyResult)
    }
    
    // Query event processing rate
    rateQuery := "rate(audit_events_processed_total[5m])"
    rateResult, err := pad.metricsClient.Query(context.Background(), rateQuery, time.Now())
    if err == nil {
        metrics.EventProcessingRate = pad.extractFloat(rateResult)
    }
    
    // Query validation success rate
    validationQuery := `
        sum(rate(audit_scenario_validation_results_total{status="passed"}[5m])) /
        sum(rate(audit_scenario_validation_results_total[5m]))
    `
    validationResult, err := pad.metricsClient.Query(context.Background(), validationQuery, time.Now())
    if err == nil {
        metrics.ValidationSuccessRate = pad.extractFloat(validationResult)
    }
    
    // Query service-specific metrics
    services := []string{"risk-monitor", "trading-engine", "exchange-simulator", "custodian-simulator"}
    for _, service := range services {
        serviceMetric, err := pad.getServiceMetrics(service)
        if err == nil {
            metrics.ServiceMetrics[service] = serviceMetric
        }
    }
    
    return metrics, nil
}
```

## 🚀 Performance

### Benchmarks
- **Event Ingestion**: >10,000 events/second from multiple sources
- **Correlation Detection**: <100ms for simple correlations, <1s for complex causal chains
- **Timeline Reconstruction**: <5s for 24-hour timelines with >50k events
- **Scenario Validation**: <2s for comprehensive scenario validation

### Resource Usage
- **Memory**: ~500MB baseline + ~1GB for 24 hours of event history
- **CPU**: <40% single core during normal operation, <80% during high load
- **Network**: <50MB/hour telemetry ingestion
- **Disk**: ~100MB/day event storage (compressed), ~10MB/day correlation data

### Scalability Features
- **Horizontal Scaling**: Event processing can be sharded by service or time range
- **Data Partitioning**: Time-based partitioning for efficient historical queries
- **Caching Strategy**: Multi-level caching for frequently accessed correlations
- **Streaming Processing**: Real-time correlation detection with backpressure handling

## 🤝 Contributing

### Development Guidelines
1. Maintain comprehensive event correlation coverage
2. Ensure all correlations include confidence scoring
3. Add validation for new scenario types and expected behaviors
4. Test correlation accuracy with synthetic and real data
5. Document all correlation rules and validation logic

### Adding New Correlation Rules
1. Define correlation rule in YAML configuration
2. Implement custom correlation logic if needed
3. Add unit tests for correlation detection
4. Validate against historical data
5. Document expected system behaviors

### Extending Validation Framework
1. Implement new assertion validators for custom behaviors
2. Add SLA requirements for new system components
3. Create test scenarios for validation logic
4. Update audit report templates
5. Ensure regulatory compliance requirements are met

## 📚 References

- **Event Correlation Theory**: [Link to correlation analysis documentation]
- **Chaos Engineering Validation**: [Link to scenario validation best practices]
- **Audit Trail Standards**: [Link to regulatory compliance requirements]
- **OpenTelemetry Integration**: [Link to observability implementation guide]

---

**Status**: 🚧 Development Phase  
**Maintainer**: [Your team]  
**Last Updated**: September 2025
