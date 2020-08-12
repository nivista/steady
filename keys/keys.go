package keys

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

type (
	// Key represents a kafka key.
	Key interface {
		sarama.Encoder
	}

	// CreateTimer is a create timer or delete timer event, value should be a messaging.CreateTimer or nil.
	CreateTimer struct {
		domain    string
		timerUUID uuid.UUID
	}

	// ExecuteTimer is an execute timer event, value should be a common.ExecuteTimer or nil.
	ExecuteTimer struct {
		domain    string
		timerUUID uuid.UUID
	}

	// Dummy is a dummy event so a node knows when its caught up after a repartition. Value should be nil.
	Dummy struct{}
)

const (
	createTimerLabel  = "create"
	createTimerFields = 2

	executeTimerLabel  = "execute"
	executeTimerFields = 2

	dummyLabel  = "dummy"
	dummyFields = 0
)

// NewCreateTimer creates a new CreateTimer key.
func NewCreateTimer(domain string, timerUUID uuid.UUID) Key {
	return CreateTimer{domain: domain, timerUUID: timerUUID}
}

// Encode returns an encoded timer.
func (c CreateTimer) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("%v:%v:%v", createTimerLabel, c.domain, c.timerUUID)), nil
}

// Length returns the Length of an encoded timer.
func (c CreateTimer) Length() int {
	return createTimerFields + len(createTimerLabel) + len(c.domain) + len(c.timerUUID.String())
}

// TimerUUID returns the TimerUUID.
func (c CreateTimer) TimerUUID() uuid.UUID {
	return c.timerUUID
}

// Domain returns the domain.
func (c CreateTimer) Domain() string {
	return c.domain
}

// NewExecuteTimer creates a new ExecuteTimer key.
func NewExecuteTimer(domain string, timerUUID uuid.UUID) Key {
	return ExecuteTimer{domain: domain, timerUUID: timerUUID}
}

// Encode returns an encoded TimerProgress.
func (e ExecuteTimer) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("%v:%v:%v", executeTimerLabel, e.domain, e.timerUUID)), nil
}

// Length returns the length of an encoded TimerProgress.
func (e ExecuteTimer) Length() int {
	return executeTimerFields + len(executeTimerLabel) + len(e.domain) + len(e.timerUUID.String())
}

// TimerUUID returns the TimerUUID.
func (e ExecuteTimer) TimerUUID() uuid.UUID {
	return e.timerUUID
}

// Domain returns the Domain.
func (e ExecuteTimer) Domain() string {
	return e.domain
}

// NewDummy returns a new Dummy.
func NewDummy() Dummy {
	return Dummy{}
}

// Encode returns an encoded Dummy.
func (Dummy) Encode() ([]byte, error) {
	return []byte(dummyLabel), nil
}

// Length returns the length of an encoded Dummy.
func (Dummy) Length() int {
	return len(dummyLabel)
}

// ParseKey parses a timer key.
func ParseKey(key []byte) (Key, error) {
	var s = string(key)
	var parts = strings.Split(s, ":")

	switch parts[0] {
	case createTimerLabel:
		if len(parts) != createTimerFields+1 {
			return nil, fmt.Errorf("wrong number of fields for key %v", s)
		}

		domain := parts[1]
		id, err := uuid.Parse(parts[2])
		if err != nil {
			return nil, fmt.Errorf("parsing id key %v: %w", s, err)
		}

		return NewCreateTimer(domain, id), nil

	case executeTimerLabel:
		if len(parts) != executeTimerFields+1 {
			return nil, fmt.Errorf("wrong number of fields for key: %v", s)
		}

		domain := parts[1]
		id, err := uuid.Parse(parts[2])
		if err != nil {
			return nil, fmt.Errorf("parsing id key %v: %w", s, err)
		}

		return NewExecuteTimer(domain, id), nil
	case dummyLabel:
		if len(parts) != dummyFields+1 {
			return nil, fmt.Errorf("wrong number of fields for key: %v", s)
		}

		return NewDummy(), nil
	}

	return nil, fmt.Errorf("unknown key type")
}
