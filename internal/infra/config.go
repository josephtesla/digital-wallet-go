package infra

import (
	"os"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	PaystackSecretKey string
	PaystackPublicKey string
	Port              string
	LogLevel          string
	GinMode           string
	JWTSecret         string
	Environment       string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		PaystackSecretKey: getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey: getEnv("PAYSTACK_PUBLIC_KEY", ""),
		Port:              getEnv("PORT", "8080"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		GinMode:           getEnv("GIN_MODE", "debug"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		Environment:       getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
