package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsInitialization(t *testing.T) {
	pm := NewPrometheusMetrics()

	if testutil.ToFloat64(pm.profilesIndexed) != 0 {
		t.Errorf("Expected profilesIndexed to start at 0, got %f", testutil.ToFloat64(pm.profilesIndexed))
	}
	if testutil.ToFloat64(pm.indexingErrors) != 0 {
		t.Errorf("Expected indexingErrors to start at 0, got %f", testutil.ToFloat64(pm.indexingErrors))
	}

	// Test Incrementing
	pm.IncIndexed()
	pm.IncError()

	if testutil.ToFloat64(pm.profilesIndexed) != 1 {
		t.Errorf("Expected profilesIndexed to be 1, got %f", testutil.ToFloat64(pm.profilesIndexed))
	}
}

func TestMetricsServerEndpoint(t *testing.T) {
	// Start server on a test port
	pm := NewPrometheusMetrics()
	go pm.StartServer(":8083") // Use a different port for testing

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Note: In a real scenario, you'd use the default mux where promhttp is registered
	http.DefaultServeMux.ServeHTTP(w, req)

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
