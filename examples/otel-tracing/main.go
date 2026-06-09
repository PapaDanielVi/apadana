// Command otel-tracing demonstrates OpenTelemetry integration with tenant awareness.
//
// It shows:
//   - TenantIDProcessor adds tenant_id attribute to spans
//   - Integration with OTel tracer provider
//   - Automatic tenant context propagation in traces
//
// Run with OpenTelemetry collector available.
package main

import (
	"context"
	"log/slog"

	"github.com/PapaDanielVi/apadana/pkg/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	slog.Info("otel-tracing example")

	// Create a tracer provider with tenant ID processor.
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(otel.NewTenantIDProcessor()),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("tracer shutdown failed", "error", err)
		}
	}()

	slog.Info("tracer provider configured with TenantIDProcessor")
	slog.Info("spans created from tenant context will include tenant_id attribute")
}

// Note: In production, you would typically:
// 1. Configure the tracer provider with your exporter
// 2. Use the tracer in your handlers
// 3. Tenant ID flows automatically from context to spans
//
// Example:
//
//	tracer := tp.Tracer("myapp")
//	ctx := tctx.WithTenantID(context.Background(), "acme")
//	ctx, span := tracer.Start(ctx, "handle-request")
//	defer span.End()
//	// The span now has tenant_id="acme" attribute