package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// 1. Test Data Integrity
// Ensures that every generated profile has the required fields populated.
func TestGenerateSmartProfile_StructIntegrity(t *testing.T) {
	// Run multiple times to ensure randomness doesn't break things
	for i := 0; i < 50; i++ {
		profile := generateSmartProfile()

		if profile.ID == "" {
			t.Error("Expected ID to be set, got empty string")
		}
		if profile.Username == "" {
			t.Error("Expected Username to be set, got empty string")
		}
		if profile.Bio == "" {
			t.Error("Expected Bio to be set, got empty string")
		}
		if profile.Followers < 1000 || profile.Followers > 5000000 {
			t.Errorf("Followers %d out of expected range (1000-5M)", profile.Followers)
		}

		// Validate Platform is one of the allowed list
		validPlatform := false
		allowedPlatforms := []string{"Instagram", "TikTok", "YouTube"}
		for _, p := range allowedPlatforms {
			if profile.Platform == p {
				validPlatform = true
				break
			}
		}
		if !validPlatform {
			t.Errorf("Invalid platform generated: %s", profile.Platform)
		}
	}
}

// 2. Test Business Logic ("Smart" Generation)
// Ensures the Category matches the Bio keywords.
// A "Food" influencer MUST have food-related keywords.
func TestGenerateSmartProfile_CategoryConsistency(t *testing.T) {
	for i := 0; i < 50; i++ {
		profile := generateSmartProfile()

		// 1. Check if Category is valid
		validCategory := false
		for _, cat := range categories {
			if profile.Category == cat {
				validCategory = true
				break
			}
		}
		if !validCategory {
			t.Fatalf("Generated unknown category: %s", profile.Category)
		}

		// 2. Check strict correlation between Category and Bio
		// The bio format is: "%s | Loves %s | #%s"
		// We expect the bio to contain the Category name as a hashtag
		expectedHashtag := "#" + profile.Category
		if !strings.Contains(profile.Bio, expectedHashtag) {
			t.Errorf("Logic Fail: Category is '%s' but Bio '%s' is missing hashtag '%s'",
				profile.Category, profile.Bio, expectedHashtag)
		}

		// 3. Check Engagement Rate Logic
		// Code logic: gofakeit.Number(10, 80) / 10.0 -> Result should be 1.0 to 8.0
		if profile.EngagementRate < 1.0 || profile.EngagementRate > 8.0 {
			t.Errorf("Engagement Rate %.2f is out of realistic bounds (1.0 - 8.0)", profile.EngagementRate)
		}
	}
}

// 3. Test Observability (Prometheus)
// Verifies that the metrics server exposes the correct counter.
func TestMetricsEndpoint(t *testing.T) {
	// A. Reset or increment the counter to a known state for testing
	initialCount := testutil.ToFloat64(profilesDiscovered)
	profilesDiscovered.Inc()
	newCount := testutil.ToFloat64(profilesDiscovered)

	if newCount != initialCount+1 {
		t.Errorf("Metric counter failed to increment. Got %.1f, expected %.1f", newCount, initialCount+1)
	}

	// B. Simulate an HTTP request to /metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Use the standard Prometheus handler
	promhttp.Handler().ServeHTTP(w, req)

	// C. Assert the response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	// --- FIX STARTS HERE ---
	// Use io.ReadAll instead of ReadFrom
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	body := string(bodyBytes)
	// --- FIX ENDS HERE ---

	// D. Check if our specific metric name is present in the text output
	expectedMetric := "influencers_discovered_total"
	if !strings.Contains(body, expectedMetric) {
		t.Errorf("Metrics endpoint response missing expected metric: %s", expectedMetric)
	}
}

// 4. Test Metric Registration
// Ensures the Init function ran and registered the collector without panic
func TestMetricsRegistration(t *testing.T) {
	// Attempting to register the same metric again would panic if not handled,
	// but here we just verify the Gatherer contains it.
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "influencers_discovered_total" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Metric 'influencers_discovered_total' was not found in DefaultGatherer")
	}
}
