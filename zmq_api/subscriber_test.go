package zmq_api

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kholmanskikh/home_sensors/zmq_api/zmq"
)

type sendWorker struct {
	stopFlag int32
	stopChan chan struct{}
}

func newSendWorker(endpoint string, dataToSend string) (*sendWorker, error) {
	var w sendWorker
	w.stopChan = make(chan struct{})

	initErrorChan := make(chan error)

	go func() {
		defer close(w.stopChan)

		ctx, err := zmq.NewContext()
		if err != nil {
			initErrorChan <- fmt.Errorf("NewContext failed: %v", err)
			return
		}
		defer ctx.Terminate()

		sock, err := zmq.NewSocket(ctx, zmq.SocketPUB)
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
			if atomic.LoadInt32(&w.stopFlag) != 0 {
				break
			}

			if err = sock.Send([]byte(dataToSend)); err != nil {
				fmt.Printf("Send('%s') failed: %v\n", dataToSend, err)
			}

			time.Sleep(time.Millisecond * 200)
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

func TestSubscriber(t *testing.T) {
	const endpoint = "tcp://127.0.0.1:9000"

	measurementData := `{"device_id": 99, "type": "Some type", "value": 90.7, "timestamp": 12}`
	expectedMeasurement := Measurement{DeviceId: 99, Type: "Some type", Value: 90.7, Timestamp: 12}

	sender, err := newSendWorker(endpoint, measurementData)
	if err != nil {
		t.Fatalf("newSendWorker() failed: %v", err)
	}
	defer sender.Destroy()

	s, err := NewSubscriber(endpoint)
	if err != nil {
		t.Fatalf("NewSubscriber() failed: %v", err)
	}

	var measurements []*Measurement
	for i := 0; (i < 10) && (len(measurements) == 0); i++ {
		measurements, err = s.RecvMeasurement(time.Millisecond * 100)
		if err != nil {
			t.Fatalf("RecvMeasurment() failed: %v", err)
		}
	}

	if len(measurements) == 0 {
		t.Fatalf("No measurement received")
	}

	if *measurements[0] != expectedMeasurement {
		t.Fatalf("Got '%#v', expected '%#v'", *measurements[0], expectedMeasurement)
	}

	err = s.Destroy()
	if err != nil {
		t.Fatalf("Destroy() failed: %v", err)
	}
}
