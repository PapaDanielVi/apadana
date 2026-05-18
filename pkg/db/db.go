// Package db provides multi-tenant database registry.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Model represents the multi-tenancy model.
type Model int

const (
	ColumnModel   Model = iota // tenant_id column in tables
	DatabaseModel              // separate DB per tenant
	InstanceModel              // separate connection per tenant
)

// Registry manages multi-tenant DB access.
type Registry struct {
	model  Model
	driver string
	dsn    string
	dbs    map[string]*sql.DB
}

// New creates a new Registry.
func New(model Model, driver, dsn string) *Registry {
	return &Registry{
		model:  model,
		driver: driver,
		dsn:    dsn,
		dbs:    make(map[string]*sql.DB),
	}
}

// ColumnMiddleware returns a query modifier that injects tenant_id.
// For ColumnModel: appends WHERE tenant_id = 'id' to queries.
func (r *Registry) ColumnMiddleware() func(context.Context, string) string {
	if r.model != ColumnModel {
		return func(_ context.Context, q string) string { return q }
	}
	return func(ctx context.Context, query string) string {
		tenantID, ok := tctx.TenantIDFromContext(ctx)
		if !ok {
			return query
		}
		return fmt.Sprintf("%s WHERE tenant_id = '%s'", query, tenantID)
	}
}

// DatabaseName returns the DB name for tenantID (DatabaseModel).
func (r *Registry) DatabaseName(tenantID string) string {
	// Assumes DSN has placeholder {tenant_id}
	return fmt.Sprintf(r.dsn, tenantID)
}

// Connection returns a DB handle for the tenant in ctx (InstanceModel).
func (r *Registry) Connection(ctx context.Context) (*sql.DB, error) {
	if r.model != InstanceModel {
		return nil, errors.New("Connection only valid for InstanceModel")
	}
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("no tenant ID in context")
	}

	db, ok := r.dbs[tenantID]
	if ok {
		return db, nil
	}

	db, err := sql.Open(r.driver, r.dsn)
	if err != nil {
		return nil, err
	}

	r.dbs[tenantID] = db
	return db, nil
}
