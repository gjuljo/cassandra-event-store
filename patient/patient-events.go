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

func (e PatientAdmitted) GetEventType() store.EventType    { return PatientAdmittedEventType }
func (e PatientTransferred) GetEventType() store.EventType { return PatientTransferredEventType }
func (e PatientDischarged) GetEventType() store.EventType  { return PatientDischargedEventType }

// PatientAdmitted event.
type PatientAdmitted struct {
	ID   store.EventID `json:"id"`
	Name Name          `json:"name"`
	Ward WardNumber    `json:"ward"`
	Age  Age           `json:"age"`
}

// PatientTransferred event.
type PatientTransferred struct {
	ID            store.EventID `json:"id"`
	NewWardNumber WardNumber    `json:"new_ward"`
}

// PatientDischarged event.
type PatientDischarged struct {
	ID store.EventID `json:"id"`
}
