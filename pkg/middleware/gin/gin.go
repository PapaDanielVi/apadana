// Package gin provides Gin middleware for tenant identification.
package gin

import (
	"log/slog"
	"net/http"

	gingonic "github.com/gin-gonic/gin"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	"github.com/PapaDanielVi/apadana/v2/pkg/resolver"
)

// Config configures the Tenant middleware.
type Config struct {
	// Chain resolves the tenant ID from the request. Required.
	Chain *resolver.Chain
	// Required aborts with 400 when no tenant is resolved.
	Required bool
	// Validator, when set, normalizes and validates the resolved tenant ID.
	Validator *tctx.Validator
	// Logger logs resolution failures at debug level. If nil, no logging.
	Logger *slog.Logger
}

// Tenant returns Gin middleware that resolves the tenant ID and stores it in
// the request context.
func Tenant(cfg Config) gingonic.HandlerFunc {
	return func(c *gingonic.Context) {
		tenantID, err := cfg.Chain.ResolveValidated(c.Request, cfg.Validator)
		if err != nil || tenantID == "" {
			if cfg.Logger != nil {
				cfg.Logger.DebugContext(c.Request.Context(), "tenant resolution failed",
					slog.String("path", c.Request.URL.Path), slog.Any("error", err))
			}
			if cfg.Required {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Next()
			return
		}
		ctx := tctx.WithTenantID(c.Request.Context(), tenantID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
