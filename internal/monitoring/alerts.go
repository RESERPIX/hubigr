package monitoring

import (
	"context"
	"time"

	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/RESERPIX/hubigr/internal/metrics"
)

type AlertLevel string

const (
	AlertCritical AlertLevel = "critical"
	AlertWarning  AlertLevel = "warning"
)

type Alert struct {
	Level   AlertLevel `json:"level"`
	Service string     `json:"service"`
	Message string     `json:"message"`
	Time    time.Time  `json:"time"`
}

type AlertManager struct {
	alerts chan Alert
}

func NewAlertManager() *AlertManager {
	am := &AlertManager{
		alerts: make(chan Alert, 100),
	}
	go am.processAlerts()
	return am
}

func (am *AlertManager) SendAlert(level AlertLevel, service, message string) {
	select {
	case am.alerts <- Alert{
		Level:   level,
		Service: service,
		Message: message,
		Time:    time.Now(),
	}:
	default:
		logger.Error("Alert queue full, dropping alert", "service", service, "message", message)
	}
}

func (am *AlertManager) processAlerts() {
	for alert := range am.alerts {
		logger.Error("ALERT", 
			"level", alert.Level,
			"service", alert.Service,
			"message", alert.Message,
			"time", alert.Time)
	}
}

func (am *AlertManager) CheckMetrics(ctx context.Context) {
	m := metrics.GetMetrics()
	snapshot := m.GetSnapshot()

	// Проверка failed logins
	if failedLogins, ok := snapshot["failed_logins"].(int64); ok && failedLogins > 10 {
		am.SendAlert(AlertWarning, "auth", "High number of failed logins detected")
	}

	// Проверка uptime
	if uptime, ok := snapshot["uptime_seconds"].(float64); ok && uptime < 60 {
		am.SendAlert(AlertWarning, "system", "Service recently restarted")
	}
}