package patient

import (
	"encoding/json"
	"my/esexample/store"
)

type patientEventStore struct {
	EventStore             store.EventStore
	EventTypeToEventMapper store.EventTypeToEventMapper
}

func NewPatientEventStore(store store.EventStore) *patientEventStore {
	return &patientEventStore{
		EventStore:             store,
		EventTypeToEventMapper: PatientEventFromType,
	}
}

func (es *patientEventStore) Find(guid store.EventID) (*Patient, error) {
	var storeEvents []store.Event
	events, err := es.EventStore.Find(guid)

	if err != nil {
		return nil, err
	}

	for _, e := range events {

		tmp, err := es.EventTypeToEventMapper(e.Type)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(e.Payload), &tmp)

		if err != nil {
			return nil, err
		}

		storeEvents = append(storeEvents, tmp)
	}

	p := NewFromEvents(storeEvents)
	return p, nil
}

func (es *patientEventStore) Update(p *Patient) error {
	var events []store.StoreEvent

	id := store.EventID(p.ID())

	for _, e := range p.Events() {

		b, err := json.Marshal(e)

		if err != nil {
			return err
		}

		events = append(events, store.StoreEvent{Payload: store.EventPayload(b), Type: e.GetEventType(), ID: id})
	}

	return es.EventStore.Update(id, p.Version(), events)
}
