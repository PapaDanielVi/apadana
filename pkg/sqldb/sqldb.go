// Package sqldb provides a per-tenant database/sql connection manager.
package sqldb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

// Manager returns the *sql.DB for the tenant in context. Connections are opened
// lazily on first use and shared per tenant. It is safe for concurrent use.
type Manager struct {
	sdk mt.ISDKE[*sql.DB]
}

// NewManager creates a Manager. driver is a registered database/sql driver
// (the driver package must be imported for its side effects by the caller).
// dsns maps tenant ID to data source name. Connections are not opened until
// the first Get for a tenant, so this never fails.
func NewManager(driver string, dsns map[string]string) *Manager {
	initFn := func(_ context.Context, dsn string) (*sql.DB, error) {
		return sql.Open(driver, dsn)
	}
	// Plain string configs are lazy, so no connection is opened here and the
	// returned error is always nil.
	sdk, _ := mt.NewSDKMgrE(dsns, initFn)
	return &Manager{sdk: sdk}
}

// Get returns the *sql.DB for the tenant in ctx, opening it on first use.
func (m *Manager) Get(ctx context.Context) (*sql.DB, error) {
	return m.sdk.Get(ctx)
}

// Close closes every opened connection and joins any errors.
func (m *Manager) Close() error {
	var errs []error
	m.sdk.Map(func(_ string, db *sql.DB) {
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}
	})
	return errors.Join(errs...)
}
