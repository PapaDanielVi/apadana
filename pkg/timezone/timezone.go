// Package timezone provides per-tenant timezone management.
package timezone

import (
	"context"
	"errors"
	"sync"
	"time"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

var tenantLocations sync.Map // tenantID string → *time.Location

// Set stores the timezone location for tenantID.
func Set(tenantID string, loc *time.Location) {
	tenantLocations.Store(tenantID, loc)
}

// Get returns the timezone for the tenant in ctx.
// Defaults to UTC if the tenant has no timezone set.
// Returns an error if no tenant ID is present in ctx.
func Get(ctx context.Context) (*time.Location, error) {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("no tenant ID in context")
	}

	v, ok := tenantLocations.Load(tenantID)
	if !ok {
		return time.UTC, nil
	}

	loc, ok := v.(*time.Location)
	if !ok {
		return time.UTC, nil
	}

	return loc, nil
}

// Now returns the current time in the tenant's timezone.
// Defaults to UTC if tenant has no timezone set or no tenant ID in ctx.
func Now(ctx context.Context) time.Time {
	loc, err := Get(ctx)
	if err != nil {
		return time.Now().UTC()
	}
	return time.Now().In(loc)
}
