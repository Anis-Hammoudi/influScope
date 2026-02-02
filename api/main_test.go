package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

// --- MOCK TRANSPORT (Reused logic) ---
type MockElasticsearchTransport struct {
	ResponseStatusCode int
	ResponseBody       string
}

func (m *MockElasticsearchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Standard Headers required by the client
	header := make(http.Header)
	header.Set("X-Elastic-Product", "Elasticsearch")
	header.Set("Content-Type", "application/json")

	// 1. Handle Handshake (GET /)
	if req.Method == "GET" && req.URL.Path == "/" {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"version":{"number":"7.17.0"}}`)),
			Header:     header,
		}, nil
	}

	// 2. Return our custom mock response for search queries
	return &http.Response{
		StatusCode: m.ResponseStatusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.ResponseBody)),
		Header:     header,
	}, nil
}

// --- HELPER TO CREATE MOCK CLIENT ---
func getMockClient(statusCode int, body string) *elasticsearch.Client {
	mockTrans := &MockElasticsearchTransport{
		ResponseStatusCode: statusCode,
		ResponseBody:       body,
	}
	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTrans,
	})
	return client
}

// --- TEST CASES ---

func TestSearchEndpoint_MissingQuery(t *testing.T) {
	// Setup Gin to Test Mode (quieter logs)
	gin.SetMode(gin.TestMode)

	// Create a client (doesn't matter what response, we won't hit it)
	client := getMockClient(200, "{}")
	router := setupRouter(client)

	// Perform Request without "q" param
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != 400 {
		t.Errorf("Expected status 400 for missing query, got %d", w.Code)
	}
}

func TestSearchEndpoint_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock valid Elasticsearch Response
	// This mimics the structure: hits -> hits -> _source -> {influencer}
	mockESResponse := `{
		"hits": {
			"hits": [
				{
					"_source": {
						"username": "test_guru",
						"category": "Tech",
						"followers": 5000
					}
				}
			]
		}
	}`

	client := getMockClient(200, mockESResponse)
	router := setupRouter(client)

	// Perform Request WITH "q" param
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?q=tech", nil)
	router.ServeHTTP(w, req)

	// Assert Status
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Assert Body contains our data
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Check "count"
	if count, ok := response["count"].(float64); !ok || count != 1 {
		t.Errorf("Expected count 1, got %v", response["count"])
	}

	// Check data content
	data := response["data"].([]interface{})
	firstUser := data[0].(map[string]interface{})
	if firstUser["username"] != "test_guru" {
		t.Errorf("Expected username 'test_guru', got %v", firstUser["username"])
	}
}

func TestSearchEndpoint_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock empty response from ES
	mockESResponse := `{
		"hits": {
			"hits": []
		}
	}`

	client := getMockClient(200, mockESResponse)
	router := setupRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?q=unknown", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["count"].(float64) != 0 {
		t.Errorf("Expected count 0, got %v", response["count"])
	}
}

func TestSearchEndpoint_ElasticError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock a 500 Internal Server Error from ES
	client := getMockClient(500, `{"error": "something exploded"}`)
	router := setupRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?q=crash", nil)
	router.ServeHTTP(w, req)

	// We expect the API to handle the ES failure gracefully (return 500)
	if w.Code != 500 {
		t.Errorf("Expected status 500 when ES fails, got %d", w.Code)
	}
}
