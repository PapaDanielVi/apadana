// Command mt-sdk demonstrates per-tenant SDK management.
//
// It shows:
//   - Managing per-tenant SDK instances
//   - Lazy initialization support
//   - Centralized SDK pattern
//
// Run to see SDK initialization patterns.
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/PapaDanielVi/apadana/pkg/mt"
)

// SDKConfig holds configuration for a payment SDK.
type SDKConfig struct {
	APIKey string
	Lazy   bool
}

func (c SDKConfig) LazyInit() bool { return c.Lazy }

func (c SDKConfig) IsCentralized() bool { return false }

// PaymentClient is a per-tenant SDK example.
type PaymentClient struct {
	apiKey string
}

func main() {
	slog.Info("mt-sdk example")

	// Example: Per-tenant SDK manager.
	sdkConfigs := map[string]SDKConfig{
		"acme":   {APIKey: "key-acme"},
		"globex": {APIKey: "key-globex"},
	}

	sdkMgr := mt.NewSDKMgr(sdkConfigs, func(ctx context.Context, cfg SDKConfig) *PaymentClient {
		slog.Info("initializing SDK", "api_key", cfg.APIKey)
		return &PaymentClient{apiKey: cfg.APIKey}
	})

	// Get SDK for specific tenant context.
	tenantCtx := mt.InjectTID(context.Background(), "acme")
	client := sdkMgr.Get(tenantCtx)
	slog.Info("got client for tenant", "api_key", client.apiKey)

	// Example 2: Lazy initialization.
	lazyConfigs := map[string]struct {
		APIKey string
		Lazy   bool
	}{
		"lazy": {APIKey: "lazy-key", Lazy: true},
	}

	_ = lazyConfigs // Used to show lazy pattern.

	slog.Info("lazy SDKs are initialized on first Get() call, not at startup")

	// Example 3: Centralized SDK pattern.
	// When you have a single SDK instance for all tenants.
	centralConfigs := map[string]struct {
		APIKey string
	}{
		"default": {APIKey: "central-key"},
	}

	_ = centralConfigs // Used to show centralized pattern.

	slog.Info("centralized SDKs use a single instance for all tenants")
	_ = sdkMgr
	fmt.Println("mt-sdk example complete")
}

// Note: In production, you would:
// 1. Load tenant configs from your database
// 2. Initialize SDKs per tenant
// 3. Use the SDK manager in your request handlers