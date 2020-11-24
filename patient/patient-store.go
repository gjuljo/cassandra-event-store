package patient

import "my/esexample/store"

type patientEventStore struct {
	EventStore             store.EventStore
	EventTypeToEventMapper store.EventTypeToEventMapper
	EventToEventTypeMapper store.EventToEventTypeMapper
}

func NewPatientEventStore(store store.EventStore) *patientEventStore {
	return &patientEventStore{
		EventStore:             store,
		EventToEventTypeMapper: PatientEventTypeFromEvent,
		EventTypeToEventMapper: PatientEventFromType,
	}
}

func (es *patientEventStore) Find(guid string) (*Patient, error) {
	events, err := es.EventStore.Find(guid, es.EventTypeToEventMapper)

	if err != nil {
		return nil, err
	}

	p := NewFromEvents(events)
	return p, nil
}

func (es *patientEventStore) Update(p *Patient) error {
	return es.EventStore.Update(p.ID(), p.Version(), p.Events(), es.EventToEventTypeMapper)
}
