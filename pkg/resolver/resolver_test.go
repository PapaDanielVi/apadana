package resolver_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PapaDanielVi/apadana/v2/pkg/resolver"
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

func TestFromUnsafeJWTClaim(t *testing.T) {
	t.Parallel()

	r := resolver.FromUnsafeJWTClaim("tid")
	token := buildJWT(map[string]any{"tid": "acme", "sub": "user1"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	id, err := r.Resolve(req)
	if err != nil {
		t.Fatalf("FromUnsafeJWTClaim() error = %v", err)
	}
	if id != "acme" {
		t.Errorf("FromUnsafeJWTClaim() = %q, want %q", id, "acme")
	}
}

func TestFromVerifiedJWT(t *testing.T) {
	t.Parallel()

	// verify accepts the token only when it equals the expected value,
	// standing in for a real signature check.
	const goodToken = "good-token"
	verify := func(token string) (map[string]any, error) {
		if token != goodToken {
			return nil, errors.New("bad signature")
		}
		return map[string]any{"tid": "acme"}, nil
	}

	r := resolver.FromVerifiedJWT("tid", verify)

	t.Run("valid token resolves", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+goodToken)
		id, err := r.Resolve(req)
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if id != "acme" {
			t.Errorf("Resolve() = %q, want %q", id, "acme")
		}
	})

	t.Run("tampered token rejected", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tampered")
		if _, err := r.Resolve(req); err == nil {
			t.Error("Resolve() should reject a token that fails verification")
		}
	})
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

func TestRegistry_GetOrLoad(t *testing.T) {
	t.Parallel()

	var loads atomic.Int32
	loader := func(_ context.Context) (map[string]resolver.Tenant, error) {
		loads.Add(1)
		return map[string]resolver.Tenant{"acme": {ID: "acme"}}, nil
	}

	reg := resolver.NewRegistry(loader)

	// First call loads because the cache has never been populated.
	tenant, ok, err := reg.GetOrLoad(context.Background(), "acme")
	if err != nil {
		t.Fatalf("GetOrLoad() error = %v", err)
	}
	if !ok || tenant.ID != "acme" {
		t.Fatalf("GetOrLoad() = %+v, %v; want acme, true", tenant, ok)
	}

	// Without a TTL the cache never goes stale, so no second load.
	if _, _, err = reg.GetOrLoad(context.Background(), "acme"); err != nil {
		t.Fatalf("GetOrLoad() error = %v", err)
	}
	if got := loads.Load(); got != 1 {
		t.Errorf("loader ran %d times, want 1", got)
	}
}

func TestRegistry_GetOrLoad_TTLReload(t *testing.T) {
	t.Parallel()

	var loads atomic.Int32
	loader := func(_ context.Context) (map[string]resolver.Tenant, error) {
		loads.Add(1)
		return map[string]resolver.Tenant{"acme": {ID: "acme"}}, nil
	}

	reg := resolver.NewRegistryWithTTL(loader, time.Millisecond)

	if _, _, err := reg.GetOrLoad(context.Background(), "acme"); err != nil {
		t.Fatalf("GetOrLoad() error = %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, _, err := reg.GetOrLoad(context.Background(), "acme"); err != nil {
		t.Fatalf("GetOrLoad() error = %v", err)
	}
	if got := loads.Load(); got != 2 {
		t.Errorf("loader ran %d times after TTL expiry, want 2", got)
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
		// {"::1", true}, // IPv6
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

func TestChain_MiddlewareWithConfig_Required(t *testing.T) {
	t.Parallel()

	chain := resolver.NewChain(resolver.FromHeader("X-Tenant-Id"))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"missing tenant rejected", "", http.StatusBadRequest},
		{"present tenant passes", "acme", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := chain.MiddlewareWithConfig(resolver.MiddlewareConfig{Required: true})(next)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("X-Tenant-Id", tt.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
