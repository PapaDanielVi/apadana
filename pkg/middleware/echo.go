package middleware

import (
	"context"
	"net/http"

	echoprom "github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	"github.com/PapaDanielVi/apadana/pkg/mt"
)

const (
	PromTenantLabel = "tenant_id"
)

// InjectTenantHTTP sets tenant ID header from ctx on req.
func InjectTenantHTTP(ctx context.Context, req *http.Request) {
	tid, _ := tctx.TenantIDFromContext(ctx)
	req.Header.Set("X-Tenant-Id", tid)
}

// DelTenantHTTP removes tenant ID header from req.
func DelTenantHTTP(req *http.Request) {
	req.Header.Del("X-Tenant-Id")
}

// TenantEchoMiddleware returns Echo middleware that extracts tenant ID from X-Tenant-Id header.
// If setDefault is true, missing header uses default tenant ID.
func TenantEchoMiddleware(setDefault bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if mt.ExtractTID(c.Request().Context()) != mt.ExtractTID(context.Background()) {
				return next(c)
			}
			tid := c.Request().Header.Get("X-Tenant-Id")
			if tid == "" {
				if !setDefault {
					return c.JSON(462, "Invalid Request")
				}
				tid = mt.ExtractTID(context.Background())
			}
			ctx := mt.InjectTID(c.Request().Context(), tid)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// PrometheusEchoMiddleware sets up prometheus metrics with tenant_id label.
// metricsPath is typically "/metrics", subsystem is your service name.
func PrometheusEchoMiddleware(e *echo.Echo, subsystem, metricsPath string, reg prometheus.Registerer) {
	labelFuncs := map[string]echoprom.LabelValueFunc{
		PromTenantLabel: func(c echo.Context, _ error) string {
			return mt.ExtractTID(c.Request().Context())
		},
	}
	prom := echoprom.NewMiddlewareWithConfig(echoprom.MiddlewareConfig{
		Subsystem:  subsystem,
		LabelFuncs: labelFuncs,
		Registerer: reg,
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/healthz", "/health", "/metrics":
				return true
			}
			return false
		},
	})
	e.Use(prom)
	e.GET(metricsPath, echoprom.NewHandler())
}
