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
	
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	
	return &RedisLimiter{client: client}, nil
}

// Allow проверяет лимит согласно ТЗ: 5 попыток входа/мин
func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Проверяем текущий счетчик
	current, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		// Первый запрос - создаем ключ с TTL
		pipe := r.client.Pipeline()
		pipe.Set(ctx, key, 1, window)
		_, err := pipe.Exec(ctx)
		return err == nil, err
	}
	if err != nil {
		return false, err
	}
	
	// Проверяем лимит
	if current >= int64(limit) {
		return false, nil
	}
	
	// Увеличиваем счетчик БЕЗ сброса TTL
	newCount, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	
	return newCount <= int64(limit), nil
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