package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(redisURL string) (*RedisLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	
	// Настройка connection pool для предотвращения memory leaks
	opts.PoolSize = 10                    // Максимальное количество соединений
	opts.MinIdleConns = 2                 // Минимум idle соединений
	opts.MaxIdleConns = 5                 // Максимум idle соединений
	opts.ConnMaxLifetime = 30 * time.Minute // Максимальное время жизни соединения
	opts.ConnMaxIdleTime = 5 * time.Minute  // Максимальное время idle
	opts.PoolTimeout = 4 * time.Second      // Таймаут получения соединения
	opts.ReadTimeout = 3 * time.Second      // Таймаут чтения
	opts.WriteTimeout = 3 * time.Second     // Таймаут записи
	
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	
	return &RedisLimiter{client: client}, nil
}

// Lua скрипт для атомарного rate limiting
var rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current == false then
    redis.call('SET', key, 1, 'EX', window)
    return 1
end

current = tonumber(current)
if current >= limit then
    return 0
end

local new_count = redis.call('INCR', key)
return new_count <= limit and 1 or 0
`

// Allow проверяет лимит согласно ТЗ: 5 попыток входа/мин
func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	result, err := r.client.Eval(ctx, rateLimitScript, []string{key}, limit, int(window.Seconds())).Result()
	if err != nil {
		return false, err
	}
	
	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result")
	}
	
	return allowed == 1, nil
}

// GetRemaining возвращает оставшиеся попытки
func (r *RedisLimiter) GetRemaining(ctx context.Context, key string, limit int) (int, error) {
	count, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return limit, nil
	}
	if err != nil {
		return 0, err
	}
	
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// GetTTL возвращает время до сброса лимита
func (r *RedisLimiter) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// LoginKey генерирует ключ для rate limiting входа по IP
func LoginKey(ip string) string {
	return fmt.Sprintf("login:%s", ip)
}

// Ping - проверка соединения с Redis
func (r *RedisLimiter) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close закрывает соединение с Redis
func (r *RedisLimiter) Close() error {
	return r.client.Close()
}

// GetClient возвращает Redis клиент для мониторинга
func (r *RedisLimiter) GetClient() *redis.Client {
	return r.client
}