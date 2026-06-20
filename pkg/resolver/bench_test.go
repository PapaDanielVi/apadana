package resolver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PapaDanielVi/apadana/v2/pkg/resolver"
)

func BenchmarkChain_Resolve(b *testing.B) {
	chain := resolver.NewChain(
		resolver.FromHeader("X-Tenant-Id"),
		resolver.FromQuery("tenant"),
		resolver.FromSubdomain(),
	)
	req := httptest.NewRequest(http.MethodGet, "/?tenant=acme", nil)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := chain.Resolve(req); err != nil {
			b.Fatal(err)
		}
	}
}
