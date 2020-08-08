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
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/steady/persisttimers/consumer"
	"github.com/nivista/steady/persisttimers/db"
	"github.com/spf13/viper"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	postgresURL  = "POSTGRES_URL"
	kafkaTopic   = "KAFKA_TOPIC"
	kafkaBrokers = "KAFKA_BROKERS"
	kafkaVersion = "KAFKA_VERSION"
	kafkaGroupID = "KAFKA_GROUP_ID"
)

func main() {
	viper.SetDefault(postgresURL, "postgresql://")
	viper.SetDefault(kafkaTopic, "timers")
	viper.SetDefault(kafkaBrokers, "localhost:9092")
	viper.SetDefault(kafkaVersion, "2.2.1")
	viper.SetDefault(kafkaGroupID, "persisters")

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

	// set up database
	ctx, cancel := context.WithCancel(context.Background())

	dbConn, err := pgxpool.Connect(ctx, viper.GetString(postgresURL))
	if err != nil {
		panic(err)
	}

	db := db.NewClient(dbConn)

	// Get Kafka Version
	version, err := sarama.ParseKafkaVersion(viper.GetString("KAFKA_VERSION"))
	if err != nil {
		panic(err)
	}

	// Kafka configuration
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// set up consumer group
	client, err := sarama.NewConsumerGroup(viper.GetStringSlice(kafkaBrokers), viper.GetString(kafkaGroupID), config)
	if err != nil {
		panic(err)
	}

	// set up consumer
	consumer := consumer.NewConsumer(db)

	// consume
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{viper.GetString(kafkaTopic)}, consumer); err != nil {
				// panic if the consumer stops working
				// maybe instead send a signal and then cancel context?
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
