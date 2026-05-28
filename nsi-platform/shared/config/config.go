package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	LLMApiKey           string
	LLMApiEndpoint      string
	ActuaryHTTPAddr     string
	LLMGatewayURL       string
	StorageEndpoint     string
	StorageAccessKey    string
	StorageSecretKey    string
	StorageBucket       string
	ServerPort          int
	AllowedOrigins      string
	WebClientAPIBaseURL string
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://localhost:39432/nsi?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:39479/0"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		ActuaryHTTPAddr:     getEnv("ACTUARY_HTTP_ADDR", "localhost:39402"),
		LLMGatewayURL:       getEnv("LLM_GATEWAY_URL", "http://localhost:39404"),
		StorageEndpoint:  getEnv("STORAGE_ENDPOINT", "http://localhost:39490"),
		ServerPort:       getEnvInt("SERVER_PORT", 39401),
		WebClientAPIBaseURL: getEnv("WEBCLIENT_API_BASE_URL", "http://127.0.0.1:39401"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
