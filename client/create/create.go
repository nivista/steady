package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/.gen/protos/services"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	c := services.NewSteadyClient(conn)

	req := services.CreateTimerRequest{
		Domain: "Yaniv",
		Task: &common.Task{
			Task: &common.Task_HttpConfig{
				HttpConfig: &common.HTTPConfig{
					Url:     "www.example.com/bug",
					Method:  common.Method_GET,
					Body:    "get requests technically shouldn't have bodies",
					Headers: map[string]string{"set-cookie": "yum"},
				},
			},
		},
		Schedule: &common.Schedule{
			Cron:          "@every 5s",
			MaxExecutions: 5,
		},
	}

	res, err := c.CreateTimer(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
