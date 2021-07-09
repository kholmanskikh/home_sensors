package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/kholmanskikh/home_sensors/zmq_api"
)

type WebPublisher struct {
	BaseUrl             string
	UpdateTypesInterval time.Duration

	lastUpdated time.Time

	mtypesUrl       string
	measurementsUrl string

	mtypeToId    map[string]int
	mtypeToIdMux *sync.Mutex
}

func NewWebPublisher(baseUrl string, updateTypesInterval time.Duration) (*WebPublisher, error) {
	publisher := WebPublisher{BaseUrl: baseUrl,
		mtypesUrl:       baseUrl + "/mtypes/",
		measurementsUrl: baseUrl + "/measurements/"}

	publisher.mtypeToId = make(map[string]int)
	publisher.mtypeToIdMux = &sync.Mutex{}

	err := publisher.updateMtypeIds()
	if err != nil {
		return nil, fmt.Errorf("UpdateMtypeIds() failed: %v", err)
	}

	publisher.lastUpdated = time.Now()

	return &publisher, nil
}

func (publisher *WebPublisher) SupportedTypes() []string {
	ret := make([]string, 0)

	publisher.mtypeToIdMux.Lock()
	for mtype, _ := range publisher.mtypeToId {
		ret = append(ret, mtype)
	}
	publisher.mtypeToIdMux.Unlock()

	return ret
}

func (publisher *WebPublisher) updateMtypeIds() error {
	type mtype struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	type mtypeList struct {
		Mtypes []mtype `json:"mtypes"`
	}

	resp, err := http.Get(publisher.mtypesUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read the body: %v", err)
	}

	var mtypes mtypeList
	err = json.Unmarshal(body, &mtypes)
	if err != nil {
		return err
	}

	publisher.mtypeToIdMux.Lock()
	publisher.mtypeToId = make(map[string]int)
	for _, mtype := range mtypes.Mtypes {
		publisher.mtypeToId[mtype.Name] = mtype.Id
	}
	publisher.mtypeToIdMux.Unlock()

	return nil
}

func (publisher *WebPublisher) PublishMeasurement(m zmq_api.Measurement) error {
	if time.Since(publisher.lastUpdated) > publisher.UpdateTypesInterval {
		err := publisher.updateMtypeIds()
		if err != nil {
			return nil
		}

		publisher.lastUpdated = time.Now()
	}

	publisher.mtypeToIdMux.Lock()
	mtypeId, found := publisher.mtypeToId[m.Type]
	publisher.mtypeToIdMux.Unlock()

	if !found {
		return fmt.Errorf("unsupported measurement type '%s'", m.Type)
	}

	type MeasurementToPost struct {
		SensorId  int     `json:"sensor_id"`
		MtypeId   int     `json:"mtype_id"`
		Timestamp int     `json:"timestamp"`
		Value     float64 `json:"value"`
	}

	measurementToPost := MeasurementToPost{SensorId: m.DeviceId,
		MtypeId:   mtypeId,
		Value:     m.Value,
		Timestamp: m.Timestamp}

	data, err := json.Marshal(measurementToPost)
	if err != nil {
		return err
	}

	resp, err := http.Post(publisher.measurementsUrl, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("returned HTTP status %d", resp.StatusCode)
	}
	resp.Body.Close()

	return nil
}

func (publisher *WebPublisher) Description() string {
	return fmt.Sprintf("Web Publisher (url '%s')", publisher.BaseUrl)
}

func (publisher *WebPublisher) Destroy() error {
	return nil
}
