package mt_test

import (
	"context"
	"testing"

	"github.com/PapaDanielVi/apadana/pkg/mt"
)

func TestNewConfigMgr(t *testing.T) {
	t.Parallel()

	def := map[string]any{"timeout": 30}
	configs := map[string]any{
		"acme":   map[string]any{"timeout": 60},
		"globex": map[string]any{"timeout": 15},
	}

	mgr := mt.NewConfigMgr[any](configs, def)

	// Verify original map is not mutated — unknown tenant gets default
	configs["new"] = map[string]any{"timeout": 99}
	ctx := mt.InjectTID(context.Background(), "new")
	got := mgr.Get(ctx)
	if got == nil {
		t.Fatal("NewConfigMgr.Get() should return default, got nil")
	}
	gotMap, ok := got.(map[string]any)
	if !ok || gotMap["timeout"] != 30 {
		t.Errorf("NewConfigMgr should return default config for unknown tenant, got %v", got)
	}
}

func TestConfigMgr_Get(t *testing.T) {
	t.Parallel()

	def := "default-config"
	configs := map[string]string{
		"acme":   "acme-config",
		"globex": "globex-config",
	}

	mgr := mt.NewConfigMgr[string](configs, def)

	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "returns tenant config",
			ctx:      mt.InjectTID(context.Background(), "acme"),
			expected: "acme-config",
		},
		{
			name:     "returns default for unknown tenant",
			ctx:      mt.InjectTID(context.Background(), "unknown"),
			expected: "default-config",
		},
		{
			name:     "returns default for nil context",
			ctx:      nil,
			expected: "default-config",
		},
		{
			name:     "returns default for empty context",
			ctx:      context.Background(),
			expected: "default-config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mgr.Get(tt.ctx)
			if got != tt.expected {
				t.Errorf("Get() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigMgr_Map(t *testing.T) {
	t.Parallel()

	configs := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	mgr := mt.NewConfigMgr[int](configs, 0)

	visited := make(map[string]int)
	mgr.Map(func(tenant string, cfg int) {
		visited[tenant] = cfg
	})

	if len(visited) != 3 {
		t.Errorf("Map() visited %d tenants, want 3", len(visited))
	}
	for k, v := range configs {
		if visited[k] != v {
			t.Errorf("Map() tenant %s = %d, want %d", k, visited[k], v)
		}
	}
}

func TestConfigMgr_Tenants(t *testing.T) {
	t.Parallel()

	configs := map[string]bool{
		"x": true,
		"y": true,
	}

	mgr := mt.NewConfigMgr[bool](configs, false)
	tenants := mgr.Tenants()

	if len(tenants) != 2 {
		t.Errorf("Tenants() returned %d tenants, want 2", len(tenants))
	}

	tenantSet := make(map[string]bool)
	for _, t := range tenants {
		tenantSet[t] = true
	}
	for k := range configs {
		if !tenantSet[k] {
			t.Errorf("Tenants() missing tenant %q", k)
		}
	}
}
