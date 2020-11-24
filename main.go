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
		Hosts:    []string{"localhost"},
		Keyspace: "eventstore",
	})

	cmdhandler := patient.NewPatientCommandHandler(patient.NewPatientEventStore(store))

	if err := cmdhandler.HandleAdmitPatient(&patient.AdmitPatient{
		ID:   guid1.String(),
		Name: "John Doe",
		Age:  33,
		Ward: "AA",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	if err := cmdhandler.HandleTransferPatient(&patient.TransferPatient{
		ID:            guid1.String(),
		NewWardNumber: "BB",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	if err := cmdhandler.HandleTransferPatient(&patient.TransferPatient{
		ID:            guid1.String(),
		NewWardNumber: "CC",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	showPatient(guid1.String(), store)

	guid2 := uuid.New()

	if err := cmdhandler.HandleAdmitPatient(&patient.AdmitPatient{
		ID:   guid2.String(),
		Name: "Pinco Pallino",
		Age:  22,
		Ward: "AA",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	guid3 := uuid.New()

	if err := cmdhandler.HandleAdmitPatient(&patient.AdmitPatient{
		ID:   guid3.String(),
		Name: "Paolino Paperino",
		Age:  33,
		Ward: "BB",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	calcNumberOfPatients(store)

	cmdhandler.HandleDischargePatient(&patient.DischargePatient{
		ID: guid3.String(),
	})

	calcNumberOfPatients(store)

	cmdhandler.HandleDischargePatient(&patient.DischargePatient{
		ID: guid2.String(),
	})

	calcNumberOfPatients(store)

	// GENERIAMO UN PROBLEMA DI CONCORRENZA

	showPatient(guid1.String(), store)

	evt, _ := store.Find(guid1.String(), patient.PatientEventFromType)
	patient1 := patient.NewFromEvents(evt)

	patient1.Transfer("XX")
	log.Printf("patient1 version %v\n", patient1)

	showPatient(guid1.String(), store)

	if err := cmdhandler.HandleTransferPatient(&patient.TransferPatient{
		ID:            guid1.String(),
		NewWardNumber: "DD",
	}); err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	showPatient(guid1.String(), store)

	if err := store.Update(patient1.ID(), patient1.Version(), patient1.Events(), patient.PatientEventTypeFromEvent); err != nil {
		log.Printf("ERROR, unable to update patient1: %+v", err)
	}

	showPatient(guid1.String(), store)
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
