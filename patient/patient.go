package patient

import (
	"errors"
	"my/esexample/store"
)

var ErrPatientDischarged = errors.New("patient already discharged")

// type ID uuid.UUID
type WardNumber string
type Name string
type Age int
type Version int

// Patient aggregate.
type Patient struct {
	id         string
	ward       WardNumber
	name       Name
	age        Age
	discharged bool

	changes []store.Event
	version int
}

// NewFromEvents is a helper method that creates a new patient
// from a series of events.
func NewFromEvents(events []store.Event) *Patient {
	p := &Patient{}

	for _, event := range events {
		p.On(event, false)
	}

	return p
}

// Ward returns the patient's ward number.
func (p Patient) Ward() WardNumber {
	return p.ward
}

// Name returns the patient's name.
func (p Patient) Name() Name {
	return p.name
}

// Age returns the patient's age.
func (p Patient) Age() Age {
	return p.age
}

// Discharged returns wether or not the patient has been discharged.
func (p Patient) Discharged() bool {
	return p.discharged
}

// ID returns the id of the patient. Duh.
func (p Patient) ID() string {
	return p.id
}

// New creates a new Patient from id, name, age and ward number.
func New(id string, name Name, age Age, ward WardNumber) *Patient {
	p := &Patient{}

	p.raise(&PatientAdmitted{
		ID:   id,
		Name: name,
		Age:  age,
		Ward: ward,
	})

	return p
}

// Transfer transfers a patient to a new ward.
func (p *Patient) Transfer(newWard WardNumber) error {
	if p.discharged {
		return ErrPatientDischarged
	}

	p.raise(&PatientTransferred{
		ID:            p.id,
		NewWardNumber: newWard,
	})

	return nil
}

// Discharge discharges a patient
func (p *Patient) Discharge() error {
	if p.discharged {
		return ErrPatientDischarged
	}

	p.raise(&PatientDischarged{
		ID: p.id,
	})

	return nil
}

// On handles patient events on the patient aggregate.
func (p *Patient) On(event store.Event, new bool) {
	switch e := event.(type) {
	case *PatientAdmitted:
		p.id = e.ID
		p.name = e.Name
		p.age = e.Age
		p.ward = e.Ward
		p.version = 0

	case *PatientDischarged:
		p.discharged = true

	case *PatientTransferred:
		p.ward = e.NewWardNumber
	}

	if !new {
		p.version++
	}
}

// Events returns the uncommitted events from the patient aggregate.
func (p Patient) Events() []store.Event {
	return p.changes
}

// Version returns the last version of the aggregate before changes.
func (p Patient) Version() int {
	return p.version
}

func (p *Patient) raise(event store.Event) {
	p.changes = append(p.changes, event)
	p.On(event, true)
}
