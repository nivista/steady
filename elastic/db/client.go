package db

import (
	"bytes"
	"context"
	"encoding/json"
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
		ExecuteTimer(ctx context.Context, id string, kafkaTimestamp time.Time, value *messaging.Execute) error
		CreateTimer(ctx context.Context, id string, value *messaging.Create)
	}

	client struct {
		elastic              *elasticsearch.Client
		execIndex, progIndex string
	}
)

func NewClient(elastic *elasticsearch.Client, executionsIndex, progressIndex string) Client {
	return &client{
		elastic:   elastic,
		execIndex: executionsIndex,
		progIndex: progressIndex,
	}
}

func (c *client) ExecuteTimer(ctx context.Context, id string, kafkaTimestamp time.Time, value *messaging.Execute) error {
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

	res, err := c.elastic.Create(c.execIndex, strings.Join([]string{id, kafkaTimestamp.String()}, "-"), bytes.NewReader(doc))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(ioutil.Discard, res.Body)
	progJSON, err := protojson.Marshal(value.Progress)
	fmt.Println(fmt.Sprint(`{
		"doc": `, progJSON,
		`,
		"doc_as_upsert": true	
	}`))
	res2, err := c.elastic.Update(c.progIndex, id,
		strings.NewReader(fmt.Sprint(`{
			"doc": `, string(progJSON),
			`,
			"doc_as_upsert": true	
		}`)),
	)
	if err != nil {
		return err
	}
	defer res2.Body.Close()

	var b bytes.Buffer
	io.Copy(&b, res2.Body)
	fmt.Println(b.String())
	return nil
}
