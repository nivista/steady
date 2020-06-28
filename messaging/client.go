package messaging

import (
	"os"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/nivista/steady/timer"
	"github.com/nivista/steady/utils"
)

// Client is a synchronous messaging client to the queue.
type Client struct {
	producer   sarama.SyncProducer
	partitions int
	topic      string
}

// NewClient returns a new Client.
func NewClient() *Client {
	client := Client{}

	client.topic = os.Getenv("TIMER_TOPIC")

	partitions, err := strconv.Atoi(os.Getenv("TIMER_PARTITIONS"))
	if err != nil {
		panic(err)
	}
	client.partitions = partitions

	prodConfig := sarama.NewConfig()
	prodConfig.Producer.RequiredAcks = sarama.WaitForAll
	prodConfig.Producer.Retry.Max = 10
	prodConfig.Producer.Return.Successes = true

	brokerAddrs := strings.Split(os.Getenv("KAFKA_PEERS"), ":")
	producer, err := sarama.NewSyncProducer(brokerAddrs, prodConfig)
	if err != nil {
		panic(err)
	}
	client.producer = producer

	return &client
}

// PublishCreate publishes a timer creation to the queue synchronously.
func (c *Client) PublishCreate(t *timer.Timer) error {
	return c.upsertTimer(t)
}

// PublishUpdate publishes a timer update to the queue synchronously.
func (c *Client) PublishUpdate(t *timer.Timer) error {
	return c.upsertTimer(t)
}

// PublishDelete publishes a timer deletion to the queue synchronously.
func (c *Client) PublishDelete(domain string, id uuid.UUID) error {
	key, err := utils.NewTimerKey(domain, id)
	if err != nil {
		return err
	}

	partition := c.idToPartition(id)

	m := sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       key,
		Partition: partition,
	}

	_, _, err = c.producer.SendMessage(&m)
	return err
}

func (c *Client) upsertTimer(t *timer.Timer) error {
	key, err := utils.NewTimerKey(t.Domain, t.ID)
	if err != nil {
		return err
	}

	timerBytes, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	partition := c.idToPartition(t.ID)

	m := sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       key,
		Value:     sarama.ByteEncoder(timerBytes),
		Partition: partition,
	}

	_, _, err = c.producer.SendMessage(&m)
	return err
}

func (c *Client) idToPartition(id [16]byte) int32 {
	var partition int

	for b := range id {
		partition <<= 8
		partition += b
		partition %= c.partitions
	}

	return int32(partition)
}
