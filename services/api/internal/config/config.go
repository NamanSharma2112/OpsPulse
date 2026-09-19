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
	Env            string
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	JWTTTL         time.Duration
	CORSOrigins    []string
	GitHubClientID string
	GitHubSecret   string
	// GitHubRedirectURL must match the callback registered on the OAuth app.
	GitHubRedirectURL string
	// Base URLs, overridable for GitHub Enterprise and for tests.
	GitHubOAuthBase string
	GitHubAPIBase   string
	// AppURL is where the browser lands after a successful sign-in.
	AppURL string
	// PublicAPIURL is the address GitHub delivers webhooks to. It must be
	// reachable from the internet, so localhost will not do in practice.
	PublicAPIURL string
	// TokenEncryptionKey seals GitHub access tokens at rest, 32 bytes hex.
	TokenEncryptionKey string
	// SecureCookies marks the session cookie Secure; off for plain-HTTP dev.
	SecureCookies   bool
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, applying defaults suited to
// local development. It fails fast when a production-critical value is
// missing.
func Load() (Config, error) {
	cfg := Config{
		Env:                env("OPSPULSE_ENV", "development"),
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://opspulse:opspulse@localhost:5432/opspulse?sslmode=disable"),
		JWTSecret:          env("JWT_SECRET", ""),
		JWTTTL:             duration("JWT_TTL", 24*time.Hour),
		CORSOrigins:        list("CORS_ORIGINS", "http://localhost:3000"),
		GitHubClientID:     env("GITHUB_CLIENT_ID", ""),
		GitHubSecret:       env("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  env("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/github/callback"),
		GitHubOAuthBase:    env("GITHUB_OAUTH_BASE", "https://github.com"),
		GitHubAPIBase:      env("GITHUB_API_BASE", "https://api.github.com"),
		AppURL:             env("APP_URL", "http://localhost:3000"),
		PublicAPIURL:       env("PUBLIC_API_URL", "http://localhost:8080"),
		TokenEncryptionKey: env("TOKEN_ENCRYPTION_KEY", ""),
		SecureCookies:      boolean("SECURE_COOKIES", false),
		LogLevel:           env("LOG_LEVEL", "info"),
		ShutdownTimeout:    duration("SHUTDOWN_TIMEOUT", 15*time.Second),
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

	if cfg.TokenEncryptionKey == "" {
		if cfg.IsProduction() {
			return Config{}, fmt.Errorf("TOKEN_ENCRYPTION_KEY is required outside development")
		}
		// A fixed development key. Tokens sealed with it are readable by
		// anyone with this source, which is the point of failing in
		// production without a real one.
		cfg.TokenEncryptionKey = strings.Repeat("00", 32)
	}
	if cfg.IsProduction() && !cfg.SecureCookies {
		return Config{}, fmt.Errorf("SECURE_COOKIES must be true in production")
	}
	return cfg, nil
}

// GitHubConfigured reports whether sign-in with GitHub is available.
func (c Config) GitHubConfigured() bool {
	return c.GitHubClientID != "" && c.GitHubSecret != ""
}

func boolean(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch raw {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
