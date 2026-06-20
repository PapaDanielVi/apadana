// Package fiber provides Fiber middleware for tenant identification.
package fiber

import (
	"log/slog"

	fiberv2 "github.com/gofiber/fiber/v2"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
)

// Config configures the Tenant middleware. Fiber runs on fasthttp rather than
// net/http, so the tenant ID is read directly from the fiber context.
type Config struct {
	// Header is the request header to read. Defaults to tctx.HeaderKey.
	Header string
	// Query, when set, is checked after the header.
	Query string
	// Cookie, when set, is checked after the query.
	Cookie string
	// Required rejects the request with 400 when no tenant is resolved.
	Required bool
	// Validator, when set, normalizes and validates the resolved tenant ID.
	Validator *tctx.Validator
	// Logger logs resolution failures at debug level. If nil, no logging.
	Logger *slog.Logger
}

// extract reads the tenant ID from the configured sources in order.
func (cfg Config) extract(c *fiberv2.Ctx) string {
	header := cfg.Header
	if header == "" {
		header = tctx.HeaderKey
	}
	if v := c.Get(header); v != "" {
		return v
	}
	if cfg.Query != "" {
		if v := c.Query(cfg.Query); v != "" {
			return v
		}
	}
	if cfg.Cookie != "" {
		if v := c.Cookies(cfg.Cookie); v != "" {
			return v
		}
	}
	return ""
}

// Tenant returns Fiber middleware that resolves the tenant ID and stores it in
// the user context, retrievable with c.UserContext().
func Tenant(cfg Config) fiberv2.Handler {
	return func(c *fiberv2.Ctx) error {
		tenantID := cfg.extract(c)
		var err error
		if tenantID != "" && cfg.Validator != nil {
			tenantID, err = cfg.Validator.Validate(tenantID)
		}
		if tenantID == "" || err != nil {
			if cfg.Logger != nil {
				cfg.Logger.DebugContext(c.UserContext(), "tenant resolution failed",
					slog.String("path", c.Path()), slog.Any("error", err))
			}
			if cfg.Required {
				return c.SendStatus(fiberv2.StatusBadRequest)
			}
			return c.Next()
		}
		c.SetUserContext(tctx.WithTenantID(c.UserContext(), tenantID))
		return c.Next()
	}
}
