package store

type EventStore interface {
	// find all events for given ID (aggregate).
	// returns event list as well as aggregate version
	Find(guid string, mapper EventTypeToEventMapper) ([]Event, error)

	// Update an aggregate with new events. If the version specified
	// does not match with the version in the Event Store, an error is returned
	Update(guid string, expectedVersion int, events []Event, mapper EventToEventTypeMapper) error

	// Get events of a given type from Event Store
	GetEventsByType(etype EventType, since int64, batchSize int, mapper EventTypeToEventMapper) ([]Event, int64, error)
}
