package coordinator

import (
	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/runtimers/db"
)

// Coordinator hands out managers and stops them for repartitions.
// Maybe the coordinator should just hand out managers and we deal w/ repartitions in the consumer.
type Coordinator struct {
	producer                  sarama.AsyncProducer
	db                        db.Client
	managers                  map[int]*Manager
	createTopic, executeTopic string
	clock                     clockwork.Clock
}

func NewCoordinator(producer sarama.AsyncProducer, db db.Client, createTopic, executeTopic string, clock clockwork.Clock) *Coordinator {
	coord := &Coordinator{
		producer:     producer,
		db:           db,
		managers:     map[int]*Manager{},
		createTopic:  createTopic,
		executeTopic: executeTopic,
		clock:        clock,
	}

	return coord
}

func (c *Coordinator) Stop() {
	for _, manager := range c.managers {
		manager.stop()
	}

	c.producer.AsyncClose()
}

func (c *Coordinator) HandleRepartition(newPartitions []int32) {
	// put new partitions in set
	newPartitionsSet := make(map[int]struct{})
	for _, partition := range newPartitions {
		newPartitionsSet[int(partition)] = struct{}{}
	}

	// cleanup managers for lost partitions
	for partition, man := range c.managers {
		if _, ok := newPartitionsSet[partition]; !ok {
			man.stop()
			delete(c.managers, partition)
		}
	}
}

func (c *Coordinator) GetManager(partition int) *Manager {
	if _, ok := c.managers[partition]; !ok {
		c.managers[partition] = newManager(c.producer.Input(), c.db, c.createTopic, c.executeTopic, partition, c.clock)
	}
	return c.managers[partition]
}

func (c *Coordinator) HasPartition(partition int) bool {
	_, ok := c.managers[partition]
	return ok
}
