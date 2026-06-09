// Command resolver-chain demonstrates tenant resolution via chain-of-responsibility.
//
// It shows:
//   - Building a resolver chain with multiple resolution strategies
//   - Priority ordering: header → query → subdomain → cookie
//   - Per-tenant metadata caching with Registry
//
// Run it and try:
//
//	curl -H "X-Tenant-Id: acme" http://localhost:8081/info
//	curl "http://localhost:8081/info?tenant=globex"
//	curl -H "Cookie: tenant_id=initech" http://localhost:8081/info
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/PapaDanielVi/apadana/pkg/resolver"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func main() {
	// Create a tenant registry.
	reg := resolver.NewRegistry(func(ctx context.Context) (map[string]resolver.Tenant, error) {
		return map[string]resolver.Tenant{
			"acme":   {ID: "acme", Name: "Acme Corp", Plan: "enterprise"},
			"globex": {ID: "globex", Name: "Globex Inc", Plan: "pro"},
			"initech": {ID: "initech", Name: "Initech", Plan: "basic"},
		}, nil
	})

	if err := reg.Load(context.Background()); err != nil {
		slog.Error("failed to load tenants", "error", err)
		return
	}

	// Build a resolver chain with priority order.
	chain := resolver.NewChain(
		resolver.FromHeader("X-Tenant-Id"),      // First priority: header
		resolver.FromQuery("tenant"),              // Second priority: query param
		resolver.FromSubdomain(),                  // Third priority: subdomain
		resolver.FromCookie("tenant_id"),          // Fourth priority: cookie
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := tctx.TenantIDFromContext(r.Context())
		log := slog.Default().With("tenant_id", tenantID)
		log.Info("info request")

		tenant, ok := reg.Get(tenantID)
		if !ok {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":   tenant.ID,
			"name": tenant.Name,
			"plan": tenant.Plan,
		})
	})

	handler := chain.Middleware()(mux)

	slog.Info("starting resolver-chain server on :8081")
	if err := http.ListenAndServe(":8081", handler); err != nil {
		slog.Error("server failed", "error", err)
	}
}