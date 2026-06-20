// Command basic-http demonstrates a minimal multi-tenant HTTP server.
//
// It shows:
//   - Tenant extraction from the X-Tenant-Id header
//   - Per-tenant configuration via ConfigMgr
//   - Tenant-aware logging and rate limiting
//
// Run it and try:
//
//	curl -H "X-Tenant-Id: acme" http://localhost:8080/config
//	curl -H "X-Tenant-Id: globex" http://localhost:8080/config
//	curl -H "X-Tenant-Id: acme" http://localhost:8080/hello
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/PapaDanielVi/apadana/v2/pkg/burst"
	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	"github.com/PapaDanielVi/apadana/v2/pkg/logger"
	"github.com/PapaDanielVi/apadana/v2/pkg/middleware"
	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

type appConfig struct {
	Name    string
	Timeout time.Duration
}

func main() {
	configs := map[string]appConfig{
		"acme": {
			Name:    "Acme Corp",
			Timeout: 30 * time.Second,
		},
		"globex": {
			Name:    "Globex Inc",
			Timeout: 15 * time.Second,
		},
	}

	mgr := mt.NewConfigMgr(configs, appConfig{
		Name:    "Default App",
		Timeout: 10 * time.Second,
	})

	rateLimiter := burst.New(10, 20)

	mux := http.NewServeMux()
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := tctx.TenantIDFromContext(r.Context())
		log := logger.New(r.Context())
		log.Info("config request", slog.String("tenant_id", tenantID))

		if !rateLimiter.Allow(r.Context()) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		cfg := mgr.Get(r.Context())
		fmt.Fprintf(w, "Tenant: %s\nName: %s\nTimeout: %s\n", tenantID, cfg.Name, cfg.Timeout)
	})

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := tctx.TenantIDFromContext(r.Context())
		log := logger.New(r.Context())
		log.Info("hello request", slog.String("tenant_id", tenantID))

		if !rateLimiter.Allow(r.Context()) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		cfg := mgr.Get(r.Context())
		fmt.Fprintf(w, "Hello from %s (%s)! Your config timeout is %s.\n", tenantID, cfg.Name, cfg.Timeout)
	})

	stack := middleware.TenantMiddleware(
		middleware.FromHeader("X-Tenant-Id"),
	)(mux)

	slog.Info("starting server on :8080")
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      stack,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
	}
}
