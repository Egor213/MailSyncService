package elastic

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client обёртка над официальным клиентом Elasticsearch
type Client struct {
	*elasticsearch.Client
}

// NewClient создаёт новый клиент Elasticsearch
func NewClient(addresses []string, username, password string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client: %w", err)
	}
	// Проверяем соединение
	res, err := es.Ping()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch ping: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch ping error: %s", res.String())
	}
	return &Client{Client: es}, nil
}

// Index индексирует документ с указанным ID
func (c *Client) Index(index, documentID string, body io.Reader, opts ...func(*esapi.IndexRequest)) (*esapi.Response, error) {
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       body,
	}
	for _, opt := range opts {
		opt(&req)
	}
	return req.Do(context.Background(), c.Client)
}

// Update обновляет документ (частичное обновление)
func (c *Client) Update(index, documentID string, body io.Reader, opts ...func(*esapi.UpdateRequest)) (*esapi.Response, error) {
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       body,
	}
	for _, opt := range opts {
		opt(&req)
	}
	return req.Do(context.Background(), c.Client)
}

// Get возвращает документ по ID
func (c *Client) Get(index, documentID string, opts ...func(*esapi.GetRequest)) (*esapi.Response, error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: documentID,
	}
	for _, opt := range opts {
		opt(&req)
	}
	return req.Do(context.Background(), c.Client)
}

// Search выполняет поиск
func (c *Client) Search(index string, body io.Reader, opts ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  body,
	}
	for _, opt := range opts {
		opt(&req)
	}
	return req.Do(context.Background(), c.Client)
}
