package metrics

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// MetricsMiddleware собирает метрики HTTP запросов
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := c.Method()
		
		// Увеличиваем счетчик запросов
		IncrementRequests(method)
		
		// Выполняем запрос
		err := c.Next()
		
		// Записываем метрики после выполнения
		duration := time.Since(start)
		status := c.Response().StatusCode()
		
		RecordDuration(method, duration)
		IncrementStatus(status)
		
		return err
	}
}