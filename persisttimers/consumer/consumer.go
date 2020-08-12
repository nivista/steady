package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/keys"
	"github.com/nivista/steady/persisttimers/db"
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

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var k, err = keys.ParseKey(msg.Key, msg.Partition)
		if err != nil {
			fmt.Println("handling message error:", err.Error())
		}

		switch key := k.(type) {
		// A timer Create or Delete
		case keys.CreateTimer:
			var (
				id     = key.TimerUUID()
				domain = key.Domain()
			)

			// is delete message
			if msg.Value == nil {
				c.db.FinishTimer(session.Context(), domain, id)
				break
			}

			//is create message
			var createTimer messaging.CreateTimer
			err := proto.Unmarshal(msg.Value, &createTimer)
			if err != nil {
				fmt.Println("store consumer unmarshal timer err:", err.Error())
			}

			err = c.db.CreateTimer(session.Context(), domain, id, &createTimer)
			if err != nil {
				fmt.Println("store consumer CreateTimer err:", err.Error())
			}

		// A Progress update or delete
		case keys.ExecuteTimer:
			var (
				id     = key.TimerUUID()
				domain = key.Domain()
			)

			var progress common.Progress
			err := proto.Unmarshal(msg.Value, &progress)
			if err != nil {
				fmt.Println("store consumer unmarshal progress err:", err.Error())
			}

			err = c.db.UpdateTimerProgress(session.Context(), domain, id, &progress)
			if err != nil {
				fmt.Println("store consumer UpdateProgress err:", err.Error())
			}
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
