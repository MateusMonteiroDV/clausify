package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment    string
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	MigrationsPath string
	GeminiAPIKey   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Environment:    getEnv("ENVIRONMENT", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://clausify:clausify@localhost:5432/clausify?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-change-in-production"),
		JWTExpireHours: 24,
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://migrations"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
