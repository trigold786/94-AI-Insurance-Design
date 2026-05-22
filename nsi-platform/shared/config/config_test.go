package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.DatabaseURL != "postgres://postgres:postgres123@localhost:35432/nsi?sslmode=disable" {
		t.Errorf("unexpected DatabaseURL default: %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://localhost:36379/0" {
		t.Errorf("unexpected RedisURL default: %s", cfg.RedisURL)
	}
	if cfg.JWTSecret != "dev-secret-change-me" {
		t.Errorf("unexpected JWTSecret default: %s", cfg.JWTSecret)
	}
	if cfg.ServerPort != 30001 {
		t.Errorf("unexpected ServerPort default: %d", cfg.ServerPort)
	}
	if cfg.StorageBucket != "nsi-reports" {
		t.Errorf("unexpected StorageBucket default: %s", cfg.StorageBucket)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:9999/test")
	os.Setenv("REDIS_URL", "redis://localhost:9999/1")
	os.Setenv("JWT_SECRET", "production-secret")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("STORAGE_BUCKET", "test-bucket")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("STORAGE_BUCKET")
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://test:test@localhost:9999/test" {
		t.Errorf("expected overridden DatabaseURL, got %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://localhost:9999/1" {
		t.Errorf("expected overridden RedisURL, got %s", cfg.RedisURL)
	}
	if cfg.JWTSecret != "production-secret" {
		t.Errorf("expected overridden JWTSecret, got %s", cfg.JWTSecret)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("expected overridden ServerPort, got %d", cfg.ServerPort)
	}
	if cfg.StorageBucket != "test-bucket" {
		t.Errorf("expected overridden StorageBucket, got %s", cfg.StorageBucket)
	}
}

func TestGetEnvIntInvalid(t *testing.T) {
	os.Setenv("SERVER_PORT", "not-a-number")
	defer os.Unsetenv("SERVER_PORT")

	cfg := Load()
	if cfg.ServerPort != 30001 {
		t.Errorf("expected fallback for invalid SERVER_PORT, got %d", cfg.ServerPort)
	}
}

func TestGetEnvEmptyReturnsFallback(t *testing.T) {
	result := getEnv("NONEXISTENT_VAR_12345", "fallback-value")
	if result != "fallback-value" {
		t.Errorf("expected fallback, got %s", result)
	}
}
