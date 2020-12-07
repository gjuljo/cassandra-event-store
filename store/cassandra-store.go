package store

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gocql/gocql"
)

type CassandraEventStoreConfig struct {
	Keyspace    string
	WriteQuorum string
	ReadQuorum  string
	Hosts       []string
	Port        int
	Auth        struct {
		Username string
		Password string
	}
	TraceSession bool
}

type CassandraEventStore struct {
	session     *gocql.Session
	config      *CassandraEventStoreConfig
	readQuorum  gocql.Consistency
	writeQuorum gocql.Consistency
}

// @see EventStore.Find
func (es *CassandraEventStore) Find(guid EventID) ([]StoreEvent, error) {
	var events []StoreEvent
	var event string
	var etype int

	// ATTENTION: we need to parse the guid into the atual type we use in the table
	stringGuid := string(guid)

	iter := es.session.
		Query(`SELECT type, payload FROM events WHERE id = ?`, stringGuid).
		Consistency(es.readQuorum).
		Iter()

	for iter.Scan(&etype, &event) {
		events = append(events, StoreEvent{ID: EventID(guid), Type: EventType(etype), Payload: EventPayload(event)})
	}

	return events, nil
}

func (es *CassandraEventStore) Update(guid EventID, expectedVersion int, events []StoreEvent) error {
	batch := es.session.NewBatch(gocql.UnloggedBatch)
	quorum := es.writeQuorum
	numbEvents := len(events)
	newVersion := expectedVersion + numbEvents

	// ATTENTION: we need to parse the guid into the actual type we use in the table
	stringGuid := string(guid)

	batch.SetConsistency(quorum)
	if expectedVersion == 0 {
		batch.Query("INSERT INTO events (id, current_version) VALUES (?,?) IF NOT EXISTS", stringGuid, numbEvents)
	} else {
		batch.Query("UPDATE events SET current_version = ? WHERE id = ? IF current_version = ?", newVersion, stringGuid, expectedVersion)
	}

	stmt := "INSERT INTO events (id, version, type, payload, savetime) VALUES (?,?,?,?,toTimeStamp(now()))"

	for i, event := range events {
		eventVersion := expectedVersion + 1 + i
		batch.Query(stmt, stringGuid, eventVersion, event.Type, event.Payload)
	}

	// here we can get an error only if we are unable to run the query or it is invalid
	casMap := make(map[string]interface{})
	applied, _, err := es.session.MapExecuteBatchCAS(batch, casMap)

	if err != nil {
		return fmt.Errorf("CQL ERROR: %+v", err)
	}

	if !applied {
		return fmt.Errorf("%+v", "optimistic locking failed")
	}

	return nil
}

func (es *CassandraEventStore) GetEventsByType(etype EventType, sinceMillis int64, batchSize int) (events []StoreEvent, latest int64, theError error) {
	var payload string
	var id string
	var query *gocql.Query

	if batchSize <= 0 {
		batchSize = 1000 // TODO: set a default value at CassandraEventStore level
	}

	if sinceMillis > 0 {
		query = es.session.Query(`SELECT savetime, payload, id FROM events_by_type WHERE type=? AND savetime > ? LIMIT ?`, etype, sinceMillis, batchSize)
	} else {
		query = es.session.Query(`SELECT savetime, payload, id FROM events_by_type WHERE type=? LIMIT ?`, etype, batchSize)
	}

	iter := query.Consistency(es.readQuorum).Iter()

	for iter.Scan(&latest, &payload, &id) {
		events = append(events, StoreEvent{ID: EventID(id), Type: etype, Payload: EventPayload(payload), TimeStamp: latest})
	}

	return events, latest, nil
}

// initializer for event store
func NewCassandraEventStore(config *CassandraEventStoreConfig) (*CassandraEventStore, error) {

	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace

	// set port if provided
	if config.Port > 0 {
		cluster.Port = config.Port
	}

	cluster.Consistency = gocql.Quorum
	// cluster.SerialConsistency = gocql.LocalSerial

	var readQuorum gocql.Consistency
	if err := readQuorum.UnmarshalText([]byte(strings.ToUpper(config.ReadQuorum))); err != nil {
		readQuorum = gocql.Quorum
	}

	var writeQuorum gocql.Consistency
	if err := writeQuorum.UnmarshalText([]byte(strings.ToUpper(config.WriteQuorum))); err != nil {
		writeQuorum = gocql.Quorum
	}

	session, err := cluster.CreateSession()

	if err != nil {
		return nil, err
	}

	if config.TraceSession {
		tracer := gocql.NewTraceWriter(session, log.With().Logger().Level(zerolog.InfoLevel))
		session.SetTrace(tracer)
	}

	return &CassandraEventStore{session: session, config: config, readQuorum: readQuorum, writeQuorum: writeQuorum}, nil
}

// Add events to the store and send them down the channel
func (es *CassandraEventStore) Dispose() {
	es.session.Close()
}
