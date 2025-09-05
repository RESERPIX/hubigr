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
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost/hubigr?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		SMTPHost:    getEnv("SMTP_HOST", ""),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		SMTPFrom:    getEnv("SMTP_FROM", "noreply@hubigr.com"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:3000"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		TurnstileSecret: getEnv("TURNSTILE_SECRET", ""),
	}
	
	// Проверка критически важных настроек
	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be set and at least 32 characters long for security")
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}