package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/elastic/go-elasticsearch"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/frontend/db"
	"github.com/nivista/steady/frontend/queue"
	"github.com/nivista/steady/frontend/services/rest"
	"github.com/nivista/steady/frontend/services/rpc"
	"github.com/soheilhy/cmux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	elasticURL             = "ELASTIC_URL"
	elasticExecutionsIndex = "EXECUTIONS"
	elasticTimersIndex     = "TIMERS"
	elasticUsersIndex      = "USERS"
	postgresURL            = "POSTGRES_URL"
	createTopic            = "KAFKA_TOPIC"
	partitions             = "PARTITIONS"
	kafkaBrokers           = "KAFKA_BROKERS"
	kafkaVersion           = "KAFKA_VERSION"
	addr                   = "ADDR"
)

func init() {
	viper.SetDefault(elasticURL, "http://localhost:9200")
	viper.SetDefault(elasticExecutionsIndex, "executions")
	viper.SetDefault(elasticTimersIndex, "timers")
	viper.SetDefault(elasticUsersIndex, "users")
	viper.SetDefault(postgresURL, "postgresql://")
	viper.SetDefault(createTopic, "create")
	viper.SetDefault(partitions, 1)
	viper.SetDefault(kafkaBrokers, "localhost:9092")
	viper.SetDefault(kafkaVersion, "2.2.1")
	viper.SetDefault(addr, "127.0.0.1:8080")

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

	// Kafka configuration
	version, err := sarama.ParseKafkaVersion(viper.GetString("KAFKA_VERSION"))
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	// Queue setup
	producer, err := sarama.NewSyncProducer(viper.GetStringSlice(kafkaBrokers), config)
	if err != nil {
		panic(err)
	}
	queueClient := queue.NewClient(producer, viper.GetInt(partitions), viper.GetString(createTopic))

	// DB setup
	elasticClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}
	dbClient := db.NewClient(elasticClient, viper.GetString(elasticUsersIndex))

	l, err := net.Listen("tcp", viper.GetString(addr))
	if err != nil {
		panic(err)
	}

	// demultiplex connections to two listeners, one that recieves all grpc connections, one that recieves rest connections.
	m := cmux.New(l)
	m.MatchWithWriters()
	grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	restListener := m.Match(cmux.Any())

	// server for handling grpc requests
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(rpc.GetAuth(dbClient)))
	steadyService := rpc.NewServer(queueClient)
	services.RegisterSteadyServer(grpcServer, steadyService)

	// server for auth and elastic redirects
	elasticURL, err := url.Parse(viper.GetString(elasticURL))
	if err != nil {
		panic(err)
	}
	restServer := rest.NewApp(dbClient, queueClient, viper.GetString(elasticTimersIndex), viper.GetString(elasticExecutionsIndex), elasticURL)

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := http.Serve(restListener, restServer); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := m.Serve(); err != nil {
			panic(err)
		}
	}()

	// catch os signals
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// block for signals
	<-sigterm
	log.Println("terminating: via signal")
}
