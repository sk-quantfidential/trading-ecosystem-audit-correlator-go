package services

import (
	"github.com/sirupsen/logrus"
)

type AuditService struct {
	logger *logrus.Logger
}

func NewAuditService(logger *logrus.Logger) *AuditService {
	return &AuditService{
		logger: logger,
	}
}

func (s *AuditService) LogEvent(eventType, source, message string) error {
	s.logger.WithFields(logrus.Fields{
		"eventType": eventType,
		"source":    source,
		"message":   message,
	}).Info("Logging audit event")
	return nil
}

func (s *AuditService) CorrelateEvents(timeWindow string) ([]string, error) {
	s.logger.WithField("timeWindow", timeWindow).Info("Correlating events")
	return []string{"correlation-1", "correlation-2"}, nil
}