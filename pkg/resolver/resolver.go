// Package resolver provides tenant resolution via chain-of-responsibility.
package resolver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"net"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
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
		if net.ParseIP(host) != nil {
			return "", errors.New("host is an ip address, cannot extract subdomain")
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

// Verifier validates a raw JWT and returns its claims. Implement it with a
// library such as github.com/golang-jwt/jwt or a JWKS client. Return a non-nil
// error to reject the token.
type Verifier func(token string) (map[string]any, error)

// FromVerifiedJWT resolves a tenant ID from a JWT bearer token after verifying
// it with verify. The token signature is checked before the claim is read, so
// this is safe to use on untrusted requests.
func FromVerifiedJWT(claimName string, verify Verifier) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		token, err := bearerToken(r)
		if err != nil {
			return "", err
		}
		claims, err := verify(token)
		if err != nil {
			return "", fmt.Errorf("verify jwt: %w", err)
		}
		return claimString(claims, claimName)
	})
}

// FromUnsafeJWTClaim resolves a tenant ID from a JWT bearer token WITHOUT
// verifying the signature. A forged token can set any tenant ID, so use this
// only when the token has already been validated by upstream middleware.
// Prefer FromVerifiedJWT on untrusted requests.
func FromUnsafeJWTClaim(claimName string) Resolver {
	return ResolverFunc(func(r *http.Request) (string, error) {
		token, err := bearerToken(r)
		if err != nil {
			return "", err
		}
		claim, err := extractJWTClaim(token, claimName)
		if err != nil {
			return "", errors.New("jwt claim " + claimName + ": " + err.Error())
		}
		return claim, nil
	})
}

// bearerToken extracts the raw token from an Authorization: Bearer header.
func bearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("no bearer token")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}

// MiddlewareConfig configures Chain.MiddlewareWithConfig.
type MiddlewareConfig struct {
	// Required rejects the request with 400 when no tenant is resolved,
	// instead of continuing without a tenant ID.
	Required bool
	// Logger logs resolution failures at debug level. If nil, no logging.
	Logger *slog.Logger
	// Validator, when set, normalizes and validates the resolved tenant ID.
	// A failed validation is treated like a resolution failure.
	Validator *tctx.Validator
}

// Middleware returns HTTP middleware that resolves tenant ID using the chain.
// Resolution failures are ignored and the request continues without a tenant ID.
func (c *Chain) Middleware() func(http.Handler) http.Handler {
	return c.MiddlewareWithConfig(MiddlewareConfig{})
}

// ResolveValidated resolves the tenant ID from r and, when v is non-nil,
// normalizes and validates it. It returns an empty string and an error when
// no valid tenant is found, so framework adapters can share one code path.
func (c *Chain) ResolveValidated(r *http.Request, v *tctx.Validator) (string, error) {
	tenantID, err := c.Resolve(r)
	if err != nil {
		return "", err
	}
	if v != nil {
		return v.Validate(tenantID)
	}
	return tenantID, nil
}

// MiddlewareWithConfig returns HTTP middleware configured by cfg.
func (c *Chain) MiddlewareWithConfig(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, err := c.ResolveValidated(r, cfg.Validator)
			if err != nil || tenantID == "" {
				if cfg.Logger != nil {
					cfg.Logger.DebugContext(r.Context(), "tenant resolution failed",
						slog.String("path", r.URL.Path), slog.Any("error", err))
				}
				if cfg.Required {
					http.Error(w, "missing tenant ID", http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(tctx.WithTenantID(r.Context(), tenantID))
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

	return claimString(claims, claimName)
}

// claimString reads a string claim from a decoded JWT claim set.
func claimString(claims map[string]any, claimName string) (string, error) {
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
	mu       sync.RWMutex
	tenants  map[string]Tenant
	loader   func(context.Context) (map[string]Tenant, error)
	ttl      time.Duration
	loadedAt time.Time
}

// NewRegistry creates a Registry that loads tenants via loader.
func NewRegistry(loader func(context.Context) (map[string]Tenant, error)) *Registry {
	return &Registry{
		tenants: make(map[string]Tenant),
		loader:  loader,
	}
}

// NewRegistryWithTTL creates a Registry whose cache is considered stale after
// ttl. GetOrLoad reloads automatically once the cache is stale.
func NewRegistryWithTTL(
	loader func(context.Context) (map[string]Tenant, error),
	ttl time.Duration,
) *Registry {
	r := NewRegistry(loader)
	r.ttl = ttl
	return r
}

// Load fetches tenants from the loader and atomically replaces the cache.
func (r *Registry) Load(ctx context.Context) error {
	tenants, err := r.loader(ctx)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.tenants = tenants
	r.loadedAt = time.Now()
	r.mu.Unlock()
	return nil
}

// stale reports whether the cache should be reloaded.
func (r *Registry) stale() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.loadedAt.IsZero() {
		return true
	}
	return r.ttl > 0 && time.Since(r.loadedAt) > r.ttl
}

// GetOrLoad returns the tenant by ID, reloading the cache first if it has
// never been loaded or has gone stale past the TTL.
func (r *Registry) GetOrLoad(ctx context.Context, tenantID string) (Tenant, bool, error) {
	if r.stale() {
		if err := r.Load(ctx); err != nil {
			return Tenant{}, false, err
		}
	}
	t, ok := r.Get(tenantID)
	return t, ok, nil
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
