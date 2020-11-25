package main

import (
	"log"
	"my/esexample/patient"
	"my/esexample/store"
	"time"

	"github.com/google/uuid"
)

func main() {
	log.Println("PATIENT TEST")

	guid1 := uuid.New()

	// store := store.NewInMemStore()
	store, _ := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
		Hosts:       []string{"localhost"},
		Keyspace:    "eventstore",
		WriteQuorum: "QUORUM",
		ReadQuorum:  "LOCAL_QUORUM",
	})

	cmdhandler := patient.NewPatientCommandHandler(patient.NewPatientEventStore(store))

	admitPatient(*cmdhandler, guid1.String(), "John Doe", 34, "AA")
	transferPatient(*cmdhandler, guid1.String(), "BB")
	transferPatient(*cmdhandler, guid1.String(), "CC")

	showPatient(guid1.String(), store)

	guid2 := uuid.New()

	admitPatient(*cmdhandler, guid2.String(), "Jane Doe", 29, "AA")

	guid3 := uuid.New()

	admitPatient(*cmdhandler, guid3.String(), "Johnny Doe", 12, "BB")

	calcNumberOfPatients(store)

	dischargePatient(*cmdhandler, guid3.String())

	calcNumberOfPatients(store)

	dischargePatient(*cmdhandler, guid2.String())

	calcNumberOfPatients(store)

	// VERIFY OPTIMISTICK LOCKING ENFORCEMENT
	showPatient(guid1.String(), store)

	evt, _ := store.Find(guid1.String(), patient.PatientEventFromType)
	patient1 := patient.NewFromEvents(evt)

	patient1.Transfer("XX")
	log.Printf("patient1 version %v\n", patient1)

	showPatient(guid1.String(), store)

	transferPatient(*cmdhandler, guid1.String(), "DD")

	showPatient(guid1.String(), store)

	// try to store the user, expect an error
	if err := store.Update(patient1.ID(), patient1.Version(), patient1.Events(), patient.PatientEventTypeFromEvent); err != nil {
		log.Printf("ERROR, unable to update patient1: %+v", err)
	}

	showPatient(guid1.String(), store)
}

func admitPatient(cmdhandler patient.PatientCommandHandler, uuid string, name string, age int, ward string) {
	if err := cmdhandler.HandleAdmitPatient(&patient.AdmitPatient{
		ID:   uuid,
		Name: patient.Name(name),
		Age:  patient.Age(age),
		Ward: patient.WardNumber(ward),
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}
}

func transferPatient(cmdhandler patient.PatientCommandHandler, uuid string, ward string) {
	if err := cmdhandler.HandleTransferPatient(&patient.TransferPatient{
		ID:            uuid,
		NewWardNumber: patient.WardNumber(ward),
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}
}

func dischargePatient(cmdhandler patient.PatientCommandHandler, uuid string) {
	cmdhandler.HandleDischargePatient(&patient.DischargePatient{
		ID: uuid,
	})
}

func showPatient(id string, store store.EventStore) {
	events, _ := store.Find(id, patient.PatientEventFromType)
	p := patient.NewFromEvents(events)
	log.Printf("%+v\n", p)
}

func calcNumberOfPatients(store store.EventStore) {
	since := time.Now().UnixNano()/int64(time.Millisecond) - int64(10000) // millis
	log.Printf("SINCE %v = %v\n", since, time.Unix(0, since*int64(time.Millisecond)))

	admittedEvents, _, _ := store.GetEventsByType(patient.PatientAdmittedEventType, since, 100, patient.PatientEventFromType)
	dichargedEvents, _, _ := store.GetEventsByType(patient.PatientDischargedEventType, since, 100, patient.PatientEventFromType)

	admitted := len(admittedEvents)
	dicharged := len(dichargedEvents)

	log.Printf("TOTAL PATIENTS ADMITTED = %d\n", admitted)
	log.Printf("TOTAL PATIENTS DISCHARGED = %d\n", dicharged)

	log.Printf("TOTAL PATIENTS = %d\n", admitted-dicharged)
}
