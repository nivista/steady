package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

type kafkaKey struct {
	id      uuid.UUID
	domain  string
	encoded []byte
}

func NewTimerKey(domain string, id uuid.UUID) (sarama.ByteEncoder, error) {
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bytes := append([]byte(domain+":"), idBytes...)
	return sarama.ByteEncoder(bytes), nil
}

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
