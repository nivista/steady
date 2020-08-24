package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	// Client is a client to the database.
	Client interface {
		CreateTimer(ctx context.Context, domain string, id uuid.UUID, timer *messaging.CreateTimer) error
		UpdateTimerProgress(ctx context.Context, domain string, id uuid.UUID, progress *messaging.Progress) error
		FinishTimer(ctx context.Context, domain string, id uuid.UUID) error
	}

	client struct {
		conn *pgxpool.Pool
	}
)

const (
	updateProgressString = `update timers
		set progress=$1
		where domain=$2 and id=$3`

	createTimerString = `insert into timers (id, domain, progress, task, schedule, meta) 
		values ($1, $2, $3, $4, $5, $6) 
		on conflict do nothing`

	finishTimerString = `update timers
		set executing=false
		where domain=$1 and id=$2`
)

// NewClient returns a new Client.
func NewClient(conn *pgxpool.Pool) Client {
	return &client{
		conn: conn,
	}
}

func (c *client) CreateTimer(ctx context.Context, domain string, id uuid.UUID, timer *messaging.CreateTimer) error {
	task, err := protojson.Marshal(timer.Task)
	if err != nil {
		return err
	}
	sched, err := protojson.Marshal(timer.Schedule)
	if err != nil {
		return err
	}

	// initialized with empty progress
	prog, err := protojson.Marshal(&messaging.Progress{})
	if err != nil {
		return err
	}
	meta, err := protojson.Marshal(timer.Meta)
	if err != nil {
		return err
	}
	_, err = c.conn.Exec(ctx, createTimerString, id, domain, prog, task, sched, meta)
	if err != nil {
		return err
	}

	return err
}

func (c *client) UpdateTimerProgress(ctx context.Context, domain string, id uuid.UUID, progress *messaging.Progress) error {
	progBytes, err := protojson.Marshal(progress)
	if err != nil {
		return err
	}
	_, err = c.conn.Exec(ctx, updateProgressString, progBytes, domain, id)
	return err
}

func (c *client) FinishTimer(ctx context.Context, domain string, id uuid.UUID) error {
	_, err := c.conn.Exec(ctx, finishTimerString, domain, id)
	return err
}
