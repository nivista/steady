package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/webservice/db"
	"github.com/nivista/steady/webservice/queue"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	db    db.Client
	queue queue.Client
}

// NewServer returns a services.SteadyServer
func NewServer(db db.Client, queue queue.Client) services.SteadyServer {
	return &server{
		db:    db,
		queue: queue,
	}
}

func (s *server) CreateTimer(ctx context.Context, req *services.CreateTimerRequest) (*services.CreateTimerResponse, error) {
	timerID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	err = s.queue.PublishCreate(req.Domain, timerID, &messaging.CreateTimer{
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

	err = s.queue.PublishDelete(req.Domain, id)
	if err != nil {
		return nil, err
	}

	return &services.DeleteTimerResponse{}, nil
}

func (s *server) GetTimer(ctx context.Context, req *services.GetTimerRequest) (*services.GetTimerResponse, error) {
	return s.db.GetTimer(ctx, req)
}
