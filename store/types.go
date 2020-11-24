package store

// Event is a domain event marker.
type Event interface {
	IsEvent()
}

type EventType int
type EventToEventTypeMapper func(e Event) (EventType, error)
type EventTypeToEventMapper func(e EventType) (Event, error)

// Command is a domain event marker.
type Command interface {
	IsCommand()
}
