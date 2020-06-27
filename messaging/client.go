package messaging

import (
	"fmt"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/nivista/steady/timer"
	"github.com/nivista/steady/utils"
)

type Client struct {
	prod       sarama.SyncProducer
	partitions int32
}

func NewClient() *Client {
	brokerList := strings.Split(os.Getenv("KAFKA_PEERS"), ",")

	broker := sarama.NewBroker(brokerList[0])
	brokerConfig := sarama.NewConfig()
	err := broker.Open(brokerConfig)
	if err != nil {
		panic(err)
	}

	req := sarama.MetadataRequest{
		Topics: []string{"timers"},
	}

	res, err := broker.GetMetadata(&req)
	if err != nil {
		panic(err)
	}
	partitions := len(res.Topics[0].Partitions)
	fmt.Println("partitions:", partitions)

	if err != nil {
		panic(err)
	}

	prodConfig := sarama.NewConfig()
	prodConfig.Producer.RequiredAcks = sarama.WaitForAll
	prodConfig.Producer.Retry.Max = 10
	prodConfig.Producer.Return.Successes = true

	prod, err := sarama.NewSyncProducer(brokerList, prodConfig)
	if err != nil {
		panic(err)
	}
	c := Client{prod: prod, partitions: int32(partitions)}
	return &c
}

func (c *Client) PublishAdd(t *timer.Timer) error {
	return c.upsertTimer(t)
}

func (c *Client) PublishUpdate(t *timer.Timer) error {
	return c.upsertTimer(t)
}

func (c *Client) PublishDelete(domain string, id uuid.UUID) error {
	key, err := utils.NewTimerKey(domain, id)
	if err != nil {
		return err
	}

	partition := c.bytesToPartition([16]byte(id))

	m := sarama.ProducerMessage{
		Topic:     "timers",
		Key:       key,
		Partition: partition,
	}

	_, _, err = c.prod.SendMessage(&m)
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
	partition := c.bytesToPartition([16]byte(t.ID))

	m := sarama.ProducerMessage{
		Topic:     "timers",
		Key:       key,
		Value:     sarama.ByteEncoder(timerBytes),
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
