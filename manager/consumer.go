package manager

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/messaging"
	"github.com/nivista/steady/timer"
	"github.com/nivista/steady/utils"
)

// InitConsumer starts consuming timers from Kafka and executing them.
func InitConsumer(c *messaging.Client) {

	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion("2.2.1")
	if err != nil {
		panic(err)
	}
	config.Version = version

	config.Consumer.Group.Rebalance.Strategy = consistentHash(0)

	cord := newCoordinator(c)
	myConsumer := consumer{
		queue: c,
		cord:  cord,
		ready: make(chan bool),
	}

	brokers := os.Getenv("KAFKA_PEERS")
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), "timers", config)

	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {

			if err := client.Consume(ctx, []string{"timers"}, &myConsumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			myConsumer.ready = make(chan bool)
		}
	}()

	<-myConsumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}

}

type consumer struct {
	queue *messaging.Client
	ready chan bool
	cord  *coordinator
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()["timers"]

	for _, partition := range partitions {
		if !c.cord.hasPartition(partition) {
			session.ResetOffset("timers", partition, 0, "")
		}
	}

	c.cord.handleRepartition(partitions)

	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	c.cord.setHighWatermark(claim.Partition(), claim.HighWaterMarkOffset())

	for message := range claim.Messages() {

		_, id, _ := utils.ParseTimerKey(message.Key)
		fmt.Println("id:", id)
		if len(message.Value) == 0 {
			_, id, err := utils.ParseTimerKey(message.Key)
			if err != nil {
				fmt.Println("error parsing timer key: ", err)
			}
			c.cord.removeTimer(id, claim.Partition(), message.Offset)
		} else {
			var t timer.Timer
			err := t.UnmarshalBinary(message.Value)
			if err != nil {
				fmt.Printf("manager/consumer.go unmarshalbinary :: %v\n", err)
			}
			if t.ExecutionCount == 0 {
				c.cord.addTimer(&t, claim.Partition(), message.Offset)
			}
		}
		session.MarkMessage(message, "")
	}

	return nil
}
