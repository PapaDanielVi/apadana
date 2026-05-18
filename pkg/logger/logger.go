// Package logger provides a slog wrapper that includes tenant ID in log output.
package logger

import (
	"context"
	"log/slog"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// New returns a slog.Logger that includes tenant_id from ctx if present.
func New(ctx context.Context) *slog.Logger {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return slog.Default()
	}
	return slog.Default().With("tenant_id", tenantID)
}
