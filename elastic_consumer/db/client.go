package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nivista/steady/elastic"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	// Client is a client to the data store.
	Client interface {
		ExecuteTimer(ctx context.Context, domain, id string, partition int32, kafkaTimestamp time.Time, value *messaging.Execute) error
		CreateTimer(ctx context.Context, domain, id string, value *messaging.Create) error
		DeleteTimer(ctx context.Context, domain, id string) error
	}

	client struct {
		elastic                          *elasticsearch.Client
		execIndex, progIndex, timerIndex string
	}
)

var marshaller = protojson.MarshalOptions{EmitUnpopulated: true}

func NewClient(elastic *elasticsearch.Client, executionsIndex, progressIndex, timerIndex string) Client {
	return &client{
		elastic:    elastic,
		execIndex:  executionsIndex,
		progIndex:  progressIndex,
		timerIndex: timerIndex,
	}
}

func (c *client) ExecuteTimer(ctx context.Context, domain, id string, partition int32, kafkaTimestamp time.Time, value *messaging.Execute) error {
	// write execution
	// index : c.execIndex + "-" + value.Domain
	// _id   : value.TimerUUID + kafkaTimestamp

	executeTimer := elastic.ExecuteTimer{
		TimerUUID:      id,
		KafkaTimestamp: kafkaTimestamp,
		Result:         value.Result,
	}

	doc, err := json.Marshal(executeTimer)
	if err != nil {
		return fmt.Errorf("marshalling executeTimer: %v", err)
	}

	res, err := c.elastic.Create(strings.Join([]string{c.execIndex, domain}, "-"),
		strings.Join([]string{id, kafkaTimestamp.String()}, "-"), bytes.NewReader(doc))
	if err != nil {
		return errors.New("error creating execution record: " + err.Error())
	}
	defer res.Body.Close()
	io.Copy(ioutil.Discard, res.Body)

	// write progress
	// index : x.progIndex
	// _id : value.TimerUUID

	progress := elastic.Progress{
		LastExecution:       value.Progress.LastExecution.AsTime(),
		CompletedExecutions: int(value.Progress.CompletedExecutions),
	}
	doc, err = json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("marshaling progress: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      c.progIndex,
		DocumentID: id,
		Body:       bytes.NewReader(doc),
	}
	res, err = req.Do(ctx, c.elastic.Transport)
	if err != nil {
		return fmt.Errorf("indexing progress: %w", err)
	}
	defer res.Body.Close()

	defer res.Body.Close()
	io.Copy(ioutil.Discard, res.Body)

	return nil
}

func (c *client) CreateTimer(ctx context.Context, domain, id string, value *messaging.Create) error {
	// write timer
	// index : c.timerIndex + "-" + value.Domain
	// id    : value.TimerUUID

	json, err := marshaller.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.elastic.Create(strings.Join([]string{c.timerIndex, domain}, "-"), id, bytes.NewReader(json))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) DeleteTimer(ctx context.Context, domain, id string) error {
	_, err := c.elastic.Delete(strings.Join([]string{c.timerIndex, domain}, "-"), id)
	if err != nil {
		return err
	}
	return nil
}
