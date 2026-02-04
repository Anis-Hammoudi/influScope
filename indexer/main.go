package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/hammo/influScope/pkg/models"
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"
	// Prometheus Libs
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/hammo/influScope/gen/analytics"
)

const (
	exchangeName = "influencer-events"
	queueName    = "indexer-queue"
	indexName    = "influencers"
)

var (
	profilesIndexed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "influencers_indexed_total",
			Help: "Total number of profiles successfully saved to Elasticsearch",
		},
	)
	indexingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "indexer_errors_total",
			Help: "Total number of failed indexing attempts",
		},
	)
)

func init() {
	// Register metrics so Prometheus can scrape them
	prometheus.MustRegister(profilesIndexed)
	prometheus.MustRegister(indexingErrors)
}

func main() {
	// 1. START METRICS SERVER (Background)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println(" Metrics server listening on :8082")
		http.ListenAndServe(":8082", nil)
	}()

	// 2. Connect to Elasticsearch
	esCfg := elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	conn, err := grpc.Dial("analytics:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to gRPC: %v", err)
	}
	defer conn.Close()

	// Create the client stub
	analyticsClient := pb.NewAnalyticsServiceClient(conn)
	log.Println("Connected to Analytics gRPC Service")
	for i := 0; i < 10; i++ {
		res, err := es.Info()
		if err == nil && res.StatusCode == 200 {
			fmt.Println("Connected to Elasticsearch!")
			res.Body.Close()
			break
		}
		fmt.Println(" Waiting for Elasticsearch...")
		time.Sleep(3 * time.Second)
	}

	// 3. Connect to RabbitMQ
	fmt.Println(" Connecting to RabbitMQ...")
	broker := amqputil.Open()
	defer broker.Close()
	ctx := context.Background()

	// 4. Setup Queue
	_ = broker.DeclareQueue(ctx, queueName, amqp.DeclareQueueOptions{Durable: true})
	_ = broker.BindQueue(ctx, queueName, "#", exchangeName, amqp.BindQueueOptions{})

	// 5. Start Consuming
	consumer, err := broker.Consume(ctx, queueName, amqp.ConsumeOptions{
		Consumer: "indexer-worker-1",
		AutoACK:  false,
	})
	if err != nil {
		log.Fatalf("Could not start consumer: %v", err)
	}
	defer consumer.Close()

	fmt.Println(" Indexer listening for profiles...")

	// 6. The Processing Loop
	for {
		// A. Get the next message
		delivery, err := consumer.Next(ctx)
		if err != nil {
			log.Printf("Consumer error: %v", err)
			continue
		}
		var influencer models.Influencer
		if err := json.Unmarshal(delivery.Message.Body, &influencer); err != nil {
			log.Printf("cdJSON Error: %v", err)
			// Important: Ack the bad message so we don't process it forever
			consumer.Ack(ctx, delivery.DeliveryTag, amqp.AckOptions{})
			continue
		}
		grpcCtx, cancel := context.WithTimeout(context.Background(), time.Second)

		// Call the remote function
		resp, err := analyticsClient.CalculateEngagement(grpcCtx, &pb.EngagementRequest{
			Username:  influencer.Username,
			Followers: int64(influencer.Followers),
			Platform:  influencer.Platform,
		})
		cancel()

		if err != nil {
			log.Printf(" Analytics Service failed: %v", err)
			influencer.EngagementRate = 0.0
		} else {
			influencer.EngagementRate = resp.EngagementRate
		}

		// B. Re-Serialize for Elasticsearch
		enrichedBody, _ := json.Marshal(influencer)

		// C. Index to Elasticsearch
		res, err := es.Index(
			indexName,
			bytes.NewReader(enrichedBody),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			log.Printf("Elastic Error: %s", err)
			indexingErrors.Inc()
			continue
		}

		if res.IsError() {
			log.Printf(" Indexing Failed: %s", res.String())
			indexingErrors.Inc()
			res.Body.Close()
			continue
		}
		res.Body.Close()

		// C. Acknowledge the message
		if err := consumer.Ack(ctx, delivery.DeliveryTag, amqp.AckOptions{}); err != nil {
			log.Printf("Failed to ACK: %v", err)
		} else {
			profilesIndexed.Inc()
			fmt.Print(".")
		}
	}
}
