package mt

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// CloneCtx creates a new context from parent with tenant ID and OpenTelemetry values preserved.
// Cancellation is not inherited; metadata is copied.
func CloneCtx(ctx context.Context) context.Context {
	newCtx := context.Background()
	tid := ExtractTID(ctx)
	newCtx = InjectTID(newCtx, tid)

	if bag := baggage.FromContext(ctx); bag.Len() > 0 {
		newCtx = baggage.ContextWithBaggage(newCtx, bag)
	}

	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		newCtx = trace.ContextWithSpanContext(newCtx, span.SpanContext())
	}
	return newCtx
}
