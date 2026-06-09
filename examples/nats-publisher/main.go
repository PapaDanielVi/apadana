// Command nats-publisher demonstrates NATS publisher and subscriber with tenant awareness.
//
// It shows:
//   - Publishing messages with X-Tenant-Id header
//   - Subscribing to messages and extracting tenant ID for context
//   - Per-tenant message processing
//
// Run with a NATS server available at nats://localhost:4222.
package main

import (
	"context"
	"log/slog"
	"time"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	"github.com/PapaDanielVi/apadana/pkg/nats"
	natsgo "github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS server.
	nc, err := natsgo.Connect("nats://localhost:4222")
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		return
	}
	defer nc.Close()

	ctx := context.Background()

	// Publisher: Publishes messages with tenant ID in header.
	publisher := nats.NewPublisher(nc)

	// Example: Publishing with tenant context.
	tenantCtx := tctx.WithTenantID(ctx, "acme")
	if err := publisher.Publish(tenantCtx, "events", []byte("order created")); err != nil {
		slog.Error("publish failed", "error", err)
	}

	// Subscriber: Listens and processes with tenant context.
	subscriber := nats.NewSubscriber(nc)

	// Start subscribing in background.
	go func() {
		if err := subscriber.Subscribe(ctx, "events", func(msgCtx context.Context, msg *natsgo.Msg) {
			tenantID, _ := tctx.TenantIDFromContext(msgCtx)
			slog.Info("processing event", "tenant_id", tenantID, "subject", msg.Subject, "data", string(msg.Data))

			// Process message based on tenant.
			// Each message carries the tenant ID from the header.
		}); err != nil {
			slog.Error("subscribe error", "error", err)
		}
	}()

	slog.Info("nats example started - publishing and subscribing with tenant awareness")
	time.Sleep(2 * time.Second) // Let it run briefly.
}
