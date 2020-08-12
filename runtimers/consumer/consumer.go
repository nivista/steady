package consumer

import (
	"fmt"
	"strconv"

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
		// if we already have this partition, do nothing
		if c.coord.HasPartition(partition) {
			continue
		}

		// read from beggining
		session.ResetOffset(c.topic, partition, -1, "")

		// send dummy message to know when "present" is
		c.producer.Input() <- &sarama.ProducerMessage{
			Topic: c.topic,
			Key:   keys.NewDummy(),
			Value: nil,
			Headers: []sarama.RecordHeader{{
				Key:   []byte("generationID"),
				Value: []byte(strconv.Itoa(int(session.GenerationID()))),
			}},
			Partition: partition,
		}
	}

	// drop and stop whatever partitions you need to.
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
		man = c.coord.GetManager(claim.Partition())

		// set of the keys of all the potentially stale progresses
		staleProgressCandiates = map[string]struct{}{}
	)

	man.GenerationID = strconv.Itoa(int(session.GenerationID()))

	for msg := range claim.Messages() {
		var k, err = keys.ParseKey(msg.Key)
		if err != nil {
			fmt.Println("handling message error:", err.Error())
			// we didn't mark the message here oops
			continue
		}

		var headers = map[string]string{}

		for _, header := range msg.Headers {
			headers[string(header.Key)] = string(header.Value)
		}

		switch key := k.(type) {
		// A timer Create or Delete
		case keys.CreateTimer:
			var (
				id     = key.TimerUUID()
				domain = key.Domain()
			)

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
		case keys.ExecuteTimer:
			if msg.Value == nil {
				// the timer associated with this progress must've been deleted.
				delete(staleProgressCandiates, string(msg.Key))
				break
			}

			var prog common.Progress
			err := proto.Unmarshal(msg.Value, &prog)
			if err != nil {
				fmt.Println("consume claim unmarshal progress:", err.Error())
				break
			}

			var id = key.TimerUUID()

			// we're reading historical progress updates to get to the current state
			if man.HasTimer(id) && !man.Active {
				man.UpdateTimerProgress(id, &prog)
				break
			}

			// this progress update is from a timer we already have running
			if man.HasTimer(id) {
				break
			}

			// in this case we definitley already sent a progress delete
			if man.GenerationID == headers["generationID"] {
				break
			}

			// in this case this progress update probably doesn't have a delete coming yet.
			if man.Active {
				c.producer.Input() <- &sarama.ProducerMessage{
					Topic: c.topic,
					Key:   sarama.ByteEncoder(msg.Key),
					Value: nil,
					Headers: []sarama.RecordHeader{{
						Key:   []byte("generationID"),
						Value: []byte(man.GenerationID),
					}},
					Partition: msg.Partition,
				}
				break
			}

			// there might be a delete just later in the queue
			staleProgressCandiates[string(msg.Key)] = struct{}{}

		case keys.Dummy:
			// is this me
			if headers["generationID"] != man.GenerationID {
				break
			}

			// if so start the manager
			man.Start()

			// delete stale progress updates
			for key := range staleProgressCandiates {

				c.producer.Input() <- &sarama.ProducerMessage{
					Topic: c.topic,
					Key:   sarama.ByteEncoder(key),
					Value: nil,
					Headers: []sarama.RecordHeader{{
						Key:   []byte("generationID"),
						Value: []byte(man.GenerationID),
					}},
					Partition: msg.Partition,
				}

			}

		default:
			fmt.Println("Unknown key type:", string(msg.Key))
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
