package metrics

import (
	"fmt"
	"strings"
)

// PrometheusFormat возвращает метрики в формате Prometheus
func PrometheusFormat() string {
	m := GetMetrics()
	snapshot := m.GetSnapshot()
	
	var sb strings.Builder
	
	// HTTP метрики
	sb.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
	sb.WriteString("# TYPE http_requests_total counter\n")
	if val, exists := snapshot["requests_total"]; exists {
		if requests, ok := val.(map[string]int64); ok {
			for method, count := range requests {
				sb.WriteString(fmt.Sprintf("http_requests_total{method=\"%s\"} %d\n", method, count))
			}
		}
	}
	
	// Статус коды
	sb.WriteString("# HELP http_responses_total Total number of HTTP responses by status\n")
	sb.WriteString("# TYPE http_responses_total counter\n")
	if val, exists := snapshot["response_status"]; exists {
		if statuses, ok := val.(map[int]int64); ok {
			for status, count := range statuses {
				sb.WriteString(fmt.Sprintf("http_responses_total{status=\"%d\"} %d\n", status, count))
			}
		}
	}
	
	// Длительность запросов
	sb.WriteString("# HELP http_request_duration_ms Average request duration in milliseconds\n")
	sb.WriteString("# TYPE http_request_duration_ms gauge\n")
	if val, exists := snapshot["request_duration_ms"]; exists {
		if durations, ok := val.(map[string]float64); ok {
			for method, duration := range durations {
				sb.WriteString(fmt.Sprintf("http_request_duration_ms{method=\"%s\"} %.2f\n", method, duration))
			}
		}
	}
	
	// Бизнес метрики с безопасными проверками
	sb.WriteString("# HELP users_registered_total Total number of registered users\n")
	sb.WriteString("# TYPE users_registered_total counter\n")
	if val, exists := snapshot["users_registered"]; exists {
		if count, ok := val.(int64); ok {
			sb.WriteString(fmt.Sprintf("users_registered_total %d\n", count))
		}
	}
	
	sb.WriteString("# HELP emails_sent_total Total number of emails sent\n")
	sb.WriteString("# TYPE emails_sent_total counter\n")
	if val, exists := snapshot["emails_sent"]; exists {
		if count, ok := val.(int64); ok {
			sb.WriteString(fmt.Sprintf("emails_sent_total %d\n", count))
		}
	}
	
	sb.WriteString("# HELP login_attempts_total Total number of login attempts\n")
	sb.WriteString("# TYPE login_attempts_total counter\n")
	if val, exists := snapshot["login_attempts"]; exists {
		if count, ok := val.(int64); ok {
			sb.WriteString(fmt.Sprintf("login_attempts_total %d\n", count))
		}
	}
	
	sb.WriteString("# HELP failed_logins_total Total number of failed login attempts\n")
	sb.WriteString("# TYPE failed_logins_total counter\n")
	if val, exists := snapshot["failed_logins"]; exists {
		if count, ok := val.(int64); ok {
			sb.WriteString(fmt.Sprintf("failed_logins_total %d\n", count))
		}
	}
	
	// Uptime
	sb.WriteString("# HELP uptime_seconds Service uptime in seconds\n")
	sb.WriteString("# TYPE uptime_seconds gauge\n")
	if val, exists := snapshot["uptime_seconds"]; exists {
		if uptime, ok := val.(float64); ok {
			sb.WriteString(fmt.Sprintf("uptime_seconds %.2f\n", uptime))
		}
	}
	
	return sb.String()
}