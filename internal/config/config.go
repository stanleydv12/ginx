package config

import (
	"errors"
	"os"
	"path/filepath"

	"ginx/pkg/logger"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Server struct {
		Address         string   `yaml:"address"`
		Port            int      `yaml:"port"`
		AsyncMethod     string   `yaml:"async_method"`
		LoadBalancer    string   `yaml:"load_balancer"`
		UpstreamServers []string `yaml:"upstream_servers"`
	} `yaml:"server"`
}

func LoadConfig() (*ServerConfig, error) {

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found or error loading .env file", "error", err)
	}

	// Get config path from environment variable, default to "config/development.yaml"
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/development.yaml"
	}

	// Convert to absolute path if it's not already
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get working directory", "error", err)
			return nil, err
		}
		configPath = filepath.Join(wd, configPath)
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Failed to read config file", "path", configPath, "error", err)
		return nil, err
	}

	// Parse the YAML into our config struct
	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.Error("Failed to parse config file", "path", configPath, "error", err)
		return nil, err
	}

	// Validate required fields
	if cfg.Server.Port == 0 {
		logger.Error("server.port is required")
		return nil, errors.New("server.port is required")
	}
	if len(cfg.Server.UpstreamServers) == 0 {
		logger.Error("at least one upstream server is required")
		return nil, errors.New("at least one upstream server is required")
	}

	return &cfg, nil
}
