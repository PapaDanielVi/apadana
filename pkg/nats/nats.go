// Package nats provides multi-tenant NATS tools.
package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Publisher publishes messages with tenant ID header.
type Publisher struct {
	nc *nats.Conn
}

// NewPublisher creates a Publisher using the given connection.
func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

// Publish sends a message with X-Tenant-ID header from ctx.
func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{"X-Tenant-ID": []string{tenantID}},
	}
	return p.nc.PublishMsg(msg)
}

// Subscriber subscribes to messages and injects tenant ID into context.
type Subscriber struct {
	nc *nats.Conn
}

// NewSubscriber creates a Subscriber using the given connection.
func NewSubscriber(nc *nats.Conn) *Subscriber {
	return &Subscriber{nc: nc}
}

// Subscribe listens on subject and calls handler with tenant-aware context.
func (s *Subscriber) Subscribe(ctx context.Context, subject string, handler func(context.Context, *nats.Msg)) error {
	_, err := s.nc.Subscribe(subject, func(msg *nats.Msg) {
		tenantID := msg.Header.Get("X-Tenant-ID")
		ctx := tctx.WithTenantID(ctx, tenantID)
		handler(ctx, msg)
	})
	return err
}
