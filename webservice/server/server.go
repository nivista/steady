package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/.gen/protos/services"

	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/webservice/queue"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	queue queue.Client
}

// NewServer returns a services.SteadyServer
func NewServer(queue queue.Client) services.SteadyServer {
	return &server{
		queue: queue,
	}
}

func (s *server) CreateTimer(ctx context.Context, req *services.CreateTimerRequest) (*services.CreateTimerResponse, error) {
	// TODO validate schedule execter, etc.

	// if starttime is unset, starttime = now
	if req.Schedule.StartTime == nil {
		req.Schedule.StartTime = timestamppb.Now()
	}

	timerID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	err = s.queue.PublishCreate(timerID, &messaging.Create{
		Task:     req.Task,
		Schedule: req.Schedule,
		Meta: &common.Meta{
			CreateTime: timestamppb.Now(),
		},
	})

	if err != nil {
		return nil, err
	}
	return &services.CreateTimerResponse{
		TimerUuid: timerID.String(),
	}, nil
}
func (s *server) DeleteTimer(ctx context.Context, req *services.DeleteTimerRequest) (*services.DeleteTimerResponse, error) {
	id, err := uuid.Parse(req.TimerUuid)
	if err != nil {
		return nil, err
	}

	err = s.queue.PublishDelete(id)
	if err != nil {
		return nil, err
	}

	return &services.DeleteTimerResponse{}, nil
}
