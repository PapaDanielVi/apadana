package mt_test

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

func TestCloneCtx_PreservesTenantID(t *testing.T) {
	t.Parallel()

	original := mt.InjectTID(context.Background(), "acme")
	cloned := mt.CloneCtx(original)

	got := mt.ExtractTID(cloned)
	if got != "acme" {
		t.Errorf("CloneCtx() tenant = %q, want %q", got, "acme")
	}
}

func TestCloneCtx_PreservesDefaultTenant(t *testing.T) {
	t.Parallel()

	original := context.Background()
	cloned := mt.CloneCtx(original)

	got := mt.ExtractTID(cloned)
	if got != "default" {
		t.Errorf("CloneCtx() with empty context tenant = %q, want %q", got, "default")
	}
}

func TestCloneCtx_PreservesBaggage(t *testing.T) {
	t.Parallel()

	member, err := baggage.NewMember("key", "value")
	if err != nil {
		t.Fatalf("baggage.NewMember failed: %v", err)
	}
	bag, err := baggage.New(member)
	if err != nil {
		t.Fatalf("baggage.New failed: %v", err)
	}
	original := mt.InjectTID(context.Background(), "acme")
	original = baggage.ContextWithBaggage(original, bag)

	cloned := mt.CloneCtx(original)
	gotBag := baggage.FromContext(cloned)
	if gotBag.Len() != 1 {
		t.Errorf("CloneCtx() baggage len = %d, want 1", gotBag.Len())
	}
	if gotBag.Member("key").Value() != "value" {
		t.Errorf("CloneCtx() baggage key = %q, want %q", gotBag.Member("key").Value(), "value")
	}
}

func TestCloneCtx_PreservesSpanContext(t *testing.T) {
	t.Parallel()

	traceID := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := [8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}
	fakeSC := trace.NewSpanContext(sc)

	original := mt.InjectTID(context.Background(), "acme")
	original = trace.ContextWithSpanContext(original, fakeSC)

	cloned := mt.CloneCtx(original)
	gotSC := trace.SpanContextFromContext(cloned)
	if !gotSC.Equal(fakeSC) {
		t.Errorf("CloneCtx() span context not preserved")
	}
}

func TestCloneCtx_CancellationNotInherited(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		mt.InjectTID(context.Background(), "acme"),
	)
	cancel()

	cloned := mt.CloneCtx(ctx)
	if err := cloned.Err(); err != nil {
		t.Errorf("CloneCtx() cloned context should not be cancelled, got %v", err)
	}
}

func TestCloneCtxWithDeadline_PreservesDeadline(t *testing.T) {
	t.Parallel()

	want := time.Now().Add(time.Hour)
	parent, cancel := context.WithDeadline(mt.InjectTID(context.Background(), "acme"), want)
	defer cancel()

	cloned, cancelClone := mt.CloneCtxWithDeadline(parent)
	defer cancelClone()

	got, ok := cloned.Deadline()
	if !ok {
		t.Fatal("CloneCtxWithDeadline() should carry the parent deadline")
	}
	if !got.Equal(want) {
		t.Errorf("deadline = %v, want %v", got, want)
	}
	if tid := mt.ExtractTID(cloned); tid != "acme" {
		t.Errorf("tenant = %q, want %q", tid, "acme")
	}
}

func TestCloneCtxWithDeadline_NoParentDeadline(t *testing.T) {
	t.Parallel()

	cloned, cancel := mt.CloneCtxWithDeadline(mt.InjectTID(context.Background(), "acme"))
	defer cancel()

	if _, ok := cloned.Deadline(); ok {
		t.Error("CloneCtxWithDeadline() should have no deadline when parent has none")
	}
}
