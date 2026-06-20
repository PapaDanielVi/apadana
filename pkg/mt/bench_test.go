package mt_test

import (
	"context"
	"testing"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

func BenchmarkSDKMgr_Get(b *testing.B) {
	initFn := func(_ context.Context, cfg string) string {
		return "sdk-" + cfg
	}
	mgr := mt.NewSDKMgr[string, string](map[string]string{"acme": "config-acme"}, initFn)
	ctx := mt.InjectTID(context.Background(), "acme")

	// Warm the cache so the benchmark measures the steady-state read path.
	_ = mgr.Get(ctx)

	b.ReportAllocs()
	for b.Loop() {
		_ = mgr.Get(ctx)
	}
}
