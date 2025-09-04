package http

import (
	"strings"
	"time"

	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/ratelimit"
	"github.com/RESERPIX/hubigr/internal/security"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware - проверка JWT токена
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(domain.NewError("unauthorized", "Токен отсутствует"))
		}

		// Извлечение токена из "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(domain.NewError("unauthorized", "Неверный формат токена"))
		}

		token := parts[1]
		claims, err := security.VerifyJWT(token, jwtSecret)
		if err != nil {
			return c.Status(401).JSON(domain.NewError("unauthorized", "Недействительный токен"))
		}

		// Сохранение данных пользователя в контексте
		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)
		c.Locals("user_nick", claims.Nick)

		return c.Next()
	}
}

// RoleMiddleware - проверка роли пользователя
func RoleMiddleware(allowedRoles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role").(string)
		
		for _, role := range allowedRoles {
			if string(role) == userRole {
				return c.Next()
			}
		}

		return c.Status(403).JSON(domain.NewError("forbidden", "Недостаточно прав"))
	}
}

// LoginRateLimitMiddleware - rate limiting для входа согласно ТЗ (5 попыток/мин)
func LoginRateLimitMiddleware(limiter *ratelimit.RedisLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := ratelimit.LoginKey(ip)
		
		allowed, err := limiter.Allow(c.Context(), key, 5, time.Minute)
		if err != nil {
			return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка проверки лимита"))
		}
		
		if !allowed {
			ttl, _ := limiter.GetTTL(c.Context(), key)
			return c.Status(429).JSON(domain.NewError("rate_limit_exceeded", 
				"Слишком много попыток входа. Попробуйте через "+ttl.String()))
		}
		
		return c.Next()
	}
}