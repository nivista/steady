package db

import (
	"context"
)

type (
	// Client is a client to the database.
	Client interface {
		UpsertUser(ctx context.Context, apiToken, apiKey string) error
		AuthenticateUser(ctx context.Context, apiToken, apiSecret string) error
	}
)
