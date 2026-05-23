# apadana: Multi-Tenant SDK for Go

[![CI](https://github.com/PapaDanielVi/apadana/actions/workflows/ci.yml/badge.svg)](https://github.com/PapaDanielVi/apadana/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/PapaDanielVi/apadana)](https://github.com/PapaDanielVi/apadana)
[![License](https://img.shields.io/github/license/PapaDanielVi/apadana)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/PapaDanielVi/apadana)](https://goreportcard.com/report/github.com/PapaDanielVi/apadana)
[![GoDoc](https://pkg.go.dev/badge/github.com/PapaDanielVi/apadana)](https://pkg.go.dev/github.com/PapaDanielVi/apadana)

A Go SDK providing building blocks for multi-tenant applications. Handles tenant identification, context propagation, per-tenant configuration, metrics, logging, and instrumentation across common infrastructure.

## Features

- **Tenant Isolation** — Extract, propagate, and inject tenant IDs across HTTP, gRPC, Kafka, NATS, and RabbitMQ
- **SaaS-Ready** — Per-tenant configuration, singletons, timezones, and rate limiting out of the box
- **Context Propagation** — Thread-safe tenant context with `context.Context` integration
- **Multi-Tenant Middleware** — HTTP (standard lib & Echo) middlewares for header, query, and subdomain extraction
- **Per-Tenant Infrastructure** — Kafka/NATS/RabbitMQ header injection
- **Observability** — OpenTelemetry span processor, structured logging with `log/slog`
- **Generic SDK Managers** — Type-safe `ConfigMgr[T]` and `SDKMgr[T,C]` for any multi-tenant resource
- **YAML Config Expansion** — Merge default and per-tenant configs with `${tenant}` placeholder replacement

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

	// Extract tenant ID from X-Tenant-Id header
	mw := middleware.TenantMiddleware(middleware.FromHeader("X-Tenant-Id"))(handler)

	http.ListenAndServe(":8080", mw)
}
```

## Packages

| Package | Description | Key Functions / Types |
|---------|-------------|----------------------|
| `pkg/context` | Tenant ID in `context.Context` | `WithTenantID`, `TenantIDFromContext`, `HasTenantID` |
| `pkg/mt` | Generic multi-tenant tools | `SetDefTenant`, `ExtractTID`, `InjectTID`, `ConfigMgr[T]`, `SDKMgr[T,C]`, `CloneCtx`, `ExpandConfigReader` |
| `pkg/middleware` | HTTP middlewares (std, Echo) | `TenantMiddleware`, `FromHeader`, `FromQuery`, `FromSubdomain`, `TenantEchoMiddleware`, `UserAuthEchoMiddleware`, `PrometheusEchoMiddleware` |
| `pkg/timezone` | Per-tenant timezone settings | `Set`, `Now` — returns time in tenant's timezone |
| `pkg/logger` | Logger with `tenant_id` field | `New(ctx)` — wraps `log/slog` with tenant ID |
| `pkg/otel` | OpenTelemetry span processor | `NewTenantIDProcessor()` — adds `tenant_id` to spans |
| `pkg/rabbitmq` | RabbitMQ with `X-Tenant-Id` header | `Publisher`, `Consumer` |
| `pkg/kafka` | Kafka with `X-Tenant-Id` header | `Producer`, `Consumer` |
| `pkg/nats` | NATS with `X-Tenant-Id` header | `Publisher`, `Subscriber` |
| `pkg/httpclient` | HTTP client with tenant header injection | `Do(ctx, req)` — auto-injects `X-Tenant-Id` |
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
