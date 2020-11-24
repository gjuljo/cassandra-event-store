package patient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	p := New("uuid", "name", 66, "ward1")

	nevents := len(p.Events())

	if nevents != 1 {
		t.Errorf("wrong number of events, got: %d, expected: %d.", nevents, 1)
	}

	event := &PatientAdmitted{
		ID:   "uuid",
		Name: "name",
		Age:  66,
		Ward: "ward1",
	}

	assert.Equal(t, p.Events()[0], event)
}

func TestTansferred(t *testing.T) {
	p := New("uuid", "name", 66, "ward1")

	if err := p.Transfer("ward2"); err != nil {
		t.Errorf("error transfering patient: %+v", err)
	}

	if err := p.Transfer("ward3"); err != nil {
		t.Errorf("error transfering patient: %+v", err)
	}

	nevents := len(p.Events())

	if nevents != 3 {
		t.Errorf("wrong number of events, got: %d, expected: %d.", nevents, 3)
	}

	event1 := &PatientAdmitted{
		ID:   "uuid",
		Name: "name",
		Age:  66,
		Ward: "ward1",
	}

	event2 := &PatientTransferred{
		ID:            "uuid",
		NewWardNumber: "ward2",
	}

	event3 := &PatientTransferred{
		ID:            "uuid",
		NewWardNumber: "ward3",
	}

	assert.Equal(t, p.Events()[0], event1)
	assert.Equal(t, p.Events()[1], event2)
	assert.Equal(t, p.Events()[2], event3)
}

func TestDischarged(t *testing.T) {
	p := New("uuid", "name", 66, "ward1")

	if err := p.Discharge(); err != nil {
		t.Errorf("error discharghing patient: %+v", err)
	}

	nevents := len(p.Events())

	if nevents != 2 {
		t.Errorf("wrong number of events, got: %d, expected: %d.", nevents, 2)
	}

	event1 := &PatientAdmitted{
		ID:   "uuid",
		Name: "name",
		Age:  66,
		Ward: "ward1",
	}

	event2 := &PatientDischarged{
		ID: "uuid",
	}

	assert.Equal(t, p.Events()[0], event1)
	assert.Equal(t, p.Events()[1], event2)
}

func TestCannotDischargeTwice(t *testing.T) {
	p := New("uuid", "name", 66, "ward1")

	if err := p.Discharge(); err != nil {
		t.Errorf("error discharghing patient: %+v", err)
	}

	nevents := len(p.Events())

	if nevents != 2 {
		t.Errorf("wrong number of events, got: %d, expected: %d.", nevents, 2)
	}

	event1 := &PatientAdmitted{
		ID:   "uuid",
		Name: "name",
		Age:  66,
		Ward: "ward1",
	}

	event2 := &PatientDischarged{
		ID: "uuid",
	}

	assert.Equal(t, p.Events()[0], event1)
	assert.Equal(t, p.Events()[1], event2)

	if err := p.Discharge(); err != ErrPatientDischarged {
		t.Errorf("expected error %+v, found %+v", ErrPatientDischarged, err)
	}

}

func TestCannotTransferDischargedPatient(t *testing.T) {
	p := New("uuid", "name", 66, "ward1")

	if err := p.Discharge(); err != nil {
		t.Errorf("error discharghing patient: %+v", err)
	}

	nevents := len(p.Events())

	if nevents != 2 {
		t.Errorf("wrong number of events, got: %d, expected: %d.", nevents, 2)
	}

	event1 := &PatientAdmitted{
		ID:   "uuid",
		Name: "name",
		Age:  66,
		Ward: "ward1",
	}

	event2 := &PatientDischarged{
		ID: "uuid",
	}

	assert.Equal(t, p.Events()[0], event1)
	assert.Equal(t, p.Events()[1], event2)

	if err := p.Transfer("ward2"); err != ErrPatientDischarged {
		t.Errorf("expected error %+v, found %+v", ErrPatientDischarged, err)
	}

}
