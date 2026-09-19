// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds every knob the API service reads at startup.
type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	JWTSecret       string
	JWTTTL          time.Duration
	CORSOrigins     []string
	GitHubClientID  string
	GitHubSecret    string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, applying defaults suited to
// local development. It fails fast when a production-critical value is
// missing.
func Load() (Config, error) {
	cfg := Config{
		Env:             env("OPSPULSE_ENV", "development"),
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://opspulse:opspulse@localhost:5432/opspulse?sslmode=disable"),
		JWTSecret:       env("JWT_SECRET", ""),
		JWTTTL:          duration("JWT_TTL", 24*time.Hour),
		CORSOrigins:     list("CORS_ORIGINS", "http://localhost:3000"),
		GitHubClientID:  env("GITHUB_CLIENT_ID", ""),
		GitHubSecret:    env("GITHUB_CLIENT_SECRET", ""),
		LogLevel:        env("LOG_LEVEL", "info"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 15*time.Second),
	}

	if cfg.JWTSecret == "" {
		if cfg.IsProduction() {
			return Config{}, fmt.Errorf("JWT_SECRET is required outside development")
		}
		cfg.JWTSecret = "dev-only-insecure-secret-change-me"
	}
	if len(cfg.JWTSecret) < 16 && cfg.IsProduction() {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

// IsProduction reports whether the service runs with production guardrails.
func (c Config) IsProduction() bool { return c.Env == "production" }

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(raw); err == nil {
		return time.Duration(secs) * time.Second
	}
	return fallback
}

func list(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
