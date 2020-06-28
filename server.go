package main

import (
	"context"
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/messaging"
	"github.com/nivista/steady/store"
	"github.com/nivista/steady/timer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type steadyServer struct {
	messenger *messaging.Client
	database  *store.Client
}

func (s *steadyServer) CreateTimer(ctx context.Context, timerReq *services.CreateTimerRequest) (*services.CreateTimerResponse, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	task, err := protoTaskToTask(timerReq.GetTaskConfig())
	if err != nil {
		return nil, err
	}

	sched, err := protoSchedToSched(timerReq.GetScheduleConfig())
	if err != nil {
		return nil, err
	}

	meta := timer.Meta{
		CreationTime: time.Now().UTC(),
	}

	t := timer.Timer{
		ID:       id,
		Domain:   timerReq.Domain,
		Task:     task,
		Schedule: sched,
		Meta:     meta,
	}

	err = s.messenger.PublishCreate(&t)
	if err != nil {
		return nil, err
	}

	return &services.CreateTimerResponse{TimerUuid: id.String()}, nil
}

func (s *steadyServer) DeleteTimer(ctx context.Context, deleteReq *services.DeleteTimerRequest) (*services.DeleteTimerResponse, error) {
	id, err := uuid.Parse(deleteReq.GetTimerUuid())
	if err != nil {
		return nil, err
	}

	s.messenger.PublishDelete(deleteReq.Domain, id)
	return &services.DeleteTimerResponse{}, nil
}

func (s *steadyServer) GetTimer(ctx context.Context, getTimerReq *services.GetTimerRequest) (*services.GetTimerResponse, error) {
	id, err := uuid.Parse(getTimerReq.TimerUuid)
	if err != nil {
		return nil, err
	}

	domain := getTimerReq.Domain

	timer, err := s.database.GetTimer(ctx, domain, id)
	if err != nil {
		return nil, err
	}
	task, err := taskToProtoTask(timer.Task)
	if err != nil {
		return nil, err
	}

	sched, err := schedToProtoSched(timer.Schedule)
	if err != nil {
		return nil, err
	}

	createTime, err := ptypes.TimestampProto(timer.Meta.CreationTime)
	if err != nil {
		return nil, err
	}

	meta := &services.Meta{
		CreateTime: createTime,
	}

	res := services.GetTimerResponse{
		TimerUuid:      timer.ID.String(),
		ExecutionCount: timer.ExecutionCount,
		TaskConfig:     task,
		ScheduleConfig: sched,
		Meta:           meta,
	}
	return &res, nil
}

func protoTaskToTask(proto *services.TaskConfig) (timer.Task, error) {
	switch p := proto.TaskConfig.(type) {
	case *services.TaskConfig_HttpConfig:
		m, err := protoMethodToMethod(p.HttpConfig.Method)
		if err != nil {
			return nil, err
		}

		return &timer.HTTP{
			URL:     p.HttpConfig.Url,
			Method:  m,
			Body:    p.HttpConfig.Body,
			Headers: p.HttpConfig.Headers,
		}, nil

	default:
		return nil, errors.New("Unknown Task type")
	}
}

func protoSchedToSched(proto *services.ScheduleConfig) (timer.Schedule, error) {
	switch p := proto.ScheduleConfig.(type) {
	case *services.ScheduleConfig_CronConfig:
		var start time.Time
		if p.CronConfig.StartTime != nil {
			start = p.CronConfig.StartTime.AsTime().UTC()
		} else {
			start = time.Now().UTC()
		}
		return &timer.Cron{
			Start:      start,
			Cron:       p.CronConfig.Cron,
			Executions: p.CronConfig.Executions,
		}, nil
	case *services.ScheduleConfig_IntervalConfig:
		var start time.Time
		if p.IntervalConfig.StartTime != nil {
			start = p.IntervalConfig.StartTime.AsTime().UTC()
		} else {
			start = time.Now().UTC()
		}
		return &timer.Interval{
			Start:      start,
			Interval:   p.IntervalConfig.Interval,
			Executions: p.IntervalConfig.Executions,
		}, nil
	default:
		return nil, errors.New("Unknown Schedule type")
	}
}

func taskToProtoTask(task timer.Task) (*services.TaskConfig, error) {
	switch t := task.(type) {
	case *timer.HTTP:
		m, err := methodToProtoMethod(t.Method)
		if err != nil {
			return nil, err
		}
		return &services.TaskConfig{TaskConfig: &services.TaskConfig_HttpConfig{HttpConfig: &services.HTTPConfig{
			Url:     t.URL,
			Method:  m, // TODO: maybe method should just be a case insensitive string?
			Body:    t.Body,
			Headers: t.Headers,
		}}}, nil
	default:
		return nil, errors.New("No such task")
	}
}

func schedToProtoSched(sched timer.Schedule) (*services.ScheduleConfig, error) {
	switch s := sched.(type) {
	case *timer.Cron:
		var t *timestamppb.Timestamp
		if !s.Start.IsZero() {
			var err error
			t, err = ptypes.TimestampProto(s.Start)
			if err != nil {
				return nil, err
			}
		}

		return &services.ScheduleConfig{ScheduleConfig: &services.ScheduleConfig_CronConfig{CronConfig: &services.CronConfig{
			StartTime:  t,
			Cron:       s.Cron,
			Executions: s.Executions,
		}}}, nil
	case *timer.Interval:
		var t *timestamppb.Timestamp
		if !s.Start.IsZero() {
			var err error
			t, err = ptypes.TimestampProto(s.Start)
			if err != nil {
				return nil, err
			}
		}

		return &services.ScheduleConfig{ScheduleConfig: &services.ScheduleConfig_IntervalConfig{IntervalConfig: &services.IntervalConfig{
			StartTime:  t,
			Interval:   s.Interval,
			Executions: s.Executions,
		}}}, nil
	default:
		return nil, errors.New("No such task")

	}
}

func methodToProtoMethod(m timer.Method) (services.Method, error) {
	switch m {
	case timer.GET:
		return services.Method_GET, nil
	case timer.POST:
		return services.Method_POST, nil
	default:
		return 0, errors.New("No such method")
	}
}

func protoMethodToMethod(m services.Method) (timer.Method, error) {
	switch m {
	case services.Method_GET:
		return timer.GET, nil
	case services.Method_POST:
		return timer.POST, nil
	default:
		return 0, errors.New("No such method")
	}
}
