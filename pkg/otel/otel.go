// Package otel provides OpenTelemetry instrumentation for multi-tenant apps.
package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
)

// TenantIDProcessor is a span processor that adds tenant_id to spans.
type TenantIDProcessor struct{}

// OnStart adds tenant_id attribute if present in context.
func (p TenantIDProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return
	}
	s.SetAttributes(attribute.String("tenant_id", tenantID))
}

// OnEnd is a no-op.
func (p TenantIDProcessor) OnEnd(s sdktrace.ReadOnlySpan) {}

// Shutdown is a no-op.
func (p TenantIDProcessor) Shutdown(ctx context.Context) error { return nil }

// ForceFlush is a no-op.
func (p TenantIDProcessor) ForceFlush(ctx context.Context) error { return nil }

// NewTenantIDProcessor returns a span processor that injects tenant ID.
func NewTenantIDProcessor() sdktrace.SpanProcessor {
	return TenantIDProcessor{}
}
