//go:build linux

package main

import (
	"os"

	"github.com/stanleydv12/ginx/internal/config"
	"github.com/stanleydv12/ginx/internal/parser"
	"github.com/stanleydv12/ginx/internal/entity"
	"github.com/stanleydv12/ginx/internal/socket"
	"github.com/stanleydv12/ginx/internal/socket/linux"
	"github.com/stanleydv12/ginx/internal/async/epoll"
	"github.com/stanleydv12/ginx/pkg/logger"

	"golang.org/x/sys/unix"
)

type Server struct {
	listenFd    int
	config      config.ServerConfig
	socket      socket.SocketManager
	epoll       epoll.EpollHandler
	httpParser  parser.HTTPParser
}

func NewServer(config config.ServerConfig, socket socket.SocketManager, epoll epoll.EpollHandler, httpParser parser.HTTPParser) *Server {
	return &Server{
		config: config,
		socket: socket,
		epoll:  epoll,
		httpParser: httpParser,
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

	if err := s.epoll.Add(s.listenFd); err != nil {
		logger.Error("Failed to add socket to epoll", "error", err)
		return err
	}

	for {
		fds, err := s.epoll.Wait()
		if err != nil {
			logger.Error("Failed to wait for events", "error", err)
			return err
		}

		for _, fd := range fds {
			if fd == s.listenFd {
				connFd, err := s.socket.AcceptConnection(s.listenFd)
				if err != nil {
					if err == unix.EINTR {
						continue
					}
					logger.Error("Failed to accept connection", "error", err)
					return err
				}

				if connFd == -1 {
					continue
				}

				if err := s.epoll.Add(connFd); err != nil {
					logger.Error("Failed to add connection to epoll", "error", err)
					return err
				}
			} else {
				// TODO: Handle connection
				buf := make([]byte, 1024)
				n, err := s.socket.ReadFromSocket(fd, buf)
				if err != nil {
					if err == unix.EINTR {
						continue
					}
					logger.Error("Failed to read from socket", "error", err)
					return err
				}

				logger.Info("Read from socket", "fd", fd, "n", n, "data", string(buf[:n]))

				req, err := s.httpParser.ParseHTTPRequest(buf[:n])
				if err != nil {
					logger.Error("Failed to parse HTTP request", "error", err)
					return err
				}

				logger.Info("Parsed HTTP request",
					"method", req.Method,
					"path", req.Path,
					"protocol", req.Protocol,
					"headers", req.Headers,
					"body", string(req.Body),
					"raw", string(req.Raw),
				)

				resp := s.httpParser.RebuildResponse(&entity.HTTPResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "text/plain",
					},
					Body: []byte("Hello, World!"),
				})

				if _, err := s.socket.WriteToSocket(fd, resp); err != nil {
					logger.Error("Failed to write to socket", "error", err)
					return err
				}

				if err := s.epoll.Remove(fd); err != nil {
					logger.Error("Failed to remove socket from epoll", "error", err)
					return err
				}

				if err := s.socket.CloseSocket(fd); err != nil {
					logger.Error("Failed to close socket", "error", err)
					return err
				}
			}
		}
	}
}

func (s *Server) Stop() {
	if err := s.epoll.Remove(s.listenFd); err != nil {
		logger.Error("Failed to remove socket from epoll", "error", err)
	}

	if err := s.socket.CloseSocket(s.listenFd); err != nil {
		logger.Error("Failed to close socket", "error", err)
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

	// Initialize epoll
	ep := epoll.NewEpoll()

	// Initialize HTTP parser
	httpParser := parser.NewHTTPParser()

	// Initialize server
	server := NewServer(*cfg, socketManager, ep, httpParser)

	// Start server
	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	defer server.Stop()

	logger.Info("Server started successfully")
}
