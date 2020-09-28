package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/nivista/steady/elastic"
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

// UpsertUser upserts a user to the database.
func (c *client) UpsertUser(ctx context.Context, apiToken, hashedAPISecret string) error {
	data, err := json.Marshal(elastic.Index{
		Doc: elastic.User{
			HashedAPIKey: hashedAPISecret,
		},
	})
	if err != nil {
		return err
	}

	indexRequest := esapi.IndexRequest{
		Index:      c.usersIndex,
		DocumentID: apiToken,
		Body:       bytes.NewReader(data),
	}
	_, err = indexRequest.Do(ctx, c.elastic.Transport)
	return err

}

func (c *client) AuthenticateUser(ctx context.Context, apiToken, apiSecret string) error {
	res, err := c.elastic.Get(c.usersIndex, apiToken)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var get elastic.Get
	err = json.Unmarshal(data, &get)
	if err != nil {
		return err
	}

	if get.Found == false {
		return errInvalidAPIToken
	}

	var user elastic.User
	err = json.Unmarshal(get.Source.Doc, &user)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedAPIKey), []byte(apiSecret)); err != nil {
		return errInvalidAPISecret
	}

	return nil
}
