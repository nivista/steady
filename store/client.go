package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/tasktimer/timer"
)

type Client struct {
	*pgxpool.Pool
}

type (
	taskType     string
	scheduleType string
)

const (
	http taskType = "HTTP"
)

const (
	cron     scheduleType = "CRON"
	interval scheduleType = "INTERVAL"
)

func (t *taskType) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Expected []byte in Scan of taskType")
	}

	*t = taskType(string(bytes)) // could i go straight to tasktype?

	return nil
}

func (t *scheduleType) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Expected []byte in Scan of scheduleType")
	}

	*t = scheduleType(string(bytes))

	return nil
}

func NewClient(ctx context.Context) (*Client, error) {
	dbpool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_CONN"))
	if err != nil {
		return nil, err
	}

	return &Client{Pool: dbpool}, nil
}

func (c *Client) CreateTimer(ctx context.Context, t timer.Timer) error {
	var tt taskType
	switch task := t.Task.(type) {
	case *timer.HTTP:
		tt = http
	default:
		panic(fmt.Sprintf("CreateTimer doesn't know about %T", task))
	}

	var st scheduleType
	switch schedule := t.Schedule.(type) {
	case *timer.Cron:
		st = cron
	case *timer.Interval:
		st = interval
	default:
		panic(fmt.Sprintf("CreateTimer doesn't know about %T", schedule))
	}

	task, err := t.Task.MarshalBinary()
	if err != nil {
		return err
	}
	sched, err := t.Schedule.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = c.Query(ctx,
		`insert into timers (id, account, executionCount, taskType, task, scheduleType, schedule, meta) 
			values ($1, $2, $3, $4, $5, $6, $7, $8)
			on conflict do nothing`,
		t.ID, t.Account, t.ExecutionCount, tt, task, st, sched, t.Meta)

	return err
}

func (c *Client) DeleteTimer(ctx context.Context, account string, id uuid.UUID) error {
	_, err := c.Query(ctx, //TODO check if a timer was deleted
		`delete from timers where account=$1 and id=$2`,
		account, id)
	return err
}

func (c *Client) GetTimer(ctx context.Context, account string, id uuid.UUID) (*timer.Timer, error) {
	t := timer.Timer{}

	var tt taskType
	var st scheduleType
	var taskBytes, scheduleBytes []byte

	err := c.QueryRow(ctx,
		`select id, account, executionCount, taskType, task, scheduleType, schedule, meta from timers 
			where account=$1 and id=$2`,
		account, id).Scan(&t.ID, &t.Account, &t.ExecutionCount, &tt, &taskBytes, &st, &scheduleBytes, &t.Meta)

	if err != nil {
		return nil, err
	}

	switch tt {
	case http:
		var task timer.HTTP
		err = task.UnmarshalBinary(taskBytes)
		if err != nil {
			return nil, err // TODO, make custom error types
		}
		t.Task = &task
	default:
		return nil, errors.New("Unknown type")
	}

	switch st {
	case cron:
		var schedule timer.Cron
		err = schedule.UnmarshalBinary(taskBytes)
		if err != nil {
			return nil, err
		}
		t.Schedule = &schedule
	case interval:
		var schedule timer.Interval
		err := schedule.UnmarshalBinary(scheduleBytes)
		if err != nil {
			return nil, err
		}
		t.Schedule = &schedule
	default:
		return nil, errors.New("Unknown type")
	}
	return &t, nil
}

func (c *Client) SetExecCount(ctx context.Context, account string, id uuid.UUID, count int) error {
	_, err := c.Query(ctx,
		`update timers
			set executionCount=$1
			where account=$2 and id=$3`,
		count, account, id)
	return err
}
