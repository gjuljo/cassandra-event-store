package patient

import (
	"my/esexample/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatientInStore(t *testing.T) {
	// create patient
	pstore := NewPatientEventStore(store.NewInMemStore())
	pnew := New("uuid", "name", 66, "ward1")

	if err := pstore.Update(pnew); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	pfind, err := pstore.Find(store.EventID("uuid"))

	if err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	assert.Equal(t, pfind.ID(), pnew.ID())
	assert.Equal(t, pfind.Name(), pnew.Name())
	assert.Equal(t, pfind.Age(), pnew.Age())
	assert.Equal(t, pfind.Ward(), pnew.Ward())

	// transfer patient
	if err := pfind.Transfer("ward2"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pstore.Update(pfind); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	pfind, err = pstore.Find(store.EventID("uuid"))

	if err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	assert.Equal(t, pfind.Ward(), WardNumber("ward2"))

	// tranfer patient again
	if err := pfind.Transfer("ward3"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pfind.Transfer("ward4"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pstore.Update(pfind); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	pfind, err = pstore.Find(store.EventID("uuid"))

	if err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	assert.Equal(t, pfind.Ward(), WardNumber("ward4"))

	// discharge patient
	if err := pfind.Transfer("ward5"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pfind.Discharge(); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pstore.Update(pfind); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	pfind, err = pstore.Find(store.EventID("uuid"))

	if err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	assert.Equal(t, pfind.Discharged(), true)
}

func TestOptimisticLockingInStore(t *testing.T) {

	// create patient
	pstore := NewPatientEventStore(store.NewInMemStore())
	pnew := New("uuid", "name", 66, "ward1")

	if err := pstore.Update(pnew); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	pfind, err := pstore.Find(store.EventID("uuid"))

	if err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pfind.Transfer("ward2"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	if err := pstore.Update(pfind); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	// try to update the old patient
	if err := pfind.Transfer("ward3"); err != nil {
		t.Errorf("got expected error: %+v", err)
	}

	// expect optimistic locking error
	if err := pstore.Update(pfind); err == nil {
		t.Errorf("expected optimistic lock error")
	}
}
