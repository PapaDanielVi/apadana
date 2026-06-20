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

// CloneCtxWithDeadline behaves like CloneCtx but also carries over the parent
// deadline if it has one. The returned cancel function must be called to
// release resources; it is a no-op when the parent has no deadline.
func CloneCtxWithDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
	newCtx := CloneCtx(ctx)
	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(newCtx, deadline)
	}
	return newCtx, func() {}
}
