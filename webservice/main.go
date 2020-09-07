package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/elastic/go-elasticsearch"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/nivista/steady/webservice/app"
	"github.com/nivista/steady/webservice/db"
	"github.com/nivista/steady/webservice/queue"
	"github.com/nivista/steady/webservice/server"
	"github.com/nivista/steady/webservice/util"
	"github.com/soheilhy/cmux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	configURL = flag.String("config", "", "path of config file")
)

const (
	elasticURL             = "ELASTIC_URL"
	elasticExecutionsIndex = "EXECUTIONS"
	elasticTimersIndex     = "TIMERS"
	postgresURL            = "POSTGRES_URL"
	createTopic            = "KAFKA_TOPIC"
	partitions             = "PARTITIONS"
	kafkaBrokers           = "KAFKA_BROKERS"
	kafkaVersion           = "KAFKA_VERSION"
	addr                   = "ADDR"
)

func main() {
	viper.SetDefault(elasticURL, "http://localhost:9200")
	viper.SetDefault(elasticExecutionsIndex, "executions")
	viper.SetDefault(elasticTimersIndex, "timers")
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

	// set up db
	elasticClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	dbClient := db.NewClient(elasticClient)

	// set up server
	l, err := net.Listen("tcp", viper.GetString(addr))
	if err != nil {
		panic(err)
	}

	m := cmux.New(l)
	m.MatchWithWriters()
	grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	restListener := m.Match(cmux.Any())

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.Internal, "")
		}

		s := md.Get("Authorization")
		if len(s) == 0 {
			return nil, grpc.Errorf(codes.Unauthenticated, "")
		}

		clientID, clientSecret, ok := util.ParseBasicAuth(s[0])
		if !ok {
			return nil, grpc.Errorf(codes.Unauthenticated, "")
		}

		err = dbClient.AuthenticateUser(ctx, clientID, clientSecret)
		if err != nil {
			switch err.(type) {
			case db.InvalidAPIToken:
				return nil, grpc.Errorf(codes.Unauthenticated, "")

			case db.InvalidAPISecret:
				return nil, grpc.Errorf(codes.Unauthenticated, "")
			}

			return nil, grpc.Errorf(codes.Internal, "")
		}

		return handler(util.SetClientID(ctx, clientID), req)

	}))

	steadyServer := server.NewServer(queueClient)
	services.RegisterSteadyServer(grpcServer, steadyServer)

	restServer := app.NewApp(dbClient, queueClient, viper.GetString(elasticTimersIndex), viper.GetString(elasticExecutionsIndex), viper.GetString(elasticURL))

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

// the problem
// i can either use a mux, or a tree thing
// tree thing is complicated
// mux i have a hard time doing basically a middleware for some endpoints and not others

// how would i do tree thing
// app has all dependencies and servehttp func
// servehttp func parses the first thing in the path

// the semantics i want is the later endpoints require an authenticated req
// an authenticated request should basically close a userid
// if its auth, it creates and calls an auth
// if its not auth, it calls auth and then

type dummyserver struct{}

func (_ dummyserver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("dumy")
}

type wrapper struct {
	grpc *grpc.Server
}

func (wr wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("wrapper")
	wr.grpc.ServeHTTP(w, r)
}
