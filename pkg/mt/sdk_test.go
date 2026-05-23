package mt_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/PapaDanielVi/apadana/pkg/mt"
)

// eagerConfig implements lazyIniter returning false for eager initialization.
type eagerConfig struct {
	value string
}

func (e eagerConfig) LazyInit() bool { return false }

// centralizedConfig implements centralizedSDKMer for centralized SDK mode.
type centralizedConfig struct {
	centralized bool
}

func (c centralizedConfig) IsCentralized() bool {
	return c.centralized
}

// centralizedEagerConfig implements both interfaces.
type centralizedEagerConfig struct {
	centralized bool
}

func (c centralizedEagerConfig) IsCentralized() bool {
	return c.centralized
}

func (c centralizedEagerConfig) LazyInit() bool { return false }

func TestNewSDKMgr_LazyInit(t *testing.T) {
	t.Parallel()

	var initCount atomic.Int32
	initFn := func(_ context.Context, _ string) string {
		initCount.Add(1)
		return "sdk-acme"
	}

	configs := map[string]string{
		"acme": "config-acme",
	}

	mgr := mt.NewSDKMgr[string, string](configs, initFn)

	// Plain string doesn't implement lazyIniter, so lazyInit returns true by default
	if initCount.Load() != 0 {
		t.Errorf("NewSDKMgr should lazy init, got %d inits", initCount.Load())
	}

	ctx := mt.InjectTID(context.Background(), "acme")
	got := mgr.Get(ctx)
	if got != "sdk-acme" {
		t.Errorf("Get() = %q, want %q", got, "sdk-acme")
	}
	if initCount.Load() != 1 {
		t.Errorf("Get() should trigger init, got %d inits", initCount.Load())
	}
}

func TestNewSDKMgr_EagerInit(t *testing.T) {
	t.Parallel()

	var initCount atomic.Int32
	initFn := func(_ context.Context, _ eagerConfig) string {
		initCount.Add(1)
		return "sdk"
	}

	configs := map[string]eagerConfig{
		"acme": {value: "config-acme"},
	}

	mgr := mt.NewSDKMgr[string, eagerConfig](configs, initFn)
	if initCount.Load() != 1 {
		t.Errorf("NewSDKMgr with eager init should call init once, got %d", initCount.Load())
	}

	// Second Get should not re-init
	_ = mgr.Get(mt.InjectTID(context.Background(), "acme"))
	if initCount.Load() != 1 {
		t.Errorf("Get() should not re-init, got %d inits", initCount.Load())
	}
}

func TestSDKMgr_Get_Centralized(t *testing.T) {
	t.Parallel()

	initFn := func(_ context.Context, _ centralizedConfig) string {
		return "central-sdk"
	}

	configs := map[string]centralizedConfig{
		"default": {centralized: true},
	}

	mgr := mt.NewSDKMgr[string, centralizedConfig](configs, initFn)

	ctx := mt.InjectTID(context.Background(), "any-tenant")
	got := mgr.Get(ctx)
	if got != "central-sdk" {
		t.Errorf("Get() = %q, want %q", got, "central-sdk")
	}
}

func TestSDKMgr_Map(t *testing.T) {
	t.Parallel()

	initFn := func(_ context.Context, cfg eagerConfig) string {
		return "sdk-" + cfg.value
	}

	configs := map[string]eagerConfig{
		"a": {value: "cfg-a"},
		"b": {value: "cfg-b"},
	}

	mgr := mt.NewSDKMgr[string, eagerConfig](configs, initFn)

	visited := make(map[string]string)
	mgr.Map(func(tenant string, sdk string) {
		visited[tenant] = sdk
	})

	if len(visited) != 2 {
		t.Errorf("Map() visited %d entries, want 2", len(visited))
	}
	if visited["a"] != "sdk-cfg-a" {
		t.Errorf("Map() tenant a = %q, want %q", visited["a"], "sdk-cfg-a")
	}
	if visited["b"] != "sdk-cfg-b" {
		t.Errorf("Map() tenant b = %q, want %q", visited["b"], "sdk-cfg-b")
	}
}

func TestNewSDKMgrE_Error(t *testing.T) {
	t.Parallel()

	initFn := func(_ context.Context, _ eagerConfig) (string, error) {
		return "", context.DeadlineExceeded
	}

	configs := map[string]eagerConfig{
		"acme": {value: "config-acme"},
	}

	_, err := mt.NewSDKMgrE[string, eagerConfig](configs, initFn)
	if err == nil {
		t.Error("NewSDKMgrE should return error when init fails")
	}
}

func TestNewSDKMgrE_Success(t *testing.T) {
	t.Parallel()

	initFn := func(_ context.Context, cfg string) (string, error) {
		return "sdk-" + cfg, nil
	}

	configs := map[string]string{
		"acme": "config-acme",
	}

	mgr, err := mt.NewSDKMgrE[string, string](configs, initFn)
	if err != nil {
		t.Fatalf("NewSDKMgrE failed: %v", err)
	}

	ctx := mt.InjectTID(context.Background(), "acme")
	got, err := mgr.Get(ctx)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != "sdk-config-acme" {
		t.Errorf("Get() = %q, want %q", got, "sdk-config-acme")
	}
}

func TestNewSDKMgrE_Centralized(t *testing.T) {
	t.Parallel()

	initFn := func(_ context.Context, _ centralizedConfig) (string, error) {
		return "central", nil
	}

	configs := map[string]centralizedConfig{
		"default": {centralized: true},
	}

	mgr, err := mt.NewSDKMgrE[string, centralizedConfig](configs, initFn)
	if err != nil {
		t.Fatalf("NewSDKMgrE failed: %v", err)
	}

	if !mgr.IsCentralized() {
		t.Error("IsCentralized() should return true")
	}

	ctx := mt.InjectTID(context.Background(), "any")
	got, err := mgr.Get(ctx)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != "central" {
		t.Errorf("Get() = %q, want %q", got, "central")
	}
}

func TestNewSDKMgrWMet(t *testing.T) {
	t.Parallel()

	type metrics struct {
		calls atomic.Int32
	}

	initFn := func(_ context.Context, cfg string, m *metrics) string {
		m.calls.Add(1)
		return "sdk-" + cfg
	}

	configs := map[string]string{
		"acme": "config-acme",
	}

	m := &metrics{}
	mgr := mt.NewSDKMgrWMet[string, string, *metrics](configs, initFn, m)

	ctx := mt.InjectTID(context.Background(), "acme")
	got := mgr.Get(ctx)
	if got != "sdk-config-acme" {
		t.Errorf("Get() = %q, want %q", got, "sdk-config-acme")
	}
	if m.calls.Load() != 1 {
		t.Errorf("initFn called %d times, want 1", m.calls.Load())
	}
}
