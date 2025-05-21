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

	"sync"
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
	connections  map[int]*connection.Connection
	mu           sync.Mutex
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
			logger.Error("Failed to wait for events", "error", err)
			return err
		}

		for _, event := range events {
			fd := int(event.Fd)
			eventType := event.Events

			if fd == s.listenFd {
				if err := s.handleNewConnection(); err != nil {
					return err
				}
			} else {
				if eventType&unix.EPOLLERR != 0{
					logger.Error("EPOLL error", "fd", fd)
					if err := s.socket.CheckSocketState(fd); err != nil {
						logger.Error("Failed to check socket state", "error", err)
					}
					s.cleanupConnection(fd)
					continue
				}

				if eventType&unix.EPOLLHUP != 0 {
					logger.Error("EPOLL hangup", "fd", fd)
					if err := s.socket.CheckSocketState(fd); err != nil {
						logger.Error("Failed to check socket state", "error", err)
					}
					s.cleanupConnection(fd)
					continue
				}

				switch s.connections[fd].State {
				case connection.StateClientAccepted:
					if eventType&unix.EPOLLIN != 0 {
						if err := s.handleClientRequest(fd); err != nil {
							return err
						}

						if err := s.handleConnectUpstream(fd); err != nil {
							return err
						}
						continue
					}
				case connection.StateConnectingUpstream:
					if eventType&unix.EPOLLOUT != 0 {
						if err := s.handleForwardUpstream(fd); err != nil {
							return err
						}
						continue
					}
				case connection.StateForwardingRequest:
					if eventType&unix.EPOLLIN != 0 {
						if err := s.handleUpstreamResponse(fd); err != nil {
							return err
						}
						s.cleanupConnection(fd)
						continue
					}
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

func (s *Server) handleNewConnection() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	connFd, err := s.socket.AcceptConnection(s.listenFd)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
		logger.Error("Failed to accept connection", "error", err)
		return err
	}

	if connFd == -1 {
		return nil
	}

	if err := s.epoll.Add(connFd, unix.EPOLLIN); err != nil {
		logger.Error("Failed to add connection to epoll", "error", err)
		return err
	}

	s.connections[connFd] = &connection.Connection{
		ClientFD: connFd,
		State:    connection.StateClientAccepted,
	}

	return nil
}

func (s *Server) handleClientRequest(clientFd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Info("handleClientRequest", "fd", clientFd)

	buf := make([]byte, 4096)

	n, err := s.socket.ReadFromSocket(clientFd, buf)
	if err != nil {
		if err == unix.EINTR {
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

	s.connections[clientFd].ClientAddress = req.Headers["Host"]
	s.connections[clientFd].Request = req
	s.connections[clientFd].State = connection.StateRequestReceived

	logger.Info("Client read", "fd", clientFd, "connection", s.connections[clientFd])

	return nil
}

func (s *Server) handleConnectUpstream(clientFd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Info("handleConnectUpstream", "fd", clientFd)

	upstreamServer, err := s.loadBalancer.SelectServer()
	if err != nil {
		logger.Error("Failed to select upstream server", "error", err)
		return err
	}

	logger.Info("Selected upstream server", "url", upstreamServer.URL.Host)

	// Change headers
	s.connections[clientFd].Request.Headers["Host"] = upstreamServer.URL.Host

	address := upstreamServer.URL.Hostname()
	port, _ := strconv.Atoi(upstreamServer.URL.Port())

	upstreamFd, err := s.socket.ConnectToSocket(address, port)
	if err != nil {
		logger.Error("Failed to connect to upstream server", "error", err)
		return err
	}

	if err := s.epoll.Add(upstreamFd, unix.EPOLLOUT); err != nil {
		logger.Error("Failed to add upstream server to epoll", "error", err)
		return err
	}

	s.connections[clientFd].UpstreamFD = upstreamFd
	s.connections[clientFd].UpstreamServer = upstreamServer
	s.connections[clientFd].State = connection.StateConnectingUpstream

	s.connections[upstreamFd] = s.connections[clientFd]

	return nil
}

func (s *Server) handleForwardUpstream(upstreamFd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Info("handleForwardUpstream", "fd", upstreamFd)

	request := s.connections[upstreamFd].Request

	// Change header
	request.Headers["Host"] = s.connections[upstreamFd].UpstreamServer.URL.Host

	_, err := s.socket.WriteToSocket(upstreamFd, request.Raw)
	if err != nil {
		logger.Error("Failed to write to upstream server", "error", err)
		return err
	}

	if err := s.epoll.Modify(upstreamFd, unix.EPOLLIN); err != nil {
		logger.Error("Failed to modify upstream server to epoll", "error", err)
		return err
	}

	s.connections[upstreamFd].State = connection.StateForwardingRequest

	clientFd := s.connections[upstreamFd].ClientFD
	s.connections[clientFd].State = connection.StateForwardingRequest

	return nil
}

func (s *Server) handleUpstreamResponse(upstreamFd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := make([]byte, 4096)

	n, err := s.socket.ReadFromSocket(upstreamFd, buf)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
		logger.Error("Failed to read from socket", "error", err)
		return err
	}

	logger.Info("handleUpstreamResponse", "fd", upstreamFd)

	response, err := s.httpParser.ParseHTTPResponse(buf[:n])
	if err != nil {
		logger.Error("Failed to parse HTTP response", "error", err)
		return err
	}

	s.connections[upstreamFd].Response = response
	s.connections[upstreamFd].State = connection.StateWaitingResponse

	clientFd := s.connections[upstreamFd].ClientFD
	s.connections[clientFd].State = connection.StateWaitingResponse

	// Modify Response Headers
	response.Headers["Server"] = "ginx"
	response.Headers["X-Forwarded-For"] = s.connections[upstreamFd].ClientAddress
	response.Headers["X-Forwarded-Proto"] = "http"
	response.Headers["Via"] = "ginx/1.0"
	response.Headers["Connection"] = "close"
	response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))

	response.Raw = s.httpParser.RebuildResponse(response)

	_, err = s.socket.WriteToSocket(clientFd, response.Raw)
	if err != nil {
		logger.Error("Failed to write to client", "error", err)
		return err
	}

	logger.Info("Finished handling upstream response", "fd", upstreamFd, "response", response)

	s.connections[upstreamFd].State = connection.StateCompleted
	s.connections[clientFd].State = connection.StateCompleted

	return nil
}

func (s *Server) cleanupConnection(fd int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Info("cleanupConnection", "fd", fd)

	conn, exists := s.connections[fd]
	if !exists {
		logger.Error("Connection not found", "fd", fd)
		return
	}

	// Clean up the other end of the connection
	var otherFd int
	if conn.ClientFD == fd {
		otherFd = conn.UpstreamFD
	} else {
		otherFd = conn.ClientFD
	}

	// Close both FDs
	s.socket.CloseSocket(fd)
	if otherFd != 0 {
		s.socket.CloseSocket(otherFd)
	}

	// Remove from epoll
	s.epoll.Remove(fd)
	if otherFd != 0 {
		s.epoll.Remove(otherFd)
	}

	// Clean up map entries
	delete(s.connections, fd)
	if otherFd != 0 {
		delete(s.connections, otherFd)
	}
}
