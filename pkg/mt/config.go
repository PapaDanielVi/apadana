package mt

import (
	"context"
	"maps"
)

// ConfigMgr holds per-tenant configs with a fallback default.
type ConfigMgr[T any] struct {
	configs map[string]T
	defCfg   T
}

// NewConfigMgr creates a new manager with copied tenant configs and a default.
func NewConfigMgr[T any](configs map[string]T, defCfg T) *ConfigMgr[T] {
	copied := make(map[string]T, len(configs))
	maps.Copy(copied, configs)
	return &ConfigMgr[T]{
		configs: copied,
		defCfg:   defCfg,
	}
}

// Get returns config for the tenant in ctx, or default if not found.
func (m *ConfigMgr[T]) Get(ctx context.Context) T {
	if ctx == nil {
		return m.defCfg
	}
	tid := ExtractTID(ctx)
	cfg, exists := m.configs[tid]
	if !exists {
		return m.defCfg
	}
	return cfg
}

// Map calls f for each tenant config.
func (m *ConfigMgr[T]) Map(f func(string, T)) {
	for tenant, cfg := range m.configs {
		f(tenant, cfg)
	}
}

// Tenants returns all configured tenant IDs.
func (m *ConfigMgr[T]) Tenants() []string {
	tenants := make([]string, 0, len(m.configs))
	for tenant := range m.configs {
		tenants = append(tenants, tenant)
	}
	return tenants
}
