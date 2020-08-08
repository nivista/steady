package coordinator

import (
	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
)

type Coordinator struct {
	producer sarama.AsyncProducer
	managers map[int32]*Manager
	topic    string
	clock    clockwork.Clock
}

func NewCoordinator(producer sarama.AsyncProducer, topic string, clock clockwork.Clock) *Coordinator {
	coord := &Coordinator{
		producer: producer,
		managers: map[int32]*Manager{},
		topic:    topic,
		clock:    clock,
	}

	return coord
}

func (c *Coordinator) Stop() {
	c.producer.AsyncClose()
}

func (c *Coordinator) HandleRepartition(newPartitions []int32) {
	// put new partitions in set
	newPartitionsSet := make(map[int32]struct{})
	for _, partition := range newPartitions {
		newPartitionsSet[partition] = struct{}{}
	}

	// cleanup managers for lost partitions
	for partition, man := range c.managers {
		if _, ok := newPartitionsSet[partition]; !ok {
			man.stop()
			delete(c.managers, partition)
		}
	}
}

func (c *Coordinator) GetManager(partition int32) *Manager {
	if _, ok := c.managers[partition]; !ok {
		c.managers[partition] = newManager(c.producer.Input(), c.topic, partition, c.clock)
	}
	return c.managers[partition]
}

func (c *Coordinator) HasPartition(partition int32) bool {
	_, ok := c.managers[partition]
	return ok
}
