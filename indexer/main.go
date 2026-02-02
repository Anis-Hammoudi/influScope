package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	_ "strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"
)

const (
	exchangeName = "influencer-events"
	queueName    = "indexer-queue"
	indexName    = "influencers"
)

func main() {
	// 1. Connect to Elasticsearch
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
			fmt.Println(" Connected to Elasticsearch!")
			res.Body.Close()
			break
		}
		fmt.Println("Waiting for Elasticsearch...")
		time.Sleep(3 * time.Second)
	}

	// 2. Connect to RabbitMQ (Upfluence Style)
	fmt.Println(" Connecting to RabbitMQ...")
	broker := amqputil.Open()
	defer broker.Close()
	ctx := context.Background()

	// We declare a Queue and Bind it to the Exchange.
	// This ensures that even if Indexer is down, RabbitMQ holds the messages.
	_ = broker.DeclareQueue(ctx, queueName, amqp.DeclareQueueOptions{Durable: true})
	_ = broker.BindQueue(ctx, queueName, "#", exchangeName, amqp.BindQueueOptions{})

	// 4. Start Consuming
	consumer, err := broker.Consume(ctx, queueName, amqp.ConsumeOptions{
		Consumer: "indexer-worker-1",
		AutoACK:  false,
	})
	if err != nil {
		log.Fatalf("Could not start consumer: %v", err)
	}
	defer consumer.Close()

	fmt.Println("🎧 Indexer listening for profiles...")

	// 5. The Processing Loop
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
			log.Printf(" Elastic Error: %s", err)
			// Don't ACK if DB fails! RabbitMQ will retry later.
			continue
		}
		res.Body.Close()

		// C. Acknowledge the message (Tell RabbitMQ "Done!")
		if err := consumer.Ack(ctx, delivery.DeliveryTag, amqp.AckOptions{}); err != nil {
			log.Printf("Failed to ACK: %v", err)
		} else {
			// Print a subtle dot for every success to show activity
			fmt.Print(".")
		}
	}
}
