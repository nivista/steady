package queue

import (
	"sync"

	"github.com/Shopify/sarama"
)

type (
	// Client is a client to the queue.
	Client interface {
		Publish(key, value sarama.Encoder, partition int32)
		UpdateGenerationID(generationID string)
	}

	client struct {
		topic    string
		producer sarama.AsyncProducer
		nodeID   string

		genID    string
		genIDMux sync.RWMutex
	}
)

const (
	generationID = "generation_id"
	nodeID       = "node_id"
)

func (c *client) Publish(key, value sarama.Encoder, partition int32) {
	c.genIDMux.RLock()
	defer c.genIDMux.RUnlock()

	c.producer.Input() <- &sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       key,
		Value:     sarama.ByteEncoder(bytes),
		Partition: partition,
		Headers: []sarama.RecordHeader{{
			Key:   []byte(generationID),
			Value: []byte(c.genID),
		},
			{
				Key:   []byte(nodeID),
				Value: []byte(c.nodeID),
			}},
	}
}

func (c *client) UpdateGenerationID(generationID string) {
	c.genIDMux.Lock()
	defer c.genIDMux.Unlock()

	c.genID = generationID
}
