package store

import (
	"context"
	"fmt"
	"my/esexample/storegrpc"
	"time"

	"google.golang.org/grpc"
)

const defaultTimeout = 2000 * time.Millisecond

type GrpcEventStoreConfig struct {
	Host            string
	TimeoutInMillis int
}

type GrpcEventStore struct {
	Config  *GrpcEventStoreConfig
	Client  storegrpc.EventStoreServiceClient
	Timeout time.Duration
}

func (es *GrpcEventStore) Find(guid EventID) ([]StoreEvent, error) {
	client, err := es.getClient()

	if err != nil {
		return nil, err
	}

	request := &storegrpc.FindByIDRequest{Id: string(guid)}

	ctx, cancelFunc := es.createContext()
	defer cancelFunc()
	response, err := client.FindByID(ctx, request)

	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("ERROR: %+v", response.Error)
	}

	var result []StoreEvent

	for _, e := range response.Events {
		result = append(result, StoreEvent{
			ID:      EventID(e.Id),
			Payload: EventPayload(e.Payload),
			Type:    EventType(e.Type)})
	}

	return result, nil
}

func (es *GrpcEventStore) Update(guid EventID, expectedVersion int, events []StoreEvent) error {
	client, err := es.getClient()

	if err != nil {
		return err
	}

	var updateRequestEvents []*storegrpc.UpdateRequest_Event

	for _, e := range events {
		updateRequestEvents = append(updateRequestEvents, &storegrpc.UpdateRequest_Event{
			Type:    int32(e.Type),
			Payload: string(e.Payload),
		})
	}

	request := &storegrpc.UpdateRequest{
		Id:      string(guid),
		Version: int32(expectedVersion),
		Events:  updateRequestEvents}

	ctx, cancelFunc := es.createContext()
	defer cancelFunc()
	response, err := client.Update(ctx, request)

	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("ERROR: %+v", response.Error)
	}

	return nil
}

func (es *GrpcEventStore) GetEventsByType(etype EventType, sinceMillis int64, batchSize int) (events []StoreEvent, latest int64, theError error) {
	client, err := es.getClient()

	if err != nil {
		theError = err
		return
	}

	request := &storegrpc.FindByTypeRequest{
		Type:      int32(etype),
		Since:     sinceMillis,
		BatchSize: int32(batchSize),
	}

	ctx, cancelFunc := es.createContext()
	defer cancelFunc()
	response, err := client.FindByType(ctx, request)

	if err != nil {
		theError = err
		return
	}

	if !response.Success {
		theError = fmt.Errorf("ERROR: %+v", response.Error)
		return
	}

	for _, e := range response.Events {
		events = append(events, StoreEvent{
			ID:        EventID(e.Id),
			Payload:   EventPayload(e.Payload),
			Type:      EventType(e.Type),
			TimeStamp: e.Savetime})
	}

	latest = response.Latest

	return
}

func (es *GrpcEventStore) createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), es.Timeout)
}

func (es *GrpcEventStore) getClient() (storegrpc.EventStoreServiceClient, error) {

	if es.Client != nil {
		return es.Client, nil
	}

	conn, err := grpc.Dial(es.Config.Host, grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	es.Client = storegrpc.NewEventStoreServiceClient(conn)

	return es.Client, nil
}

// initializer for event store
func NewGrpcEventStore(config *GrpcEventStoreConfig) *GrpcEventStore {

	timeout := defaultTimeout

	if config.TimeoutInMillis > 0 {
		timeout = time.Duration(config.TimeoutInMillis) * time.Millisecond
	}

	return &GrpcEventStore{Config: config, Timeout: timeout}
}
