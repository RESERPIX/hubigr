package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RESERPIX/hubigr/internal/config"
	"github.com/RESERPIX/hubigr/internal/email"
	"github.com/RESERPIX/hubigr/internal/http"
	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/RESERPIX/hubigr/internal/ratelimit"
	"github.com/RESERPIX/hubigr/internal/store"
	"github.com/RESERPIX/hubigr/internal/upload"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()
	
	// Инициализация логгера
	logger.Init(cfg.LogLevel)
	logger.Info("Starting Hubigr Auth Service", "version", "1.0.0")

	// Подключение к базе данных
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Проверка подключения
	if err := db.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database", "error", err)
		log.Fatal("Failed to ping database:", err)
	}
	logger.Info("Database connected successfully")

	// Инициализация репозиториев
	userRepo := store.NewUserRepo(db)

	// Инициализация Redis rate limiter
	limiter, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		log.Fatal("Failed to connect to Redis:", err)
	}
	logger.Info("Redis connected successfully")

	// Инициализация email sender
	var emailSender http.EmailSender
	if cfg.SMTPHost != "" && cfg.SMTPUser != "" {
		// Продакшен: реальный SMTP
		emailSender = email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, cfg.BaseURL)
	} else {
		// Разработка: mock sender
		emailSender = email.NewMockSender()
		logger.Info("Using mock email sender for development")
	}

	// Инициализация avatar uploader
	avatarUploader := upload.NewAvatarUploader(cfg.BaseURL)
	
	// Инициализация handlers
	handlers := http.NewHandlers(userRepo, limiter, emailSender, avatarUploader, cfg.JWTSecret)

	// Создание Fiber приложения
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "internal_error",
					"message": err.Error(),
				},
			})
		},
	})

	// Настройка маршрутов
	http.SetupRoutes(app, handlers, cfg.JWTSecret)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
	}()

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}