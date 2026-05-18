package db

import (
	"context"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestColumnMiddleware(t *testing.T) {
	reg := New(ColumnModel, "mysql", "user:pass@/dbname")
	mw := reg.ColumnMiddleware()

	ctx := tctx.WithTenantID(context.Background(), "acme")
	query := "SELECT * FROM orders"
	result := mw(ctx, query)

	expected := "SELECT * FROM orders WHERE tenant_id = 'acme'"
	if result != expected {
		t.Fatalf("got %q, want %q", result, expected)
	}
}

func TestDatabaseName(t *testing.T) {
	reg := New(DatabaseModel, "mysql", "db_%s")
	name := reg.DatabaseName("acme")
	if name != "db_acme" {
		t.Fatalf("got %q, want %q", name, "db_acme")
	}
}

func TestConnection_InstanceModel(t *testing.T) {
	reg := New(InstanceModel, "mysql", "user:pass@/dbname")
	ctx := tctx.WithTenantID(context.Background(), "acme")

	db, err := reg.Connection(ctx)
	if err != nil {
		t.Skip("skipping: no MySQL server available")
	}
	if db == nil {
		t.Fatal("expected non-nil DB")
	}
}
