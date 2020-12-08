package main

import (
	"fmt"
	"os"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"my/esexample/patient"
	"my/esexample/store"
	"time"
)

func main() {
	initLog()
	log.Info().Msg("POLLING TEST")

	host := getEnv("STORE_HOST", "localhost")
	port := getEnv("STORE_PORT", "8080")

	// eventstore := store.NewInMemStore()

	/*
		eventstore, _ := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
			Hosts:       []string{"localhost"},
			Keyspace:    "eventstore",
			WriteQuorum: "QUORUM",
			ReadQuorum:  "LOCAL_QUORUM",
		})
	*/

	// eventstore := store.NewRemoteEventStore(&store.RemoteEventStoreConfig{Host: "http://localhost:8080"})

	eventstore := store.NewGrpcEventStore(&store.GrpcEventStoreConfig{Host: host + ":" + port})
	eventTypes := [...]store.EventType{patient.PatientAdmittedEventType, patient.PatientDischargedEventType}

	for _, etype := range eventTypes {
		log.Info().Msgf("Listening for event type %v", etype)

		go func(t store.EventType) {
			var err error
			var interval time.Duration

			// start polling starting from 1 second ago
			since := time.Now().UnixNano()/int64(time.Millisecond) - 1*time.Second.Milliseconds()

			defaultInterval := 500 * time.Millisecond

			boff := backoff.NewExponentialBackOff()
			boff.InitialInterval = defaultInterval
			boff.MaxElapsedTime = 0
			boff.MaxInterval = 10 * time.Second

			for {
				since, err = listenForEvents(t, eventstore, since)

				if err != nil {
					log.Error().Msgf("Unable to get events: %v", err)
					interval = boff.NextBackOff()
				} else {
					interval = defaultInterval
					boff.Reset()
				}

				time.Sleep(interval)

			}
		}(etype)

	}

	log.Info().Msg("Listening...")
	<-make(chan bool)
}

func listenForEvents(etype store.EventType, store store.EventStore, since int64) (int64, error) {
	events, latest, err := store.GetEventsByType(etype, since, 100)

	if err != nil {
		log.Info().Msgf("ERROR POLLING EVENT %v: %v", etype, err)
		return since, err
	}

	if nevents := len(events); nevents > 0 {
		log.Info().Msgf("FOUND %d EVENTS OF EVENT TYPE %+v", len(events), etype)
		return latest, nil
	}

	return since, nil
}

func initLog() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: "2006-01-02T15:04:05.999Z07:00"}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%+v:", i)
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Get the value of the environment variable key or the fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
