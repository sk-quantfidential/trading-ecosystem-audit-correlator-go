package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/adapters"
	"github.com/quantfidential/trading-ecosystem/audit-data-adapter-go/pkg/models"
)

type AuditService struct {
	dataAdapter adapters.DataAdapter
	logger      *logrus.Logger
}

func NewAuditService(logger *logrus.Logger) *AuditService {
	return &AuditService{
		logger: logger,
	}
}

func NewAuditServiceWithDataAdapter(dataAdapter adapters.DataAdapter, logger *logrus.Logger) *AuditService {
	return &AuditService{
		dataAdapter: dataAdapter,
		logger:      logger,
	}
}

func (s *AuditService) LogEvent(eventType, source, message string) error {
	ctx := context.Background()

	// If data adapter is not available, fall back to logging only
	if s.dataAdapter == nil {
		s.logger.WithFields(logrus.Fields{
			"eventType": eventType,
			"source":    source,
			"message":   message,
		}).Info("Logging audit event (no data adapter)")
		return nil
	}

	// Create audit event
	event := &models.AuditEvent{
		ID:          fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		TraceID:     fmt.Sprintf("trace-%d", time.Now().UnixNano()),
		SpanID:      fmt.Sprintf("span-%d", time.Now().UnixNano()),
		ServiceName: "audit-correlator",
		EventType:   eventType,
		Timestamp:   time.Now(),
		Status:      models.AuditEventStatusPending,
		Metadata: json.RawMessage(`{"source":"` + source + `","message":"` + message + `"}`),
		Tags: []string{eventType, source},
	}

	// Store in data adapter
	if err := s.dataAdapter.Create(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to store audit event")
		return fmt.Errorf("failed to store audit event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"eventType":  eventType,
		"source":     source,
		"message":    message,
	}).Info("Audit event stored successfully")

	return nil
}

func (s *AuditService) CorrelateEvents(timeWindow string) ([]string, error) {
	ctx := context.Background()

	// If data adapter is not available, return mock data
	if s.dataAdapter == nil {
		s.logger.WithField("timeWindow", timeWindow).Info("Correlating events (no data adapter)")
		return []string{"correlation-1", "correlation-2"}, nil
	}

	// Parse time window (for now, use last hour as default)
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	// Query events in time window using repository pattern
	query := models.AuditQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     100,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	events, err := s.dataAdapter.Query(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to query audit events")
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}

	// Enhanced correlation logic using repository patterns
	correlationIDs, err := s.performCorrelationAnalysis(ctx, events)
	if err != nil {
		s.logger.WithError(err).Error("Failed to perform correlation analysis")
		return nil, fmt.Errorf("failed to perform correlation analysis: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"timeWindow":     timeWindow,
		"events_found":   len(events),
		"correlations":   len(correlationIDs),
	}).Info("Event correlation completed")

	return correlationIDs, nil
}

// performCorrelationAnalysis performs sophisticated correlation analysis using repository patterns
func (s *AuditService) performCorrelationAnalysis(ctx context.Context, events []*models.AuditEvent) ([]string, error) {
	if len(events) == 0 {
		return []string{}, nil
	}

	var correlationIDs []string

	// 1. Trace-based correlation - group by trace ID
	traceCorrelations := s.correlateByTraceID(events)
	for traceID, eventIDs := range traceCorrelations {
		if len(eventIDs) > 1 {
			correlationIDs = append(correlationIDs, fmt.Sprintf("trace-%s-%d-events", traceID, len(eventIDs)))
		}
	}

	// 2. Service-based correlation - group by service name and event type
	serviceCorrelations := s.correlateByServiceAndType(events)
	for key, eventIDs := range serviceCorrelations {
		if len(eventIDs) > 1 {
			correlationIDs = append(correlationIDs, fmt.Sprintf("service-%s-%d-events", key, len(eventIDs)))
		}
	}

	// 3. Temporal correlation - find events within close time proximity
	temporalCorrelations := s.correlateByTemporalProximity(events, 5*time.Second)
	for i, group := range temporalCorrelations {
		if len(group) > 1 {
			correlationIDs = append(correlationIDs, fmt.Sprintf("temporal-group-%d-%d-events", i, len(group)))
		}
	}

	s.logger.WithFields(logrus.Fields{
		"trace_correlations":    len(traceCorrelations),
		"service_correlations":  len(serviceCorrelations),
		"temporal_correlations": len(temporalCorrelations),
		"total_correlations":    len(correlationIDs),
	}).Debug("Correlation analysis completed")

	return correlationIDs, nil
}

// correlateByTraceID groups events by trace ID
func (s *AuditService) correlateByTraceID(events []*models.AuditEvent) map[string][]string {
	correlations := make(map[string][]string)
	for _, event := range events {
		if event.TraceID != "" {
			correlations[event.TraceID] = append(correlations[event.TraceID], event.ID)
		}
	}
	return correlations
}

// correlateByServiceAndType groups events by service name and event type
func (s *AuditService) correlateByServiceAndType(events []*models.AuditEvent) map[string][]string {
	correlations := make(map[string][]string)
	for _, event := range events {
		key := fmt.Sprintf("%s-%s", event.ServiceName, event.EventType)
		correlations[key] = append(correlations[key], event.ID)
	}
	return correlations
}

// correlateByTemporalProximity groups events that occur within a time window
func (s *AuditService) correlateByTemporalProximity(events []*models.AuditEvent, window time.Duration) [][]string {
	if len(events) == 0 {
		return [][]string{}
	}

	var groups [][]string
	var currentGroup []string
	var groupStartTime time.Time

	for i, event := range events {
		if i == 0 {
			currentGroup = []string{event.ID}
			groupStartTime = event.Timestamp
			continue
		}

		// If within time window, add to current group
		if event.Timestamp.Sub(groupStartTime) <= window {
			currentGroup = append(currentGroup, event.ID)
		} else {
			// Start new group
			if len(currentGroup) > 1 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []string{event.ID}
			groupStartTime = event.Timestamp
		}
	}

	// Add the last group if it has multiple events
	if len(currentGroup) > 1 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// GetEventsByTraceID retrieves all events for a specific trace ID
func (s *AuditService) GetEventsByTraceID(traceID string) ([]*models.AuditEvent, error) {
	ctx := context.Background()

	if s.dataAdapter == nil {
		s.logger.WithField("traceID", traceID).Info("Getting events by trace ID (no data adapter)")
		return []*models.AuditEvent{}, nil
	}

	query := models.AuditQuery{
		TraceID:   &traceID,
		Limit:     1000, // Higher limit for trace-based queries
		SortBy:    "timestamp",
		SortOrder: "asc", // Chronological order for trace analysis
	}

	events, err := s.dataAdapter.Query(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to query events by trace ID")
		return nil, fmt.Errorf("failed to query events by trace ID: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"trace_id":     traceID,
		"events_found": len(events),
	}).Debug("Retrieved events by trace ID")

	return events, nil
}

// GetEventsByServiceType retrieves events for a specific service and event type
func (s *AuditService) GetEventsByServiceType(serviceName, eventType string, timeWindow time.Duration) ([]*models.AuditEvent, error) {
	ctx := context.Background()

	if s.dataAdapter == nil {
		s.logger.WithFields(logrus.Fields{
			"serviceName": serviceName,
			"eventType":   eventType,
		}).Info("Getting events by service type (no data adapter)")
		return []*models.AuditEvent{}, nil
	}

	endTime := time.Now()
	startTime := endTime.Add(-timeWindow)

	query := models.AuditQuery{
		ServiceName: &serviceName,
		EventType:   &eventType,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Limit:       500,
		SortBy:      "timestamp",
		SortOrder:   "desc",
	}

	events, err := s.dataAdapter.Query(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to query events by service type")
		return nil, fmt.Errorf("failed to query events by service type: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"event_type":   eventType,
		"time_window":  timeWindow,
		"events_found": len(events),
	}).Debug("Retrieved events by service type")

	return events, nil
}

// CreateCorrelation creates a correlation record between events
func (s *AuditService) CreateCorrelation(sourceEventID, targetEventID, correlationType string, confidence float64) error {
	ctx := context.Background()

	if s.dataAdapter == nil {
		s.logger.WithFields(logrus.Fields{
			"sourceEventID":   sourceEventID,
			"targetEventID":   targetEventID,
			"correlationType": correlationType,
			"confidence":      confidence,
		}).Info("Creating correlation (no data adapter)")
		return nil
	}

	correlation := &models.AuditCorrelation{
		ID:              fmt.Sprintf("corr-%d", time.Now().UnixNano()),
		SourceEventID:   sourceEventID,
		TargetEventID:   targetEventID,
		CorrelationType: correlationType,
		Confidence:      confidence,
		CreatedAt:       time.Now(),
	}

	if err := s.dataAdapter.CreateCorrelation(ctx, correlation); err != nil {
		s.logger.WithError(err).Error("Failed to create correlation")
		return fmt.Errorf("failed to create correlation: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"correlation_id":  correlation.ID,
		"source_event":    sourceEventID,
		"target_event":    targetEventID,
		"type":            correlationType,
		"confidence":      confidence,
	}).Info("Correlation created successfully")

	return nil
}

// GetHealthStatus returns the health status of the audit service
func (s *AuditService) GetHealthStatus() map[string]string {
	status := make(map[string]string)

	if s.dataAdapter == nil {
		status["data_adapter"] = "unavailable"
		status["audit_service"] = "stub_mode"
	} else {
		status["data_adapter"] = "connected"
		status["audit_service"] = "operational"
	}

	status["last_check"] = time.Now().Format(time.RFC3339)
	return status
}