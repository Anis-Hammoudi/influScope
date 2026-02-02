package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
	"github.com/hammo/influScope/pkg/models"
	"log"
	_ "strings"
)

func main() {
	// 1. Connect to Elastic
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// 2. Setup Web Server (Gin)
	r := gin.Default()

	// 3. Define the Search Endpoint
	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		// 4. Build the Elastic Query
		var buf bytes.Buffer //
		queryJSON := map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":     query,
					"fields":    []string{"bio", "category", "username"},
					"fuzziness": "AUTO", // Handles typos!
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(queryJSON); err != nil {
			c.JSON(500, gin.H{"error": "Failed to build query"})
			return
		}

		// 5. Execute Search
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

		// 6. Parse Results
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			c.JSON(500, gin.H{"error": "Error parsing response"})
			return
		}

		// Transform Elastic response into clean JSON for the user
		var influencers []models.Influencer
		hits := r["hits"].(map[string]interface{})["hits"].([]interface{})

		for _, hit := range hits {
			source := hit.(map[string]interface{})["_source"]
			// Quick Hack: Marshall/Unmarshal to map to struct
			tmp, _ := json.Marshal(source)
			var inf models.Influencer
			json.Unmarshal(tmp, &inf)
			influencers = append(influencers, inf)
		}

		c.JSON(200, gin.H{
			"count": len(influencers),
			"data":  influencers,
		})
	})

	r.Run(":8080")
}
