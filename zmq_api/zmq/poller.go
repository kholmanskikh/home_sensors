package zmq

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

type ReadPoller struct {
	fdToSock map[int]*Socket

	pollFds []unix.PollFd
}

func NewReadPoller(sockets ...*Socket) (*ReadPoller, error) {
	if (sockets == nil) || (len(sockets) == 0) {
		return nil, fmt.Errorf("empty socket list")
	}

	poller := ReadPoller{}
	poller.fdToSock = make(map[int]*Socket)

	for _, sock := range sockets {
		fd, err := sock.GetFd()
		if err != nil {
			return nil, fmt.Errorf("failed to get file descriptor for socket %v", sock)
		}

		poller.fdToSock[fd] = sock

		poller.pollFds = append(poller.pollFds, unix.PollFd{Fd: int32(fd), Events: unix.POLLIN})
	}

	return &poller, nil
}

func (p *ReadPoller) Poll(timeout time.Duration) ([]*Socket, error) {
	ret := make([]*Socket, 0)

	var n int
	var err error

	for {
		n, err = unix.Poll(p.pollFds, int(timeout.Milliseconds()))
		if !errors.Is(err, unix.EINTR) {
			break
		}
	}

	if n == -1 {
		return nil, fmt.Errorf("poll() failed with: %v", err)
	}

	processed := 0
	for _, pollFd := range p.pollFds {
		if processed == n {
			break
		}

		if pollFd.Revents&unix.POLLIN != 0 {
			sock := p.fdToSock[int(pollFd.Fd)]

			unblocked, err := sock.IsUnblockedForRecv()
			if err != nil {
				return ret, fmt.Errorf("IsUnblockedForRecv(%v) failed: %v", sock, err)
			}

			if unblocked {
				ret = append(ret, sock)
			}

			processed += 1
		}
	}

	return ret, nil
}
