package tctx

import (
	"context"
	"testing"
)

func TestWithTenantIDAndFromContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tenantID := "acme-corp"

	ctx = WithTenantID(ctx, tenantID)

	got, ok := TenantIDFromContext(ctx)
	if !ok {
		t.Fatal("expected tenant ID to be present")
	}
	if got != tenantID {
		t.Fatalf("got %q, want %q", got, tenantID)
	}
}

func TestTenantIDFromContext_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, ok := TenantIDFromContext(ctx)
	if ok {
		t.Fatal("expected ok=false for context without tenant ID")
	}
}

func TestHasTenantID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	if HasTenantID(ctx) {
		t.Fatal("expected HasTenantID=false for empty context")
	}

	ctx = WithTenantID(ctx, "tenant-1")
	if !HasTenantID(ctx) {
		t.Fatal("expected HasTenantID=true after setting tenant ID")
	}
}
