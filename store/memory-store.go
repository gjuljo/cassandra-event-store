package store

import (
	"fmt"
	"time"
)

type MemEventStore struct {
	eventsByGuid map[EventID][]StoreEvent
	eventsByType map[EventType][]StoreEvent
}

// @see EventStore.Find
func (es *MemEventStore) Find(guid EventID) ([]StoreEvent, error) {
	result := es.eventsByGuid[guid]
	return result, nil
}

// @see EventStore.Update
func (es *MemEventStore) Update(guid EventID, expectedVersion int, events []StoreEvent) error {

	// create a list of the event instance if missing
	eventsListByGuid, okByGuid := es.eventsByGuid[guid]
	if !okByGuid {
		// initialize if not exists
		eventsListByGuid = []StoreEvent{}
	}

	// naive implementation
	if len(eventsListByGuid) == expectedVersion {
		for _, e := range events {

			if e.TimeStamp == 0 {
				e.TimeStamp = time.Now().UnixNano() / int64(time.Millisecond)
			}

			es.eventsByGuid[guid] = append(es.eventsByGuid[guid], e)

			if evts, ok := es.eventsByType[e.Type]; ok {
				es.eventsByType[e.Type] = append(evts, e)
			} else {
				es.eventsByType[e.Type] = append([]StoreEvent{}, e)
			}
		}
	} else {
		return fmt.Errorf("OPTIMISTIC LOCKING EXCEPTION - client has version %v, but store %v", expectedVersion, len(eventsListByGuid))
	}
	return nil
}

// @see EventStore.GetEventsByType
func (es *MemEventStore) GetEventsByType(etype EventType, since int64, batchSize int) ([]StoreEvent, int64, error) {
	events := es.eventsByType[etype]
	result := []StoreEvent{}
	next := 0
	latestTime := int64(0)
	for _, e := range events {
		if since == 0 || e.TimeStamp > since {
			result = append(result, e)
			latestTime = e.TimeStamp
			next++
		}

		if next >= batchSize {
			break
		}
	}

	return result, latestTime, nil
}

// initializer for event store
func NewInMemStore() *MemEventStore {
	return &MemEventStore{
		eventsByGuid: map[EventID][]StoreEvent{},
		eventsByType: map[EventType][]StoreEvent{},
	}
}
