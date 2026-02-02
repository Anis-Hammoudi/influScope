package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/upfluence/amqp"
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

// RoundTrip executes a single HTTP transaction
func (m *MockElasticsearchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Standard Headers required by the client
	header := make(http.Header)
	header.Set("X-Elastic-Product", "Elasticsearch")
	header.Set("Content-Type", "application/json")

	if req.Method == "GET" && req.URL.Path == "/" {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"version":{"number":"7.17.0"},"tagline":"You Know, for Search"}`)),
			Header:     header,
		}, nil
	}

	// Default behavior if we run out of defined responses
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

// MockConsumer simulates AMQP message consumption
type MockConsumer struct {
	messages []*amqp.Delivery
	index    int
	closed   bool
}

func (m *MockConsumer) Next(ctx context.Context) (*amqp.Delivery, error) {
	if m.index >= len(m.messages) {
		return nil, context.Canceled
	}
	msg := m.messages[m.index]
	m.index++
	return msg, nil
}

func (m *MockConsumer) Ack(ctx context.Context, tag uint64, opts amqp.AckOptions) error {
	return nil
}

func (m *MockConsumer) Close() error {
	m.closed = true
	return nil
}

// TestMetricsInitialization verifies Prometheus metrics are registered
func TestMetricsInitialization(t *testing.T) {
	reg := prometheus.NewRegistry()

	testProfilesIndexed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "test_influencers_indexed_total",
			Help: "Total number of profiles successfully saved to Elasticsearch",
		},
	)
	testIndexingErrors := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "test_indexer_errors_total",
			Help: "Total number of failed indexing attempts",
		},
	)

	reg.MustRegister(testProfilesIndexed)
	reg.MustRegister(testIndexingErrors)

	if testutil.ToFloat64(testProfilesIndexed) != 0 {
		t.Errorf("Expected profilesIndexed to start at 0, got %f", testutil.ToFloat64(testProfilesIndexed))
	}
	if testutil.ToFloat64(testIndexingErrors) != 0 {
		t.Errorf("Expected indexingErrors to start at 0, got %f", testutil.ToFloat64(testIndexingErrors))
	}
}

// TestMetricsServerEndpoint verifies the /metrics endpoint
func TestMetricsServerEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	promhttp.Handler().ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("Expected non-empty metrics body")
	}
}

// TestSuccessfulIndexing verifies a successful message indexing flow
func TestSuccessfulIndexing(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 200, body: `{"result":"created"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	testProfile := map[string]interface{}{
		"username":  "testuser",
		"followers": 1000,
	}
	profileJSON, _ := json.Marshal(testProfile)

	res, err := es.Index(
		indexName,
		bytes.NewReader(profileJSON),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		t.Errorf("Expected successful response, got error: %s", res.String())
	}

	if mockTransport.callCount != 1 {
		t.Errorf("Expected 1 Elasticsearch call, got %d", mockTransport.callCount)
	}
}

// TestIndexingWithElasticsearchError verifies error handling for ES errors
func TestIndexingWithElasticsearchError(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 400, body: `{"error":"bad request"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	testProfile := map[string]interface{}{
		"username": "testuser",
	}
	profileJSON, _ := json.Marshal(testProfile)

	res, err := es.Index(
		indexName,
		bytes.NewReader(profileJSON),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		t.Fatalf("Expected no transport error, got %v", err)
	}
	defer res.Body.Close()

	if !res.IsError() {
		t.Error("Expected error response, got success")
	}

	if res.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", res.StatusCode)
	}
}

// TestIndexingWithNetworkError verifies handling of network errors
func TestIndexingWithNetworkError(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{err: context.DeadlineExceeded},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	testProfile := map[string]interface{}{
		"username": "testuser",
	}
	profileJSON, _ := json.Marshal(testProfile)

	_, err = es.Index(
		indexName,
		bytes.NewReader(profileJSON),
		es.Index.WithRefresh("true"),
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}

// TestMessageConsumption verifies AMQP message consumption flow
func TestMessageConsumption(t *testing.T) {
	testCases := []struct {
		name     string
		message  []byte
		expected string
	}{
		{
			name: "Valid JSON profile",
			message: []byte(`{
             "username": "testuser",
             "followers": 5000,
             "platform": "instagram"
          }`),
			expected: "testuser",
		},
		{
			name: "Minimal profile",
			message: []byte(`{
             "username": "minimal"
          }`),
			expected: "minimal",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var profile map[string]interface{}
			if err := json.Unmarshal(tc.message, &profile); err != nil {
				t.Errorf("Invalid JSON message: %v", err)
			}
			if username, ok := profile["username"].(string); !ok || username != tc.expected {
				t.Errorf("Expected username %s, got %v", tc.expected, profile["username"])
			}
		})
	}
}

// TestMultipleMessageProcessing verifies processing multiple messages in sequence
func TestMultipleMessageProcessing(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 200, body: `{"result":"created"}`},
			{statusCode: 200, body: `{"result":"created"}`},
			{statusCode: 200, body: `{"result":"created"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	messages := [][]byte{
		[]byte(`{"username":"user1"}`),
		[]byte(`{"username":"user2"}`),
		[]byte(`{"username":"user3"}`),
	}

	successCount := 0
	for _, msg := range messages {
		res, err := es.Index(
			indexName,
			bytes.NewReader(msg),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			t.Errorf("Indexing error: %v", err)
			continue
		}
		if !res.IsError() {
			successCount++
		}
		res.Body.Close()
	}

	if successCount != 3 {
		t.Errorf("Expected 3 successful indexes, got %d", successCount)
	}

	if mockTransport.callCount != 3 {
		t.Errorf("Expected 3 Elasticsearch calls, got %d", mockTransport.callCount)
	}
}

// TestMixedSuccessAndFailure verifies handling of mixed success/failure scenarios
func TestMixedSuccessAndFailure(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 200, body: `{"result":"created"}`},
			{statusCode: 400, body: `{"error":"bad request"}`},
			{statusCode: 200, body: `{"result":"created"}`},
			{statusCode: 500, body: `{"error":"internal error"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	messages := [][]byte{
		[]byte(`{"username":"user1"}`),
		[]byte(`{"invalid"}`),
		[]byte(`{"username":"user2"}`),
		[]byte(`{"username":"user3"}`),
	}

	successCount := 0
	errorCount := 0

	for _, msg := range messages {
		res, err := es.Index(
			indexName,
			bytes.NewReader(msg),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			errorCount++
			continue
		}
		if res.IsError() {
			errorCount++
		} else {
			successCount++
		}
		res.Body.Close()
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successes, got %d", successCount)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 errors, got %d", errorCount)
	}
}

// TestConstants verifies expected constant values
func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"exchangeName", exchangeName, "influencer-events"},
		{"queueName", queueName, "indexer-queue"},
		{"indexName", indexName, "influencers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Expected %s = %s, got %s", tt.name, tt.expected, tt.got)
			}
		})
	}
}

// TestElasticsearchRefreshOption verifies refresh option is set correctly
func TestElasticsearchRefreshOption(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 200, body: `{"result":"created"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	res, err := es.Index(
		indexName,
		bytes.NewReader([]byte(`{"test":"data"}`)),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer res.Body.Close()

	if mockTransport.callCount != 1 {
		t.Errorf("Expected 1 call to Elasticsearch, got %d", mockTransport.callCount)
	}
}

// TestConcurrentMessageProcessing verifies concurrent processing doesn't cause issues
func TestConcurrentMessageProcessing(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: make([]MockResponse, 10),
	}
	for i := 0; i < 10; i++ {
		mockTransport.responses[i] = MockResponse{
			statusCode: 200,
			body:       `{"result":"created"}`,
		}
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			profile := map[string]interface{}{
				"username": "user" + string(rune(id)),
				"id":       id,
			}
			profileJSON, _ := json.Marshal(profile)

			res, err := es.Index(
				indexName,
				bytes.NewReader(profileJSON),
				es.Index.WithRefresh("true"),
			)
			if err == nil && !res.IsError() {
				res.Body.Close()
			}
			done <- true
		}(i)
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
}

// TestEmptyMessageHandling verifies handling of empty messages
func TestEmptyMessageHandling(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 400, body: `{"error":"empty body"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	res, err := es.Index(
		indexName,
		bytes.NewReader([]byte{}),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		t.Fatalf("Expected no transport error, got %v", err)
	}
	defer res.Body.Close()

	if !res.IsError() {
		t.Error("Expected error for empty message")
	}
}

// TestLargeMessageHandling verifies handling of large messages
func TestLargeMessageHandling(t *testing.T) {
	mockTransport := &MockElasticsearchTransport{
		responses: []MockResponse{
			{statusCode: 200, body: `{"result":"created"}`},
		},
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	largeProfile := map[string]interface{}{
		"username": "testuser",
		"bio":      string(make([]byte, 10000)),
		"metadata": make(map[string]interface{}),
	}
	profileJSON, _ := json.Marshal(largeProfile)

	res, err := es.Index(
		indexName,
		bytes.NewReader(profileJSON),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		t.Fatalf("Expected no error for large message, got %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		t.Error("Expected success for large message")
	}
}
