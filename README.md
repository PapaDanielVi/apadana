# Apadana: Multi-Tenant SDK for Go

[![Test](https://github.com/PapaDanielVi/apadana/actions/workflows/test.yml/badge.svg)](https://github.com/PapaDanielVi/apadana/actions/workflows/test.yml)
[![Lint](https://github.com/PapaDanielVi/apadana/actions/workflows/lint.yml/badge.svg)](https://github.com/PapaDanielVi/apadana/actions/workflows/lint.yml)
[![Security](https://github.com/PapaDanielVi/apadana/actions/workflows/security.yml/badge.svg)](https://github.com/PapaDanielVi/apadana/actions/workflows/security.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/PapaDanielVi/apadana)](https://github.com/PapaDanielVi/apadana)
[![License](https://img.shields.io/github/license/PapaDanielVi/apadana)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/PapaDanielVi/apadana)](https://goreportcard.com/report/github.com/PapaDanielVi/apadana)
[![GoDoc](https://pkg.go.dev/badge/github.com/PapaDanielVi/apadana)](https://pkg.go.dev/github.com/PapaDanielVi/apadana)

A Go SDK providing building blocks for multi-tenant applications. Handles tenant identification, context propagation, per-tenant configuration, metrics, logging, and instrumentation across common infrastructure.

## Features

- **Tenant Isolation** — Extract, propagate, and inject tenant IDs across HTTP, gRPC, Kafka, NATS, and RabbitMQ
- **SaaS-Ready** — Per-tenant configuration, singletons, timezones, and rate limiting out of the box
- **Context Propagation** — Thread-safe tenant context with `context.Context` integration
- **Multi-Tenant Middleware** — HTTP (standard lib & Echo) middlewares for header, query, subdomain, cookie, and JWT extraction
- **Tenant Resolver Chain** — Chain-of-responsibility pattern for flexible tenant resolution
- **Per-Tenant Infrastructure** — Kafka/NATS/RabbitMQ header injection
- **Observability** — OpenTelemetry span processor, structured logging with `log/slog`
- **Generic SDK Managers** — Type-safe `ConfigMgr[T]` and `SDKMgr[T,C]` for any multi-tenant resource
- **YAML Config Expansion** — Merge default and per-tenant configs with `${tenant}` placeholder replacement
- **gRPC Support** — Unary and stream interceptors for tenant metadata propagation

## Installation

```bash
go get github.com/PapaDanielVi/apadana
```

## Quick Start

```go
package main

import (
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

	mw := middleware.TenantMiddleware(middleware.FromHeader("X-Tenant-Id"))(handler)
	http.ListenAndServe(":8080", mw)
}
```

See [examples/basic-http](examples/basic-http) for a complete runnable example.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│ HTTP/gRPC   │────▶│  Middleware  │────▶│ context.Context │
│ Request     │     │  /Resolver   │     │ (tenant ID)     │
└─────────────┘     └──────────────┘     └────────┬────────┘
                                                   │
                    ┌──────────────────────────────┼──────────────┐
                    │              │               │              │
              ┌─────▼─────┐ ┌──────▼──────┐ ┌─────▼─────┐ ┌─────▼─────┐
              │ ConfigMgr │ │  SDKMgr     │ │  Logger   │ │  OTel     │
              │ [T]       │ │  [S,C]      │ │  (slog)   │ │  SpanProc │
              └───────────┘ └─────────────┘ └───────────┘ └───────────┘
                    │              │               │              │
              ┌─────▼──────────────▼───────────────▼──────────────▼─────┐
              │              Kafka / NATS / RabbitMQ                    │
              │              (X-Tenant-Id header injection)             │
              └────────────────────────────────────────────────────────┘
```

## Packages

| Package          | Description                              | Key Functions / Types                                                                                              |
| ---------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `pkg/context`    | Tenant ID in `context.Context`           | `WithTenantID`, `TenantIDFromContext`, `HasTenantID`                                                               |
| `pkg/mt`         | Generic multi-tenant tools               | `SetDefTenant`, `ExtractTID`, `InjectTID`, `ConfigMgr[T]`, `SDKMgr[T,C]`, `CloneCtx`, `ExpandConfigReader`         |
| `pkg/middleware` | HTTP middlewares (std, Echo)             | `TenantMiddleware`, `FromHeader`, `FromQuery`, `FromSubdomain`, `TenantEchoMiddleware`, `PrometheusEchoMiddleware` |
| `pkg/resolver`   | Tenant resolver chain                    | `Chain`, `FromHeader`, `FromQuery`, `FromSubdomain`, `FromCookie`, `FromJWTClaim`, `Tenant`, `Registry`            |
| `pkg/grpc`       | gRPC interceptors                        | `UnaryServerInterceptor`, `StreamServerInterceptor`, `UnaryClientInterceptor`, `StreamClientInterceptor`           |
| `pkg/timezone`   | Per-tenant timezone settings             | `Set`, `Now` — returns time in tenant's timezone                                                                   |
| `pkg/logger`     | Logger with `tenant_id` field            | `New(ctx)` — wraps `log/slog` with tenant ID                                                                       |
| `pkg/otel`       | OpenTelemetry span processor             | `NewTenantIDProcessor()` — adds `tenant_id` to spans                                                               |
| `pkg/rabbitmq`   | RabbitMQ with `X-Tenant-Id` header       | `Publisher`, `Consumer`                                                                                            |
| `pkg/kafka`      | Kafka with `X-Tenant-Id` header          | `Producer`, `Consumer`                                                                                             |
| `pkg/nats`       | NATS with `X-Tenant-Id` header           | `Publisher`, `Subscriber`                                                                                          |
| `pkg/httpclient` | HTTP client with tenant header injection | `Do(ctx, req)` — auto-injects `X-Tenant-Id`                                                                        |
| `pkg/burst`      | Per-tenant token bucket rate limiter     | `New(rate, burst)`, `Allow(ctx)`                                                                                   |

## Design Decisions

See [docs/adr](docs/adr) for architecture decision records:

- [ADR-001: Use Generics for Type-Safe Multi-Tenant Resources](docs/adr/001-generics.md)
- [ADR-002: Context Propagation Strategy](docs/adr/002-context-propagation.md)

## License

MIT
