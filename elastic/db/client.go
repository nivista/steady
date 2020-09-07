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
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	Client interface {
		ExecuteTimer(ctx context.Context, domain, id string, kafkaTimestamp time.Time, value *messaging.Execute) error
		CreateTimer(ctx context.Context, domain, id string, value *messaging.Create) error
		DeleteTimer(ctx context.Context, domain, id string, value *messaging.Create) error
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

func (c *client) ExecuteTimer(ctx context.Context, domain, id string, kafkaTimestamp time.Time, value *messaging.Execute) error {
	// write execution
	// index : c.execIndex + "-" + value.Domain
	// _id   : value.TimerUUID + kafkaTimestamp

	var result map[string]interface{}
	err := json.Unmarshal(value.Result, &result)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["TimerUUID"] = id
	data["KafkaTimestamp"] = kafkaTimestamp
	data["Result"] = result

	doc, err := json.Marshal(data)
	if err != nil {
		return err
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

	progJSON, err := marshaller.Marshal(value.Progress)

	res2, err := c.elastic.Update(c.progIndex, id,
		strings.NewReader(fmt.Sprint(`{
			"doc": `, string(progJSON),
			`,
			"doc_as_upsert": true	
		}`)),
	)
	if err != nil {
		return errors.New("error updating progress: " + err.Error())
	}
	defer res2.Body.Close()

	var b bytes.Buffer
	io.Copy(&b, res2.Body)

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

// big realization:: to do the delete i need the domain. This implies key should be domain-id
func (c *client) DeleteTimer(ctx context.Context, domain, id string, value *messaging.Create) error {
	_, err := c.elastic.Delete(strings.Join([]string{c.timerIndex, domain}, "-"), id)
	if err != nil {
		return err
	}
	return nil
}

// i've done this like 5 times, how should i make a key include a uuid and a string
// is it better to use protos or is it better to use custom encoding
// pro for protos: more consistent w/ everything else
// con for protos: have to go from []byte to string to use as key
// i think protos is better

// TODO

// change key to be proto w/ domain
// finish this junk
// do elasticdb user logic
// do oauth stuff
// get it working
// start working on making it nice for andrew
