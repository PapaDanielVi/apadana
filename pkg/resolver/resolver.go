// Package resolver provides tenant resolution via chain-of-responsibility.
package resolver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"strings"
	"sync"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Tenant holds metadata for a tenant.
type Tenant struct {
	ID       string         `json:"id"                 yaml:"id"`
	Name     string         `json:"name"               yaml:"name"`
	Plan     string         `json:"plan"               yaml:"plan"`
	Metadata map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Resolver extracts a tenant ID from an HTTP request.
type Resolver interface {
	Resolve(r *http.Request) (string, error)
}

// ResolverFunc is an adapter to use a function as a Resolver.
type ResolverFunc func(r *http.Request) (string, error)

// Resolve calls f(r).
func (f ResolverFunc) Resolve(r *http.Request) (string, error) {
	return f(r)
}

// Chain tries each resolver in order, returning the first successful result.
type Chain struct {
	resolvers []Resolver
}

// NewChain creates a resolver chain.
func NewChain(resolvers ...Resolver) *Chain {
	return &Chain{resolvers: resolvers}
}

// Resolve tries each resolver in order.
func (c *Chain) Resolve(r *http.Request) (string, error) {
	for _, res := range c.resolvers {
		id, err := res.Resolve(r)
		if err == nil && id != "" {
			return id, nil
		}
	}
	return "", ErrNoTenant
}

// ErrNoTenant is returned when no resolver in the chain finds a tenant ID.
var ErrNoTenant = errors.New("no tenant ID found in request")

// FromHeader resolves tenant ID from an HTTP header.
func FromHeader(headerName string) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		v := r.Header.Get(headerName)
		if v == "" {
			return "", errors.New("header " + headerName + " not found")
		}
		return v, nil
	})
}

// FromQuery resolves tenant ID from a URL query parameter.
func FromQuery(paramName string) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		v := r.URL.Query().Get(paramName)
		if v == "" {
			return "", errors.New("query param " + paramName + " not found")
		}
		return v, nil
	})
}

// FromSubdomain resolves tenant ID from the request hostname subdomain.
func FromSubdomain() Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		host := r.Host
		if i := strings.Index(host, ":"); i != -1 {
			host = host[:i]
		}
		parts := strings.Split(host, ".")
		if len(parts) < 2 {
			return "", errors.New("no subdomain found")
		}
		return parts[0], nil
	})
}

// FromCookie resolves tenant ID from a cookie.
func FromCookie(cookieName string) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return "", errors.New("cookie " + cookieName + " not found")
		}
		return cookie.Value, nil
	})
}

// FromJWTClaim resolves tenant ID from a JWT Authorization header.
// It extracts the claim without validating the signature — use only when
// the token has already been validated by upstream middleware.
func FromJWTClaim(claimName string) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return "", errors.New("no bearer token")
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claim, err := extractJWTClaim(token, claimName)
		if err != nil {
			return "", errors.New("jwt claim " + claimName + ": " + err.Error())
		}
		return claim, nil
	})
}

// Middleware returns HTTP middleware that resolves tenant ID using the chain.
func (c *Chain) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, err := c.Resolve(r)
			if err == nil && tenantID != "" {
				r = r.WithContext(tctx.WithTenantID(r.Context(), tenantID))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractJWTClaim extracts a claim from an unvalidated JWT token.
// This is a convenience for when the token is already validated upstream.
func extractJWTClaim(token, claimName string) (string, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return "", errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return "", err
	}

	var claims map[string]any
	if err = json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}

	v, ok := claims[claimName]
	if !ok {
		return "", errors.New("claim not found")
	}

	s, ok := v.(string)
	if !ok {
		return "", errors.New("claim is not a string")
	}
	return s, nil
}

// Registry loads and caches tenant metadata.
type Registry struct {
	mu      sync.RWMutex
	tenants map[string]Tenant
	loader  func(context.Context) (map[string]Tenant, error)
}

// NewRegistry creates a Registry that loads tenants via loader.
func NewRegistry(loader func(context.Context) (map[string]Tenant, error)) *Registry {
	return &Registry{
		tenants: make(map[string]Tenant),
		loader:  loader,
	}
}

// Load fetches tenants from the loader and caches them.
func (r *Registry) Load(ctx context.Context) error {
	tenants, err := r.loader(ctx)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.tenants = tenants
	r.mu.Unlock()
	return nil
}

// Get returns the tenant by ID.
func (r *Registry) Get(tenantID string) (Tenant, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tenants[tenantID]
	return t, ok
}

// All returns all registered tenants.
func (r *Registry) All() map[string]Tenant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]Tenant, len(r.tenants))
	maps.Copy(out, r.tenants)
	return out
}
