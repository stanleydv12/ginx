//go:build linux

package linux

import (
	"github.com/stanleydv12/ginx/internal/socket"
	"github.com/stanleydv12/ginx/pkg/logger"

	"net"
	"fmt"

	"golang.org/x/sys/unix"
)

const (
	domain   = unix.AF_INET     // IPv4
	tcpType  = unix.SOCK_STREAM // TCP
	udpType  = unix.SOCK_DGRAM  // UDP
	protocol = 0                // Default protocol
)

type LinuxSocketManager struct{}

func NewLinuxSocketManager() socket.SocketManager {
	return &LinuxSocketManager{}
}

func (s *LinuxSocketManager) CreateSocket(options *socket.SocketOptions) (fd int, err error) {
	opts := options
	if options == nil {
		opts = &socket.SocketOptions{
			// Default values
			NonBlocking: true,
			ReuseAddr:   true,
			Type:        tcpType,
		}
	}

	fd, err = unix.Socket(domain, opts.Type, protocol)
	if err != nil {
		logger.Error("Failed to create socket", "error", err)
		return fd, err
	}

	defer func() {
		if err != nil {
			s.CloseSocket(fd)
		}
	}()

	if opts.NonBlocking {
		if err = unix.SetNonblock(fd, true); err != nil {
			logger.Error("Failed to set socket to non-blocking", "error", err)
			return fd, err
		}
	}

	if opts.ReuseAddr {
		if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			logger.Error("Failed to set socket to reuse address", "error", err)
			return fd, err
		}
	}

	return fd, nil
}

func (s *LinuxSocketManager) CloseSocket(fd int) error {
	if err := unix.Close(fd); err != nil {
		logger.Error("Failed to close socket",
			"fd", fd,
			"error", err,
		)
		return err
	}
	return nil
}

func (s *LinuxSocketManager) BindSocket(fd int, address string, port int) error {
	var ip net.IP
	if address == "" {
		ip = net.IPv4zero
	} else {
		ip = net.ParseIP(address)
		if ip == nil {
			return fmt.Errorf("Invalid IP address: %s", address)
		}
	}

	var addr [4]byte
	copy(addr[:], ip)

	if err := unix.Bind(fd, &unix.SockaddrInet4{
		Port: port,
		Addr: addr,
	}); err != nil {
		logger.Error("Failed to bind socket",
			"fd", fd,
			"error", err,
		)
		return err
	}

	return nil
}

func (s *LinuxSocketManager) StartListening(fd int) error {
	if err := unix.Listen(fd, 128); err != nil {
		logger.Error("Failed to listen on socket",
			"fd", fd,
			"error", err,
		)
		return err
	}
	return nil
}

func (s *LinuxSocketManager) AcceptConnection(fd int) (int, error) {
	connFd, _, err := unix.Accept(fd)
	if err != nil {
		return -1, err
	}
	return connFd, nil
}

func (s *LinuxSocketManager) ReadFromSocket(fd int, buf []byte) (int, error) {
	return unix.Read(fd, buf)
}

func (s *LinuxSocketManager) WriteToSocket(fd int, buf []byte) (int, error) {
	return unix.Write(fd, buf)
}
