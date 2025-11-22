package main

import (
	"log"
	"time"

	"scaler/pkg/config"
	"scaler/pkg/engine"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: could not load .env file: %v\n", err)
	}

	cfg := config.LoadConfig()

	if cfg.ServerManagerAPI == "" || len(cfg.AvailableAgents) == 0 {
		log.Fatal("Server manager api or agents not provided")
	}

	scalerEngine := engine.NewScalerEngine(cfg)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		scalerEngine.EvaluateScaling()
	}
}
