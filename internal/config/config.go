package config

import (
	"fmt"
	"os"
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
}

func Load() (*Config, error) {
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
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}