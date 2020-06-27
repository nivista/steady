package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/manager"
	"github.com/nivista/steady/messaging"
	"github.com/nivista/steady/store"
	"google.golang.org/grpc"
)

func main() {
	database, err := store.NewClient(context.Background())
	if err != nil {
		panic(err)
	}

	go store.InitConsumer(database)

	messenger := messaging.NewClient()

	go manager.InitConsumer(messenger)

	addr := fmt.Sprintf(":%d", 8001) //TODO make configurable
	conn, err := net.Listen("tcp", "localhost:8001")

	if err != nil {
		panic(fmt.Sprintf("Cannot listen to address %s", addr))
	}

	s := grpc.NewServer()

	services.RegisterSteadyServer(s, &steadyServer{messenger: messenger, database: database})
	go func() {
		if err := s.Serve(conn); err != nil {
			panic(fmt.Sprintf("failed to serve: %v", err))
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	s.GracefulStop()
}
