package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"context"
	"my/esexample/storegrpc"
	"net"
	"os"

	"google.golang.org/grpc"
)

type Server struct {
	storegrpc.UnimplementedEventStoreServiceServer
}

func (me *Server) FindByID(c context.Context, in *storegrpc.FindByIDRequest) (*storegrpc.FindResponse, error) {
	log.Info().Msgf("GRPC SINK - FIND BY ID %v", in.Id)
	result := &storegrpc.FindResponse{Success: true}
	return result, nil
}

func (me *Server) FindByType(c context.Context, in *storegrpc.FindByTypeRequest) (*storegrpc.FindResponse, error) {
	log.Debug().Msgf("GRPC SINK - FIND BY TYPE %v", in.Type)
	result := &storegrpc.FindResponse{Success: true}
	return result, nil
}

func (me *Server) Update(c context.Context, in *storegrpc.UpdateRequest) (*storegrpc.UpdateResponse, error) {
	log.Info().Msgf("GRPC SINK - UPDATE %v", in.Id)
	response := &storegrpc.UpdateResponse{Success: true}
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
	log.Info().Msg("GRPC SINK")
	server := &Server{}

	// Listen
	port := getEnv("PORT", "8080")
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
