package main

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"context"
	"my/esexample/store"
	"my/esexample/storegrpc"
	"net"
	"os"

	"google.golang.org/grpc"
)

type Server struct {
	EventStore store.EventStore
	storegrpc.UnimplementedEventStoreServiceServer
}

func (me *Server) FindByID(c context.Context, in *storegrpc.FindByIDRequest) (*storegrpc.FindResponse, error) {
	log.Info().Msgf("FindByID: %v", in.Id)

	events, err := me.EventStore.Find(store.EventID(in.Id))

	if err != nil {
		return nil, err
	}

	var findResponseEvents []*storegrpc.FindResponse_Event

	for _, e := range events {
		findResponseEvents = append(findResponseEvents, &storegrpc.FindResponse_Event{
			Id:       string(e.ID),
			Type:     int32(e.Type),
			Payload:  string(e.Payload),
			Savetime: e.TimeStamp})
	}

	result := &storegrpc.FindResponse{Success: true, Events: findResponseEvents}

	return result, nil
}

func (me *Server) FindByType(c context.Context, in *storegrpc.FindByTypeRequest) (*storegrpc.FindResponse, error) {
	log.Debug().Msgf("FindByTYPE: %v", in.Type)

	events, latest, err := me.EventStore.GetEventsByType(store.EventType(in.Type), in.Since, int(in.BatchSize))
	if err != nil {
		return nil, err
	}

	var findResponseEvents []*storegrpc.FindResponse_Event

	for _, e := range events {
		findResponseEvents = append(findResponseEvents, &storegrpc.FindResponse_Event{
			Id:       string(e.ID),
			Type:     int32(e.Type),
			Payload:  string(e.Payload),
			Savetime: e.TimeStamp})
	}

	result := &storegrpc.FindResponse{
		Success: true,
		Events:  findResponseEvents,
		Latest:  latest,
	}

	return result, nil
}

func (me *Server) Update(c context.Context, in *storegrpc.UpdateRequest) (*storegrpc.UpdateResponse, error) {
	log.Info().Msgf("Update: %v", in.Id)

	var events []store.StoreEvent

	for _, e := range in.Events {
		events = append(events, store.StoreEvent{
			ID:      store.EventID(in.Id),
			Payload: store.EventPayload(e.Payload),
			Type:    store.EventType(e.Type),
		})
	}

	err := me.EventStore.Update(store.EventID(in.Id), int(in.Version), events)

	if err != nil {
		return nil, err
	}

	response := &storegrpc.UpdateResponse{
		Success: true,
	}

	return response, nil
}

// Get the value of the environment variable key or the fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("GRPC EVENT-STORE")

	port := getEnv("PORT", "8080")
	hosts := getEnv("CASSANDRA_HOSTS", "localhost")
	keyspace := getEnv("CASSANDRA_KEYSPACE", "eventstore")
	writeQuorum := getEnv("CASSANDRA_WRITE_QUORUM", "QUORUM")
	readQuorum := getEnv("CASSANDRA_WRITE_QUORUM", "LOCAL_QUORUM")

	store, err := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
		Hosts:       []string{hosts},
		Keyspace:    keyspace,
		WriteQuorum: strings.ToUpper(writeQuorum),
		ReadQuorum:  strings.ToUpper(readQuorum),
	})

	if err != nil {
		log.Fatal().Msgf("unable connect to database: %+v", err)
	}

	server := &Server{EventStore: store}

	// Listen
	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal().Msgf("unable to listen: %+v", err)
	}

	grpcServer := grpc.NewServer()

	storegrpc.RegisterEventStoreServiceServer(grpcServer, server)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Msgf("failed to serve: %s", err)
	}
}
