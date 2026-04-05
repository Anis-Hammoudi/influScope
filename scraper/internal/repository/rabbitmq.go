package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hammo/influScope/pkg/models"
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"
)

type RabbitMQPublisher struct {
	broker       amqp.Broker
	exchangeName string
}

func NewRabbitMQPublisher(exchangeName string, maxRetries int) (*RabbitMQPublisher, error) {
	var broker amqp.Broker
	var err error

	for i := 1; i <= maxRetries; i++ {
		log.Printf("Connecting to RabbitMQ (Attempt %d/%d)...", i, maxRetries)
		broker = amqputil.Open()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = broker.DeclareExchange(ctx, exchangeName, amqp.Fanout, amqp.DeclareExchangeOptions{Durable: true})
		cancel()

		if err == nil {
			log.Println("Connected to RabbitMQ successfully!")
			return &RabbitMQPublisher{broker: broker, exchangeName: exchangeName}, nil
		}

		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to RabbitMQ after %d attempts: %w", maxRetries, err)
}

func (r *RabbitMQPublisher) PublishProfile(ctx context.Context, profile models.Influencer) error {
	body, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	return r.broker.Publish(
		ctx,
		r.exchangeName,
		"",
		amqp.Message{
			Body:        body,
			ContentType: "application/json",
			Headers: map[string]interface{}{
				"service": "scraper-v1",
				"version": "1.0.0",
			},
		},
		amqp.PublishOptions{},
	)
}

func (r *RabbitMQPublisher) Close() error {
	return r.broker.Close()
}
