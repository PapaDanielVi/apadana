// Package httpclient provides an HTTP client that injects tenant ID headers.
package httpclient

import (
	"context"
	"net/http"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// Client wraps http.Client to inject X-Tenant-Id header.
type Client struct {
	client *http.Client
}

// New creates a new Client.
func New() *Client {
	return &Client{client: &http.Client{}}
}

// Do executes req with X-Tenant-Id header from ctx.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	if tenantID != "" {
		req.Header.Set("X-Tenant-Id", tenantID)
	}
	return c.client.Do(req.WithContext(ctx))
}
