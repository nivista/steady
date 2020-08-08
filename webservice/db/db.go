package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/steady/.gen/protos/services"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	// Client represents a client to the database.
	Client interface {
		GetTimer(ctx context.Context, req *services.GetTimerRequest) (*services.GetTimerResponse, error)
	}

	client struct {
		conn *pgxpool.Pool
	}
)

const getTimerString = `select progress, task, schedule, meta from timers 
					where domain=$1 and id=$2`

// NewClient returns a new Client.
func NewClient(conn *pgxpool.Pool) Client {
	return &client{conn}
}

// GetTimer gets a timer from the database.
func (c *client) GetTimer(ctx context.Context, req *services.GetTimerRequest) (*services.GetTimerResponse, error) {
	var (
		domain, id               = req.Domain, req.TimerUuid
		res                      services.GetTimerResponse
		taskBytes, scheduleBytes []byte
	)
	err := c.conn.QueryRow(ctx, getTimerString,
		domain, id).Scan(&res.Progress, &taskBytes, &scheduleBytes, &res.Meta)

	if err != nil {
		return nil, err
	}

	if err = protojson.Unmarshal(taskBytes, res.Task); err != nil {
		return nil, err
	}

	if err = protojson.Unmarshal(scheduleBytes, res.Schedule); err != nil {
		return nil, err
	}

	return &res, nil
}
