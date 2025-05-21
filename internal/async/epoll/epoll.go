//go:build linux

package epoll

import (
	"sync"

	"golang.org/x/sys/unix"
)

type EpollHandler interface {
	Initialize() error
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]unix.EpollEvent, error)
	Close() error
}

type Epoll struct {
	epfd   int
	events []unix.EpollEvent
	mu     sync.Mutex
}

var defaultEpoll *Epoll

func NewEpoll() EpollHandler {
	e := &Epoll{}
	e.events = make([]unix.EpollEvent, 1024)

	if err := e.Initialize(); err != nil {
		return nil
	}

	defaultEpoll = e
	return e
}

func (e *Epoll) Initialize() error {
	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		return err
	}
	e.epfd = epfd
	return nil
}

func (e *Epoll) Add(fd int) error {
	return unix.EpollCtl(e.epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(fd),
	})
}

func (e *Epoll) Remove(fd int) error {
	return unix.EpollCtl(e.epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func (e *Epoll) Wait() ([]unix.EpollEvent, error) {
	n, err := unix.EpollWait(e.epfd, e.events, -1)
	if err != nil {
		return nil, err
	}

	return e.events[:n], nil
}

func (e *Epoll) Close() error {
	return unix.Close(e.epfd)
}
