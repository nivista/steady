package manager

import (
	"github.com/google/uuid"
	"github.com/nivista/steady/messaging"
	"github.com/nivista/steady/timer"
)

type coordinator struct {
	queue         *messaging.Client
	partitionData map[int32]*partitionData // TODO : figure out better names
}

type partitionData struct {
	*manager
	fireOffset int64
}

func newCoordinator(c *messaging.Client) *coordinator {
	coord := &coordinator{queue: c}
	coord.partitionData = make(map[int32]*partitionData)
	return coord
}

func (c *coordinator) handleRepartition(newPartitions []int32) {
	newPartitionsSet := make(map[int32]bool)
	for _, partition := range newPartitions {
		if !c.hasPartition(partition) {
			c.partitionData[partition] = &partitionData{newManager(c.queue), -1}
		}
		newPartitionsSet[partition] = true
	}

	for partition := range c.partitionData {
		if !newPartitionsSet[partition] {
			c.partitionData[partition].manager.stop()
			delete(c.partitionData, partition)
		}
	}
}

func (c *coordinator) setHighWatermark(partition int32, offset int64) {
	data := c.partitionData[partition]
	if data.fireOffset == -1 {
		data.fireOffset = offset
	}
}

func (c *coordinator) addTimer(t *timer.Timer, partition int32, offset int64) {
	data := c.partitionData[partition]
	man := data.manager
	if offset == data.fireOffset {
		man.start()
	}
	man.addTimer(t)
}

func (c *coordinator) removeTimer(id uuid.UUID, partition int32, offset int64) {
	data := c.partitionData[partition]
	man := data.manager
	if offset == data.fireOffset {
		man.start()
	}
	man.removeTimer(id)
}

func (c *coordinator) hasPartition(partition int32) bool {
	_, ok := c.partitionData[partition]
	return ok
}
