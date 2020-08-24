package consumer

import (
	"fmt"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/internal/.gen/protos/messaging/create"

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
			Key:   nil,
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
	)

	// this happens concurrently with the firing of timers and affects messages produced.
	// we want the gaurantee that we will send producer messages w/ monotonically increasing generationIDs.
	// anyways this is bad because it's undefined behavior
	// instead we can use a RWLock() for updating generation ID
	man.GenerationID = strconv.Itoa(int(session.GenerationID()))

	// TODO error handling
	for msg := range claim.Messages() {
		// get headers
		var headers = map[string]string{}

		for _, header := range msg.Headers {
			headers[string(header.Key)] = string(header.Value)
		}

		if msg.Key == nil && headers["generationID"] == man.GenerationID {
			// this is a dummy key that i sent

			man.Start(session.Context())

			session.MarkMessage(msg, "")
			continue
		}

		keyStr := string(msg.Key)

		if msg.Value == nil {
			man.RemoveTimer(keyStr)
			continue
		}

		var timer create.Value
		err := proto.Unmarshal(msg.Value, &timer)
		if err != nil {
			fmt.Println("consume claim unmarshal timer:", err.Error())
			continue
		}

		err = man.AddTimer(keyStr, &timer)
		if err != nil {
			fmt.Println("consume claim add timer:", err.Error())
			continue
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
