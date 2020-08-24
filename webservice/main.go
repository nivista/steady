package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/webservice/queue"
	"github.com/nivista/steady/webservice/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	postgresURL  = "POSTGRES_URL"
	createTopic  = "KAFKA_TOPIC"
	partitions   = "PARTITIONS"
	kafkaBrokers = "KAFKA_BROKERS"
	kafkaVersion = "KAFKA_VERSION"
	addr         = "ADDR"
)

func main() {
	viper.SetDefault(postgresURL, "postgresql://")
	viper.SetDefault(createTopic, "create")
	viper.SetDefault(partitions, 1)
	viper.SetDefault(kafkaBrokers, "localhost:9092")
	viper.SetDefault(kafkaVersion, "2.2.1")
	viper.SetDefault(addr, "localhost:8080")

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
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	// set up queue
	producer, err := sarama.NewSyncProducer(viper.GetStringSlice(kafkaBrokers), config)
	if err != nil {
		panic(err)
	}

	queueClient := queue.NewClient(producer, viper.GetInt32(partitions), viper.GetString(createTopic))

	// set up server
	steadyServer := server.NewServer(queueClient)

	conn, err := net.Listen("tcp", viper.GetString(addr))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	services.RegisterSteadyServer(s, steadyServer)

	go func() {
		if err := s.Serve(conn); err != nil {
			// maybe just cancel context here?
			panic(fmt.Sprintf("failed to serve: %v", err))
		}
	}()

	// catch os signals
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// block for signals
	<-sigterm
	log.Println("terminating: via signal")
}
