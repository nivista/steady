package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/keys"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Consumer consumes from Kafka and logs the contents.
type Consumer struct {
	nodeID string
	topic  string
}

// NewConsumer returns a new Consumer.
func NewConsumer(nodeID, topic string) sarama.ConsumerGroupHandler {
	return &Consumer{
		nodeID: nodeID,
		topic:  topic,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()[c.topic]

	for _, partition := range partitions {
		session.ResetOffset(c.topic, partition, -1, "")

	}

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Println(claim.InitialOffset())
	for msg := range claim.Messages() {
		var k, err = keys.ParseKey(msg.Key, msg.Partition)
		if err != nil {
			fmt.Println("handling message error:", err.Error())
		}

		fmt.Printf("TOPIC %v, PARTITION %v, OFFSET %v\n", claim.Topic(), claim.Partition(), msg.Offset)

		headers := map[string]string{}
		for _, h := range msg.Headers {
			headers[string(h.Key)] = string(h.Value)
		}
		fmt.Println("-- HEADERS:", headers)
		fmt.Println("-- KEY:", string(msg.Key))
		switch k.(type) {
		// A timer Create or Delete
		case keys.Timer:
			if msg.Value == nil {
				fmt.Println("-- VALUE: <nil>")
				break
			}

			var timer messaging.CreateTimer
			err := proto.Unmarshal(msg.Value, &timer)
			if err != nil {
				fmt.Println("consume claim unmarshal timer:", err.Error())
				break
			}

			json, err := protojson.Marshal(&timer)
			if err != nil {
				fmt.Println("err marshalling to jsaon")
			}

			fmt.Println("-- VALUE:", string(json))

		// A Progress update or delete
		case keys.TimerProgress:
			if msg.Value == nil {
				// the timer associated with this progress must've been deleted.
				fmt.Println("-- VALUE: <nil>")
				break
			}

			var prog common.Progress
			err := proto.Unmarshal(msg.Value, &prog)
			if err != nil {
				fmt.Println("consume claim unmarshal progress:", err.Error())
				break
			}

			json, err := protojson.Marshal(&prog)
			if err != nil {
				fmt.Println("err marshalling to jsaon")
			}

			fmt.Println("-- VALUE:", string(json))
		case keys.Dummy:
		default:
			panic("unknown key type")
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
