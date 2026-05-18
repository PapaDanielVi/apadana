package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestFromHeader(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Tenant-Id", "acme")

	extractor := FromHeader("X-Tenant-Id")
	tenantID, err := extractor(r)
	if err != nil {
		t.Fatalf("extractor error = %v", err)
	}
	if tenantID != "acme" {
		t.Fatalf("got %q, want %q", tenantID, "acme")
	}
}

func TestFromQuery(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "/?tenant=acme", nil)

	extractor := FromQuery("tenant")
	tenantID, err := extractor(r)
	if err != nil {
		t.Fatalf("extractor error = %v", err)
	}
	if tenantID != "acme" {
		t.Fatalf("got %q, want %q", tenantID, "acme")
	}
}

func TestFromSubdomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host string
		want string
	}{
		{"acme.example.com", "acme"},
		{"tenant1.localhost:8080", "tenant1"},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/", nil)
		extractor := FromSubdomain()
		tenantID, err := extractor(r)
		if err != nil {
			t.Fatalf("extractor error = %v", err)
		}
		if tenantID != tt.want {
			t.Fatalf("got %q, want %q", tenantID, tt.want)
		}
	}
}

func TestTenantMiddleware_SetsTenantID(t *testing.T) {
	t.Parallel()
	var gotTenantID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := tctx.TenantIDFromContext(r.Context())
		if ok {
			gotTenantID = tenantID
		}
	})

	middleware := TenantMiddleware(FromHeader("X-Tenant-Id"))(handler)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Tenant-Id", "acme")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, r)

	if gotTenantID != "acme" {
		t.Fatalf("got %q, want %q", gotTenantID, "acme")
	}
}

func TestTenantMiddleware_NoTenantID(t *testing.T) {
	t.Parallel()
	var hasTenantID bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasTenantID = tctx.HasTenantID(r.Context())
	})

	middleware := TenantMiddleware(FromHeader("X-Tenant-Id"))(handler)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, r)

	if hasTenantID {
		t.Fatal("expected no tenant ID in context")
	}
}
