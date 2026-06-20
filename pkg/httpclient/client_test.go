package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
)

func TestDo_InjectsTenantHeader(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Tenant-Id")
	}))
	defer server.Close()

	client := New()
	ctx := tctx.WithTenantID(context.Background(), "acme")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	client.Do(ctx, req)

	if gotHeader != "acme" {
		t.Fatalf("header = %q, want %q", gotHeader, "acme")
	}
}

func TestDo_NoTenantID(t *testing.T) {
	var hasHeader bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasHeader = r.Header["X-Tenant-Id"]
	}))
	defer server.Close()

	client := New()
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	client.Do(ctx, req)

	if hasHeader {
		t.Fatal("expected no X-Tenant-Id header")
	}
}
