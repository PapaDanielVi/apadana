package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestTenantEchoMiddleware_MissingHeaderStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        TenantEchoConfig
		header     string
		wantStatus int
	}{
		{"missing rejected with 400 by default", TenantEchoConfig{}, "", http.StatusBadRequest},
		{
			"missing rejected with custom status",
			TenantEchoConfig{MissingStatus: http.StatusUnauthorized},
			"",
			http.StatusUnauthorized,
		},
		{"present header passes", TenantEchoConfig{}, "acme", http.StatusOK},
		{"missing allowed with default tenant", TenantEchoConfig{SetDefault: true}, "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("X-Tenant-Id", tt.header)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := TenantEchoMiddlewareWithConfig(tt.cfg)(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})
			if err := handler(c); err != nil {
				t.Fatalf("handler error = %v", err)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
