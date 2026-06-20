// Command httpclient demonstrates an HTTP client with tenant ID injection.
//
// It shows:
//   - Creating tenant-aware HTTP client
//   - Automatic X-Tenant-Id header injection from context
//
// Run the basic-http example server first to test against.
package main

import (
	"context"
	"log/slog"
	"net/http"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	"github.com/PapaDanielVi/apadana/v2/pkg/httpclient"
)

func main() {
	// Create a tenant-aware HTTP client.
	client := httpclient.New()

	// Example: Making a request with tenant context.
	tenantCtx := tctx.WithTenantID(context.Background(), "acme")

	req, err := http.NewRequestWithContext(tenantCtx, http.MethodGet, "http://localhost:8080/config", nil)
	if err != nil {
		slog.Error("request creation failed", "error", err)
		return
	}

	resp, err := client.Do(tenantCtx, req)
	if err != nil {
		slog.Error("request failed", "error", err)
		return
	}
	defer resp.Body.Close()

	slog.Info("httpclient made request with X-Tenant-Id header", "tenant_id", "acme")
}

// Note: The Client.Do method automatically injects X-Tenant-Id header
// from the context into outgoing requests.
