package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RESERPIX/hubigr/internal/captcha"
	"github.com/RESERPIX/hubigr/internal/config"
	"github.com/RESERPIX/hubigr/internal/email"
	"github.com/RESERPIX/hubigr/internal/errors"
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
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		os.Exit(1)
	}
	
	// Инициализация логгера
	logger.Init(cfg.LogLevel)
	logger.Info("Starting Hubigr Auth Service", "version", "1.0.0")

	// Конфигурация connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to parse database URL", "error", err)
		os.Exit(1)
	}
	
	// Настройка pool параметров
	poolConfig.MaxConns = 25                // Максимум подключений
	poolConfig.MinConns = 5                 // Минимум подключений
	poolConfig.MaxConnLifetime = time.Hour  // Время жизни подключения
	poolConfig.MaxConnIdleTime = time.Minute * 30 // Время простоя
	poolConfig.HealthCheckPeriod = time.Minute * 5 // Проверка здоровья
	
	// Подключение к базе данных
	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверка подключения
	if err := db.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("Database connected successfully")

	// Инициализация репозиториев
	userRepo := store.NewUserRepo(db)

	// Инициализация Redis rate limiter
	limiter, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
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
	
	// Инициализация Turnstile
	turnstile := captcha.NewTurnstileService(cfg.TurnstileSecret)
	
	// Инициализация handlers
	handlers := http.NewHandlers(userRepo, limiter, emailSender, avatarUploader, cfg.JWTSecret, turnstile)

	// Создание Fiber приложения
	app := fiber.New(fiber.Config{
		ErrorHandler: errors.ErrorHandler,
	})

	// Настройка маршрутов
	http.SetupRoutes(app, handlers, cfg.JWTSecret, cfg.CORSOrigins)

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