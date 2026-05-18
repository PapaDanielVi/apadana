package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	echoprom "github.com/labstack/echo-contrib/echoprometheus"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	"github.com/PapaDanielVi/apadana/pkg/mt"
)

const (
	UserCtxKey      = "user"
	UserIDCtxKey    = "user_id"
	UserAuthHeader  = "X-User-Data"
	OAuth2Header    = "x-oauth2-data"
	PromTenantLabel = "tenant_id"
)

// UserInfo holds decoded user data from auth header.
type UserInfo struct {
	UserID      uint   `json:"user_id"`
	Sub         string `json:"sub"`
	Exp         int    `json:"exp"`
	Iat         int    `json:"iat"`
	Iss         int    `json:"iss"`
	JTI         string `json:"jti"`
	Group       string `json:"group"`
	TrustedAuth bool   `json:"trustedauth"`
	QA          bool   `json:"qa"`
	SID         string `json:"sid"`
	TenantID    string `json:"tid"`
}

// OAuth2Info holds decoded OAuth2 data from auth header.
type OAuth2Info struct {
	UserID   uint   `json:"user_id"`
	Sub      string `json:"sub"`
	Exp      int    `json:"exp"`
	Iat      int    `json:"iat"`
	Iss      string `json:"iss"`
	JTI      string `json:"jti"`
	Client   string `json:"client"`
	Scope    string `json:"scope"`
	TenantID string `json:"tid"`
}

// InjectTenantHTTP sets tenant ID header from ctx on req.
func InjectTenantHTTP(ctx context.Context, req *http.Request) {
	tid, _ := tctx.TenantIDFromContext(ctx)
	req.Header.Set("X-Tenant-ID", tid)
}

// DelTenantHTTP removes tenant ID header from req.
func DelTenantHTTP(req *http.Request) {
	req.Header.Del("X-Tenant-ID")
}

// TenantEchoMiddleware returns Echo middleware that extracts tenant ID from X-Tenant-ID header.
// If setDefault is true, missing header uses default tenant ID.
func TenantEchoMiddleware(setDefault bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if mt.ExtractTID(c.Request().Context()) != mt.ExtractTID(context.Background()) {
				return next(c)
			}
			tid := c.Request().Header.Get("X-Tenant-ID")
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

// extractAndInjectUser is reusable logic for user auth headers.
func extractAndInjectUser(c echo.Context) (*UserInfo, error) {
	raw := c.Request().Header.Get(UserAuthHeader)
	if raw == "" {
		return nil, echo.ErrUnauthorized
	}
	var user UserInfo
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return nil, echo.ErrUnauthorized
	}
	if user.UserID == 0 {
		return nil, echo.ErrUnauthorized
	}
	c.Set(UserIDCtxKey, user.UserID)
	c.Set(UserCtxKey, user)
	if user.TenantID != "" && user.TenantID != mt.ExtractTID(c.Request().Context()) {
		ctx := mt.InjectTID(c.Request().Context(), user.TenantID)
		c.SetRequest(c.Request().WithContext(ctx))
	}
	return &user, nil
}

// UserAuthEchoMiddleware decodes X-User-Data and injects user info into echo context.
func UserAuthEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := extractAndInjectUser(c)
		if err != nil {
			return err
		}
		return next(c)
	}
}

// UserOptAuthEchoMiddleware optionally decodes X-User-Data if present.
func UserOptAuthEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if raw := c.Request().Header.Get(UserAuthHeader); raw != "" {
			_, _ = extractAndInjectUser(c)
		}
		return next(c)
	}
}

// ExtractUserID extracts user ID from echo context.
func ExtractUserID(c echo.Context) (uint, bool) {
	if c == nil {
		return 0, false
	}
	v := c.Get(UserIDCtxKey)
	if v == nil {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}

func extractAndInjectOAuth2(c echo.Context) (*OAuth2Info, error) {
	raw := c.Request().Header.Get(OAuth2Header)
	if raw == "" {
		return nil, echo.ErrUnauthorized
	}
	var user OAuth2Info
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return nil, echo.ErrUnauthorized
	}
	if user.UserID == 0 {
		return nil, echo.ErrUnauthorized
	}
	c.Set(UserIDCtxKey, user.UserID)
	c.Set(UserCtxKey, user)
	if user.TenantID != "" && user.TenantID != mt.ExtractTID(c.Request().Context()) {
		ctx := mt.InjectTID(c.Request().Context(), user.TenantID)
		c.SetRequest(c.Request().WithContext(ctx))
	}
	return &user, nil
}

// OAuth2EchoMiddleware decodes x-oauth2-data and injects user info into echo context.
func OAuth2EchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := extractAndInjectOAuth2(c)
		if err != nil {
			return err
		}
		return next(c)
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
