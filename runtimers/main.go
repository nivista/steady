package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/runtimers/consumer"
	"github.com/nivista/steady/runtimers/coordinator"
	"github.com/spf13/viper"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	kafkaTopic   = "KAFKA_TOPIC"
	partitions   = "PARTITIONS"
	nodeID       = "NODE_ID"
	kafkaBrokers = "KAFKA_BROKERS"
	kafkaVersion = "KAFKA_VERSION"
	kafkaGroupID = "KAFKA_GROUP_ID"
)

func main() {
	viper.SetDefault(kafkaTopic, "timers")
	viper.SetDefault(partitions, 3)
	viper.SetDefault(nodeID, "")
	viper.SetDefault(kafkaBrokers, "localhost:9092")
	viper.SetDefault(kafkaVersion, "2.2.1")
	viper.SetDefault(kafkaGroupID, "executers")

	if *configURL != "" {
		viper.AddConfigPath(*configURL)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				panic("config file not found")
			} else {
				panic(fmt.Sprint("err reading config:", err.Error()))
			}
		}
	}
	// Get Kafka Version
	version, err := sarama.ParseKafkaVersion(viper.GetString("KAFKA_VERSION"))
	if err != nil {
		panic(err)
	}

	// Kafka configuration
	config := sarama.NewConfig()
	config.Version = version
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Consumer.Group.Rebalance.Strategy = consumer.ConsistentHash(0)
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// set up producer
	producer, err := sarama.NewAsyncProducer(viper.GetStringSlice("KAFKA_BROKERS"), config)
	if err != nil {
		panic(err)
	}

	go func() {
		for err := range producer.Errors() {
			fmt.Printf("producer error w/ key '%v': %v\n", err.Msg.Key, err.Err.Error())
		}
	}()

	// set up coordinator
	coord := coordinator.NewCoordinator(producer, viper.GetString(kafkaTopic), clockwork.NewRealClock())

	// set up consumer
	consumer := consumer.NewConsumer(producer, coord, viper.GetInt32(partitions), viper.GetString(nodeID), viper.GetString(kafkaTopic))

	// set up consumer group
	client, err := sarama.NewConsumerGroup(viper.GetStringSlice(kafkaBrokers), viper.GetString(kafkaGroupID), config)
	if err != nil {
		panic(err)
	}

	// consume
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{viper.GetString(kafkaTopic)}, consumer); err != nil {
				// panic if the consumer stops working
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		fmt.Println("terminating via signal")
	}
	cancel()

	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}
