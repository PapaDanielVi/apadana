// Package tctx provides context utilities for multi-tenant applications.
package tctx

import "context"

// TenantID is a string alias for type-safe tenant identifiers.
type TenantID string

const (
	// HeaderKey is the canonical HTTP header carrying the tenant ID.
	// It is used for HTTP requests, message-queue headers, and the
	// httpclient. gRPC lowercases metadata keys, so use MetadataKey there.
	HeaderKey = "X-Tenant-Id"

	// MetadataKey is the gRPC metadata key carrying the tenant ID.
	// gRPC metadata keys are case-insensitive and stored lowercase.
	MetadataKey = "x-tenant-id"
)

type contextKey struct{}

// WithTenantID returns a copy of ctx with tenantID stored.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, contextKey{}, TenantID(tenantID))
}

// TenantIDFromContext retrieves the tenant ID from ctx.
// The second return value is false if no tenant ID is set.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKey{}).(TenantID)
	return string(v), ok
}

// HasTenantID returns true if ctx carries a tenant ID.
func HasTenantID(ctx context.Context) bool {
	_, ok := ctx.Value(contextKey{}).(TenantID)
	return ok
}
