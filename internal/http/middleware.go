package http

import (
	"strings"
	"time"

	"github.com/RESERPIX/hubigr/internal/captcha"
	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/logger"
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
		userRole, ok := c.Locals("user_role").(string)
		if !ok {
			return c.Status(403).JSON(domain.NewError("forbidden", "Ошибка проверки роли"))
		}
		
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
	return RateLimitMiddleware(limiter, "auth", 5, time.Minute)
}

// RateLimitMiddleware - универсальный rate limiting
func RateLimitMiddleware(limiter *ratelimit.RedisLimiter, prefix string, limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := prefix + ":" + ip
		
		allowed, err := limiter.Allow(c.Context(), key, limit, window)
		if err != nil {
			return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка проверки лимита"))
		}
		
		if !allowed {
			ttl, _ := limiter.GetTTL(c.Context(), key)
			return c.Status(429).JSON(domain.NewError("rate_limit_exceeded", 
				"Слишком много запросов. Попробуйте через "+ttl.String()))
		}
		
		return c.Next()
	}
}

// LoggingMiddleware - логирование действий пользователей
func LoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID != nil {
			logger.Info("User action",
				"user_id", userID,
				"method", c.Method(),
				"path", c.Path(),
				"ip", c.IP(),
				"user_agent", c.Get("User-Agent"),
			)
		}
		return c.Next()
	}
}

// CaptchaMiddleware - проверка Turnstile капчи
func CaptchaMiddleware(turnstile *captcha.TurnstileService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			CaptchaToken string `json:"captcha_token"`
		}
		
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
		}
		
		if req.CaptchaToken == "" {
			return c.Status(400).JSON(domain.NewError("captcha_required", "Требуется пройти проверку капчи"))
		}
		
		valid, err := turnstile.Verify(req.CaptchaToken, c.IP())
		if err != nil {
			logger.Error("Captcha verification error", "error", err, "ip", c.IP())
			return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка проверки капчи"))
		}
		
		if !valid {
			return c.Status(400).JSON(domain.NewError("captcha_invalid", "Неверная капча"))
		}
		
		return c.Next()
	}
}