//go:build linux

package main

import (
	"os"

	"ginx/internal/config"
	"ginx/internal/socket"
	"ginx/internal/socket/linux"
	"ginx/pkg/logger"
)

type Server struct {
	listenFd    int
	config      *config.ServerConfig
	socket      socket.SocketManager
}

func NewServer(config *config.ServerConfig, socket socket.SocketManager) *Server {
	return &Server{
		config: config,
		socket: socket,
	}
}

func (s *Server) Start() error {
	// Test open and close socket
	fd, err := s.socket.CreateSocket(nil)
	if err != nil {
		logger.Error("Failed to create socket", "error", err)
		return err
	}

	if err := s.socket.BindSocket(fd, s.config.Server.Address, s.config.Server.Port); err != nil {
		logger.Error("Failed to bind socket", "error", err)
		return err
	}

	if err := s.socket.StartListening(fd); err != nil {
		logger.Error("Failed to listen on socket", "error", err)
		return err
	}
	s.listenFd = fd

	for {
		connFd, err := s.socket.AcceptConnection(s.listenFd)
		if err != nil {
			continue
		}

		if connFd == -1 {
			continue
		}

		logger.Info("Accepted connection", "fd", connFd)

		go func(connFd int) {
			defer s.socket.CloseSocket(connFd)

			// TODO: Handle connection
		}(connFd)
	}
}

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

	// Initialize server
	server := NewServer(cfg, socketManager)

	// Start server
	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	defer server.socket.CloseSocket(server.listenFd)

	logger.Info("Server started successfully")
}
