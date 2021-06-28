package zmq_api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kholmanskikh/home_sensors/zmq_api/zmq"
)

type Subscriber struct {
	ctx    *zmq.Context
	sock   *zmq.Socket
	poller *zmq.ReadPoller

	Endpoint string
}

func NewSubscriber(endpoint string) (*Subscriber, error) {
	var err error
	s := Subscriber{Endpoint: endpoint}

	defer func() {
		if err != nil {
			s.cleanupResources()
		}
	}()

	s.ctx, err = zmq.NewContext()
	if err != nil {
		err = fmt.Errorf("NewContext() failed: %v", err)
		return nil, err
	}

	s.sock, err = zmq.NewSocket(s.ctx, zmq.SocketSUB)
	if err != nil {
		err = fmt.Errorf("NewSocket() failed: %v", err)
		return nil, err
	}

	if err = s.sock.Connect(s.Endpoint); err != nil {
		err = fmt.Errorf("Connect('%s') failed: %v", s.Endpoint, err)
		return nil, err
	}

	if err = s.sock.AddSubscribeFilter(nil); err != nil {
		err = fmt.Errorf("AddSubscribeFilter('') failed: %v", err)
		return nil, err
	}

	s.poller, err = zmq.NewReadPoller(s.sock)
	if err != nil {
		err := fmt.Errorf("NewReadPoller() failed: %v", err)
		return nil, err
	}

	return &s, err
}

func (s *Subscriber) cleanupResources() error {
	var err error

	if s.sock != nil {
		sockErr := s.sock.Close()
		if (err == nil) && (sockErr != nil) {
			err = fmt.Errorf("socket Close() failed: %v", sockErr)
		}

		s.sock = nil
	}

	if s.ctx != nil {
		ctxErr := s.ctx.Terminate()
		if (err == nil) && (ctxErr != nil) {
			err = fmt.Errorf("ctx Terminate() failed: %v", ctxErr)
		}

		s.ctx = nil
	}

	return err
}

func (s *Subscriber) Destroy() error {
	return s.cleanupResources()
}

// returns the first error encountered
func (s *Subscriber) RecvMeasurement(timeout time.Duration) ([]*Measurement, error) {
	measurements := make([]*Measurement, 0)

	// as Poll() is edge-triggered, and RecvMeasurement() returns the first error
	// encountered, Poll() may return an empty socket list if the previous call
	// to Poll() failed with an error. That's why we have to always check the
	// socket for available messages.
	_, err := s.poller.Poll(timeout)
	if err != nil {
		return measurements, fmt.Errorf("Poll() failed: %v", err)
	}

	var recvM struct {
		DeviceId  int     `json:"device_id"`
		Type      string  `json:"type"`
		Value     float64 `json:"value"`
		Timestamp int     `json:"timestamp"`
	}

	recvBuf := make([]byte, 1024)
	for {
		recvLen, err, received := s.sock.RecvNonBlocking(recvBuf)
		if err != nil {
			return measurements, fmt.Errorf("RecvNonBlocking() failed: %v", err)
		}

		if !received {
			break
		}

		recvData := recvBuf[:recvLen]

		err = json.Unmarshal(recvData, &recvM)
		if err != nil {
			return measurements, fmt.Errorf("Unmarshal('%s') failed: %v", string(recvData), err)
		}

		m := Measurement{DeviceId: recvM.DeviceId,
			Type:      recvM.Type,
			Value:     recvM.Value,
			Timestamp: recvM.Timestamp}
		measurements = append(measurements, &m)
	}

	return measurements, nil
}
