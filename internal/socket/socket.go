//go:build linux
package socket

type SocketOptions struct {
	NonBlocking bool
	ReuseAddr   bool
	Type        int
}

type SocketManager interface {
	CreateSocket(options *SocketOptions) (fd int, err error)
	CloseSocket(fd int) error
	BindSocket(fd int, address string, port int) error
	StartListening(fd int) error
	AcceptConnection(fd int) (int, error)
	ReadFromSocket(fd int, buf []byte) (int, error)
	WriteToSocket(fd int, buf []byte) (int, error)
}