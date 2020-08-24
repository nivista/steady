package queue

import (
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/nivista/steady/internal/.gen/protos/messaging"

	"google.golang.org/protobuf/proto"
)

type (
	// Client represents a synchrounous client to the queue.
	Client interface {
		PublishCreate(timerID uuid.UUID, timer *messaging.Create) error
		PublishDelete(timerID uuid.UUID) error
	}

	client struct {
		producer   sarama.SyncProducer
		partitions int32
		topic      string
	}
)

// NewClient returns a new Client.
func NewClient(producer sarama.SyncProducer, partitions int32, topic string) Client {
	return &client{
		producer:   producer,
		partitions: partitions,
		topic:      topic,
	}
}

func (c *client) PublishCreate(timerID uuid.UUID, timer *messaging.Create) error {
	bytes, err := proto.Marshal(timer)
	if err != nil {
		return err
	}

	_, _, err = c.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       sarama.StringEncoder(timerID.String()),
		Value:     sarama.ByteEncoder(bytes),
		Partition: c.bytesToPartition(timerID),
	})

	return err
}

func (c *client) PublishDelete(timerID uuid.UUID) error {
	_, _, err := c.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       sarama.StringEncoder(timerID.String()),
		Value:     nil,
		Partition: c.bytesToPartition(timerID),
	})

	return err
}

func (c *client) bytesToPartition(id [16]byte) int32 {
	var partition int32

	for _, b := range id {
		partition = partition << 8
		partition += int32(b)
		partition = partition % c.partitions
	}

	return partition
}
