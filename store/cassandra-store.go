package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

type CassandraEventStoreConfig struct {
	Keyspace    string
	WriteQuorum string
	ReadQuorum  string
	Hosts       []string
	Auth        struct {
		Username string
		Password string
	}
}

type CassandraEventStore struct {
	session     *gocql.Session
	config      *CassandraEventStoreConfig
	readQuorum  gocql.Consistency
	writeQuorum gocql.Consistency
}

// @see EventStore.Find
func (es *CassandraEventStore) Find(guid string, mapper EventTypeToEventMapper) ([]Event, error) {
	var events []Event
	var jevent string
	var etype EventType

	iter := es.session.
		Query(`SELECT type, payload FROM events WHERE id = ?`, guid).
		Consistency(es.readQuorum).
		Iter()

	for iter.Scan(&etype, &jevent) {
		event, err := mapper(etype) // TODO: handle error

		// unknown event type
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(jevent), &event)

		// unable to unmarshal the events
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (es *CassandraEventStore) Update(guid string, expectedVersion int, events []Event, mapper EventToEventTypeMapper) error {
	batch := es.session.NewBatch(gocql.UnloggedBatch)
	quorum := es.writeQuorum
	numbEvents := len(events)
	newVersion := expectedVersion + numbEvents

	batch.SetConsistency(quorum)
	if expectedVersion == 0 {
		batch.Query("INSERT INTO events (id, current_version) VALUES (?,?) IF NOT EXISTS", guid, numbEvents)
	} else {
		batch.Query("UPDATE events SET current_version = ? WHERE id = ? IF current_version = ?", newVersion, guid, expectedVersion)
	}

	stmt := "INSERT INTO events (id, version, type, payload, savetime) VALUES (?,?,?,?,toTimeStamp(now()))"

	for i, event := range events {
		etype, err := mapper(event)

		// error, uknown event type
		if err != nil {
			return err
		}

		jevent, err := json.Marshal(event)

		// error, unable to marshal json
		if err != nil {
			return fmt.Errorf("CQL ERROR: %+v", err)
		}

		eventVersion := expectedVersion + 1 + i
		batch.Query(stmt, guid, eventVersion, etype, jevent)
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

func (es *CassandraEventStore) GetEventsByType(etype EventType, sinceMillis int64, batchSize int, mapper EventTypeToEventMapper) (events []Event, latest int64, theError error) {
	var jevent string
	var event Event
	var query *gocql.Query

	if batchSize <= 0 {
		batchSize = 1000 // TODO: set a default value at CassandraEventStore level
	}

	if sinceMillis > 0 {
		query = es.session.Query(`SELECT savetime, payload FROM events_by_type WHERE type=? AND savetime > ? LIMIT ?`, etype, sinceMillis, batchSize)
	} else {
		query = es.session.Query(`SELECT savetime, payload FROM events_by_type WHERE type=? LIMIT ?`, etype, batchSize)
	}

	iter := query.Consistency(es.readQuorum).Iter()

	for iter.Scan(&latest, &jevent) {
		event, theError = mapper(etype)

		// block if uknown event type
		if theError != nil {
			return
		}

		theError = json.Unmarshal([]byte(jevent), &event)

		// error, unable to unmarhal json
		if theError != nil {
			return
		}

		events = append(events, event)
	}

	return events, latest, nil
}

// initializer for event store
func NewCassandraEventStore(config *CassandraEventStoreConfig) (*CassandraEventStore, error) {

	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
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

	return &CassandraEventStore{session: session, config: config, readQuorum: readQuorum, writeQuorum: writeQuorum}, nil
}

// Add events to the store and send them down the channel
func (es *CassandraEventStore) Dispose() {
	es.session.Close()
}
