package zmq_api

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func sendGoroutine(tp socketType, endpoint string, dataChan <-chan []byte) (error, <-chan error) {
	initErrorChan := make(chan error)
	errorChan := make(chan error, 1)

	go func() {
		defer close(errorChan)

		ctx, err := NewContext()
		if err != nil {
			initErrorChan <- fmt.Errorf("NewContext() failed: %v", err)
			return
		}
		defer ctx.Terminate()

		sock, err := NewSocket(ctx, tp)
		if err != nil {
			initErrorChan <- fmt.Errorf("NewSocket() failed: %v", err)
			return
		}
		defer sock.Close()

		if err = sock.Bind(endpoint); err != nil {
			initErrorChan <- fmt.Errorf("Bind() failed: %v", err)
			return
		}

		initErrorChan <- nil

		for {
			buf, ok := <-dataChan
			if !ok {
				break
			}

			errorChan <- sock.Send(buf)
		}
	}()

	return <-initErrorChan, errorChan
}

func initSendGoroutine(t *testing.T, dataChan chan<- []byte, errorChan <-chan error) {
	// Even following all the recommendations from the "Missing Message Problem Solver"
	// chapter of the ZMQ guide does not solve the problem that the first message get lost
	// That's why we sleep to let ZMQ create all necessary internal queues
	dataChan <- []byte("this message will be lost")
	if err := <-errorChan; err != nil {
		t.Fatalf("The first Send() failed: %v", err)
	}
	time.Sleep(time.Second)
}

func sendAndRecvMsgs(t *testing.T, dataChan chan<- []byte, errorChan <-chan error,
	recvSock *Socket, sendMsgs, recvMsgs []string) {
	for _, msg := range sendMsgs {
		dataChan <- []byte(msg)
		err := <-errorChan
		if err != nil {
			t.Fatalf("Send('%s') failed: %v", msg, err)
		}
	}

	recvBuf := make([]byte, 512)
	for _, msg := range recvMsgs {
		recvLen, err := recvSock.Recv(recvBuf)
		if err != nil {
			t.Fatalf("Recv('%s') failed: %v", msg, err)
		}

		recvData := recvBuf[:recvLen]
		expData := []byte(msg)

		if !bytes.Equal(recvData, expData) {
			t.Fatalf("Received '%s', expected '%s'", string(recvData), string(expData))
		}
	}
}

func TestSendRecv(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:5555"

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

	initSendGoroutine(t, dataChan, errorChan)

	sendMsgs := []string{"hello", "my", "dear", "friend"}
	recvMsgs := sendMsgs
	sendAndRecvMsgs(t, dataChan, errorChan, sock,
		sendMsgs, recvMsgs)
}

func TestSendRecvNonWildcard(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:5556"

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	sock := createSocketOrFail(t, ctx, SocketSUB)
	defer closeSocketOrFail(t, sock)

	if err := sock.Connect(endpoint); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	prefix := "prefix:"

	if err := sock.AddSubscribeFilter([]byte(prefix)); err != nil {
		t.Fatalf("AddSubscribeFilter('%s') failed: %v", prefix, err)
	}

	dataChan := make(chan []byte)
	defer close(dataChan)

	err, errorChan := sendGoroutine(SocketPUB, endpoint, dataChan)
	if err != nil {
		t.Fatalf("Failed to initialize the send goroutine: %v", <-errorChan)
	}

	initSendGoroutine(t, dataChan, errorChan)

	msgWithPrefix := prefix + "some content"
	sendMsgs := []string{"one", "two", msgWithPrefix}
	recvMsgs := []string{msgWithPrefix}
	sendAndRecvMsgs(t, dataChan, errorChan, sock,
		sendMsgs, recvMsgs)
}

func TestSendRecv2NonWildcards(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:5557"

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	sock := createSocketOrFail(t, ctx, SocketSUB)
	defer closeSocketOrFail(t, sock)

	if err := sock.Connect(endpoint); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	prefix1 := "prefix1:"
	prefix2 := "prefix2:"
	ignoredPrefix := "ignored_prefix:"

	if err := sock.AddSubscribeFilter([]byte(prefix1)); err != nil {
		t.Fatalf("AddSubscribeFilter('%s') failed: %v", prefix1, err)
	}
	if err := sock.AddSubscribeFilter([]byte(prefix2)); err != nil {
		t.Fatalf("AddSubscribeFilter('%s') failed: %v", prefix2, err)
	}

	if err := sock.AddSubscribeFilter([]byte(ignoredPrefix)); err != nil {
		t.Fatalf("AddSubscribeFilter('%s') failed: %v", ignoredPrefix, err)
	}
	if err := sock.RemoveSubscribeFilter([]byte(ignoredPrefix)); err != nil {
		t.Fatalf("RemoveSubscribeFilter('%s') failed: %v", ignoredPrefix, err)
	}

	dataChan := make(chan []byte)
	defer close(dataChan)

	err, errorChan := sendGoroutine(SocketPUB, endpoint, dataChan)
	if err != nil {
		t.Fatalf("Failed to initialize the send goroutine: %v", <-errorChan)
	}

	initSendGoroutine(t, dataChan, errorChan)

	msgWithPrefix1 := prefix1 + "some content"
	msgWithPrefix2 := prefix2 + "other content"
	msgWithIgnoredPrefix := ignoredPrefix + "content"
	sendMsgs := []string{"one", "two", msgWithPrefix1, msgWithIgnoredPrefix, msgWithPrefix2}
	recvMsgs := []string{msgWithPrefix1, msgWithPrefix2}
	sendAndRecvMsgs(t, dataChan, errorChan, sock,
		sendMsgs, recvMsgs)
}

func TestTruncatedRecv(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:5558"

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

	initSendGoroutine(t, dataChan, errorChan)

	sentData := []byte("data")
	recvData := make([]byte, 2)

	dataChan <- sentData
	if err = <-errorChan; err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	_, err = sock.Recv(recvData)
	if (err == nil) || (!strings.Contains(err.Error(), "truncated")) {
		t.Fatalf("Unexpected error from Recv(): %v", err)
	}
}
