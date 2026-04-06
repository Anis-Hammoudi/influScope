package repository

import (
	"context"
	"fmt"

	"github.com/hammo/influScope/indexer/internal/domain"
	"github.com/upfluence/amqp"
	"github.com/upfluence/amqp/amqputil"
)

type upfluenceConsumer struct {
	broker   amqp.Broker
	consumer amqp.Consumer
	ctx      context.Context
}

type upfluenceMessage struct {
	ctx      context.Context
	consumer amqp.Consumer
	delivery *amqp.Delivery
}

func (m *upfluenceMessage) Body() []byte { return m.delivery.Message.Body }
func (m *upfluenceMessage) Ack() error {
	return m.consumer.Ack(m.ctx, m.delivery.DeliveryTag, amqp.AckOptions{})
}

func NewRabbitMQConsumer(ctx context.Context, exchange, queue string) (domain.MessageConsumer, error) {
	broker := amqputil.Open()

	if err := broker.DeclareQueue(ctx, queue, amqp.DeclareQueueOptions{Durable: true}); err != nil {
		return nil, fmt.Errorf("declare queue failed: %w", err)
	}
	if err := broker.BindQueue(ctx, queue, "#", exchange, amqp.BindQueueOptions{}); err != nil {
		return nil, fmt.Errorf("bind queue failed: %w", err)
	}

	consumer, err := broker.Consume(ctx, queue, amqp.ConsumeOptions{
		Consumer: "indexer-worker-1",
		AutoACK:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("consume failed: %w", err)
	}

	return &upfluenceConsumer{
		broker:   broker,
		consumer: consumer,
		ctx:      ctx,
	}, nil
}

func (c *upfluenceConsumer) Next(ctx context.Context) (domain.Message, error) {
	delivery, err := c.consumer.Next(ctx)
	if err != nil {
		return nil, err
	}
	return &upfluenceMessage{
		ctx:      ctx,
		consumer: c.consumer,
		delivery: delivery,
	}, nil
}

func (c *upfluenceConsumer) Close() error {
	c.consumer.Close()
	return c.broker.Close()
}
