package fiber_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fiberv2 "github.com/gofiber/fiber/v2"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	fibermw "github.com/PapaDanielVi/apadana/v2/pkg/middleware/fiber"
)

func newApp(cfg fibermw.Config) *fiberv2.App {
	app := fiberv2.New()
	app.Use(fibermw.Tenant(cfg))
	app.Get("/", func(c *fiberv2.Ctx) error {
		tid, _ := tctx.TenantIDFromContext(c.UserContext())
		return c.SendString(tid)
	})
	return app
}

func TestTenant_InjectsTenant(t *testing.T) {
	t.Parallel()

	app := newApp(fibermw.Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "acme" {
		t.Errorf("body = %q, want %q", string(body), "acme")
	}
}

func TestTenant_RequiredRejectsMissing(t *testing.T) {
	t.Parallel()

	app := newApp(fibermw.Config{Required: true})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestTenant_QueryFallback(t *testing.T) {
	t.Parallel()

	app := newApp(fibermw.Config{Query: "tenant"})
	req := httptest.NewRequest(http.MethodGet, "/?tenant=globex", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "globex" {
		t.Errorf("body = %q, want %q", string(body), "globex")
	}
}
