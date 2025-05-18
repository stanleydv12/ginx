package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Server struct {
		Port            int      `yaml:"port"`
		AsyncMethod     string   `yaml:"async_method"`
		LoadBalancer    string   `yaml:"load_balancer"`
		UpstreamServers []string `yaml:"upstream_servers"`
	} `yaml:"server"`
}

func LoadConfig(path string) (*ServerConfig, error) {
	// Convert to absolute path if it's not already
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(wd, path)
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the YAML into our config struct
	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Validate required fields
	if cfg.Server.Port == 0 {
		return nil, errors.New("server.port is required")
	}
	if len(cfg.Server.UpstreamServers) == 0 {
		return nil, errors.New("at least one upstream server is required")
	}

	return &cfg, nil
}