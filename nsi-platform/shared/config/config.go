package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	LLMApiKey        string
	LLMApiEndpoint   string
	ActuaryGRPCAddr  string
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	ServerPort       int
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres123@localhost:35432/nsi?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:36379/0"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		LLMApiKey:        getEnv("LLM_API_KEY", ""),
		LLMApiEndpoint:   getEnv("LLM_API_ENDPOINT", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"),
		ActuaryGRPCAddr:  getEnv("ACTUARY_GRPC_ADDR", "localhost:50051"),
		StorageEndpoint:  getEnv("STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageAccessKey: getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
		StorageBucket:    getEnv("STORAGE_BUCKET", "nsi-reports"),
		ServerPort:       getEnvInt("SERVER_PORT", 30001),
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
