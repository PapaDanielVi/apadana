package timezone

import (
	"context"
	"testing"
	"time"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestSetAndGet(t *testing.T) {
	tenantID := "test-tenant-tz"
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation error: %v", err)
	}

	Set(tenantID, loc)

	ctx := tctx.WithTenantID(context.Background(), tenantID)
	got, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != loc {
		t.Fatalf("Get() = %v, want %v", got, loc)
	}
}

func TestGet_DefaultsToUTC(t *testing.T) {
	tenantID := "test-tenant-no-tz"
	ctx := tctx.WithTenantID(context.Background(), tenantID)

	got, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != time.UTC {
		t.Fatalf("Get() = %v, want UTC", got)
	}
}

func TestGet_NoTenantID(t *testing.T) {
	ctx := context.Background()
	_, err := Get(ctx)
	if err == nil {
		t.Fatal("expected error for missing tenant ID")
	}
}

func TestNow_TenantTimezone(t *testing.T) {
	tenantID := "test-tenant-now"
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("LoadLocation error: %v", err)
	}

	Set(tenantID, loc)

	ctx := tctx.WithTenantID(context.Background(), tenantID)
	now := Now(ctx)

	if now.Location() != loc {
		t.Fatalf("Now() location = %v, want %v", now.Location(), loc)
	}
}

func TestNow_DefaultUTC(t *testing.T) {
	tenantID := "test-tenant-now-utc"
	ctx := tctx.WithTenantID(context.Background(), tenantID)

	now := Now(ctx)
	if now.Location() != time.UTC {
		t.Fatalf("Now() location = %v, want UTC", now.Location())
	}
}
