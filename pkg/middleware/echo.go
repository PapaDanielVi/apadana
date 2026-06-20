package middleware

import (
	"context"
	"net/http"

	echoprom "github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

const (
	PromTenantLabel = "tenant_id"
)

// InjectTenantHTTP sets tenant ID header from ctx on req.
func InjectTenantHTTP(ctx context.Context, req *http.Request) {
	tid, _ := tctx.TenantIDFromContext(ctx)
	req.Header.Set(tctx.HeaderKey, tid)
}

// DelTenantHTTP removes tenant ID header from req.
func DelTenantHTTP(req *http.Request) {
	req.Header.Del(tctx.HeaderKey)
}

// TenantEchoConfig configures TenantEchoMiddlewareWithConfig.
type TenantEchoConfig struct {
	// SetDefault uses the default tenant ID when the header is missing
	// instead of rejecting the request.
	SetDefault bool
	// MissingStatus is the HTTP status returned when the tenant header is
	// missing and SetDefault is false. Defaults to http.StatusBadRequest.
	MissingStatus int
}

// TenantEchoMiddleware returns Echo middleware that extracts the tenant ID from
// the tenant header. If setDefault is true, a missing header uses the default
// tenant ID; otherwise the request is rejected with 400.
func TenantEchoMiddleware(setDefault bool) echo.MiddlewareFunc {
	return TenantEchoMiddlewareWithConfig(TenantEchoConfig{SetDefault: setDefault})
}

// TenantEchoMiddlewareWithConfig returns Echo middleware configured by cfg.
func TenantEchoMiddlewareWithConfig(cfg TenantEchoConfig) echo.MiddlewareFunc {
	status := cfg.MissingStatus
	if status == 0 {
		status = http.StatusBadRequest
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if mt.ExtractTID(c.Request().Context()) != mt.ExtractTID(context.Background()) {
				return next(c)
			}
			tid := c.Request().Header.Get(tctx.HeaderKey)
			if tid == "" {
				if !cfg.SetDefault {
					return c.JSON(status, "missing tenant ID")
				}
				tid = mt.ExtractTID(context.Background())
			}
			ctx := mt.InjectTID(c.Request().Context(), tid)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// UnknownTenantLabel is the bucket used for tenant IDs outside the allowlist,
// bounding the cardinality of the tenant_id metric label.
const UnknownTenantLabel = "unknown"

// PrometheusConfig configures PrometheusEchoMiddlewareWithConfig.
type PrometheusConfig struct {
	// Subsystem is the metric subsystem, typically the service name.
	Subsystem string
	// MetricsPath is the path the /metrics handler is registered on.
	MetricsPath string
	// Registerer is the Prometheus registerer.
	Registerer prometheus.Registerer
	// AllowTenants, when non-empty, restricts the tenant_id label to these
	// IDs. Any other tenant is reported as UnknownTenantLabel, so a flood of
	// tenant IDs cannot explode metric cardinality. Empty means no limit.
	AllowTenants []string
}

// PrometheusEchoMiddleware sets up prometheus metrics with tenant_id label.
// metricsPath is typically "/metrics", subsystem is your service name.
func PrometheusEchoMiddleware(e *echo.Echo, subsystem, metricsPath string, reg prometheus.Registerer) {
	PrometheusEchoMiddlewareWithConfig(e, PrometheusConfig{
		Subsystem:   subsystem,
		MetricsPath: metricsPath,
		Registerer:  reg,
	})
}

// PrometheusEchoMiddlewareWithConfig sets up prometheus metrics configured by cfg.
func PrometheusEchoMiddlewareWithConfig(e *echo.Echo, cfg PrometheusConfig) {
	var allow map[string]struct{}
	if len(cfg.AllowTenants) > 0 {
		allow = make(map[string]struct{}, len(cfg.AllowTenants))
		for _, id := range cfg.AllowTenants {
			allow[id] = struct{}{}
		}
	}
	labelFuncs := map[string]echoprom.LabelValueFunc{
		PromTenantLabel: func(c echo.Context, _ error) string {
			tid := mt.ExtractTID(c.Request().Context())
			if allow == nil {
				return tid
			}
			if _, ok := allow[tid]; ok {
				return tid
			}
			return UnknownTenantLabel
		},
	}
	prom := echoprom.NewMiddlewareWithConfig(echoprom.MiddlewareConfig{
		Subsystem:  cfg.Subsystem,
		LabelFuncs: labelFuncs,
		Registerer: cfg.Registerer,
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/healthz", "/health", "/metrics":
				return true
			}
			return false
		},
	})
	e.Use(prom)
	e.GET(cfg.MetricsPath, echoprom.NewHandler())
}
