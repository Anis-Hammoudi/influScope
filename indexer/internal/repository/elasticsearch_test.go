package repository

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/hammo/influScope/pkg/models"
)

// MockElasticsearchTransport simulates Elasticsearch responses
type MockElasticsearchTransport struct {
	responses []MockResponse
	callCount int
	mu        sync.Mutex
}

type MockResponse struct {
	statusCode int
	body       string
	err        error
}

func (m *MockElasticsearchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	header := make(http.Header)
	header.Set("X-Elastic-Product", "Elasticsearch")
	header.Set("Content-Type", "application/json")

	// Handle the Ping/Info request during initialization
	if req.Method == "GET" && req.URL.Path == "/" {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"version":{"number":"7.17.0"},"tagline":"You Know, for Search"}`)),
			Header:     header,
		}, nil
	}

	if m.callCount >= len(m.responses) {
		m.callCount++
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"result":"created"}`)),
			Header:     header,
		}, nil
	}

	resp := m.responses[m.callCount]
	m.callCount++

	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(resp.body)),
		Header:     header,
	}, nil
}

func TestSuccessfulIndexing(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{{statusCode: 200, body: `{"result":"created"}`}},
	}

	esClient, _ := elasticsearch.NewClient(elasticsearch.Config{Transport: mockTransport})
	repo := &esRepository{client: esClient, indexName: "test-index"}

	profile := &models.Influencer{Username: "testuser", Followers: 1000}
	err := repo.IndexProfile(context.Background(), profile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if mockTransport.callCount != 1 {
		t.Errorf("Expected 1 Elasticsearch call, got %d", mockTransport.callCount)
	}
}

func TestIndexingWithNetworkError(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{{err: context.DeadlineExceeded}},
	}

	esClient, _ := elasticsearch.NewClient(elasticsearch.Config{Transport: mockTransport})
	repo := &esRepository{client: esClient, indexName: "test-index"}

	profile := &models.Influencer{Username: "testuser"}
	err := repo.IndexProfile(context.Background(), profile)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}
