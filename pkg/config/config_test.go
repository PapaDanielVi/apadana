package config

import (
	"context"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestSetAndGet(t *testing.T) {
	tenantID := "test-tenant"
	key := "db_host"
	value := "localhost:5432"

	Set(tenantID, key, value)

	ctx := tctx.WithTenantID(context.Background(), tenantID)
	got, err := Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != value {
		t.Fatalf("Get() = %v, want %v", got, value)
	}
}

func TestGet_NoTenantID(t *testing.T) {
	ctx := context.Background()
	_, err := Get(ctx, "any-key")
	if err == nil {
		t.Fatal("expected error for missing tenant ID")
	}
}

func TestGet_NoConfigForTenant(t *testing.T) {
	ctx := tctx.WithTenantID(context.Background(), "nonexistent-tenant")
	_, err := Get(ctx, "any-key")
	if err == nil {
		t.Fatal("expected error for nonexistent tenant")
	}
}

func TestGet_KeyNotFound(t *testing.T) {
	tenantID := "test-tenant-2"
	Set(tenantID, "existing-key", "value")

	ctx := tctx.WithTenantID(context.Background(), tenantID)
	_, err := Get(ctx, "nonexistent-key")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}
