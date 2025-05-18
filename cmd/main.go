package main

import (
	"os"

	"github.com/joho/godotenv"
	"ginx/internal/config"
	"ginx/pkg/logger"
)

func main() {
	// Initialize logger
	logger := logger.Default()
	logger.SetDefault()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found or error loading .env file", "error", err)
	}

	// Get config path from environment variable, default to "config/development.yaml"
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/development.yaml"
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("Failed to load config", "path", configPath, "error", err)
		os.Exit(1)
	}

	logger.Info("Config loaded successfully", 
		"port", cfg.Server.Port,
		"async_method", cfg.Server.AsyncMethod,
		"load_balancer", cfg.Server.LoadBalancer,
		"upstream_servers", cfg.Server.UpstreamServers,
	)

	// TODO: Initialize and start the server
}