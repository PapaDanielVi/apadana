// Package burst provides per-tenant burst rate limiting.
package burst

import (
	"context"
	"sync"
	"time"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

type tokenBucket struct {
	rate   int
	burst  int
	tokens int
	last   time.Time
	mu     sync.Mutex
}

// Controller manages per-tenant rate limiters.
type Controller struct {
	rate       int
	burst      int
	controllers sync.Map // tenantID → *tokenBucket
}

// New creates a Controller with given rate and burst per tenant.
func New(rate, burst int) *Controller {
	return &Controller{rate: rate, burst: burst}
}

// Allow checks if a request is allowed for the tenant in ctx.
func (c *Controller) Allow(ctx context.Context) bool {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok {
		return false
	}

	v, _ := c.controllers.LoadOrStore(tenantID, &tokenBucket{
		rate:   c.rate,
		burst:  c.burst,
		tokens: c.burst,
		last:   time.Now(),
	})
	bucket := v.(*tokenBucket)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.last).Seconds()
	bucket.tokens += int(elapsed * float64(bucket.rate))
	if bucket.tokens > bucket.burst {
		bucket.tokens = bucket.burst
	}
	bucket.last = now

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	return false
}
