package main

import (
	"log"

	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
)

func main() {
	cfg := config.Load()

	log.Printf("policy-crawler starting... LLM endpoint: %s", cfg.LLMApiEndpoint)

	go startScheduleLoop(cfg)
	startAdminAPI(cfg)
}

func startScheduleLoop(cfg *config.Config) {
	// Asynq scheduler — implemented in Sprint 3-4
}

func startAdminAPI(cfg *config.Config) {
	// Admin HTTP API for manual review — implemented in Sprint 3-4
}
