// Package redis provides multi-tenant Redis v9 tools.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// TenantClient wraps redis.Client with automatic key prefixing.
type TenantClient struct {
	client *redis.Client
}

// NewClient creates a new TenantClient.
func NewClient(ctx context.Context, opts *redis.Options) *TenantClient {
	return &TenantClient{client: redis.NewClient(opts)}
}

// KeyPrefix returns the tenant-specific key prefix.
func (tc *TenantClient) KeyPrefix(ctx context.Context) string {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	return fmt.Sprintf("tenant:{%s}:", tenantID)
}

func (tc *TenantClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return tc.client.Get(ctx, tc.KeyPrefix(ctx)+key)
}

func (tc *TenantClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return tc.client.Set(ctx, tc.KeyPrefix(ctx)+key, value, expiration)
}
