package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/hammo/influScope/pkg/models"
)

type esRepository struct {
	client    *elasticsearch.Client
	indexName string
}

func NewESRepository(address, index string) (*esRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Retry loop for ES startup
	for i := 0; i < 10; i++ {
		res, err := es.Info()
		if err == nil && res.StatusCode == 200 {
			res.Body.Close()
			return &esRepository{client: es, indexName: index}, nil
		}
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to elasticsearch after retries")
}

func (r *esRepository) IndexProfile(ctx context.Context, profile *models.Influencer) error {
	body, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		r.indexName,
		bytes.NewReader(body),
		r.client.Index.WithRefresh("true"),
		r.client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.String())
	}
	return nil
}
