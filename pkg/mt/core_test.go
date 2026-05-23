package mt_test

import (
	"context"
	"testing"

	"github.com/PapaDanielVi/apadana/pkg/mt"
)

func TestExtractTID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "returns tenant ID from context",
			ctx:      mt.InjectTID(context.Background(), "acme"),
			expected: "acme",
		},
		{
			name:     "returns default when no tenant in context",
			ctx:      context.Background(),
			expected: "default",
		},
		{
			name:     "returns default for empty tenant ID",
			ctx:      mt.InjectTID(context.Background(), ""),
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mt.ExtractTID(tt.ctx)
			if got != tt.expected {
				t.Errorf("ExtractTID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInjectTID(t *testing.T) {
	t.Parallel()

	ctx := mt.InjectTID(context.Background(), "test-tenant")
	got := mt.ExtractTID(ctx)
	if got != "test-tenant" {
		t.Errorf("InjectTID() then ExtractTID() = %q, want %q", got, "test-tenant")
	}
}

func TestInjectTID_EmptyUsesDefault(t *testing.T) {
	t.Parallel()

	ctx := mt.InjectTID(context.Background(), "")
	got := mt.ExtractTID(ctx)
	if got != "default" {
		t.Errorf("InjectTID(\"\") then ExtractTID() = %q, want %q", got, "default")
	}
}

type fakeTenantGetter struct {
	tenantID string
}

func (f fakeTenantGetter) TenantID() string {
	return f.tenantID
}

func TestInjectTenantFromObj(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		obj      any
		expected string
	}{
		{
			name:     "extracts from tenantIDGetter",
			obj:      fakeTenantGetter{tenantID: "from-obj"},
			expected: "from-obj",
		},
		{
			name:     "falls back to default for non-getter",
			obj:      "not a getter",
			expected: "default",
		},
		{
			name:     "falls back to default for nil",
			obj:      nil,
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := mt.InjectTenantFromObj(context.Background(), tt.obj)
			got := mt.ExtractTID(ctx)
			if got != tt.expected {
				t.Errorf("InjectTenantFromObj() then ExtractTID() = %q, want %q", got, tt.expected)
			}
		})
	}
}
