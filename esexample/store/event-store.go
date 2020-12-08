package store

type EventStore interface {
	// find all events for given ID (aggregate).
	// returns event list as well as aggregate version
	Find(guid EventID) ([]StoreEvent, error)

	// Update an aggregate with new events. If the version specified
	// does not match with the version in the Event Store, an error is returned
	Update(guid EventID, expectedVersion int, events []StoreEvent) error

	// Get events of a given type from Event Store
	GetEventsByType(etype EventType, since int64, batchSize int) ([]StoreEvent, int64, error)
}
