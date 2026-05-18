package burst

import (
	"context"
	"testing"
	"time"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestAllow(t *testing.T) {
	c := New(1, 1) // rate 1/sec, burst 1
	ctx := tctx.WithTenantID(context.Background(), "tenant1")

	if !c.Allow(ctx) {
		t.Fatal("expected first request to be allowed")
	}
	if c.Allow(ctx) {
		t.Fatal("expected second request to be denied")
	}

	// Wait for token refill
	time.Sleep(1 * time.Second)
	if !c.Allow(ctx) {
		t.Fatal("expected request to be allowed after refill")
	}
}

func TestAllow_DifferentTenants(t *testing.T) {
	c := New(1, 1)
	ctx1 := tctx.WithTenantID(context.Background(), "tenantA")
	ctx2 := tctx.WithTenantID(context.Background(), "tenantB")

	// TenantA uses its token
	c.Allow(ctx1)

	// TenantB should still have its token
	if !c.Allow(ctx2) {
		t.Fatal("expected tenantB to have separate bucket")
	}
}

func TestAllow_NoTenantID(t *testing.T) {
	c := New(1, 1)
	ctx := context.Background()
	if c.Allow(ctx) {
		t.Fatal("expected deny for missing tenant ID")
	}
}
