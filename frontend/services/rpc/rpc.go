package rpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/timer"

	"github.com/nivista/steady/frontend/db"
	"github.com/nivista/steady/frontend/queue"
	"github.com/nivista/steady/frontend/util"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	// server time on server that recieves request is the default start time.
	if req.Schedule.StartTime == nil {
		req.Schedule.StartTime = timestamppb.Now()
	}

	timerID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	domain, ok := util.GetClientID(ctx)
	if !ok {
		fmt.Println("CreateTimer got unauthenticated context.")
		return nil, grpc.Errorf(codes.Internal, "")
	}

	create := messaging.Create{
		Task:     req.Task,
		Schedule: req.Schedule,
		Meta: &common.Meta{
			CreateTime: timestamppb.Now(),
		},
	}

	err = timer.IsValid(&create)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.queue.PublishCreate(domain, timerID, &create)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
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

	domain, ok := util.GetClientID(ctx)
	if !ok {
		fmt.Println("DeleteTimer got unauthenticated context.")
		return nil, grpc.Errorf(codes.Internal, "")
	}

	err = s.queue.PublishDelete(domain, id)
	if err != nil {
		return nil, err
	}

	return &services.DeleteTimerResponse{}, nil
}

func GetAuth(client db.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.Internal, "")
		}

		s := md.Get("Authorization")
		if len(s) == 0 {
			return nil, grpc.Errorf(codes.Unauthenticated, "")
		}

		clientID, clientSecret, ok := util.ParseBasicAuth(s[0])
		if !ok {
			return nil, grpc.Errorf(codes.Unauthenticated, "")
		}

		err = client.AuthenticateUser(ctx, clientID, clientSecret)
		if err != nil {
			switch err.(type) {
			case db.InvalidAPIToken:
				return nil, grpc.Errorf(codes.Unauthenticated, "")

			case db.InvalidAPISecret:
				return nil, grpc.Errorf(codes.Unauthenticated, "")
			}

			return nil, grpc.Errorf(codes.Internal, "")
		}

		return handler(util.SetClientID(ctx, clientID), req)

	}

}
