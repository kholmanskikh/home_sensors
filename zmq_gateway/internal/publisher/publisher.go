package publisher

import (
	"github.com/kholmanskikh/home_sensors/zmq_api"
)

type Publisher interface {
	PublishMeasurement(zmq_api.Measurement) error
	Description() string
	Destroy() error
}
