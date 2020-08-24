package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/elastic/db"
	"github.com/nivista/steady/internal/.gen/protos/messaging"

	"google.golang.org/protobuf/proto"
)

type consumer struct {
	db                        db.Client
	createTopic, executeTopic string
}

// NewConsumer returns a new sarama.ConsumerGroupHandler.
func NewConsumer(db db.Client, createTopic, executeTopic string) sarama.ConsumerGroupHandler {
	return &consumer{
		db:           db,
		createTopic:  createTopic,
		executeTopic: executeTopic,
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

		if msg.Value == nil {
			//todo done event
			session.MarkMessage(msg, "")
			continue
		}

		var val messaging.Execute
		err := proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println("consumer claim unmarshal value:", err.Error())
			session.MarkMessage(msg, "")
			continue
		}

		err = c.db.AddExecuteTimer(session.Context(), string(msg.Key), msg.Timestamp, &val)
		if err != nil {
			fmt.Println("consume claim addexecute err:", err.Error())
		}
	}
	return nil
}
