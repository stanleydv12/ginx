//go:build linux

package main

import (
	"os"

	"github.com/stanleydv12/ginx/internal/config"
	"github.com/stanleydv12/ginx/internal/parser"
	"github.com/stanleydv12/ginx/internal/socket/linux"
	"github.com/stanleydv12/ginx/internal/async/epoll"
	"github.com/stanleydv12/ginx/internal/server"
	"github.com/stanleydv12/ginx/internal/loadbalancer"
	"github.com/stanleydv12/ginx/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Default()
	logger.Info("Starting ginx proxy server")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Config loaded successfully",
		"port", cfg.Server.Port,
		"async_method", cfg.Server.AsyncMethod,
		"load_balancer", cfg.Server.LoadBalancer,
		"upstream_servers", cfg.Server.UpstreamServers,
	)

	// Initialize socket manager
	socketManager := linux.NewLinuxSocketManager()

	// Initialize epoll
	ep := epoll.NewEpoll()

	// Initialize HTTP parser
	httpParser := parser.NewHTTPParser()

	// Initialize load balancer
	loadBalancer, err := loadbalancer.NewLoadBalancer(cfg)
	if err != nil {
		logger.Error("Failed to initialize load balancer", "error", err)
		os.Exit(1)
	}

	// Initialize server
	server := server.NewServer(*cfg, socketManager, ep, httpParser, loadBalancer)

	logger.Info("Server initialized successfully", "server", &server)

	// Start server
	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	defer server.Stop()

	logger.Info("Server started successfully")
}
