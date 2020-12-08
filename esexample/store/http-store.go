package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type RemoteEventStoreConfig struct {
	Host string
}

type RemoteEventStore struct {
	config *RemoteEventStoreConfig
}

type FindEventsByTypeResult struct {
	Events []StoreEvent `json:"events"`
	Latest int64        `json:"latest"`
}

// @see EventStore.Find
func (es *RemoteEventStore) Find(guid EventID) ([]StoreEvent, error) {
	api := fmt.Sprintf("%s/api/v1/events/%s", es.config.Host, guid)

	resp, err := http.Get(api)

	if err != nil {
		return nil, err
	}

	jsondata, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var tmp FindEventsByTypeResult
	err = json.Unmarshal(jsondata, &tmp)

	if err != nil {
		return nil, err
	}

	return tmp.Events, nil
}

func (es *RemoteEventStore) Update(guid EventID, expectedVersion int, events []StoreEvent) error {
	api := fmt.Sprintf("%s/api/v1/events/%s/%d", es.config.Host, guid, expectedVersion)

	var eventsArray []string
	for _, e := range events {

		b, err := json.Marshal(e)

		if err != nil {
			return err
		}

		eventsArray = append(eventsArray, string(b))
	}

	jsondata := fmt.Sprintf("[%s]", strings.Join(eventsArray, ","))
	resp, err := http.Post(api, "application/json", bytes.NewBuffer([]byte(jsondata)))

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fullerror, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		return fmt.Errorf("%s", fullerror)
	}

	return nil
}

func (es *RemoteEventStore) GetEventsByType(etype EventType, sinceMillis int64, batchSize int) (events []StoreEvent, latest int64, theError error) {
	api := fmt.Sprintf("%s/api/v1/types/%d?size=%d&since=%d", es.config.Host, int(etype), batchSize, sinceMillis)

	resp, err := http.Get(api)

	if err != nil {
		theError = err
		return
	}

	// read all the response bytes
	jsondata, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		theError = err
		return
	}

	// parse the response bytes and FindEventByType structure
	var findResult FindEventsByTypeResult
	err = json.Unmarshal(jsondata, &findResult)

	if err != nil {
		theError = err
		return
	}

	return findResult.Events, findResult.Latest, nil
}

// initializer for event store
func NewRemoteEventStore(config *RemoteEventStoreConfig) *RemoteEventStore {
	return &RemoteEventStore{config: config}
}

func parseJsonBytesToStringArry(jsondata []byte) ([]string, error) {
	// parse all input data
	var tmp []interface{}

	err := json.Unmarshal(jsondata, &tmp)

	if err != nil {
		return nil, err
	}

	var events []string

	for _, e := range tmp {

		b, err := json.Marshal(e)

		if err != nil {
			return nil, err
		}

		events = append(events, string(b))
	}

	return events, nil
}
