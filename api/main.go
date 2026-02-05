package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
	"github.com/hammo/influScope/pkg/models"
)

// setupRouter allows us to pass in the dependency (ES Client) for testing
func setupRouter(es *elasticsearch.Client) *gin.Engine {
	r := gin.Default()

	// Define the Search Endpoint
	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		// Build the Elastic Query
		var buf bytes.Buffer
		queryJSON := map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":     query,
					"fields":    []string{"bio", "category", "username"},
					"fuzziness": "AUTO",
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(queryJSON); err != nil {
			c.JSON(500, gin.H{"error": "Failed to build query"})
			return
		}

		// Execute Search
		res, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex("influencers"),
			es.Search.WithBody(&buf),
			es.Search.WithTrackTotalHits(true),
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "Elasticsearch failed"})
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			c.JSON(500, gin.H{"error": "Elasticsearch returned an error"})
			return
		}

		// Parse Results
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			c.JSON(500, gin.H{"error": "Error parsing response"})
			return
		}

		// Transform Elastic response into clean JSON
		var influencers []models.Influencer

		// Safe extraction to avoid panics if "hits" is missing
		if hitsMap, ok := r["hits"].(map[string]interface{}); ok {
			if hitsList, ok := hitsMap["hits"].([]interface{}); ok {
				for _, hit := range hitsList {
					source := hit.(map[string]interface{})["_source"]
					
					tmp, err := json.Marshal(source)
					if err != nil {
						log.Printf("Error marshalling source: %v", err)
						continue
					}
					
					var inf models.Influencer
					// Linter Fix: Check Unmarshal error
					if err := json.Unmarshal(tmp, &inf); err != nil {
						log.Printf("Error unmarshalling to struct: %v", err)
						continue
					}
					
					influencers = append(influencers, inf)
				}
			}
		}

		// Ensure we return empty array [] instead of null
		if influencers == nil {
			influencers = []models.Influencer{}
		}

		c.JSON(200, gin.H{
			"count": len(influencers),
			"data":  influencers,
		})
	})

	return r
}

func main() {
	// 1. Connect to Elastic
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// 2. Setup Web Server
	r := setupRouter(es)

	// 3. Start Server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}