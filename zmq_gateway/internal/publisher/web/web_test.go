package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/kholmanskikh/home_sensors/zmq_api"
)

func TestWebPublisherEmptyMtypesOutput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	}))
	defer ts.Close()

	_, err := NewWebPublisher(ts.URL, time.Minute)
	if (err == nil) || (!strings.Contains(err.Error(), "UpdateMtypeIds")) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestWebPublisherPublishMeasurementUnsupportedType(t *testing.T) {
	var mux http.ServeMux
	mux.HandleFunc("/mtypes/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"mtypes": [{"id": 1, "name": "Some type"}]}`)
	})

	ts := httptest.NewServer(&mux)
	defer ts.Close()

	publisher, err := NewWebPublisher(ts.URL, time.Minute)
	if err != nil {
		t.Fatalf("NewWebPublisher() failed: %v", err)
	}

	unsupportedTypeName := "Unsupported type"

	m := zmq_api.Measurement{DeviceId: 1,
		Type:      unsupportedTypeName,
		Value:     12.0,
		Timestamp: 1}

	err = publisher.PublishMeasurement(m)
	if (err == nil) || (!strings.Contains(err.Error(), unsupportedTypeName)) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestWebPublisherPublishMeasurementStatusNotOk(t *testing.T) {
	var mux http.ServeMux
	mux.HandleFunc("/mtypes/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"mtypes": [{"id": 1, "name": "Some type"}]}`)
	})
	mux.HandleFunc("/measurements/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	ts := httptest.NewServer(&mux)
	defer ts.Close()

	publisher, err := NewWebPublisher(ts.URL, time.Minute)
	if err != nil {
		t.Fatalf("NewWebPublisher() failed: %v", err)
	}

	m := zmq_api.Measurement{DeviceId: 1,
		Type:      "Some type",
		Value:     12.0,
		Timestamp: 1}

	err = publisher.PublishMeasurement(m)
	if (err == nil) || (!strings.Contains(err.Error(), "HTTP status")) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestWebPublisherPublishMeasurement(t *testing.T) {
	errorChan := make(chan error, 1)

	var mux http.ServeMux
	mux.HandleFunc("/mtypes/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"mtypes": [{"id": 1, "name": "Some type"}]}`)
	})
	mux.HandleFunc("/measurements/", func(w http.ResponseWriter, r *http.Request) {
		statusToWrite := http.StatusBadRequest

		defer func() {
			w.WriteHeader(statusToWrite)
		}()

		if r.Method != "POST" {
			errorChan <- fmt.Errorf("unexpected method '%s', 'POST' expected", r.Method)
			return
		}

		contentType := r.Header["Content-Type"][0]
		if contentType != "application/json" {
			errorChan <- fmt.Errorf("unexpected 'Content-Type' '%s', 'application/json' expected", contentType)
			return
		}

		type Data struct {
			SensorId  int     `json:"sensor_id"`
			MtypeId   int     `json:"mtype_id"`
			Timestamp int     `json:"timestamp"`
			Value     float64 `json:"value"`
		}

		expectedData := Data{SensorId: 1, MtypeId: 1, Timestamp: 1, Value: 12.0}
		var receivedData Data

		receivedBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errorChan <- fmt.Errorf("failed to recieve the body: %v", err)
			return
		}

		err = json.Unmarshal(receivedBytes, &receivedData)
		if err != nil {
			errorChan <- fmt.Errorf("failed to unmarshal json: %v", err)
			return
		}

		if expectedData != receivedData {
			errorChan <- fmt.Errorf("received '%#v', expected '%#v'", receivedData, expectedData)
			return
		}

		statusToWrite = http.StatusOK
	})

	ts := httptest.NewServer(&mux)
	defer ts.Close()

	publisher, err := NewWebPublisher(ts.URL, time.Minute)
	if err != nil {
		t.Fatalf("NewWebPublisher() failed: %v", err)
	}

	m := zmq_api.Measurement{DeviceId: 1, Type: "Some type",
		Value: 12.0, Timestamp: 1}

	var fatalMessage string
	err = publisher.PublishMeasurement(m)
	if err != nil {
		fatalMessage = fmt.Sprintf("PublishMeasurement() failed: %v", err)
	}

	if len(errorChan) != 0 {
		fatalMessage = fatalMessage + fmt.Sprintf(", the web server reported error: %v", <-errorChan)
	}

	if fatalMessage != "" {
		t.Fatalf(fatalMessage)
	}
}

func TestSupportedTypes(t *testing.T) {
	var mux http.ServeMux
	mux.HandleFunc("/mtypes/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"mtypes": [{"id": 1, "name": "Some type"}]}`)
	})

	ts := httptest.NewServer(&mux)
	defer ts.Close()

	publisher, err := NewWebPublisher(ts.URL, time.Minute)
	if err != nil {
		t.Fatalf("NewWebPublisher() failed: %v", err)
	}

	expectedTypes := []string{"Some type"}
	supportedTypes := publisher.SupportedTypes()

	if (len(expectedTypes) != 1) || (expectedTypes[0] != supportedTypes[0]) {
		t.Fatalf("Expected '%v', returned '%v'", expectedTypes, supportedTypes)
	}
}
