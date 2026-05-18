// Package obj provides lazy-initialized, centralized per-tenant object management.
package obj

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Factory creates a new instance of an object.
type Factory func() (any, error)

var (
	factories   = make(map[string]Factory)
	factoriesMu sync.RWMutex
	tenantObjs  sync.Map // tenantID string → *sync.Map (name → any)
)

// Register associates a Factory with name. Must be called before Get for that name.
func Register(name string, factory Factory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	factories[name] = factory
}

// Get returns the single instance for name under the tenant in ctx.
// Instances are lazy-initialized and shared across all calls for the same tenant+name.
func Get(ctx context.Context, name string) (any, error) {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("no tenant ID in context")
	}

	// Get or create per-tenant instance map
	v, _ := tenantObjs.LoadOrStore(tenantID, &sync.Map{})
	perTenant := v.(*sync.Map)

	// Check if instance already exists
	inst, ok := perTenant.Load(name)
	if ok {
		return inst, nil
	}

	// Get factory
	factoriesMu.RLock()
	factory, ok := factories[name]
	factoriesMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no factory registered for %q", name)
	}

	// Create instance
	obj, err := factory()
	if err != nil {
		return nil, fmt.Errorf("factory for %q failed: %w", name, err)
	}

	// Store, handling race with another goroutine
	actual, loaded := perTenant.LoadOrStore(name, obj)
	if loaded {
		return actual, nil
	}
	return obj, nil
}
