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
	pipe := r.client.Pipeline()
	
	// Увеличиваем счетчик
	incr := pipe.Incr(ctx, key)
	// Устанавливаем TTL только при первом запросе
	pipe.Expire(ctx, key, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	
	count := incr.Val()
	return count <= int64(limit), nil
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