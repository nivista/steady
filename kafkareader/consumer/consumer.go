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

// Consumer consumes from the database and runs its partition of timers.
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
	fmt.Println(session.Claims()[c.topic])

	session.GenerationID()
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

	var (
		initialLatestReplicatedMessageOffset = claim.HighWaterMarkOffset() - 1
	)

	fmt.Println("INITIAL_LATEST_REPLICATED_MESSAGE_OFFSET", initialLatestReplicatedMessageOffset)
	for msg := range claim.Messages() {
		var k, err = keys.ParseKey(msg.Key, msg.Partition)
		if err != nil {
			fmt.Println("handling message error:", err.Error())
		}

		var (
			id     = k.TimerUUID()
			domain = k.Domain()
		)
		fmt.Println("KAFKA ENTRY OFFSET ", msg.Offset)
		fmt.Println("-- DOMAIN:", domain)
		fmt.Println("-- ID:", id)

		headers := map[string]string{}
		for _, h := range msg.Headers {
			headers[string(h.Key)] = string(h.Value)
		}
		fmt.Println("-- HEADERS:", headers)
		fmt.Println("-- KEY:", string(msg.Key))
		switch k.(type) {
		// A timer Create or Delete
		case keys.Timer:
			fmt.Println("-- KEY TYPE: create or delete")
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
			fmt.Println("-- KEY TYPE: update or delete progress")
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
		default:
			panic("unknown key type")
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
