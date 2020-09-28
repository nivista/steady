package queue

import (
	"github.com/google/uuid"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	// Client represents a synchrounous client to the queue.
	Client interface {
		PublishCreate(domain string, timerID uuid.UUID, timer *messaging.Create) error
		PublishDelete(domain string, timerID uuid.UUID) error
	}
)
