package zmq_api

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

/*
create two senders
pass them 2 slices with data to send
make them send
*/

// should be less than 1000
const numberOfSenders = 4
const sendWorkerMaxSleepTimeMs = 1000
const recvSleepTimeMs = 500
const numberOfSuccessfulReads = 20

type sendWorker struct {
	Endpoint string

	stopFlag int32
	stopChan chan struct{}
}

func newSendWorker(tp socketType, endpoint string, maxSleepTimeMs int) (*sendWorker, error) {
	var w sendWorker
	w.stopChan = make(chan struct{})
	w.Endpoint = endpoint

	initErrorChan := make(chan error)

	go func() {
		defer close(w.stopChan)

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

		if err = sock.Bind(w.Endpoint); err != nil {
			initErrorChan <- fmt.Errorf("Bind() failed: %v", err)
			return
		}

		initErrorChan <- nil

		for {
			if atomic.LoadInt32(&w.stopFlag) != 0 {
				break
			}

			if err = sock.Send([]byte("some data")); err != nil {
				fmt.Printf("Send() in worker '%s' failed: %v\n", w.Endpoint, err)
			}

			time.Sleep(time.Duration(rand.Intn(maxSleepTimeMs)) * time.Millisecond)
		}
	}()

	if err := <-initErrorChan; err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *sendWorker) Destroy() {
	atomic.StoreInt32(&w.stopFlag, 1)
	<-w.stopChan
}

func TestReadPoller(t *testing.T) {
	endpoints := make([]string, numberOfSenders)
	for i, _ := range endpoints {
		endpoints[i] = fmt.Sprintf("tcp://127.0.0.1:6%03d", i)
	}

	ctx := createContextOrFail(t)
	defer terminateContextOrFail(t, ctx)

	sockets := make([]*Socket, numberOfSenders)
	for i, _ := range sockets {
		sockets[i] = createSocketOrFail(t, ctx, SocketSUB)
		defer closeSocketOrFail(t, sockets[i])

		if err := sockets[i].Connect(endpoints[i]); err != nil {
			t.Fatalf("Connect('%s') failed: %v", endpoints[i], err)
		}

		if err := sockets[i].AddSubscribeFilter(nil); err != nil {
			t.Fatalf("AddSubscribeFilter('') for socket %d failed: %v", i, err)
		}
	}

	for _, endpoint := range endpoints {
		worker, err := newSendWorker(SocketPUB, endpoint, sendWorkerMaxSleepTimeMs)
		if err != nil {
			t.Fatalf("Unable to create worker at '%s': %v", endpoint, err)
		}
		defer worker.Destroy()
	}

	poller, err := NewReadPoller(sockets...)
	if err != nil {
		t.Fatalf("NewReadPoller() failed: %v", err)
	}

	succReads := make(map[*Socket]int)
	for _, socket := range sockets {
		succReads[socket] = 0
	}

	isEnoughReads := func() bool {
		enough := true
		for _, n := range succReads {
			if n < numberOfSuccessfulReads {
				enough = false
			}
		}

		return enough
	}

	recvBuf := make([]byte, 1024)
	for {
		time.Sleep(time.Duration(rand.Intn(recvSleepTimeMs)) * time.Millisecond)

		readySocks, err := poller.Poll(sendWorkerMaxSleepTimeMs * time.Millisecond)
		if err != nil {
			t.Fatalf("Poll() failed: %v", err)
		}

		for _, s := range readySocks {
			for {
				_, err, received := s.RecvNonBlocking(recvBuf)
				if !received {
					break
				}

				if err != nil {
					t.Fatalf("Recv(%v) failed: %v", s, err)
				}

				succReads[s] += 1
			}
		}

		if isEnoughReads() {
			break
		}
	}
}

func TestReadPollerWithEmptySocketList(t *testing.T) {
	_, err := NewReadPoller()
	if (err == nil) || (!strings.Contains(err.Error(), "empty")) {
		t.Fatalf("Unexpected error from NewReadPoller(): %v", err)
	}
}
