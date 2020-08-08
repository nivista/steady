package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/webservice/db"
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
	kafkaTopic   = "KAFKA_TOPIC"
	partitions   = "PARTITIONS"
	kafkaBrokers = "KAFKA_BROKERS"
	kafkaVersion = "KAFKA_VERSION"
	addr         = "ADDR"
)

func main() {
	viper.SetDefault(postgresURL, "postgresql://")
	viper.SetDefault(kafkaTopic, "timers")
	viper.SetDefault(partitions, 3)
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

	queueClient := queue.NewClient(producer, viper.GetInt32(partitions), viper.GetString(kafkaTopic))

	// set up database
	ctx, cancel := context.WithCancel(context.Background())

	dbConn, err := pgxpool.Connect(ctx, viper.GetString(postgresURL))
	if err != nil {
		panic(err)
	}
	db := db.NewClient(dbConn)

	// set up server
	steadyServer := server.NewServer(db, queueClient)

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

	// cancel context
	cancel()
}
