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
	// Client is a client to the database.
	Client interface {
		UpsertUser(ctx context.Context, apiToken, apiKey string) error
		AuthenticateUser(ctx context.Context, apiToken, apiSecret string) error
	}

	// InvalidAPIToken is the error returned when provided with an invalid APIToken
	InvalidAPIToken error

	// InvalidAPISecret is the error returned when provided with an invalid APISecret
	InvalidAPISecret error

	client struct {
		elastic    *elasticsearch.Client
		usersIndex string
	}
)

var (
	errInvalidAPIToken  InvalidAPIToken  = errors.New("invalid api token")
	errInvalidAPISecret InvalidAPISecret = errors.New("invalid api secret")
)

// NewClient returns a new client to the database.
func NewClient(elastic *elasticsearch.Client, usersIndex string) Client {
	return &client{
		elastic:    elastic,
		usersIndex: usersIndex,
	}
}

func (c *client) UpsertUser(ctx context.Context, apiToken, hashedAPISecret string) error {
	// index : users
	// id    : just the id of the user

	indexRequest := esapi.IndexRequest{
		Index:      c.usersIndex,
		DocumentID: apiToken,
		Body: strings.NewReader(fmt.Sprint(`{
			"doc": {
				"hashed_api_key": "`, hashedAPISecret,
			`" }	
		}`)),
	}
	_, err := indexRequest.Do(ctx, c.elastic.Transport)
	return err

}

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

	var elasticRes struct {
		Found  bool `json:"found"`
		Source struct {
			Doc struct {
				HashedAPIKey string `json:"hashed_api_key"`
			} `json:"doc"`
		} `json:"_source"`
	}

	err = json.Unmarshal(data, &elasticRes)
	if err != nil {
		return err
	}

	if elRes.Found == false {
		return errInvalidAPIToken
	}

	if err = bcrypt.CompareHashAndPassword([]byte(elRes.Source.Doc.HashedAPIKey), []byte(apiSecret)); err != nil {
		return errInvalidAPISecret
	}

	return nil
}
