package config

import (
    "os"
    "io/ioutil"
    "fmt"
    "encoding/json"
)

type Config struct {
    ZMQEndpoint string `json:"zmq_endpoint"`
    Debug bool `json:"debug"`

    Publisher string `json:"publisher"`

    // Web
    WebURL string `json:"web_url"`
    WebUpdateTypesInterval int `json:"web_update_types_interval"`

    //MQTT
    MQTTBroker string `json:"mqtt_broker"`
    MQTTUser string `json:"mqtt_user"`
    MQTTPassword string `json:"mqtt_password"`
    MQTTTopic string `json:"mqtt_topic"`
}

var Format string = `{
    "zmq_endpint": "tcp://1.2.3.4:5555",
    "debug": true of false, // optional
    "publisher": "web" or "mqtt",

    // web-only options
    "web_url": "http://1.2.3.4/api", // The web API url
    "web_update_types_interval": 30, // How often (in secs) to refresh the info 
                                        regarding supported types

    // mqtt-only options
    "mqtt_broker": "...",
    "mqtt_user": "...", // optional,
    "mqtt_password": "...", // optional
    "mqtt_topic": "..." // measurements are posted to <mqtt_topic>/<device_id>/<type>
}`

func ParseFromFile(path string) (*Config, error) {
    config := Config{}

    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("unable to read: %v", err)
    }

    if err := json.Unmarshal(bytes, &config); err != nil {
        return nil, fmt.Errorf("unable to parse json: %v", err)
    }

    if config.ZMQEndpoint == "" {
        return nil, fmt.Errorf("zmq_endpoint must be set")
    }

    switch (config.Publisher) {
    case "web":
        err = validateWebConfig(&config)
    case "mqtt":
        err = validateMQTTConfig(&config)
    default:
        err = fmt.Errorf("unsupported Publisher '%s'", config.Publisher)
    }

    if err != nil {
        return nil, err
    }

    return &config, nil
}

func validateWebConfig(config *Config) error {
    if config.WebURL == "" {
        return fmt.Errorf("web_url must be set")
    }

    if config.WebUpdateTypesInterval <= 0 {
        return fmt.Errorf("invalid value for web_update_types_interval: %d",
                            config.WebUpdateTypesInterval)
    }

    return nil
}

func validateMQTTConfig(config *Config) error {
    if config.MQTTBroker == "" {
        return fmt.Errorf("mqtt_broker must be set")
    }

    if config.MQTTTopic == "" {
        return fmt.Errorf("mqtt_topic must be set")
    }

    if (config.MQTTUser == "") && (config.MQTTPassword != "") {
        return fmt.Errorf("mqtt_password is set for an empty mqtt_user")
    }

    return nil
}

