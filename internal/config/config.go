package config

import (
	"errors"
	"fmt"
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

// Load reads environment variables, validates mandatory fields, and returns config.
// Required: DB_DSN, JWT_SECRET. Others have safe defaults.
func Load() (Config, error) {
	appPort := getEnv("APP_PORT", "8080")
	dbDSN := os.Getenv("DB_DSN")
	jwtSecret := os.Getenv("JWT_SECRET")
	if dbDSN == "" {
		return Config{}, errors.New("DB_DSN is required")
	}
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	accessMinutes := getEnv("ACCESS_TOKEN_MINUTES", "20")
	refreshDays := getEnv("REFRESH_TOKEN_DAYS", "14")
	storage := getEnv("STORAGE_PATH", "./storage")

	accessTTL, err := strconv.Atoi(accessMinutes)
	if err != nil || accessTTL <= 0 {
		return Config{}, fmt.Errorf("invalid ACCESS_TOKEN_MINUTES: %v", err)
	}
	refreshTTL, err := strconv.Atoi(refreshDays)
	if err != nil || refreshTTL <= 0 {
		return Config{}, fmt.Errorf("invalid REFRESH_TOKEN_DAYS: %v", err)
	}

	return Config{
		AppPort:         appPort,
		DBDSN:           dbDSN,
		RedisAddr:       redisAddr,
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  time.Duration(accessTTL) * time.Minute,
		RefreshTokenTTL: time.Duration(refreshTTL) * 24 * time.Hour,
		StoragePath:     storage,
	}, nil
}
