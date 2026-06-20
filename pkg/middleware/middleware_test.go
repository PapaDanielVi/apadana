package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
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

func TestFromSubdomain_IPAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host    string
		wantErr bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"10.0.0.1:8080", true},
		// {"::1", true},
		// {"2001:db8::1", true},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/", nil)
		extractor := FromSubdomain()
		_, err := extractor(r)

		if tt.wantErr && err == nil {
			t.Errorf("FromSubdomain() with host %s should return error", tt.host)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("FromSubdomain() with host %s returned error: %v", tt.host, err)
		}
	}
}

func TestRequireTenant(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing tenant rejected", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		RequireTenant(next).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("present tenant passes", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(tctx.WithTenantID(req.Context(), "acme"))
		rec := httptest.NewRecorder()
		RequireTenant(next).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
