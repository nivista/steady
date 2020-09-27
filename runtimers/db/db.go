package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/esapi"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	Client interface {
		GetProgresses(ctx context.Context, partition int) (map[string]*messaging.Progress, error)
	}

	client struct {
		elastic       *elasticsearch.Client
		progressIndex string
	}
)

func NewClient(elastic *elasticsearch.Client, progressIndex string) Client {
	return &client{
		elastic:       elastic,
		progressIndex: progressIndex,
	}
}

type elasticFormat struct {
	Hits []struct {
		ID     *string             `json:"_id"`
		Source *messaging.Progress `json:"_source"`
	} `json:"hits"`
}

func (c *client) GetProgresses(ctx context.Context, partition int) (map[string]*messaging.Progress, error) {

	req := esapi.SearchRequest{
		Index: []string{c.progressIndex},
		Body:  strings.NewReader(fmt.Sprint(`{"query": { "match" : { "partition": `, partition, `}}}`)),
	}

	res, err := req.Do(ctx, c.elastic)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var b bytes.Buffer
	io.Copy(&b, res.Body)

	var eF elasticFormat
	if err = json.Unmarshal(b.Bytes(), &eF); err != nil {
		return nil, err
	}

	out := make(map[string]*messaging.Progress)

	for _, hit := range eF.Hits {
		out[*hit.ID] = hit.Source
	}

	return out, nil

}
