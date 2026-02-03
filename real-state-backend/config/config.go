package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"real-state-backend/internal/repository"
	"time"
)

type Config struct {
	ServerPort        string
	DBUrl             string
	Environment       string // dev, prod
	JWTSecret         string
	JWTPepper         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

func LoadConfig() *Config {

	// Construir a partir de variables individuales
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "realstatedb")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		DBUrl:             dbURL,
		Environment:       getEnv("GO_ENV", "development"),
		JWTSecret:         getEnv("JWT_SECRET", "my-secret-key"),
		JWTPepper:         getEnv("JWT_PEPPER", "my-pepper-key"),
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   7 * 24 * time.Hour,
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
	}
}

// LoadConfigFromDB carga las configuraciones de seguridad desde la BD
func LoadConfigFromDB(cfg *Config, db *sql.DB) *Config {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	configRepo := repository.NewSecurityConfigRepository(db)

	// AccessTokenTTL
	if ttl, err := configRepo.GetDuration(ctx, "ACCESS_TOKEN_TTL_MINUTES"); err == nil {
		cfg.AccessTokenTTL = ttl
	} else {
		slog.Warn("Failed to load ACCESS_TOKEN_TTL from DB, using default", "error", err)
	}

	// RefreshTokenTTL
	if ttl, err := configRepo.GetDuration(ctx, "REFRESH_TOKEN_TTL_DAYS"); err == nil {
		cfg.RefreshTokenTTL = ttl * 24 // Convertir días a horas
	} else {
		slog.Warn("Failed to load REFRESH_TOKEN_TTL from DB, using default", "error", err)
	}

	// MaxFailedAttempts
	if attempts, err := configRepo.GetInt(ctx, "MAX_FAILED_ATTEMPTS"); err == nil {
		cfg.MaxFailedAttempts = attempts
	} else {
		slog.Warn("Failed to load MAX_FAILED_ATTEMPTS from DB, using default", "error", err)
	}

	// LockoutDuration
	if duration, err := configRepo.GetDuration(ctx, "LOCKOUT_DURATION_MINUTES"); err == nil {
		cfg.LockoutDuration = duration
	} else {
		slog.Warn("Failed to load LOCKOUT_DURATION from DB, using default", "error", err)
	}

	slog.Info("Security config loaded from database", "access_ttl", cfg.AccessTokenTTL, "refresh_ttl", cfg.RefreshTokenTTL)
	return cfg
}
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
