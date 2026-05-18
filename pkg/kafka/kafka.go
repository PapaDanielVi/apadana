// Package kafka provides multi-tenant Kafka tools.
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Producer publishes messages with tenant ID header.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Producer using the given writer.
func NewProducer(writer *kafka.Writer) *Producer {
	return &Producer{writer: writer}
}

// Produce sends a message with X-Tenant-ID header from ctx.
func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "X-Tenant-ID", Value: []byte(tenantID)},
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
func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, kafka.Message)) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var tenantID string
		for _, h := range msg.Headers {
			if h.Key == "X-Tenant-ID" {
				tenantID = string(h.Value)
				break
			}
		}

		ctx := tctx.WithTenantID(ctx, tenantID)
		handler(ctx, msg)

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
