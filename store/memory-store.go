package store

import (
	"fmt"
	"time"
)

// MemEventStore
// In-memory event store. Uses slice for 'complete events catalogue'
// and a map for 'per aggregate' events

type eventWithTime struct {
	eventTime int64
	event     Event
}

type MemEventStore struct {
	eventsByGuid map[string][]Event
	eventsByType map[EventType][]eventWithTime
}

// @see EventStore.Find
func (es *MemEventStore) Find(guid string, _ EventTypeToEventMapper) ([]Event, error) {
	events := es.eventsByGuid[guid]
	return events, nil
}

// @see EventStore.Update
func (es *MemEventStore) Update(guid string, expectedVersion int, events []Event, mapper EventToEventTypeMapper) error {

	eventsListByGuid, okByGuid := es.eventsByGuid[guid]
	if !okByGuid {
		// initialize if not exists
		eventsListByGuid = []Event{}
	}

	// naive implementation
	if len(eventsListByGuid) == expectedVersion {
		for _, e := range events {
			if etype, err := mapper(e); err == nil {
				millis := time.Now().UnixNano() / int64(time.Millisecond)

				if evts, ok := es.eventsByType[etype]; ok {
					es.eventsByType[etype] = append(evts, eventWithTime{eventTime: millis, event: e})
				} else {
					es.eventsByType[etype] = append([]eventWithTime{}, eventWithTime{eventTime: millis, event: e})
				}
			} else {
				return fmt.Errorf("UNKNOWN ERROR TYPE %v", etype)
			}
		}

		es.eventsByGuid[guid] = append(eventsListByGuid, events...)

	} else {
		return fmt.Errorf("OPTIMISTIC LOCKING EXCEPTION - client has version %v, but store %v", expectedVersion, len(eventsListByGuid))
	}
	return nil
}

// @see EventStore.GetEventsByType
func (es *MemEventStore) GetEventsByType(etype EventType, since int64, batchSize int, mapper EventTypeToEventMapper) ([]Event, int64, error) {
	events := es.eventsByType[etype]
	result := []Event{}
	next := 0
	latestTime := int64(0)
	for _, e := range events {
		if since == 0 || e.eventTime > since {
			result = append(result, e.event)
			latestTime = e.eventTime
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
		eventsByGuid: map[string][]Event{},
		eventsByType: map[EventType][]eventWithTime{},
	}
}
