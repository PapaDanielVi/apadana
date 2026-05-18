// Package config provides thread-safe per-tenant configuration storage.
package config

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// tenantConfigs maps tenantID → per-tenant config map.
var tenantConfigs sync.Map // string → *sync.Map

// Set stores value for key under the given tenantID.
func Set(tenantID, key string, value interface{}) {
	v, _ := tenantConfigs.LoadOrStore(tenantID, &sync.Map{})
	perTenant, ok := v.(*sync.Map)
	if !ok {
		return
	}
	perTenant.Store(key, value)
}

// Get retrieves the config value for key from the tenant in ctx.
// Returns an error if tenant ID is missing or key is not found.
func Get(ctx context.Context, key string) (interface{}, error) {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("no tenant ID in context")
	}

	v, ok := tenantConfigs.Load(tenantID)
	if !ok {
		return nil, fmt.Errorf("no config for tenant %q", tenantID)
	}

	perTenant, ok := v.(*sync.Map)
	if !ok {
		return nil, fmt.Errorf("invalid config type for tenant %q", tenantID)
	}
	val, ok := perTenant.Load(key)
	if !ok {
		return nil, fmt.Errorf("key %q not found for tenant %q", key, tenantID)
	}

	return val, nil
}
