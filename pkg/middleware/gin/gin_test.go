package gin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gingonic "github.com/gin-gonic/gin"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	ginmw "github.com/PapaDanielVi/apadana/v2/pkg/middleware/gin"
	"github.com/PapaDanielVi/apadana/v2/pkg/resolver"
)

func newRouter(cfg ginmw.Config) *gingonic.Engine {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	r.Use(ginmw.Tenant(cfg))
	r.GET("/", func(c *gingonic.Context) {
		tid, _ := tctx.TenantIDFromContext(c.Request.Context())
		c.String(http.StatusOK, tid)
	})
	return r
}

func TestTenant_InjectsTenant(t *testing.T) {
	t.Parallel()

	chain := resolver.NewChain(resolver.FromHeader("X-Tenant-Id"))
	r := newRouter(ginmw.Config{Chain: chain})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "acme" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "acme")
	}
}

func TestTenant_RequiredRejectsMissing(t *testing.T) {
	t.Parallel()

	chain := resolver.NewChain(resolver.FromHeader("X-Tenant-Id"))
	r := newRouter(ginmw.Config{Chain: chain, Required: true})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
