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

	// Query events in time window
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

	// Simple correlation logic - group by trace ID
	correlations := make(map[string][]string)
	for _, event := range events {
		if event.TraceID != "" {
			correlations[event.TraceID] = append(correlations[event.TraceID], event.ID)
		}
	}

	// Return correlation IDs
	var correlationIDs []string
	for traceID, eventIDs := range correlations {
		if len(eventIDs) > 1 { // Only include traces with multiple events
			correlationIDs = append(correlationIDs, fmt.Sprintf("trace-%s-%d-events", traceID, len(eventIDs)))
		}
	}

	s.logger.WithFields(logrus.Fields{
		"timeWindow":     timeWindow,
		"events_found":   len(events),
		"correlations":   len(correlationIDs),
	}).Info("Event correlation completed")

	return correlationIDs, nil
}