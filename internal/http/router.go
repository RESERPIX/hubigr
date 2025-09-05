package http

import (
	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App, handlers *Handlers, jwtSecret string) {
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
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
	auth.Post("/signup", LoginRateLimitMiddleware(handlers.limiter), handlers.SignUp)
	auth.Post("/login", LoginRateLimitMiddleware(handlers.limiter), handlers.Login)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/verify-email", handlers.VerifyEmail)
	auth.Post("/resend-verification", LoginRateLimitMiddleware(handlers.limiter), handlers.ResendVerification)
	auth.Post("/reset-password", LoginRateLimitMiddleware(handlers.limiter), handlers.ResetPasswordRequest)
	auth.Post("/reset-password/confirm", handlers.ResetPasswordConfirm)

	// Profile routes (API-4.6 - API-4.8 из ТЗ)
	profile := api.Group("/profile", AuthMiddleware(jwtSecret), LoggingMiddleware())
	profile.Get("/", handlers.GetProfile)
	profile.Put("/", handlers.UpdateProfile)
	profile.Get("/notifications", handlers.GetNotifications)
	profile.Put("/notifications", handlers.UpdateNotifications)
	profile.Get("/submissions", handlers.GetMySubmissions)
	profile.Post("/avatar", handlers.UploadAvatar)
	
	// Статические файлы (аватары)
	app.Static("/uploads", "./uploads")

	// Admin routes (для будущего расширения)
	admin := api.Group("/admin", AuthMiddleware(jwtSecret), RoleMiddleware(domain.RoleAdmin, domain.RoleModerator))
	_ = admin // Пока не используется

	// Health check
	api.Get("/health", handlers.Health)
}