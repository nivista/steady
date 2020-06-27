package store

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
	"github.com/nivista/steady/timer"
	"github.com/nivista/steady/utils"
)

func InitConsumer(c *Client) {

	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion("2.2.1")
	if err != nil {
		panic(err)
	}
	config.Version = version

	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	myConsumer := consumer{db: c, ready: make(chan bool)}

	brokers := os.Getenv("KAFKA_PEERS")
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), "store", config)

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
	db    *Client
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
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

	for message := range claim.Messages() {
		if len(message.Value) == 0 {
			domain, id, err := utils.ParseTimerKey(message.Key)
			if err != nil {
				fmt.Println("Unable to parse key: ", err)
			}

			err = c.db.DeleteTimer(context.Background(), domain, id)
			if err != nil {
				fmt.Printf("store/consumer createtimer error :: %v\n", err)
			}
		} else {
			var t timer.Timer

			err := t.UnmarshalBinary(message.Value)
			if err != nil {
				fmt.Printf("store/consumer.go unmarshalbinary :: %v\n", err)
			}

			err = c.db.CreateTimer(context.Background(), t)
			if err != nil {
				fmt.Printf("store/consumer createtimer error :: %v\n", err)
			}
		}
		session.MarkMessage(message, "")
	}

	return nil
}
