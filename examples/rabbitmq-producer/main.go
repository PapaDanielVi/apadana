// Command rabbitmq-producer demonstrates RabbitMQ publisher and consumer with tenant awareness.
//
// It shows:
//   - Publishing messages with X-Tenant-Id header
//   - Consuming messages and extracting tenant ID for context
//   - Per-tenant message processing
//
// Run with a RabbitMQ broker available at amqp://localhost:5672.
package main

import (
	"context"
	"log/slog"
	"time"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	"github.com/PapaDanielVi/apadana/v2/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Connect to RabbitMQ.
	conn, err := amqp.Dial("amqp://localhost:5672")
	if err != nil {
		slog.Error("failed to connect to rabbitmq", "error", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to open channel", "error", err)
		return
	}
	defer ch.Close()

	ctx := context.Background()

	// Publisher: Publishes messages with tenant ID in header.
	publisher := rabbitmq.NewPublisher(ch)

	// Example: Publishing with tenant context.
	tenantCtx := tctx.WithTenantID(ctx, "acme")
	if err := publisher.Publish(tenantCtx, "", "events", []byte("order created")); err != nil {
		slog.Error("publish failed", "error", err)
	}

	// Consumer: Reads messages and processes with tenant context.
	consumer := rabbitmq.NewConsumer(ch)

	// Start consuming in background.
	go func() {
		if err := consumer.Consume(ctx, "events", func(msgCtx context.Context, d amqp.Delivery) {
			tenantID, _ := tctx.TenantIDFromContext(msgCtx)
			slog.Info("processing event", "tenant_id", tenantID, "routingKey", d.RoutingKey, "body", string(d.Body))

			// Process message based on tenant.
			// Each message carries the tenant ID from the header.
		}); err != nil {
			slog.Error("consume error", "error", err)
		}
	}()

	slog.Info("rabbitmq example started - publishing and consuming with tenant awareness")
	time.Sleep(2 * time.Second) // Let it run briefly.
}
