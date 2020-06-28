package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

// NewTimerKey creates a record key for Kafka.
func NewTimerKey(domain string, id uuid.UUID) (sarama.ByteEncoder, error) {
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bytes := append([]byte(domain+":"), idBytes...)
	return sarama.ByteEncoder(bytes), nil
}

// ParseTimerKey parses a timer key from Kafka.
func ParseTimerKey(bytes []byte) (domain string, id uuid.UUID, err error) {
	s := string(bytes)
	split := strings.Split(s, ":")
	if len(split) != 2 {
		err = errors.New(fmt.Sprint("Unexpected key format: ", s))
		return
	}

	domain = split[0]
	err = id.UnmarshalBinary([]byte(split[1]))
	return
}
