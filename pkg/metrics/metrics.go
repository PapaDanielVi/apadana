// Package metrics provides multi-tenant Prometheus instrumentation.
package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

// TenantCounter wraps a prometheus.CounterVec with automatic tenant_id label.
type TenantCounter struct {
	vec *prometheus.CounterVec
}

// NewCounter creates a counter with tenant_id label.
func NewCounter(name, help string) *TenantCounter {
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: name, Help: help},
		[]string{"tenant_id"},
	)
	prometheus.MustRegister(vec)
	return &TenantCounter{vec: vec}
}

// Inc increments the counter for the tenant in ctx.
func (tc *TenantCounter) Inc(ctx context.Context) {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	tc.vec.WithLabelValues(tenantID).Inc()
}

// TenantHistogram wraps a prometheus.HistogramVec with automatic tenant_id label.
type TenantHistogram struct {
	vec *prometheus.HistogramVec
}

// NewHistogram creates a histogram with tenant_id label.
func NewHistogram(name, help string, buckets []float64) *TenantHistogram {
	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		[]string{"tenant_id"},
	)
	prometheus.MustRegister(vec)
	return &TenantHistogram{vec: vec}
}

// Observe records a value for the tenant in ctx.
func (th *TenantHistogram) Observe(ctx context.Context, value float64) {
	tenantID, _ := tctx.TenantIDFromContext(ctx)
	th.vec.WithLabelValues(tenantID).Observe(value)
}
