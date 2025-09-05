package main

import (
	"context"
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
	"github.com/RESERPIX/hubigr/internal/monitoring"
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
		// Логгер еще не инициализирован, используем stderr
		os.Stderr.WriteString("Failed to load config: " + err.Error() + "\n")
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
	
	// Настройка pool параметров для предотвращения memory leaks
	poolConfig.MaxConns = 10                // Уменьшаем для экономии памяти
	poolConfig.MinConns = 2                 // Минимум активных подключений
	poolConfig.MaxConnLifetime = time.Minute * 30  // Короткое время жизни
	poolConfig.MaxConnIdleTime = time.Minute * 5   // Быстрое закрытие idle
	poolConfig.HealthCheckPeriod = time.Minute     // Частые проверки
	poolConfig.MaxConnLifetimeJitter = time.Minute * 5 // Jitter для равномерного обновления
	
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
	refreshRepo := store.NewRefreshTokenRepo(db)

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
	
	// Инициализация мониторинга connection pools
	poolMonitor := monitoring.NewPoolMonitor(db, limiter.GetClient(), 30*time.Second)
	go poolMonitor.Start(context.Background())
	
	// Инициализация handlers
	handlers := http.NewHandlers(userRepo, refreshRepo, limiter, emailSender, avatarUploader, cfg.JWTSecret, turnstile, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

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
		logger.Info("Gracefully shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Останавливаем мониторинг pools
		poolMonitor.Stop()
		
		// Закрываем Redis соединение
		if err := limiter.Close(); err != nil {
			logger.Error("Failed to close Redis connection", "error", err)
		}
		
		// Закрываем базу данных с принудительным таймаутом
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		
		// Запускаем закрытие БД в горутине для контроля таймаута
		done := make(chan struct{})
		go func() {
			defer close(done)
			db.Close()
		}()
		
		// Ждем завершения или таймаута
		select {
		case <-done:
			logger.Info("Database connections closed gracefully")
		case <-shutdownCtx.Done():
			logger.Warn("Database shutdown timeout - forcing close")
			// БД уже закрывается в горутине, просто логируем
		}
		
		// Закрываем HTTP сервер
		_ = app.ShutdownWithContext(ctx)
	}()

	// Запуск сервера
	logger.Info("Server starting", "port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		logger.Info("Server stopped", "error", err)
	}
}