package metrics

import (
	"context"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestCounter_Inc(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter")
	ctx := tctx.WithTenantID(context.Background(), "acme")
	counter.Inc(ctx)
	// If we get here, no panic
}

func TestHistogram_Observe(t *testing.T) {
	hist := NewHistogram("test_hist", "Test histogram", []float64{1, 2, 3})
	ctx := tctx.WithTenantID(context.Background(), "acme")
	hist.Observe(ctx, 1.5)
	// If we get here, no panic
}
