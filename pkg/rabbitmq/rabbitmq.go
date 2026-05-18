// Package rabbitmq provides multi-tenant RabbitMQ tools.
package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Publisher publishes messages with tenant ID header.
type Publisher struct {
	ch *amqp091.Channel
}

// NewPublisher creates a Publisher using the given channel.
func NewPublisher(ch *amqp091.Channel) *Publisher {
	return &Publisher{ch: ch}
}

// Publish sends a message with X-Tenant-Id header from ctx.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	headers := amqp091.Table{"X-Tenant-Id": tenantID}
	return p.ch.Publish(
		exchange, routingKey, false, false,
		amqp091.Publishing{
			Headers: headers,
			Body:    body,
		},
	)
}

// Consumer consumes messages and injects tenant ID into context.
type Consumer struct {
	ch *amqp091.Channel
}

// NewConsumer creates a Consumer using the given channel.
func NewConsumer(ch *amqp091.Channel) *Consumer {
	return &Consumer{ch: ch}
}

// Consume starts consuming and calls handler with tenant-aware context.
func (c *Consumer) Consume(ctx context.Context, queue string, handler func(context.Context, amqp091.Delivery)) error {
	deliveries, err := c.ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			tenantID, _ := d.Headers["X-Tenant-Id"].(string)
			ctx := tctx.WithTenantID(ctx, tenantID)
			handler(ctx, d)
		}
	}()
	return nil
}
