package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds environment-based server configuration.
type Config struct {
	AppEnv          string
	LogLevel        string
	ServerVersion   string
	HTTPAddr        string
	PostgresDSN     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	ShutdownTimeout time.Duration
}

// Load reads configuration from environment variables with development defaults.
func Load() (*Config, error) {
	appEnv := envOrDefault("APP_ENV", "development")

	cfg := &Config{
		AppEnv:          appEnv,
		LogLevel:        envOrDefault("LOG_LEVEL", defaultLogLevel(appEnv)),
		ServerVersion:   envOrDefault("SERVER_VERSION", "dev"),
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		PostgresDSN:     envOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/slimeyquest?sslmode=disable"),
		RedisAddr:       envOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		ShutdownTimeout: 10 * time.Second,
	}

	redisDB, err := strconv.Atoi(envOrDefault("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_DB: %w", err)
	}
	cfg.RedisDB = redisDB

	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		cfg.ShutdownTimeout, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
		}
	}

	return cfg, nil
}

// IsDevelopment reports whether the app runs in development mode.
func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.AppEnv, "development")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultLogLevel(appEnv string) string {
	if strings.EqualFold(appEnv, "production") {
		return "info"
	}
	return "debug"
}
