package zmq

import (
	"fmt"
	"testing"
	"time"
)

func createContextOrFail(t *testing.T) *Context {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	return ctx
}

func terminateContextOrFail(t *testing.T, ctx *Context) {
	err := ctx.Terminate()
	if err != nil {
		t.Fatalf("ctx.Terminate() failed: %v", err)
	}
}

func createSocketOrFail(t *testing.T, ctx *Context, tp socketType) *Socket {
	sock, err := NewSocket(ctx, tp)
	if err != nil {
		t.Fatalf("creating of a '%s' socket failed: %v", tp, err)
	}

	return sock
}

func closeSocketOrFail(t *testing.T, sock *Socket) {
	err := sock.Close()
	if err != nil {
		t.Fatalf("closing the socket failed: %v", err)
	}
}

func TestSocketTypeString(t *testing.T) {
	compareStrings := func(t *testing.T, got, expected string) {
		if got != expected {
			t.Fatalf("Got '%s', expected '%s'", got, expected)
		}
	}

	compareStrings(t, fmt.Sprint(SocketPUB), "PUB")
	compareStrings(t, fmt.Sprint(SocketSUB), "SUB")
}

func TestContext(t *testing.T) {
	ctx := createContextOrFail(t)

	terminateContextOrFail(t, ctx)
}

func TestSocketTypes(t *testing.T) {
	const endpoint = "ipc://*"

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	types := [...]socketType{SocketPUB, SocketSUB}

	for _, tp := range types {
		sock, err := NewSocket(ctx, tp)
		if err != nil {
			t.Fatalf("NewSocket(%s) failed: %v", tp, err)
		}

		if err = sock.Bind(endpoint); err != nil {
			t.Fatalf("Bind(%s) failed: %v", tp, err)
		}

		endpoint, err := sock.GetLastEndpoint()
		if err != nil {
			t.Fatalf("GetLastEndpoint(%s) failed: %v", tp, err)
		}

		if err = sock.Unbind(endpoint); err != nil {
			t.Fatalf("Unbind(%s) failed: %v", tp, err)
		}

		if err = sock.Close(); err != nil {
			t.Fatalf("Close(%s) failed: %v", tp, err)
		}
	}
}

func TestGetFd(t *testing.T) {
	const endpoint = "ipc://*"

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	sock := createSocketOrFail(t, ctx, SocketPUB)
	defer closeSocketOrFail(t, sock)

	if err := sock.Bind(endpoint); err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}

	fd, err := sock.GetFd()
	if err != nil {
		t.Fatalf("GetFd() failed: %v", err)
	}

	if fd < 0 {
		t.Fatalf("GetFd() returned '%d' which is negative", fd)
	}
}

func TestIsUnblockedForRecv(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:5556"

	checkState := func(sock *Socket, expectedState bool) {
		state, err := sock.IsUnblockedForRecv()
		if err != nil {
			t.Fatalf("IsUnblockedForRecv() failed: %v", err)
		}

		if state != expectedState {
			t.Fatalf("Socket unblocked state is '%v', expected '%v",
				state, expectedState)
		}
	}

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	sock := createSocketOrFail(t, ctx, SocketSUB)
	defer closeSocketOrFail(t, sock)

	if err := sock.Connect(endpoint); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	if err := sock.AddSubscribeFilter([]byte("")); err != nil {
		t.Fatalf("AddSubscribeFilter('') failed: %v", err)
	}

	dataChan := make(chan []byte)
	defer close(dataChan)

	err, errorChan := sendGoroutine(SocketPUB, endpoint, dataChan)
	if err != nil {
		t.Fatalf("Failed to initialize the send goroutine: %v", <-errorChan)
	}

	checkState(sock, false)

	for i := 0; i < 10; i++ {
		dataChan <- []byte("msg")
		if err = <-errorChan; err != nil {
			t.Fatalf("Send() at %d-th iteration failed: %v", i+1, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	checkState(sock, true)
}
