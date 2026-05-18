package obj

import (
	"context"
	"errors"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestGet_ReturnsSameInstance(t *testing.T) {
	tenantID := "tenant-obj-1"
	Register("db", func() (any, error) {
		return "db-conn", nil
	})

	ctx := tctx.WithTenantID(context.Background(), tenantID)

	inst1, err := Get(ctx, "db")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	inst2, err := Get(ctx, "db")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if inst1 != inst2 {
		t.Fatal("expected same instance for same tenant+name")
	}
}

func TestGet_DifferentTenantsGetDifferentInstances(t *testing.T) {
	Register("cache", func() (any, error) {
		return new(int), nil
	})

	ctx1 := tctx.WithTenantID(context.Background(), "tenant-A")
	ctx2 := tctx.WithTenantID(context.Background(), "tenant-B")

	inst1, err := Get(ctx1, "cache")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	inst2, err := Get(ctx2, "cache")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if inst1 == inst2 {
		t.Fatal("expected different instances for different tenants")
	}
}

func TestGet_UnregisteredName(t *testing.T) {
	ctx := tctx.WithTenantID(context.Background(), "any-tenant")
	_, err := Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistered name")
	}
}

func TestGet_FactoryError(t *testing.T) {
	Register("bad", func() (any, error) {
		return nil, errors.New("init failed")
	})

	ctx := tctx.WithTenantID(context.Background(), "tenant-err")
	_, err := Get(ctx, "bad")
	if err == nil {
		t.Fatal("expected error from factory")
	}
}

func TestGet_NoTenantID(t *testing.T) {
	ctx := context.Background()
	_, err := Get(ctx, "any-name")
	if err == nil {
		t.Fatal("expected error for missing tenant ID")
	}
}
