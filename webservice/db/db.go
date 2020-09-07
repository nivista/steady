package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"golang.org/x/crypto/bcrypt"
)

type (
	Client interface {
		UpsertUser(ctx context.Context, id, apiKey string) error
		AuthenticateUser(ctx context.Context, apiToken, apiSecret string) error
	}

	client struct {
		elastic *elasticsearch.Client
	}
)

func NewClient(elastic *elasticsearch.Client) Client {
	return &client{
		elastic: elastic,
	}
}

func (c *client) UpsertUser(ctx context.Context, id, hashedApiSecret string) error {
	// index : users
	// id    : just the id of the user

	indexRequest := esapi.IndexRequest{
		Index:      "users",
		DocumentID: id,
		Body: strings.NewReader(fmt.Sprint(`{
			"doc": {
				"hashed_api_key": "`, hashedApiSecret,
			`" }	
		}`)),
	}
	_, err := indexRequest.Do(ctx, c.elastic.Transport)
	return err

}

type (
	InvalidAPIToken  error
	InvalidAPISecret error
)

func (c *client) AuthenticateUser(ctx context.Context, apiToken, apiSecret string) error {
	res, err := c.elastic.Get("users", apiToken)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	type elasticResponse struct {
		Found  bool `json:"found"`
		Source struct {
			Doc struct {
				HashedAPIKey string `json:"hashed_api_key"`
			} `json:"doc"`
		} `json:"_source"`
	}

	var elRes elasticResponse
	err = json.Unmarshal(data, &elRes)
	if err != nil {
		return err
	}

	if elRes.Found == false {
		return InvalidAPIToken(errors.New("invalid api token"))
	}

	if err = bcrypt.CompareHashAndPassword([]byte(elRes.Source.Doc.HashedAPIKey), []byte(apiSecret)); err != nil {
		return InvalidAPISecret(err)
	}

	return nil
}
