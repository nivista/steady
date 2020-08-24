package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/internal/.gen/protos/messaging"

	"google.golang.org/protobuf/proto"
)

// Consumer consumes from Kafka and logs the contents.
type Consumer struct {
	nodeID, createTopic, executeTopic string
}

// NewConsumer returns a new Consumer.
func NewConsumer(nodeID, createTopic, executeTopic string) sarama.ConsumerGroupHandler {
	return &Consumer{
		nodeID:       nodeID,
		createTopic:  createTopic,
		executeTopic: executeTopic,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {

	for topic, partitions := range session.Claims() {
		fmt.Println(topic, partitions)
		for _, partition := range partitions {
			session.ResetOffset(topic, partition, -1, "")

		}
	}

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	switch claim.Topic() {
	case c.createTopic:
		c.consumeCreates(claim)

	case c.executeTopic:
		c.consumeExecutes(claim)

	default:
		fmt.Println("unknown topic")
	}
	return nil
}

func (c *Consumer) consumeCreates(claim sarama.ConsumerGroupClaim) {
	for msg := range claim.Messages() {
		fmt.Printf("TOPIC %v, PARTITION %v, OFFSET %v\n", claim.Topic(), claim.Partition(), msg.Offset)

		if msg.Key == nil {
			fmt.Println("-- DUMMY")
			continue
		}

		fmt.Println("-- ID:", string(msg.Key))

		if msg.Value == nil {
			fmt.Println("-- DELETE")
		}

		var val messaging.Create
		err := proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("-- VALUE:", val)
	}
}

func (c *Consumer) consumeExecutes(claim sarama.ConsumerGroupClaim) {
	for msg := range claim.Messages() {
		fmt.Printf("TOPIC %v, PARTITION %v, OFFSET %v\n", claim.Topic(), claim.Partition(), msg.Offset)

		if msg.Key == nil {
			fmt.Println("-- NIL KEY")
			continue
		}

		fmt.Println("-- ID:", string(msg.Key))

		if msg.Value == nil {
			fmt.Println("-- DELETE EXECUTE")
		}

		var val messaging.Execute
		err := proto.Unmarshal(msg.Value, &val)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("-- VALUE:", val)
	}
}
