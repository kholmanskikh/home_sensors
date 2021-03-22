package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type WebApi struct {
	BaseUrl string

	mtypesUrl       string
	measurementsUrl string

	mtypeToId    map[string]int
	mtypeToIdMux *sync.Mutex
}

func NewWebApi(baseUrl string) (*WebApi, error) {
	api := WebApi{BaseUrl: baseUrl,
		mtypesUrl:       baseUrl + "/mtypes/",
		measurementsUrl: baseUrl + "/measurements/"}

	api.mtypeToId = make(map[string]int)
	api.mtypeToIdMux = &sync.Mutex{}

	err := api.UpdateMtypeIds()
	if err != nil {
		return nil, fmt.Errorf("UpdateMtypeIds() failed: %v", err)
	}

	return &api, nil
}

func (api *WebApi) SupportedTypes() []string {
	ret := make([]string, 0)

	api.mtypeToIdMux.Lock()
	for mtype, _ := range api.mtypeToId {
		ret = append(ret, mtype)
	}
	api.mtypeToIdMux.Unlock()

	return ret
}

func (api *WebApi) UpdateMtypeIds() error {
	type mtype struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	type mtypeList struct {
		Mtypes []mtype `json:"mtypes"`
	}

	resp, err := http.Get(api.mtypesUrl)
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

	api.mtypeToIdMux.Lock()
	api.mtypeToId = make(map[string]int)
	for _, mtype := range mtypes.Mtypes {
		api.mtypeToId[mtype.Name] = mtype.Id
	}
	api.mtypeToIdMux.Unlock()

	return nil
}

func (api *WebApi) PublishMeasurement(m Measurement) error {
	api.mtypeToIdMux.Lock()
	mtypeId, found := api.mtypeToId[m.Type]
	api.mtypeToIdMux.Unlock()

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

	resp, err := http.Post(api.measurementsUrl, "application/json", bytes.NewReader(data))
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
