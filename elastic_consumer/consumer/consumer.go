package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/elastic_consumer/db"
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
	switch claim.Topic() {
	case c.createTopic:
		c.consumeCreateClaim(session, claim)
	case c.executeTopic:
		c.consumeExecuteClaim(session, claim)
	}

	return nil
}
func (c *consumer) consumeCreateClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {

		if msg.Key == nil {
			// dummy event, ignore
			session.MarkMessage(msg, "")
			continue
		}

		var key messaging.Key
		err := proto.Unmarshal(msg.Key, &key)
		if err != nil {
			fmt.Println("error unmarshalling keys")
			session.MarkMessage(msg, "")
			continue
		}

		if msg.Value == nil {
			//todo delete, done event
			session.MarkMessage(msg, "")
			continue
		}

		var val messaging.Create
		err = proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println("consumeCreateClaim unmarshal value:", err.Error())
			session.MarkMessage(msg, "")
			continue
		}

		err = c.db.CreateTimer(session.Context(), key.Domain, key.TimerUUID, &val)
		if err != nil {
			fmt.Println("consumeCreateClaim addexecute err:", err.Error())
		}
	}
	return nil
}

func (c *consumer) consumeExecuteClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for msg := range claim.Messages() {
		if msg.Key == nil {
			//Dummy event, ignore
			session.MarkMessage(msg, "")
			continue
		}

		var key messaging.Key
		err := proto.Unmarshal(msg.Key, &key)
		if err != nil {
			fmt.Println("consumeExecuteClaim unmarshalling keys")
			session.MarkMessage(msg, "")
			continue
		}

		if msg.Value == nil {
			c.db.DeleteTimer(session.Context(), key.Domain, key.TimerUUID)
			session.MarkMessage(msg, "")
			continue
		}

		var val messaging.Execute
		err = proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println("consumerExecuteClaim unmarshal value:", err.Error())
			session.MarkMessage(msg, "")
			continue
		}

		err = c.db.ExecuteTimer(session.Context(), key.Domain, key.TimerUUID, claim.Partition(), msg.Timestamp, &val)
		if err != nil {
			fmt.Println("consumeExecuteClaim addexecute err:", err.Error())
			session.MarkMessage(msg, "")
			continue
		}
	}
	return nil
}
