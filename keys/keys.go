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

		// Gets the UUID of the timer this relates to
		TimerUUID() uuid.UUID

		// Gets the domain this timer belongs to
		Domain() string
	}

	// Timer is a create timer or delete timer event, value should be a messaging.Timer or nil.
	Timer struct {
		domain    string
		timerUUID uuid.UUID
	}

	// TimerProgress is an update progress event, value should be a common.Progress or nil.
	TimerProgress struct {
		domain    string
		timerUUID uuid.UUID
	}

	// NodeStart is a dummy event so a node knows when its caught up after a repartition. Value should be nil.
	//NodeStart struct {
	//	NodeID    string
	//	partition int32
	//}
)

const (
	timerLabel  = "timer"
	timerFields = 2

	timerProgressLabel  = "prog"
	timerProgressFields = 2
)

// NewTimer creates a new Timer.
func NewTimer(domain string, timerUUID uuid.UUID) Key {
	return Timer{domain: domain, timerUUID: timerUUID}
}

// Encode returns an encoded timer.
func (t Timer) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("%v:%v:%v", timerLabel, t.domain, t.timerUUID)), nil
}

// Length returns the Length of an encoded timer.
func (t Timer) Length() int {
	return timerFields + len(timerLabel) + len(t.domain) + len(t.timerUUID.String())
}

// TimerUUID returns the TimerUUID.
func (t Timer) TimerUUID() uuid.UUID {
	return t.timerUUID
}

// Domain returns the domain.
func (t Timer) Domain() string {
	return t.domain
}

// NewTimerProgress creates a new NewTimerProgress.
func NewTimerProgress(domain string, timerUUID uuid.UUID) Key {
	return TimerProgress{domain: domain, timerUUID: timerUUID}
}

// Encode returns an encoded TimerProgress.
func (t TimerProgress) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("%v:%v:%v", timerProgressLabel, t.domain, t.timerUUID)), nil
}

// Length returns the length of an encoded TimerProgress.
func (t TimerProgress) Length() int {
	return timerProgressFields + len(timerProgressLabel) + len(t.domain) + len(t.timerUUID.String())
}

// TimerUUID returns the TimerUUID.
func (t TimerProgress) TimerUUID() uuid.UUID {
	return t.timerUUID
}

// Domain returns the Domain.
func (t TimerProgress) Domain() string {
	return t.domain
}

// ParseKey parses a timer key.
func ParseKey(key []byte, partition int32) (Key, error) {
	var s = string(key)
	var parts = strings.Split(s, ":")

	switch parts[0] {
	case timerLabel:
		if len(parts) != timerFields+1 {
			return nil, fmt.Errorf("wrong number of fields for key %v", s)
		}

		domain := parts[1]
		id, err := uuid.Parse(parts[2])
		if err != nil {
			return nil, fmt.Errorf("parsing id key %v: %w", s, err)
		}

		return NewTimer(domain, id), nil

	case timerProgressLabel:
		if len(parts) != timerProgressFields+1 {
			return nil, fmt.Errorf("wrong number of fields for key: %v", s)
		}

		domain := parts[1]
		id, err := uuid.Parse(parts[2])
		if err != nil {
			return nil, fmt.Errorf("parsing id key %v: %w", s, err)
		}

		return NewTimerProgress(domain, id), nil
	}

	return nil, fmt.Errorf("unknown key type")
}

func idToPartition(id [16]byte, partitions int32) int32 {
	var partition int32

	for b := range id {
		partition <<= 8
		partition += int32(b)
		partition %= partitions
	}

	return partition
}
