package otel

import (
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewTenantIDProcessor(t *testing.T) {
	p := NewTenantIDProcessor()
	if p == nil {
		t.Fatal("expected non-nil processor")
	}
	// Verify it implements SpanProcessor
	_ = sdktrace.SpanProcessor(p)
}
