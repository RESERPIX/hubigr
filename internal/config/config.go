package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
	BaseURL           string
	LogLevel          string
	TurnstileSecret   string
	CORSOrigins       string
	// TTL Policies - Политики времени жизни токенов
	AccessTokenTTL    int // Access token TTL в минутах (5-15 мин)
	RefreshTokenTTL   int // Refresh token TTL в днях (7-30 дней)
}

func Load() (*Config, error) {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		// .env файл не обязателен, продолжаем без него
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost/hubigr?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-key-32-characters-long"),
		SMTPHost:    getEnv("SMTP_HOST", ""),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		SMTPFrom:    getEnv("SMTP_FROM", "noreply@hubigr.com"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:3000"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		TurnstileSecret: getEnv("TURNSTILE_SECRET", ""),
		CORSOrigins:     getEnv("CORS_ORIGINS", "http://localhost:3000"),
		// TTL Policies
		AccessTokenTTL:  getEnvInt("ACCESS_TOKEN_TTL", 15),  // 15 минут по умолчанию
		RefreshTokenTTL: getEnvInt("REFRESH_TOKEN_TTL", 7),  // 7 дней по умолчанию
	}
	
	// Проверка критически важных настроек только в продакшене
	if getEnv("ENV", "development") == "production" {
		if cfg.JWTSecret == "dev-secret-key-32-characters-long" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		if cfg.DatabaseURL == "postgres://user:pass@localhost/hubigr?sslmode=disable" {
			return nil, fmt.Errorf("DATABASE_URL must be set in production")
		}
		if cfg.RedisURL == "redis://localhost:6379" {
			return nil, fmt.Errorf("REDIS_URL must be set in production")
		}
	}
	
	// Валидация TTL политик
	if cfg.AccessTokenTTL < 5 || cfg.AccessTokenTTL > 15 {
		return nil, fmt.Errorf("ACCESS_TOKEN_TTL must be between 5-15 minutes")
	}
	if cfg.RefreshTokenTTL < 7 || cfg.RefreshTokenTTL > 30 {
		return nil, fmt.Errorf("REFRESH_TOKEN_TTL must be between 7-30 days")
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intVal == 1 {
			return defaultValue
		}
	}
	return defaultValue
}