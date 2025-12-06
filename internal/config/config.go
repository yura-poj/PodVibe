package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AppPort         string
	DBDSN           string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	StoragePath     string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads environment variables and fills Config with defaults when missing.
func Load() Config {
	appPort := getEnv("APP_PORT", "8080")
	dbDSN := getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/podvibe?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret")
	accessMinutes := getEnv("ACCESS_TOKEN_MINUTES", "20")
	refreshDays := getEnv("REFRESH_TOKEN_DAYS", "14")
	storage := getEnv("STORAGE_PATH", "./storage")

	accessTTL, err := strconv.Atoi(accessMinutes)
	if err != nil {
		log.Printf("invalid ACCESS_TOKEN_MINUTES, using default 20: %v", err)
		accessTTL = 20
	}
	refreshTTL, err := strconv.Atoi(refreshDays)
	if err != nil {
		log.Printf("invalid REFRESH_TOKEN_DAYS, using default 14: %v", err)
		refreshTTL = 14
	}

	return Config{
		AppPort:         appPort,
		DBDSN:           dbDSN,
		RedisAddr:       redisAddr,
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  time.Duration(accessTTL) * time.Minute,
		RefreshTokenTTL: time.Duration(refreshTTL) * 24 * time.Hour,
		StoragePath:     storage,
	}
}
