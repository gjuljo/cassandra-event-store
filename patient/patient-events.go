package patient

import (
	"fmt"
	"my/esexample/store"
)

const (
	PatientAdmittedEventType store.EventType = iota + 1
	PatientTransferredEventType
	PatientDischargedEventType
)

func PatientEventFromType(e store.EventType) (store.Event, error) {
	switch e {
	case PatientAdmittedEventType:
		return &PatientAdmitted{}, nil
	case PatientTransferredEventType:
		return &PatientTransferred{}, nil
	case PatientDischargedEventType:
		return &PatientDischarged{}, nil
	}

	return nil, fmt.Errorf("Event type not found %d", e)
}

func PatientEventTypeFromEvent(e store.Event) (store.EventType, error) {
	switch e.(type) {
	case *PatientAdmitted:
		return PatientAdmittedEventType, nil
	case *PatientTransferred:
		return PatientTransferredEventType, nil
	case *PatientDischarged:
		return PatientDischargedEventType, nil
	}

	return -1, fmt.Errorf("Event type %+v not enumerated", e)
}

func (e PatientAdmitted) IsEvent()    {}
func (e PatientTransferred) IsEvent() {}
func (e PatientDischarged) IsEvent()  {}

// PatientAdmitted event.
type PatientAdmitted struct {
	ID   string     `json:"id"`
	Name Name       `json:"name"`
	Ward WardNumber `json:"ward"`
	Age  Age        `json:"age"`
}

// PatientTransferred event.
type PatientTransferred struct {
	ID            string     `json:"id"`
	NewWardNumber WardNumber `json:"new_ward"`
}

// PatientDischarged event.
type PatientDischarged struct {
	ID string `json:"id"`
}
