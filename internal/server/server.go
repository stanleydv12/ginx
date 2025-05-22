//go:build linux

package server

import (
	"github.com/stanleydv12/ginx/internal/async/epoll"
	"github.com/stanleydv12/ginx/internal/config"
	"github.com/stanleydv12/ginx/internal/connection"
	"github.com/stanleydv12/ginx/internal/loadbalancer"
	"github.com/stanleydv12/ginx/internal/parser"
	"github.com/stanleydv12/ginx/internal/socket"
	"github.com/stanleydv12/ginx/pkg/logger"

	"fmt"
	"strconv"
	"golang.org/x/sys/unix"
)

type Server struct {
	listenFd     int
	config       config.ServerConfig
	socket       socket.SocketManager
	epoll        epoll.EpollHandler
	httpParser   parser.HTTPParser
	loadBalancer loadbalancer.LoadBalancerHandler
	connections map[int]*connection.Connection
}

func NewServer(config config.ServerConfig, socket socket.SocketManager, epoll epoll.EpollHandler, httpParser parser.HTTPParser, loadBalancer loadbalancer.LoadBalancerHandler) *Server {
	return &Server{
		config:       config,
		socket:       socket,
		epoll:        epoll,
		httpParser:   httpParser,
		loadBalancer: loadBalancer,
		connections:  make(map[int]*connection.Connection),
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

	logger.Info("Server started and listening", "address", fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port), "fd", fd)
	if err := s.socket.StartListening(fd); err != nil {
		logger.Error("Failed to listen on socket", "error", err)
		return err
	}
	s.listenFd = fd

	if err := s.epoll.Add(s.listenFd, unix.EPOLLIN); err != nil {
		logger.Error("Failed to add socket to epoll", "error", err)
		return err
	}

	for {
		events, err := s.epoll.Wait()
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			logger.Error("Failed to wait for events", "error", err)
			return err
		}

		for _, event := range events {
			s.handleEvent(event)
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

func (s *Server) handleEvent(event unix.EpollEvent) {
	fd := int(event.Fd)
	eventType := event.Events

	if fd == s.listenFd {
		if eventType&unix.EPOLLIN != 0 {
			if err := s.handleNewConnection(); err != nil {
				logger.Error("Error accepting new client connection", "error", err, "fd", fd)
			}
			return
		}
	}

	conn, exists := s.connections[fd]
	if !exists {
		logger.Error("Connection not found while handling event", "fd", fd, "event_type", eventType)
		return
	}

	if eventType&unix.EPOLLERR != 0 {
		logger.Error("Socket error detected by epoll", "fd", fd, "event_type", "EPOLLERR")
		if err := s.socket.CheckSocketState(fd); err != nil {
			logger.Error("Failed to check socket state", "error", err)
		}
		s.cleanupConnection(fd)
		return
	}

	if eventType&unix.EPOLLHUP != 0 {
		logger.Error("Connection hangup detected by epoll", "fd", fd, "event_type", "EPOLLHUP")
		if err := s.socket.CheckSocketState(fd); err != nil {
			logger.Error("Failed to check socket state", "error", err)
		}
		s.cleanupConnection(fd)
		return
	}

	switch conn.State {
	case connection.StateClientAccepted:
		if eventType&unix.EPOLLIN != 0 {
			if err := s.handleClientRequest(fd); err != nil {
				logger.Error("Failed to handle client request", "fd", fd, "error", err)
				s.cleanupConnection(fd)
			}
			if err := s.handleConnectUpstream(fd); err != nil {
				logger.Error("Failed to handle connect upstream", "fd", fd, "error", err)
				s.cleanupConnection(fd)
			}
		}
	case connection.StateConnectingUpstream:
		if eventType&unix.EPOLLOUT != 0 {
			if err := s.handleForwardUpstream(fd); err != nil {
				logger.Error("Failed to handle forward upstream", "error", err)
				s.cleanupConnection(fd)
			}
		}
	case connection.StateForwardingRequest:
		if eventType&unix.EPOLLIN != 0 {
			if err := s.handleUpstreamResponse(fd); err != nil {
				logger.Error("Failed to handle upstream response", "error", err)
				s.cleanupConnection(fd)
			}
			s.cleanupConnection(fd)
		}
	}
}

func (s *Server) handleNewConnection() error {
	connFd, err := s.socket.AcceptConnection(s.listenFd)
	if err != nil {
		if err == unix.EINTR || err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		logger.Error("Failed to accept connection", "error", err)
		return err
	}

	if connFd == -1 {
		return nil
	}

	if err := s.epoll.Add(connFd, unix.EPOLLIN|unix.EPOLLET); err != nil {
		logger.Error("Failed to add connection to epoll", "fd", connFd, "error", err)
		s.socket.CloseSocket(connFd)
		return nil
	}

	s.connections[connFd] = &connection.Connection{
		ClientFD: connFd,
		State:    connection.StateClientAccepted,
	}

	logger.Info("New connection accepted", "fd", connFd)
	return nil
}

func (s *Server) handleClientRequest(clientFd int) error {
	logger.Debug("Processing client request", "client_fd", clientFd)

	conn, exists := s.connections[clientFd]

	if !exists {
		return fmt.Errorf("connection not found for fd %d", clientFd)
	}

	buf := make([]byte, 4096)

	n, err := s.socket.ReadFromSocket(clientFd, buf)
	if err != nil {
		if err == unix.EINTR {
			logger.Info("handleClientRequest: EINTR", "fd", clientFd)
			return nil
		}
		logger.Error("Failed to read from socket", "error", err)
		return err
	}

	req, err := s.httpParser.ParseHTTPRequest(buf[:n])
	if err != nil {
		logger.Error("Failed to parse HTTP request", "error", err)
		return err
	}

	conn.ClientAddress = req.Headers["Host"]
	conn.Request = req
	conn.State = connection.StateRequestReceived
	s.connections[clientFd] = conn

	logger.Info("HTTP request received", "client_fd", clientFd, "method", req.Method, "path", req.Path, "host", req.Headers["Host"])

	return nil
}

func (s *Server) handleConnectUpstream(clientFd int) error {
	conn, exists := s.connections[clientFd]

	if !exists {
		return fmt.Errorf("connection not found for fd %d", clientFd)
	}

	logger.Debug("Initiating upstream connection", "client_fd", clientFd)

	upstreamServer, err := s.loadBalancer.SelectServer()
	if err != nil {
		logger.Error("Failed to select upstream server", "error", err)
		return err
	}

	logger.Debug("Selected upstream server for request", "client_fd", clientFd, "upstream_host", upstreamServer.URL.Host)

	// Change headers
	conn.Request.Headers["Host"] = upstreamServer.URL.Host

	address := upstreamServer.URL.Hostname()
	port, _ := strconv.Atoi(upstreamServer.URL.Port())

	upstreamFd, err := s.socket.ConnectToSocket(address, port)
	if err != nil {
		logger.Error("Failed to connect to upstream server", "error", err)
		return err
	}

	if err := s.epoll.Add(upstreamFd, unix.EPOLLOUT|unix.EPOLLET); err != nil {
		logger.Error("Failed to add upstream server to epoll", "error", err)
		return err
	}

	conn.UpstreamFD = upstreamFd
	conn.UpstreamServer = upstreamServer
	conn.State = connection.StateConnectingUpstream
	s.connections[upstreamFd] = conn

	return nil
}

func (s *Server) handleForwardUpstream(upstreamFd int) error {
	conn, exists := s.connections[upstreamFd]

	if !exists {
		return fmt.Errorf("connection not found for fd %d", upstreamFd)
	}

	logger.Debug("Forwarding request to upstream", "client_fd", conn.ClientFD, "upstream_fd", upstreamFd, "upstream_host", conn.UpstreamServer.URL.Host)

	request := conn.Request

	// Change header
	request.Headers["Host"] = conn.UpstreamServer.URL.Host

	_, err := s.socket.WriteToSocket(upstreamFd, request.Raw)
	if err != nil {
		logger.Error("Failed to write to upstream server", "error", err)
		return err
	}

	if err := s.epoll.Modify(upstreamFd, unix.EPOLLIN); err != nil {
		logger.Error("Failed to modify upstream server to epoll", "error", err)
		return err
	}

	conn.State = connection.StateForwardingRequest
	if s.connections[conn.ClientFD] != nil {
		s.connections[conn.ClientFD].State = connection.StateForwardingRequest
	}

	return nil
}

func (s *Server) handleUpstreamResponse(upstreamFd int) error {
	conn, exists := s.connections[upstreamFd]

	if !exists {
		return fmt.Errorf("connection not found for fd %d", upstreamFd)
	}

	buf := make([]byte, 4096)

	n, err := s.socket.ReadFromSocket(upstreamFd, buf)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
		logger.Error("Failed to read from socket", "error", err)
		return err
	}

	logger.Debug("Received response from upstream", "upstream_fd", upstreamFd, "client_fd", conn.ClientFD)

	response, err := s.httpParser.ParseHTTPResponse(buf[:n])
	if err != nil {
		logger.Error("Failed to parse HTTP response", "error", err)
		return err
	}

	conn.Response = response
	conn.State = connection.StateWaitingResponse
	if s.connections[conn.ClientFD] != nil {
		s.connections[conn.ClientFD].State = connection.StateWaitingResponse
	}

	// Modify Response Headers
	response.Headers["Server"] = "ginx"
	response.Headers["X-Forwarded-For"] = conn.ClientAddress
	response.Headers["X-Forwarded-Proto"] = "http"
	response.Headers["Via"] = "ginx/1.0"
	response.Headers["Connection"] = "close"
	response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))

	response.Raw = s.httpParser.RebuildResponse(response)

	_, err = s.socket.WriteToSocket(conn.ClientFD, response.Raw)
	if err != nil {
		logger.Error("Failed to write to client", "error", err)
		return err
	}

	logger.Info("Request completed", "client_fd", conn.ClientFD, "status_code", response.StatusCode, "content_length", len(response.Body))

	conn.State = connection.StateCompleted
	if s.connections[conn.ClientFD] != nil {
		s.connections[conn.ClientFD].State = connection.StateCompleted
	}

	return nil
}

func (s *Server) cleanupConnection(fd int) {
    conn, exists := s.connections[fd]
    if !exists {
        logger.Error("Connection not found in cleanupConnection", "fd", fd)
        return
    }

    if conn.Closed {
        return // Already cleaning up!
    }
    conn.Closed = true

    // Now remove both sides from the map and close both fds
    delete(s.connections, conn.ClientFD)
    if conn.UpstreamFD != 0 {
        delete(s.connections, conn.UpstreamFD)
    }

    s.epoll.Remove(conn.ClientFD)
    s.socket.CloseSocket(conn.ClientFD)
    if conn.UpstreamFD != 0 {
        s.epoll.Remove(conn.UpstreamFD)
        s.socket.CloseSocket(conn.UpstreamFD)
    }
    logger.Info("Connection terminated", "client_fd", conn.ClientFD, "upstream_fd", conn.UpstreamFD)
}
