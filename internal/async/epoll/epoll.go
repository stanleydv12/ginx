//go:build linux

package epoll

import (
	"sync"

	"golang.org/x/sys/unix"
)

type EpollHandler interface {
	Initialize() error
	Add(fd int, events uint32) error
	Remove(fd int) error
	Modify(fd int, events uint32) error
	Wait() ([]unix.EpollEvent, error)
	Close() error
}

type Epoll struct {
	epfd   int
	events []unix.EpollEvent
	mu     sync.Mutex
}

var defaultEpoll *Epoll

func NewEpoll(maxEvents int) EpollHandler {
	e := &Epoll{}
	e.events = make([]unix.EpollEvent, maxEvents)

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

func (e *Epoll) Add(fd int, events uint32) error {
	return unix.EpollCtl(e.epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

func (e *Epoll) Remove(fd int) error {
	return unix.EpollCtl(e.epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func (e *Epoll) Modify(fd int, events uint32) error {
	return unix.EpollCtl(e.epfd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

func (e *Epoll) Wait() ([]unix.EpollEvent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for {
		n, err := unix.EpollWait(e.epfd, e.events, -1)
		if err != nil {
			// If the system call was interrupted, retry
			if err == unix.EINTR {
				continue
			}
			return nil, err
		}
		return e.events[:n], nil
	}
}

func (e *Epoll) Close() error {
	return unix.Close(e.epfd)
}
