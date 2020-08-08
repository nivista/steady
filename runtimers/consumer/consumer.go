package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/keys"
	"github.com/nivista/steady/runtimers/coordinator"
	"google.golang.org/protobuf/proto"
)

// Consumer consumes from the database and runs its partition of timers.
type Consumer struct {
	producer   sarama.AsyncProducer
	coord      *coordinator.Coordinator
	partitions int32
	nodeID     string
	topic      string
}

// NewConsumer returns a new Consumer.
func NewConsumer(producer sarama.AsyncProducer, coord *coordinator.Coordinator, partitions int32, nodeID, topic string) sarama.ConsumerGroupHandler {
	return &Consumer{
		producer:   producer,
		coord:      coord,
		partitions: partitions,
		nodeID:     nodeID,
		topic:      topic,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()[c.topic]

	for _, partition := range partitions {
		if !c.coord.HasPartition(partition) {
			session.ResetOffset(c.topic, partition, -1, "")
		}
	}
	// this is where i halt production.
	// i will stop sending messages, and then i will consume, but others might still be sending messages.
	// if this happens
	c.coord.HandleRepartition(partitions)

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	var (
		initialLatestReplicatedMessageOffset = claim.HighWaterMarkOffset() - 1
		man                                  = c.coord.GetManager(claim.Partition())

		// encounteredProgressDelete tracks whether a delete message (nil) for a progress with a given id has been found
		// it allows us to clear out stale progresses in Kafka that ended up in front of their original delete messages
		encounteredProgressDelete = map[string]bool{}
	)

	for msg := range claim.Messages() {
		var k, err = keys.ParseKey(msg.Key, msg.Partition)
		if err != nil {
			fmt.Println("handling message error:", err.Error())
		}

		var (
			keyStr = string(msg.Key)
			id     = k.TimerUUID()
			domain = k.Domain()
		)
		switch k.(type) {
		// A timer Create or Delete
		case keys.Timer:
			fmt.Println(msg.Value)
			if msg.Value == nil {
				man.RemoveTimer(id)
				break
			}

			var timer messaging.CreateTimer
			err := proto.Unmarshal(msg.Value, &timer)
			if err != nil {
				fmt.Println("consume claim unmarshal timer:", err.Error())
				break
			}

			err = man.AddTimer(id, domain, &timer)
			if err != nil {
				fmt.Println("consume claim add timer:", err.Error())
				break
			}

		// A Progress update or delete
		case keys.TimerProgress:
			if msg.Value == nil {
				// the timer associated with this progress must've been deleted.
				encounteredProgressDelete[keyStr] = true
				break
			}

			var prog common.Progress
			err := proto.Unmarshal(msg.Value, &prog)
			if err != nil {
				fmt.Println("consume claim unmarshal progress:", err.Error())
				break
			}

			if man.HasTimer(id) {
				if !man.IsActive() {
					man.UpdateTimerProgress(id, &prog)
				}
			} else {
				// if man doesn't know about this timer, it means the timer was created and then later deleted.
				// track all progresses in Kafka that haven't been deleted even though their timers have been.
				encounteredProgressDelete[keyStr] = false
			}

		default:
			panic("unknown key type")
		}

		if !man.IsActive() && msg.Offset >= initialLatestReplicatedMessageOffset {
			man.Start()

			// delete stale progress updates
			for key, encountered := range encounteredProgressDelete {
				if !encountered {
					c.producer.Input() <- &sarama.ProducerMessage{
						Topic:     c.topic,
						Partition: msg.Partition,
						Key:       sarama.ByteEncoder(key),
						Value:     nil,
					}
				}
			}
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
