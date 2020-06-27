package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nivista/steady/.gen/protos/services"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	c := services.NewSteadyClient(conn)

	req := services.DeleteTimerRequest{
		Domain:    "Yaniv",
		TimerUuid: "844a3c77-9806-4719-b51a-bfd7013e3c47",
	}

	res, err := c.DeleteTimer(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
