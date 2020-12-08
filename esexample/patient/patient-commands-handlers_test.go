package patient

import (
	"my/esexample/store"
	"testing"
)

func TestNewByHandler(t *testing.T) {

	cmdhandler := NewPatientCommandHandler(NewPatientEventStore(store.NewInMemStore()))

	command := &AdmitPatient{
		ID:   "uuid1",
		Name: "John Doe",
		Age:  33,
		Ward: "AA",
	}

	if err := cmdhandler.HandleAdmitPatient(command); err != nil {
		t.Errorf("unexpected error %+v", err)
	}
}

func TestTransferByHandler(t *testing.T) {

	cmdhandler := NewPatientCommandHandler(NewPatientEventStore(store.NewInMemStore()))

	adminCommand := &AdmitPatient{
		ID:   "uuid1",
		Name: "John Doe",
		Age:  33,
		Ward: "AA",
	}

	if err := cmdhandler.HandleAdmitPatient(adminCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	transferCommand := &TransferPatient{
		ID:            "uuid1",
		NewWardNumber: "BB",
	}

	if err := cmdhandler.HandleTransferPatient(transferCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}
}

func TestDischargeByHandler(t *testing.T) {

	cmdhandler := NewPatientCommandHandler(NewPatientEventStore(store.NewInMemStore()))

	adminCommand := &AdmitPatient{
		ID:   "uuid1",
		Name: "John Doe",
		Age:  33,
		Ward: "AA",
	}

	if err := cmdhandler.HandleAdmitPatient(adminCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	dischargeCommand := &DischargePatient{
		ID: "uuid1",
	}

	if err := cmdhandler.HandleDischargePatient(dischargeCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}
}

func TestDoubleDischargeByHandler(t *testing.T) {

	cmdhandler := NewPatientCommandHandler(NewPatientEventStore(store.NewInMemStore()))

	adminCommand := &AdmitPatient{
		ID:   "uuid1",
		Name: "John Doe",
		Age:  33,
		Ward: "AA",
	}

	if err := cmdhandler.HandleAdmitPatient(adminCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	dischargeCommand := &DischargePatient{
		ID: "uuid1",
	}

	if err := cmdhandler.HandleDischargePatient(dischargeCommand); err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	if err := cmdhandler.HandleDischargePatient(dischargeCommand); err != ErrPatientDischarged {
		t.Errorf("unexpected error ErrPatientDischarged, got %+v", err)
	}
}
