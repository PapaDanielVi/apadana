// Package kafka provides multi-tenant Kafka tools.
package kafka

import (
	"context"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	"github.com/segmentio/kafka-go"
)

// Producer publishes messages with tenant ID header.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Producer using the given writer.
func NewProducer(writer *kafka.Writer) *Producer {
	return &Producer{writer: writer}
}

// Produce sends a message with X-Tenant-Id header from ctx.
func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "X-Tenant-Id", Value: []byte(tenantID)},
		},
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Consumer consumes messages and injects tenant ID into context.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a Consumer using the given reader.
func NewConsumer(reader *kafka.Reader) *Consumer {
	return &Consumer{reader: reader}
}

// Consume reads messages and calls handler with tenant-aware context.
// It returns when the context is cancelled or a fatal error occurs.
func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, kafka.Message)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var tenantID string
		for _, h := range msg.Headers {
			if h.Key == "X-Tenant-Id" {
				tenantID = string(h.Value)
				break
			}
		}

		msgCtx := tctx.WithTenantID(ctx, tenantID)
		handler(msgCtx, msg)

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
