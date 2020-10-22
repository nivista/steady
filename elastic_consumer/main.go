package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/elastic/go-elasticsearch"
	"github.com/nivista/steady/elastic_consumer/consumer"
	"github.com/nivista/steady/elastic_consumer/db"

	"github.com/spf13/viper"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	elasticURL             = "ELASTIC_URL"
	elasticExecutionsIndex = "EXECUTIONS"
	elasticProgressIndex   = "PROGRESS"
	elasticTimersIndex     = "TIMERS"
	executeTopic           = "EXECUTE_KAFKA_TOPIC"
	createTopic            = "CREATE_KAFKA_TOPIC"
	kafkaBrokers           = "KAFKA_BROKERS"
	kafkaVersion           = "KAFKA_VERSION"
	kafkaGroupID           = "KAFKA_GROUP_ID"
)

func init() {
	viper.SetDefault(elasticURL, "http://localhost:9200")
	viper.SetDefault(elasticExecutionsIndex, "executions")
	viper.SetDefault(elasticProgressIndex, "progress")
	viper.SetDefault(elasticTimersIndex, "timers")
	viper.SetDefault(executeTopic, "execute")
	viper.SetDefault(createTopic, "create")
	viper.SetDefault(kafkaBrokers, "localhost:9092")
	viper.SetDefault(kafkaVersion, "2.2.1")
	viper.SetDefault(kafkaGroupID, "elastic")

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
}

func main() {
	// set up database
	elasticCfg := elasticsearch.Config{
		Addresses: []string{viper.GetString(elasticURL)},
	}
	elasticClient, err := elasticsearch.NewClient(elasticCfg)
	if err != nil {
		panic(err)
	}
	db := db.NewClient(elasticClient, viper.GetString(elasticExecutionsIndex), viper.GetString(elasticProgressIndex), viper.GetString(elasticTimersIndex))

	// get Kafka Version
	version, err := sarama.ParseKafkaVersion(viper.GetString("KAFKA_VERSION"))
	if err != nil {
		panic(err)
	}

	// Kafka configuration
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// set up consumer
	client, err := sarama.NewConsumerGroup(viper.GetStringSlice(kafkaBrokers), viper.GetString(kafkaGroupID), config)
	if err != nil {
		panic(err)
	}
	consumer := consumer.NewConsumer(db, viper.GetString(createTopic), viper.GetString(executeTopic))

	// consume
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{viper.GetString(createTopic), viper.GetString(executeTopic)}, consumer); err != nil {
				// panic if the consumer stops working
				// maybe instead send a signal and then cancel context?
				panic(fmt.Sprintf("Error from consumer: %v", err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm
	fmt.Println("terminating via signal")

	cancel()

	wg.Wait()
	if err = client.Close(); err != nil {
		panic(fmt.Sprintf("Error closing client: %v", err))
	}
}
