package monitoring

import (
	"context"
	"time"

	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PoolMonitor отслеживает состояние connection pools
type PoolMonitor struct {
	dbPool    *pgxpool.Pool
	redisPool *redis.Client
	interval  time.Duration
	stopCh    chan struct{}
}

// NewPoolMonitor создает новый монитор connection pools
func NewPoolMonitor(dbPool *pgxpool.Pool, redisPool *redis.Client, interval time.Duration) *PoolMonitor {
	return &PoolMonitor{
		dbPool:    dbPool,
		redisPool: redisPool,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start запускает мониторинг connection pools
func (pm *PoolMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.checkPools(ctx)
		case <-pm.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop останавливает мониторинг
func (pm *PoolMonitor) Stop() {
	close(pm.stopCh)
}

// checkPools проверяет состояние всех pools
func (pm *PoolMonitor) checkPools(ctx context.Context) {
	pm.checkDBPool()
	pm.checkRedisPool(ctx)
}

// checkDBPool проверяет PostgreSQL connection pool
func (pm *PoolMonitor) checkDBPool() {
	stats := pm.dbPool.Stat()
	
	// Логируем статистику
	logger.Info("DB Pool Stats",
		"acquired_conns", stats.AcquiredConns(),
		"constructing_conns", stats.ConstructingConns(),
		"idle_conns", stats.IdleConns(),
		"max_conns", stats.MaxConns(),
		"total_conns", stats.TotalConns(),
	)

	// Проверяем на потенциальные проблемы
	if stats.AcquiredConns() > int32(float64(stats.MaxConns())*0.8) {
		logger.Warn("DB Pool: High connection usage",
			"acquired", stats.AcquiredConns(),
			"max", stats.MaxConns(),
			"usage_percent", float64(stats.AcquiredConns())/float64(stats.MaxConns())*100,
		)
	}

	if stats.ConstructingConns() > 5 {
		logger.Warn("DB Pool: Many connections being constructed",
			"constructing", stats.ConstructingConns(),
		)
	}

	// Проверяем на заблокированные acquire операции
	if stats.CanceledAcquireCount() > 0 {
		logger.Warn("DB Pool: Canceled acquire operations detected",
			"canceled_count", stats.CanceledAcquireCount(),
		)
	}
}

// checkRedisPool проверяет Redis connection pool
func (pm *PoolMonitor) checkRedisPool(ctx context.Context) {
	stats := pm.redisPool.PoolStats()
	
	// Логируем статистику
	logger.Info("Redis Pool Stats",
		"hits", stats.Hits,
		"misses", stats.Misses,
		"timeouts", stats.Timeouts,
		"total_conns", stats.TotalConns,
		"idle_conns", stats.IdleConns,
		"stale_conns", stats.StaleConns,
	)

	// Проверяем на потенциальные проблемы
	if stats.Timeouts > 0 {
		logger.Warn("Redis Pool: Connection timeouts detected",
			"timeouts", stats.Timeouts,
		)
	}

	if stats.StaleConns > 0 {
		logger.Warn("Redis Pool: Stale connections detected",
			"stale_conns", stats.StaleConns,
		)
	}

	// Проверяем hit rate
	if stats.Hits+stats.Misses > 0 {
		hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
		if hitRate < 50 {
			logger.Warn("Redis Pool: Low hit rate",
				"hit_rate_percent", hitRate,
			)
		}
	}
}

// GetDBPoolStats возвращает статистику DB pool для метрик
func (pm *PoolMonitor) GetDBPoolStats() map[string]interface{} {
	stats := pm.dbPool.Stat()
	return map[string]interface{}{
		"acquired_conns":      stats.AcquiredConns(),
		"constructing_conns": stats.ConstructingConns(),
		"idle_conns":          stats.IdleConns(),
		"max_conns":           stats.MaxConns(),
		"total_conns":         stats.TotalConns(),
		"acquire_count":       stats.AcquireCount(),
		"acquire_duration_ms": stats.AcquireDuration().Milliseconds(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
		"empty_acquire_count":    stats.EmptyAcquireCount(),
	}
}

// GetRedisPoolStats возвращает статистику Redis pool для метрик
func (pm *PoolMonitor) GetRedisPoolStats() map[string]interface{} {
	stats := pm.redisPool.PoolStats()
	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}