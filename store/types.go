package store

import (
	"encoding/json"
	"fmt"
)

// Command is a domain event marker.
type Command interface {
	IsCommand()
}

// Event is a domain event marker.
type Event interface {
	GetEventType() EventType
}

type EventID string
type EventPayload string
type EventType int
type EventTypeToEventMapper func(e EventType) (Event, error)

type StoreEvent struct {
	ID        EventID      `json:"id"`
	Payload   EventPayload `json:"payload"`
	Type      EventType    `json:"type"`
	TimeStamp int64        `json:"time"`
}

func GetEventTypeFromJSON(e string) (EventType, error) {
	// get the type from the event
	var tmp map[string]interface{}

	if err := json.Unmarshal([]byte(e), &tmp); err != nil {
		return 0, err
	}

	t, ok := tmp["type"].(float64)

	if !ok {
		return 0, fmt.Errorf("Unable to find field \"type\" in input %s", e)
	}

	return EventType(t), nil
}
