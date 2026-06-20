package resolver_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PapaDanielVi/apadana/pkg/resolver"
)

func buildJWT(claims map[string]any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString(mustMarshal(claims))
	return header + "." + payload + "."
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestFromHeader(t *testing.T) {
	t.Parallel()

	r := resolver.FromHeader("X-Tenant-Id")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromHeader() error = %v", err)
	}
	if id != "acme" {
		t.Errorf("FromHeader() = %q, want %q", id, "acme")
	}
}

func TestFromHeader_Missing(t *testing.T) {
	t.Parallel()

	r := resolver.FromHeader("X-Tenant-Id")
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := r.Resolve(req)
	if err == nil {
		t.Error("FromHeader() should error on missing header")
	}
}

func TestFromQuery(t *testing.T) {
	t.Parallel()

	r := resolver.FromQuery("tenant")
	req := httptest.NewRequest(http.MethodGet, "/?tenant=globex", nil)

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromQuery() error = %v", err)
	}
	if id != "globex" {
		t.Errorf("FromQuery() = %q, want %q", id, "globex")
	}
}

func TestFromSubdomain(t *testing.T) {
	t.Parallel()

	r := resolver.FromSubdomain()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "acme.example.com"

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromSubdomain() error = %v", err)
	}
	if id != "acme" {
		t.Errorf("FromSubdomain() = %q, want %q", id, "acme")
	}
}

func TestFromCookie(t *testing.T) {
	t.Parallel()

	r := resolver.FromCookie("tenant_id")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "tenant_id", Value: "acme"})

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromCookie() error = %v", err)
	}
	if id != "acme" {
		t.Errorf("FromCookie() = %q, want %q", id, "acme")
	}
}

func TestFromJWTClaim(t *testing.T) {
	t.Parallel()

	r := resolver.FromJWTClaim("tid")
	token := buildJWT(map[string]any{"tid": "acme", "sub": "user1"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromJWTClaim() error = %v", err)
	}
	if id != "acme" {
		t.Errorf("FromJWTClaim() = %q, want %q", id, "acme")
	}
}

func TestChain(t *testing.T) {
	t.Parallel()

	chain := resolver.NewChain(
		resolver.FromHeader("X-Tenant-Id"),
		resolver.FromQuery("tenant"),
		resolver.FromSubdomain(),
	)

	tests := []struct {
		name     string
		req      *http.Request
		expected string
		wantErr  bool
	}{
		{
			name: "header takes priority",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("X-Tenant-Id", "from-header")
				return r
			}(),
			expected: "from-header",
		},
		{
			name:     "query when no header",
			req:      httptest.NewRequest(http.MethodGet, "/?tenant=from-query", nil),
			expected: "from-query",
		},
		{
			name: "subdomain when no header or query",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Host = "from-sub.example.com"
				return r
			}(),
			expected: "from-sub",
		},
		{
			name:    "error when nothing matches",
			req:     func() *http.Request { r := httptest.NewRequest(http.MethodGet, "/", nil); r.Host = ""; return r }(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := chain.Resolve(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("chain.Resolve() = %q, want %q", id, tt.expected)
			}
		})
	}
}

func TestChain_Middleware(t *testing.T) {
	t.Parallel()

	chain := resolver.NewChain(resolver.FromHeader("X-Tenant-Id"))
	mw := chain.Middleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mw(handler))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("X-Tenant-Id", "acme")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRegistry(t *testing.T) {
	t.Parallel()

	loader := func(_ context.Context) (map[string]resolver.Tenant, error) {
		return map[string]resolver.Tenant{
			"acme":   {ID: "acme", Name: "Acme Corp", Plan: "premium"},
			"globex": {ID: "globex", Name: "Globex Inc", Plan: "free"},
		}, nil
	}

	reg := resolver.NewRegistry(loader)
	if err := reg.Load(context.Background()); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tenant, ok := reg.Get("acme")
	if !ok {
		t.Fatal("Get(acme) not found")
	}
	if tenant.Name != "Acme Corp" {
		t.Errorf("Name = %q, want %q", tenant.Name, "Acme Corp")
	}
	if tenant.Plan != "premium" {
		t.Errorf("Plan = %q, want %q", tenant.Plan, "premium")
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should not be found")
	}

	all := reg.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d tenants, want 2", len(all))
	}
}

func TestRegistry_Get_Defaults(t *testing.T) {
	t.Parallel()

	loader := func(_ context.Context) (map[string]resolver.Tenant, error) {
		return map[string]resolver.Tenant{}, nil
	}

	reg := resolver.NewRegistry(loader)
	if err := reg.Load(context.Background()); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, ok := reg.Get("missing")
	if ok {
		t.Error("Get(missing) should return false for empty registry")
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
        {"::1", true}, // IPv6
    }

    for _, tt := range tests {
        req := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/", nil)
        resolver := resolver.FromSubdomain()
        _, err := resolver.Resolve(req)

        if tt.wantErr && err == nil {
            t.Errorf("FromSubdomain() with host %s should return error", tt.host)
        }
    }
}