package http

import (
	"time"

	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/metrics"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App, handlers *Handlers, jwtSecret string, corsOrigins string) {
	// Middleware
	app.Use(metrics.MetricsMiddleware())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Authorization,Content-Type",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// API группа
	api := app.Group("/api/v1")

	// Auth routes (API-4.1 - API-4.5 из ТЗ)
	auth := api.Group("/auth")
	auth.Post("/signup", LoginRateLimitMiddleware(handlers.limiter), CaptchaMiddleware(handlers.turnstile), handlers.SignUp)
	auth.Post("/login", LoginRateLimitMiddleware(handlers.limiter), CaptchaMiddleware(handlers.turnstile), handlers.Login)
	auth.Post("/logout", AuthMiddleware(jwtSecret), handlers.Logout)
	auth.Post("/refresh", LoginRateLimitMiddleware(handlers.limiter), handlers.RefreshToken)
	auth.Post("/verify-email", handlers.VerifyEmail)
	auth.Post("/resend-verification", LoginRateLimitMiddleware(handlers.limiter), CaptchaMiddleware(handlers.turnstile), handlers.ResendVerification)
	auth.Post("/reset-password", LoginRateLimitMiddleware(handlers.limiter), CaptchaMiddleware(handlers.turnstile), handlers.ResetPasswordRequest)
	auth.Post("/reset-password/confirm", handlers.ResetPasswordConfirm)

	// Profile routes (API-4.6 - API-4.8 из ТЗ)
	profile := api.Group("/profile", AuthMiddleware(jwtSecret), LoggingMiddleware(), RateLimitMiddleware(handlers.limiter, "profile", 30, time.Minute))
	profile.Get("/", handlers.GetProfile)
	profile.Put("/", handlers.UpdateProfile)
	profile.Get("/notifications", handlers.GetNotifications)
	profile.Put("/notifications", handlers.UpdateNotifications)
	profile.Get("/submissions", handlers.GetMySubmissions)
	profile.Post("/avatar", handlers.UploadAvatar)
	
	// Безопасная раздача статических файлов (аватары)
	app.Get("/uploads/*", SecureStaticHandler)

	// Admin routes (для будущего расширения)
	admin := api.Group("/admin", AuthMiddleware(jwtSecret), RoleMiddleware(domain.RoleAdmin, domain.RoleModerator))
	_ = admin // Пока не используется

	// Health check
	api.Get("/health", handlers.Health)
	
	// Metrics endpoints
	api.Get("/metrics", handlers.Metrics)
	api.Get("/metrics/prometheus", handlers.PrometheusMetrics)
}