package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestConfigFile struct {
	file *os.File
}

func NewTestConfigFile(content string) (*TestConfigFile, error) {
	file, err := os.CreateTemp("", "*")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(file.Name(), []byte(content), 0)
	if err != nil {
		os.Remove(file.Name())
		return nil, err
	}

	return &TestConfigFile{file: file}, nil
}

func (f *TestConfigFile) Name() string {
	return f.file.Name()
}

func (f *TestConfigFile) Destroy() error {
	return os.Remove(f.Name())
}

func checkError(err error, expectedErrorMessage string) error {
	if (err == nil) || (err.Error() != expectedErrorMessage) {
		return fmt.Errorf("unexpected error: %v", err)
	}

	return nil
}

func createTestFileOrFail(t *testing.T, content string) *TestConfigFile {
	tf, err := NewTestConfigFile(content)
	if err != nil {
		t.Fatalf("unable to create a test config file: %v", err)
	}

	return tf
}

func TestParseFromFileWebPublisher(t *testing.T) {
	tf := createTestFileOrFail(t, `{
"zmq_endpoint": "endpoint",
"debug": true,
"publisher": "web",
"web_url": "url",
"web_update_types_interval": 99
}`)
	defer tf.Destroy()

	config, err := ParseFromFile(tf.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if (config.ZMQEndpoint != "endpoint") || (!config.Debug) || (config.Publisher != "web") ||
		(config.WebURL != "url") || (config.WebUpdateTypesInterval != 99) {
		t.Fatalf("fields are set incorrectly")
	}
}

func TestParseFromFileMQTTPublisher(t *testing.T) {
	tf := createTestFileOrFail(t, `{
"zmq_endpoint": "endpoint",
"debug": true,
"publisher": "mqtt",
"mqtt_broker": "broker",
"mqtt_user": "user",
"mqtt_password": "password",
"mqtt_topic": "topic"
}`)
	defer tf.Destroy()

	config, err := ParseFromFile(tf.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if (config.ZMQEndpoint != "endpoint") || (!config.Debug) || (config.Publisher != "mqtt") ||
		(config.MQTTBroker != "broker") || (config.MQTTUser != "user") ||
		(config.MQTTPassword != "password") || (config.MQTTTopic != "topic") {
		t.Fatalf("fields are set incorrectly")
	}
}

func TestParseFromFileNotExistingFile(t *testing.T) {
	tf := createTestFileOrFail(t, "it is not important what we put here")
	path := tf.Name()
	if err := tf.Destroy(); err != nil {
		t.Fatalf("unable to destroy the test file '%s': %v", path, err)
	}

	_, err := ParseFromFile(path)
	if err == nil {
		t.Fatalf("ParseFromFile() did not fail")
	}
}

func TestParseFromFileNotAJSON(t *testing.T) {
	tf := createTestFileOrFail(t, "some non-json content")
	defer tf.Destroy()

	_, err := ParseFromFile(tf.Name())
	if (err == nil) || (strings.Index(err.Error(), "unable to parse json:") != 0) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFromFileNoZMQEndpoint(t *testing.T) {
	tf := createTestFileOrFail(t, `{"other_field": "other value"}`)
	defer tf.Destroy()

	_, err := ParseFromFile(tf.Name())
	if err := checkError(err, "zmq_endpoint must be set"); err != nil {
		t.Fatal(err)
	}
}

func TestParseFromFileUnsupportedPublisher(t *testing.T) {
	tf := createTestFileOrFail(t, `{"zmq_endpoint": "endpoint", "publisher": "other_publisher"}`)
	defer tf.Destroy()

	_, err := ParseFromFile(tf.Name())
	if err := checkError(err, "unsupported Publisher 'other_publisher'"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateWebConfig(t *testing.T) {
	config0 := Config{WebURL: "some_url", WebUpdateTypesInterval: 30}
	if err := validateWebConfig(&config0); err != nil {
		t.Fatalf("unexpected error for a valid config: %v", err)
	}

	config1 := Config{}
	if err := checkError(validateWebConfig(&config1), "web_url must be set"); err != nil {
		t.Fatal(err)
	}

	config2 := Config{WebURL: "some_url",
		WebUpdateTypesInterval: -1}
	if err := checkError(validateWebConfig(&config2), "invalid value for web_update_types_interval: -1"); err != nil {
		t.Fatal(err)
	}

	config3 := Config{WebURL: "some_url",
		WebUpdateTypesInterval: 0}
	if err := checkError(validateWebConfig(&config3), "invalid value for web_update_types_interval: 0"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMQTTConfig(t *testing.T) {
	config0 := Config{MQTTBroker: "some_broker", MQTTTopic: "some_topic", MQTTUser: "some_user"}
	if err := validateMQTTConfig(&config0); err != nil {
		t.Fatalf("unexpected error for a valid config: %v", err)
	}

	config1 := Config{}
	if err := checkError(validateMQTTConfig(&config1), "mqtt_broker must be set"); err != nil {
		t.Fatal(err)
	}

	config2 := Config{MQTTBroker: "some_broker"}
	if err := checkError(validateMQTTConfig(&config2), "mqtt_topic must be set"); err != nil {
		t.Fatal(err)
	}

	config3 := Config{MQTTBroker: "some_broker", MQTTTopic: "some_topic", MQTTPassword: "some_password"}
	if err := checkError(validateMQTTConfig(&config3), "mqtt_password is set for an empty mqtt_user"); err != nil {
		t.Fatal(err)
	}
}
