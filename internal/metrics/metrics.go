package metrics

import (
	"sync"
	"time"
)

// Metrics - структура для хранения метрик
type Metrics struct {
	mu sync.RWMutex
	
	// HTTP метрики
	RequestsTotal     map[string]int64 // по методам
	RequestDuration   map[string]time.Duration // средняя длительность
	ResponseStatus    map[int]int64 // по статус кодам
	
	// Бизнес метрики
	UsersRegistered   int64
	EmailsSent        int64
	LoginAttempts     int64
	FailedLogins      int64
	
	// Системные метрики
	StartTime         time.Time
	LastRequestTime   time.Time
}

var globalMetrics = &Metrics{
	RequestsTotal:   make(map[string]int64),
	RequestDuration: make(map[string]time.Duration),
	ResponseStatus:  make(map[int]int64),
	StartTime:       time.Now(),
}

// GetMetrics возвращает текущие метрики
func GetMetrics() *Metrics {
	return globalMetrics
}

// IncrementRequests увеличивает счетчик запросов
func IncrementRequests(method string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.RequestsTotal[method]++
	globalMetrics.LastRequestTime = time.Now()
}

// RecordDuration записывает длительность запроса
func RecordDuration(method string, duration time.Duration) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	// Простое скользящее среднее
	current := globalMetrics.RequestDuration[method]
	globalMetrics.RequestDuration[method] = (current + duration) / 2
}

// IncrementStatus увеличивает счетчик статус кодов
func IncrementStatus(status int) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.ResponseStatus[status]++
}

// IncrementUserRegistered увеличивает счетчик регистраций
func IncrementUserRegistered() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.UsersRegistered++
}

// IncrementEmailSent увеличивает счетчик отправленных email
func IncrementEmailSent() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.EmailsSent++
}

// IncrementLoginAttempt увеличивает счетчик попыток входа
func IncrementLoginAttempt(success bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.LoginAttempts++
	if !success {
		globalMetrics.FailedLogins++
	}
}

// GetSnapshot возвращает снимок метрик для безопасного чтения
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	uptime := time.Since(m.StartTime)
	
	return map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"requests_total": m.RequestsTotal,
		"request_duration_ms": func() map[string]float64 {
			durations := make(map[string]float64)
			for k, v := range m.RequestDuration {
				durations[k] = float64(v.Nanoseconds()) / 1e6
			}
			return durations
		}(),
		"response_status": m.ResponseStatus,
		"users_registered": m.UsersRegistered,
		"emails_sent": m.EmailsSent,
		"login_attempts": m.LoginAttempts,
		"failed_logins": m.FailedLogins,
		"last_request": m.LastRequestTime.Unix(),
	}
}