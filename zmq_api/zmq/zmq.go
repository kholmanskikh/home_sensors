package zmq

// #cgo LDFLAGS: -lzmq
// #include <zmq.h>
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

type socketType int

const (
	SocketPUB socketType = iota
	SocketSUB
)

func (tp socketType) String() string {
	return [...]string{"PUB", "SUB"}[tp]
}

type Context struct {
	ctx unsafe.Pointer
}

func NewContext() (*Context, error) {
	var p unsafe.Pointer
	var err error

	for {
		p, err = C.zmq_ctx_new()

		if (p != nil) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if p == nil {
		return nil, err
	}

	return &Context{ctx: p}, nil
}

func (ctx *Context) Terminate() error {
	var err error
	var rv C.int

	for {
		rv, err = C.zmq_ctx_term(ctx.ctx)

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return err
	}

	return nil
}

type Socket struct {
	sock unsafe.Pointer
}

func NewSocket(ctx *Context, sockType socketType) (*Socket, error) {
	var tp C.int

	switch sockType {
	case SocketPUB:
		tp = C.ZMQ_PUB
	case SocketSUB:
		tp = C.ZMQ_SUB
	default:
		return nil, fmt.Errorf("unsupported socket type %d", sockType)
	}

	var p unsafe.Pointer
	var err error

	for {
		p, err = C.zmq_socket(ctx.ctx, tp)

		if (p != nil) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if p == nil {
		return nil, err
	}

	return &Socket{sock: p}, nil
}

func (sock *Socket) Close() error {
	var err error
	var rv C.int

	for {
		rv, err = C.zmq_close(sock.sock)

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) GetFd() (int, error) {
	var fd C.int
	l := C.size_t(unsafe.Sizeof(fd))

	var err error
	var rv C.int

	for {
		rv, err = C.zmq_getsockopt(sock.sock, C.ZMQ_FD, unsafe.Pointer(&fd), &l)

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv == -1 {
		return -1, err
	}

	return int(fd), nil
}

func (sock *Socket) Bind(endpoint string) error {
	var err error
	var rv C.int

	for {
		rv, err = C.zmq_bind(sock.sock, C.CString(endpoint))

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) Unbind(endpoint string) error {
	var err error
	var rv C.int

	for {
		rv, err = C.zmq_unbind(sock.sock, C.CString(endpoint))

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) Connect(endpoint string) error {
	var err error
	var rv C.int

	for {
		rv, err = C.zmq_connect(sock.sock, C.CString(endpoint))

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) doRecv(p []byte, flags C.int) (int, error) {
	if (p == nil) || (len(p) == 0) {
		return 0, nil
	}

	var err error
	var rv C.int

	for {
		rv, err = C.zmq_recv(sock.sock, unsafe.Pointer(&p[0]), C.size_t(len(p)), flags)

		if (rv != -1) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv == -1 {
		return 0, err
	}

	readLen := int(rv)
	if readLen > len(p) {
		return readLen, fmt.Errorf("received data was truncated")
	}

	_, err = sock.updateEventsState()
	if err != nil {
		return 0, fmt.Errorf("updateEventsStatus(): %v", err)
	}

	return readLen, nil
}

func (sock *Socket) Recv(p []byte) (int, error) {
	return sock.doRecv(p, 0)
}

func (sock *Socket) RecvNonBlocking(p []byte) (int, error, bool) {
	readLen, err := sock.doRecv(p, C.ZMQ_DONTWAIT)

	if errors.Is(err, unix.EAGAIN) {
		return 0, nil, false
	}

	return readLen, err, true
}

func (sock *Socket) Send(p []byte) error {
	if (p == nil) || (len(p) == 0) {
		return nil
	}

	var err error
	var rv C.int

	for {
		rv, err = C.zmq_send(sock.sock, unsafe.Pointer(&p[0]), C.size_t(len(p)), 0)

		if (rv != -1) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv == -1 {
		return err
	}

	_, err = sock.updateEventsState()
	if err != nil {
		return fmt.Errorf("updateEventsStatus(): %v", err)
	}

	return nil
}

func (sock *Socket) AddSubscribeFilter(prefix []byte) error {
	var p unsafe.Pointer
	var l C.size_t

	if (prefix == nil) || (len(prefix) == 0) {
		p = nil
		l = 0
	} else {
		p = unsafe.Pointer(&prefix[0])
		l = C.size_t(len(prefix))
	}

	rv, err := C.zmq_setsockopt(sock.sock, C.ZMQ_SUBSCRIBE, p, l)

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) RemoveSubscribeFilter(prefix []byte) error {
	var p unsafe.Pointer
	var l C.size_t

	if (prefix == nil) || (len(prefix) == 0) {
		p = nil
		l = 0
	} else {
		p = unsafe.Pointer(&prefix[0])
		l = C.size_t(len(prefix))
	}

	rv, err := C.zmq_setsockopt(sock.sock, C.ZMQ_UNSUBSCRIBE, p, l)

	if rv != 0 {
		return err
	}

	return nil
}

func (sock *Socket) GetLastEndpoint() (string, error) {
	buf := make([]byte, 512)
	l := C.size_t(len(buf))

	// if the buf is too small, this call should fail with C.EINVAL
	rv, err := C.zmq_getsockopt(sock.sock, C.ZMQ_LAST_ENDPOINT, unsafe.Pointer(&buf[0]), &l)

	if rv != 0 {
		return "", err
	}

	if l <= 1 {
		return "", fmt.Errorf("empty endpoint?")
	}

	return string(buf[:l-1]), nil
}

func (sock *Socket) updateEventsState() (C.int, error) {
	var bitmask C.int
	l := C.size_t(unsafe.Sizeof(bitmask))

	var err error
	var rv C.int

	for {
		rv, err = C.zmq_getsockopt(sock.sock, C.ZMQ_EVENTS, unsafe.Pointer(&bitmask), &l)

		if (rv == 0) || (!errors.Is(err, unix.EINTR)) {
			break
		}
	}

	if rv != 0 {
		return 0, err
	}

	return bitmask, nil
}

func (sock *Socket) IsUnblockedForRecv() (bool, error) {
	bitmask, err := sock.updateEventsState()
	if err != nil {
		return false, fmt.Errorf("updateEventsState() failed: %v", err)
	}

	return bitmask&C.ZMQ_POLLIN != 0, nil
}
