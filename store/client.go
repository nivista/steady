package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/steady/timer"
)

type Client struct {
	*pgxpool.Pool
}

type (
	taskType     string
	scheduleType string
)

const (
	http     taskType     = "HTTP"
	cron     scheduleType = "CRON"
	interval scheduleType = "INTERVAL"
)

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

	task, err := json.Marshal(t.Task)
	if err != nil {
		return err
	}
	sched, err := json.Marshal(t.Schedule)
	if err != nil {
		return err
	}

	_, err = c.Query(ctx,
		`insert into timers (id, domain, executionCount, taskType, task, scheduleType, schedule, meta) 
			values ($1, $2, $3, $4, $5, $6, $7, $8)
			on conflict do nothing`,
		t.ID, t.Domain, t.ExecutionCount, tt, task, st, sched, t.Meta)

	return err
}

func (c *Client) DeleteTimer(ctx context.Context, domain string, id uuid.UUID) error {
	_, err := c.Query(ctx, //TODO check if a timer was deleted
		`delete from timers where domain=$1 and id=$2`,
		domain, id)
	return err
}

func (c *Client) GetTimer(ctx context.Context, domain string, id uuid.UUID) (*timer.Timer, error) {
	t := timer.Timer{}

	var tt taskType
	var st scheduleType
	var taskBytes, scheduleBytes []byte

	err := c.QueryRow(ctx,
		`select id, domain, executionCount, taskType, task, scheduleType, schedule, meta from timers 
			where domain=$1 and id=$2`,
		domain, id).Scan(&t.ID, &t.Domain, &t.ExecutionCount, &tt, &taskBytes, &st, &scheduleBytes, &t.Meta)

	if err != nil {
		return nil, err
	}

	switch tt {
	case http:
		var task timer.HTTP
		err = json.Unmarshal(taskBytes, &task)
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
		err = json.Unmarshal(scheduleBytes, &schedule)
		if err != nil {
			return nil, err
		}
		t.Schedule = &schedule
	case interval:
		var schedule timer.Interval
		err := json.Unmarshal(scheduleBytes, schedule)
		if err != nil {
			return nil, err
		}
		t.Schedule = &schedule
	default:
		return nil, errors.New("Unknown type")
	}
	return &t, nil
}

func (c *Client) SetExecCount(ctx context.Context, domain string, id uuid.UUID, count int) error {
	_, err := c.Query(ctx,
		`update timers
			set executionCount=$1
			where domain=$2 and id=$3`,
		count, domain, id)
	return err
}

func (t *taskType) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Expected []byte in Scan of taskType")
	}

	*t = taskType(bytes)

	return nil
}

func (t *scheduleType) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Expected []byte in Scan of scheduleType")
	}

	*t = scheduleType(bytes)

	return nil
}
