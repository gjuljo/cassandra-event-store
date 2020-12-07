package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"my/esexample/patient"
	"my/esexample/store"
	"time"

	"github.com/google/uuid"
)

func mainShort() {
	initLog()
	log.Info().Msg("PATIENT TEST")

	guid1 := uuid.New()

	// eventstore := store.NewInMemStore()

	/*
		eventstore, _ := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
			Hosts:       []string{"localhost"},
			Keyspace:    "eventstore",
			WriteQuorum: "QUORUM",
			ReadQuorum:  "LOCAL_QUORUM",
		})
	*/

	// eventstore := store.NewRemoteEventStore(&store.RemoteEventStoreConfig{Host: "http://localhost:8091"})

	eventstore := store.NewGrpcEventStore(&store.GrpcEventStoreConfig{Host: "localhost:8091"})

	pstore := patient.NewPatientEventStore(eventstore)
	cmdhandler := patient.NewPatientCommandHandler(pstore)

	admitPatient(cmdhandler, guid1.String(), "John Doe", 34, "AA")
	transferPatient(cmdhandler, guid1.String(), "BB")
}

func main() {
	initLog()
	log.Info().Msg("PATIENT TEST")

	guid1 := uuid.New()

	eventstore := store.NewInMemStore()

	/*
		eventstore, _ := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
			Hosts:       []string{"localhost"},
			Keyspace:    "eventstore",
			WriteQuorum: "QUORUM",
			ReadQuorum:  "LOCAL_QUORUM",
		})
	*/

	// eventstore := store.NewRemoteEventStore(&store.RemoteEventStoreConfig{Host: "http://localhost:8091"})

	// eventstore := store.NewGrpcEventStore(&store.GrpcEventStoreConfig{Host: "localhost:8091"})

	pstore := patient.NewPatientEventStore(eventstore)
	cmdhandler := patient.NewPatientCommandHandler(pstore)

	admitPatient(cmdhandler, guid1.String(), "John Doe", 34, "AA")
	transferPatient(cmdhandler, guid1.String(), "BB")
	transferPatient(cmdhandler, guid1.String(), "CC")

	guid2 := uuid.New()

	admitPatient(cmdhandler, guid2.String(), "Jane Doe", 29, "AA")

	guid3 := uuid.New()

	admitPatient(cmdhandler, guid3.String(), "Johnny Doe", 12, "BB")

	calcNumberOfPatients(eventstore)

	dischargePatient(cmdhandler, guid3.String())

	calcNumberOfPatients(eventstore)

	dischargePatient(cmdhandler, guid2.String())

	calcNumberOfPatients(eventstore)

	// VERIFY OPTIMISTICK LOCKING ENFORCEMENT
	patient1, _ := pstore.Find(store.EventID(guid1.String()))

	patient1.Transfer("XX")
	log.Info().Msgf("patient1 version %v", patient1)

	transferPatient(cmdhandler, guid1.String(), "DD")

	// try to store the user, expect an error
	if err := pstore.Update(patient1); err != nil {
		log.Info().Msgf("ERROR, unable to update patient1: %+v", err)
	}
}

func admitPatient(cmdhandler *patient.PatientCommandHandler, uuid string, name string, age int, ward string) {
	if err := cmdhandler.HandleAdmitPatient(&patient.AdmitPatient{
		ID:   uuid,
		Name: patient.Name(name),
		Age:  patient.Age(age),
		Ward: patient.WardNumber(ward),
	}); err != nil {
		log.Info().Msgf("ERROR: %+v", err)
	}
}

func transferPatient(cmdhandler *patient.PatientCommandHandler, uuid string, ward string) {
	if err := cmdhandler.HandleTransferPatient(&patient.TransferPatient{
		ID:            uuid,
		NewWardNumber: patient.WardNumber(ward),
	}); err != nil {
		log.Info().Msgf("ERROR: %+v", err)
	}
}

func dischargePatient(cmdhandler *patient.PatientCommandHandler, uuid string) {
	cmdhandler.HandleDischargePatient(&patient.DischargePatient{
		ID: uuid,
	})
}

func calcNumberOfPatients(store store.EventStore) {
	nowInMillis := time.Now().UnixNano() / int64(time.Millisecond)
	since := nowInMillis - 1*time.Second.Milliseconds()

	log.Info().Msgf("SINCE %v = %v", since, time.Unix(0, since*int64(time.Millisecond)))

	admittedEvents, _, _ := store.GetEventsByType(patient.PatientAdmittedEventType, since, 100)
	dichargedEvents, _, _ := store.GetEventsByType(patient.PatientDischargedEventType, since, 100)

	admitted := len(admittedEvents)
	dicharged := len(dichargedEvents)

	log.Info().Msgf("PATIENTS ADMITTED   = %d", admitted)
	log.Info().Msgf("PATIENTS DISCHARGED = %d", dicharged)
	log.Info().Msgf("TOTAL PATIENTS      = %d", admitted-dicharged)
}

func initLog() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: "2006-01-02T15:04:05.999Z07:00"}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%+v:", i)
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}
