package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nivista/steady/.gen/protos/services"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	conn, err := grpc.Dial("localhost:8001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	c := services.NewSteadyClient(conn)

	req := services.CreateTimerRequest{
		Domain: "Yaniv",
		TaskConfig: &services.TaskConfig{
			TaskConfig: &services.TaskConfig_HttpConfig{
				HttpConfig: &services.HTTPConfig{
					Url:     "www.example.com/bug",
					Method:  services.Method_GET,
					Body:    "get requests technically shouldn't have bodies",
					Headers: map[string]string{"set-cookie": "yum"},
				},
			},
		},
		ScheduleConfig: &services.ScheduleConfig{
			ScheduleConfig: &services.ScheduleConfig_IntervalConfig{
				IntervalConfig: &services.IntervalConfig{
					StartTime:  timestamppb.New(time.Now().UTC()),
					Interval:   10,
					Executions: 5,
				},
			},
		},
	}
	res, err := c.CreateTimer(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
