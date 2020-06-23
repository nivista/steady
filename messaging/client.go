package messaging

import (
	"os"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/gofrs/uuid"
	"github.com/nivista/tasktimer/timer"
)

type Client struct {
	prod       sarama.SyncProducer
	partitions int32
}

func NewClient() *Client {
	partitionsString := os.Getenv("PARTITIONS")
	partitions, err := strconv.Atoi(partitionsString)
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	brokerList := strings.Split(os.Getenv("KAFKA_PEERS"), ",")
	prod, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		panic(err)
	}
	c := Client{prod: prod, partitions: int32(partitions)}
	return &c
}

func (c *Client) UpsertTimer(t *timer.Timer) error {
	idBytes, err := t.ID.MarshalBinary()
	if err != nil {
		return err
	}

	timerBytes, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	partition := c.bytesToPartition([16]byte(t.ID))

	m := sarama.ProducerMessage{
		Topic:     "timers",
		Key:       sarama.ByteEncoder(idBytes),
		Value:     sarama.ByteEncoder(timerBytes),
		Partition: partition,
	}

	_, _, err = c.prod.SendMessage(&m)
	return err
}

func (c *Client) DeleteTimer(id uuid.UUID) error {
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return err
	}

	partition := c.bytesToPartition([16]byte(id))

	m := sarama.ProducerMessage{
		Topic:     "timers",
		Key:       sarama.ByteEncoder(idBytes),
		Partition: partition,
	}

	_, _, err = c.prod.SendMessage(&m)
	return err
}

func (c *Client) bytesToPartition(id [16]byte) int32 {
	var partition int32

	for b := range id {
		partition = partition << 8
		partition += int32(b)
		partition = partition % c.partitions
	}

	return partition
}
