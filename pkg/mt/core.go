// Package mt provides generic multi-tenant tools.
package mt

import (
	"context"
	"sync"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

var (
	defaultTID = "default"
	setDefOnce sync.Once
)

const fallbackTID = "default"

// SetDefTenant sets the default tenant ID. Must be called once.
func SetDefTenant(tenantID string) {
	setDefOnce.Do(func() {
		defaultTID = tenantID
	})
}

// ExtractTID returns tenant ID from context, or defaultTID if not found.
func ExtractTID(ctx context.Context) string {
	tid, ok := tctx.TenantIDFromContext(ctx)
	if !ok || tid == "" {
		return defaultTID
	}
	return tid
}

// InjectTID stores tenant ID in context, using defaultTID if empty.
func InjectTID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		tenantID = defaultTID
	}
	return tctx.WithTenantID(ctx, tenantID)
}

// tenantIDGetter is satisfied by objects that can return a tenant ID.
type tenantIDGetter interface {
	TenantID() string
}

// InjectTenantFromObj extracts tenant ID from v if it implements tenantIDGetter,
// then stores it in a new context returned.
func InjectTenantFromObj(ctx context.Context, v any) context.Context {
	var tid string
	if tg, ok := v.(tenantIDGetter); ok {
		tid = tg.TenantID()
	}
	return InjectTID(ctx, tid)
}
