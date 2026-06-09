// Command kafka-producer demonstrates Kafka producer and consumer with tenant awareness.
//
// It shows:
//   - Publishing messages with X-Tenant-Id header
//   - Consuming messages and extracting tenant ID for context
//   - Per-tenant message processing
//
// Run with a Kafka broker available.
package main

import (
	"context"
	"log/slog"
	"time"

	kafka "github.com/PapaDanielVi/apadana/pkg/kafka"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	// Producer: Publishes messages with tenant ID in header.
	// The Writer must be managed externally as producer.writer is unexported.
	writer := &kafkago.Writer{
		Addr:     kafkago.TCP("localhost:9092"),
		Topic:    "events",
		Balancer: &kafkago.Hash{},
	}
	defer writer.Close()

	producer := kafka.NewProducer(writer)

	// Example: Publishing with tenant context.
	tenantCtx := tctx.WithTenantID(ctx, "acme")
	if err := producer.Produce(tenantCtx, "events", []byte("key"), []byte("order created")); err != nil {
		slog.Error("publish failed", "error", err)
	}

	// Consumer: Reads messages and processes with tenant context.
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "events",
		GroupID: "processor",
	})
	consumer := kafka.NewConsumer(reader)

	// Start consuming in background.
	go func() {
		if err := consumer.Consume(ctx, func(msgCtx context.Context, msg kafkago.Message) {
			tenantID, _ := tctx.TenantIDFromContext(msgCtx)
			slog.Info("processing event", "tenant_id", tenantID, "topic", msg.Topic, "value", string(msg.Value))

			// Process message based on tenant.
			// Each message carries the tenant ID from the header.
		}); err != nil {
			slog.Error("consume error", "error", err)
		}
	}()

	slog.Info("kafka example started - publishing and consuming with tenant awareness")
	time.Sleep(2 * time.Second) // Let it run briefly.
}