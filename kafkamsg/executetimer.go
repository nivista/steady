package kafka

import (
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/protobuf/proto"
)

type ExecuteTimer struct {
	Msg          *sarama.ConsumerMessage
	Domain       string
	ID           string
	GenerationID string
	Value        *messaging.ExecuteTimer
}

const generationID = "generation_id"

func (e *ExecuteTimer) FromConsumerMessage(msg *sarama.ConsumerMessage) error {
	var out ExecuteTimer
	for _, header := range msg.Headers {
		if string(header.Key) == generationID {
			val, err := strconv.Atoi(string(header.Value))
			if err != nil {
				return err
			}
			out.GenerationID = int32(val)
			break
		}
	}

	if msg.Value == nil {
		*e = out
		return nil
	}

	var val messaging.ExecuteTimer
	err := proto.Unmarshal(msg.Value, &val)
	if err != nil {
		return err
	}

	*e = out
	return nil
}
