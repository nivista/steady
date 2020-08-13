package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
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
		var key messaging.Key
		err := proto.Unmarshal(msg.Key, &key)
		if err != nil {
			fmt.Println("consumer claim unmarshal key:", err.Error())
		}

		switch k := key.Key.(type) {
		// A timer Create or Delete
		case *messaging.Key_CreateTimer_:
			var (
				id, err = uuid.Parse(k.CreateTimer.TimerUuid)
				domain  = k.CreateTimer.Domain
			)

			if err != nil {
				fmt.Println("unmarshal id err:", err)
			}

			// is delete message
			if msg.Value == nil {
				c.db.FinishTimer(session.Context(), domain, id)
				break
			}

			//is create message
			var createTimer messaging.CreateTimer
			err = proto.Unmarshal(msg.Value, &createTimer)
			if err != nil {
				fmt.Println("store consumer unmarshal timer err:", err.Error())
			}

			err = c.db.CreateTimer(session.Context(), domain, id, &createTimer)
			if err != nil {
				fmt.Println("store consumer CreateTimer err:", err.Error())
			}

		// A Progress update or delete
		case *messaging.Key_ExecuteTimer_:
			var (
				id, err = uuid.Parse(k.ExecuteTimer.TimerUuid)
				domain  = k.ExecuteTimer.Domain
			)

			if err != nil {
				fmt.Println("unmarshal id err:", err)
			}
			var exec messaging.ExecuteTimer
			err = proto.Unmarshal(msg.Value, &exec)
			if err != nil {
				fmt.Println("store consumer unmarshal exec err:", err.Error())
			}

			err = c.db.UpdateTimerProgress(session.Context(), domain, id, exec.Progress)
			if err != nil {
				fmt.Println("store consumer UpdateProgress err:", err.Error())
			}
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
