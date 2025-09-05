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

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// GetMetrics возвращает текущие метрики
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			RequestsTotal:   make(map[string]int64),
			RequestDuration: make(map[string]time.Duration),
			ResponseStatus:  make(map[int]int64),
			StartTime:       time.Now(),
		}
	})
	return globalMetrics
}

// IncrementRequests увеличивает счетчик запросов
func IncrementRequests(method string) {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RequestsTotal[method]++
	m.LastRequestTime = time.Now()
}

// RecordDuration записывает длительность запроса
func RecordDuration(method string, duration time.Duration) {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	// Экспоненциальное скользящее среднее (EMA)
	if current, exists := m.RequestDuration[method]; exists {
		// EMA: new_avg = alpha * new_value + (1 - alpha) * old_avg
		m.RequestDuration[method] = time.Duration(0.1*float64(duration) + 0.9*float64(current))
	} else {
		m.RequestDuration[method] = duration
	}
}

// IncrementStatus увеличивает счетчик статус кодов
func IncrementStatus(status int) {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ResponseStatus[status]++
}

// IncrementUserRegistered увеличивает счетчик регистраций
func IncrementUserRegistered() {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UsersRegistered++
}

// IncrementEmailSent увеличивает счетчик отправленных email
func IncrementEmailSent() {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmailsSent++
}

// IncrementLoginAttempt увеличивает счетчик попыток входа
func IncrementLoginAttempt(success bool) {
	m := GetMetrics()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoginAttempts++
	if !success {
		m.FailedLogins++
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