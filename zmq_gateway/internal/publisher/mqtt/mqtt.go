package mqtt

import (
    "fmt"
    "encoding/json"
    "strings"

    MQTT "github.com/eclipse/paho.mqtt.golang"

    "github.com/kholmanskikh/home_sensors/zmq_api"
)

type MQTTPublisher struct {
    BaseTopic string

    client MQTT.Client
}

func NewMQTTPublisher(broker string, user string, password string, topic string) (*MQTTPublisher, error) {
    publisher := MQTTPublisher{BaseTopic: topic}

    clientOpts := MQTT.NewClientOptions()
    clientOpts.AddBroker(broker)

    if user != "" {
        clientOpts.SetUsername(user)
        clientOpts.SetPassword(password)
    }

    publisher.client = MQTT.NewClient(clientOpts)
    if t := publisher.client.Connect(); t.Wait() && t.Error() != nil {
        return nil, fmt.Errorf("MQTT.NewClient() failed: %v", t.Error())
    }

    return &publisher, nil
}

func (publisher *MQTTPublisher) Description() string {
    return fmt.Sprintf("MQTT Publisher (topic: '%s')", publisher.BaseTopic)
}

func (publisher *MQTTPublisher) Destroy() error {
    publisher.client.Disconnect(0)

    return nil
}

func (publisher *MQTTPublisher) PublishMeasurement(m zmq_api.Measurement) error {
    path := fmt.Sprintf("%s/%d/%s", publisher.BaseTopic, m.DeviceId, strings.ToLower(m.Type))

    type MeasurementToPost struct {
        Timestamp int `json:"timestamp"`
        Value float64 `json:"value"`
    }

    data, err := json.Marshal(MeasurementToPost{Timestamp: m.Timestamp, Value: m.Value})
    if err != nil {
        return fmt.Errorf("failed to marshal the data: %v", err)
    }

    t := publisher.client.Publish(path, 0, false, data)
    t.Wait()
    return t.Error()
}
