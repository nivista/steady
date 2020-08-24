package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/elastic/db"
	"github.com/nivista/steady/internal/.gen/protos/messaging/execute"
	"github.com/nivista/steady/internal/.gen/protos/timerpk"

	"google.golang.org/protobuf/proto"
)

type consumer struct {
	db db.Client
}

// NewConsumer returns a new sarama.ConsumerGroupHandler.
func NewConsumer(db db.Client) sarama.ConsumerGroupHandler {
	return &consumer{
		db: db,
	}
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// TODO handle errors
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var key timerpk.Key
		err := proto.Unmarshal(msg.Key, &key)
		if err != nil {
			fmt.Println("consumer claim unmarshal key:", err.Error())
		}

		if msg.Value == nil {
			//todo done event
			session.MarkMessage(msg, "")
			continue
		}

		var val execute.Value
		err = proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println("consumer claim unmarshal value:", err.Error())
			session.MarkMessage(msg, "")
			continue
		}

		err = c.db.AddExecuteTimer(session.Context(), key.Domain, key.TimerUuid, msg.Timestamp, &val)
		if err != nil {
			fmt.Println("consume claim addexecute err:", err.Error())
		}
	}
	return nil
}
