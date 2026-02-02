package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	_ "strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"

	// Prometheus Libs
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	exchangeName = "influencer-events"
	queueName    = "indexer-queue"
	indexName    = "influencers"
)

// --- METRICS DEFINITIONS ---
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

	fmt.Println("🎧 Indexer listening for profiles...")

	// 6. The Processing Loop
	for {
		// A. Get the next message
		delivery, err := consumer.Next(ctx)
		if err != nil {
			log.Printf("Consumer error: %v", err)
			continue
		}

		// B. Index into Elasticsearch
		res, err := es.Index(
			indexName,
			bytes.NewReader(delivery.Message.Body),
			es.Index.WithRefresh("true"),
		)

		if err != nil {
			log.Printf("Elastic Error: %s", err)
			indexingErrors.Inc()
			continue
		}

		// Check for Elasticsearch application-level errors (e.g. 400 Bad Request)
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
			profilesIndexed.Inc() // <--- RECORD SUCCESS
			fmt.Print(".")
		}
	}
}
