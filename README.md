# apadana

A Go SDK providing building blocks for multi-tenant applications. Handles tenant identification, context propagation, per-tenant configuration, metrics, logging, and instrumentation across common infrastructure.

## Installation

```bash
go get github.com/PapaDanielVi/apadana
```

## Quick Start

```go
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/PapaDanielVi/apadana/pkg/context"
	"github.com/PapaDanielVi/apadana/pkg/middleware"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := context.TenantIDFromContext(r.Context())
		slog.Info("request", "tenant", tenantID)
		w.Write([]byte("hello " + tenantID))
	})

	// Extract tenant ID from X-Tenant-ID header
	mw := middleware.TenantMiddleware(middleware.FromHeader("X-Tenant-ID"))(handler)

	http.ListenAndServe(":8080", mw)
}
```

## Packages

| Package | Description | Key Functions / Types |
|---------|-------------|----------------------|
| `pkg/context` | Tenant ID in `context.Context` | `WithTenantID`, `TenantIDFromContext`, `HasTenantID` |
| `pkg/mt` | Generic multi-tenant tools | `SetDefTenant`, `ExtractTID`, `InjectTID`, `ConfigMgr[T]`, `SDKMgr[T,C]`, `CloneCtx`, `ExpandConfigReader` |
| `pkg/middleware` | HTTP middlewares (std, Echo) | `TenantMiddleware`, `FromHeader`, `FromQuery`, `FromSubdomain`, `TenantEchoMiddleware`, `UserAuthEchoMiddleware`, `PrometheusEchoMiddleware` |
| `pkg/replacer` | Tenant placeholder replacement | `Replace` — replaces `{tenant_id}` in strings |
| `pkg/config` | Per-tenant config storage | `Set`, `Get`, thread-safe key-value per tenant |
| `pkg/timezone` | Per-tenant timezone settings | `Set`, `Now` — returns time in tenant's timezone |
| `pkg/obj` | Lazy-init per-tenant singletons | `Register`, `Get` — centralized or per-tenant objects |
| `pkg/logger` | Logger with `tenant_id` field | `New(ctx)` — wraps `log/slog` with tenant ID |
| `pkg/otel` | OpenTelemetry span processor | `NewTenantIDProcessor()` — adds `tenant_id` to spans |
| `pkg/metrics` | Prometheus metrics with `tenant_id` label | `NewCounter`, `NewHistogram` — auto-labeled |
| `pkg/db` | Multi-tenant database support | `ColumnModel`, `DatabaseModel`, `InstanceModel` |
| `pkg/redis` | Redis with key prefixing | `NewClient(ctx, opts)` — keys prefixed `tenant:{id}:key` |
| `pkg/rabbitmq` | RabbitMQ with `X-Tenant-ID` header | `Publisher`, `Consumer` |
| `pkg/kafka` | Kafka with `X-Tenant-ID` header | `Producer`, `Consumer` |
| `pkg/nats` | NATS with `X-Tenant-ID` header | `Publisher`, `Subscriber` |
| `pkg/httpclient` | HTTP client with tenant header injection | `Do(ctx, req)` — auto-injects `X-Tenant-ID` |
| `pkg/burst` | Per-tenant token bucket rate limiter | `New(rate, burst)`, `Allow(ctx)` |
| `pkg/mt` — Additional Tools | | |
| &nbsp;&nbsp; `core.go` | Core tenant ID extraction/injection | `SetDefTenant`, `ExtractTID`, `InjectTID`, `InjectTenantFromObj` |
| &nbsp;&nbsp; `config.go` | Generic config manager | `ConfigMgr[T]{Get, Map, Tenants}` |
| &nbsp;&nbsp; `sdk.go` | Generic SDK managers | `SDKMgr[T,C]`, `SDKMgrE[T,C]`, `SDKMgrWMet[T,C,M]`, `NewSDKMgr`, `NewSDKMgrE`, `NewSDKMgrWMet` |
| &nbsp;&nbsp; `context.go` | Context cloning with OTEL preservation | `CloneCtx` — copies tenant + baggage + span |
| &nbsp;&nbsp; `yaml.go` | YAML config expansion | `ExpandConfigReader` — merges defaults, replaces `${tenant}` |
| &nbsp;&nbsp; `template.go` | Echo template rendering + tenant replacer | `TplRenderer`, `TenantRepl`, `NewTplRenderer` |
| &nbsp;&nbsp; `assertions.go` | SDK init interfaces | `lazyIniter`, `centralizedSDKMer` |
| &nbsp;&nbsp; `mocks.go` | Mock generators for SDK interfaces | `MockISDK[T]`, `MockISDKE[T]`, `NewISDKMock`, `NewISDKEMock` |
| `pkg/middleware` — Echo Extras | | |
| &nbsp;&nbsp; `echo.go` | Echo-specific middlewares | `TenantEchoMiddleware`, `UserAuthEchoMiddleware`, `OAuth2EchoMiddleware`, `PrometheusEchoMiddleware`, `ExtractUserID` |

## License

MIT
