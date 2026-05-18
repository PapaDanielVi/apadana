# apadana

apadana is a Go SDK that provides building blocks for multi-tenant applications. It handles tenant identification, context propagation, per-tenant configuration, metrics, logging, and instrumentation across common infrastructure: databases, Redis, Kafka, RabbitMQ, NATS, and HTTP.

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

### Core

#### `pkg/context` — Context Management
Utilities for storing and retrieving tenant IDs in `context.Context`.

```go
ctx := context.WithTenantID(context.Background(), "acme-corp")
tenantID, ok := context.TenantIDFromContext(ctx)
```

#### `pkg/replacer` — Tenant Replacer
Replaces `{tenant_id}` placeholders in strings.

```go
s := replacer.Replace("db_{tenant_id}_schema", "acme")
// s == "db_acme_schema"
```

#### `pkg/config` — Config Management
Thread-safe per-tenant configuration storage.

```go
config.Set("acme", "db_host", "localhost:5432")
ctx := context.WithTenantID(context.Background(), "acme")
host, _ := config.Get(ctx, "db_host")
```

#### `pkg/timezone` — Timezone Per Tenant
Manage per-tenant timezone settings.

```go
loc, _ := time.LoadLocation("America/New_York")
timezone.Set("acme", loc)

ctx := context.WithTenantID(context.Background(), "acme")
now := timezone.Now(ctx) // time in tenant's timezone
```

#### `pkg/obj` — Object Management SDK
Lazy-initialized, centralized singleton objects per tenant.

```go
obj.Register("cache", func() (any, error) {
	return make(map[string]string), nil
})

ctx := context.WithTenantID(context.Background(), "acme")
cache, _ := obj.Get(ctx, "cache") // same instance for all calls
```

### Identification & Logging

#### `pkg/middleware` — HTTP Middlewares
Extract tenant ID from HTTP requests and inject into context.

```go
// From header
m := middleware.TenantMiddleware(middleware.FromHeader("X-Tenant-ID"))(handler)

// From query parameter
m := middleware.TenantMiddleware(middleware.FromQuery("tenant"))(handler)

// From subdomain
m := middleware.TenantMiddleware(middleware.FromSubdomain())(handler)
```

#### `pkg/logger` — Logger Wrapper
Wraps `log/slog` to automatically include `tenant_id` in log output.

```go
ctx := context.WithTenantID(context.Background(), "acme")
logger := logger.New(ctx)
logger.Info("processing order") // includes tenant_id=acme
```

### Instrumentation

#### `pkg/otel` — OpenTelemetry Tenant Injection
Automatically adds `tenant_id` as a span attribute.

```go
processor := otel.NewTenantIDProcessor()
// Register with your OTEL SDK setup
```

#### `pkg/metrics` — Prometheus Metrics
Multi-tenant counters and histograms with automatic `tenant_id` label.

```go
counter := metrics.NewCounter("http_requests_total", "Total requests")
histogram := metrics.NewHistogram("request_duration_seconds", "Request duration", nil)

ctx := context.WithTenantID(context.Background(), "acme")
counter.Inc(ctx)
histogram.Observe(ctx, 0.5)
```

### Data Layer

#### `pkg/db` — Tenant Registry
Multi-tenancy support for MySQL/PostgreSQL using three models:

- **ColumnModel**: Adds `WHERE tenant_id = 'id'` to queries
- **DatabaseModel**: Separate database per tenant
- **InstanceModel**: Separate connection per tenant

```go
reg := db.New(db.ColumnModel, "mysql", "user:pass@/db")
mw := reg.ColumnMiddleware()
query := mw(ctx, "SELECT * FROM orders")
// query == "SELECT * FROM orders WHERE tenant_id = 'acme'"
```

#### `pkg/redis` — Redis v9 Tools
Wraps `go-redis/v9` with automatic key prefixing (`tenant:{id}:key`).

```go
client := redis.NewClient(ctx, &redis.Options{Addr: "localhost:6379"})
client.Set(ctx, "mykey", "myvalue", 0)
val, _ := client.Get(ctx, "mykey").Result()
```

### Messaging

#### `pkg/rabbitmq` — RabbitMQ Tools
Publish and consume messages with `X-Tenant-ID` header.

```go
// Publish
pub := rabbitmq.NewPublisher(ch)
pub.Publish(ctx, "exchange", "routing-key", []byte("msg"))

// Consume
cons := rabbitmq.NewConsumer(ch)
cons.Consume(ctx, "queue", func(ctx context.Context, d amqp.Delivery) {
	// ctx has tenant ID from message headers
})
```

#### `pkg/kafka` — Kafka Tools
Produce and consume Kafka messages with tenant ID in headers.

```go
// Produce
prod := kafka.NewProducer(writer)
prod.Produce(ctx, "topic", []byte("key"), []byte("value"))

// Consume
cons := kafka.NewConsumer(reader)
cons.Consume(ctx, func(ctx context.Context, msg kafka.Message) {
	// ctx has tenant ID from message headers
})
```

#### `pkg/nats` — NATS Tools
Publish and subscribe with tenant ID in NATS headers.

```go
// Publish
pub := nats.NewPublisher(nc)
pub.Publish(ctx, "subject", []byte("msg"))

// Subscribe
sub := nats.NewSubscriber(nc)
sub.Subscribe(ctx, "subject", func(ctx context.Context, msg *nats.Msg) {
	// ctx has tenant ID from message headers
})
```

### HTTP & Advanced

#### `pkg/httpclient` — HTTP Client Tools
Wraps `http.Client` to inject `X-Tenant-ID` header into outgoing requests.

```go
client := httpclient.New()
req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
resp, err := client.Do(ctx, req) // X-Tenant-ID header auto-injected
```

#### `pkg/burst` — Burst Controller Per Tenant
Token bucket rate limiter with separate buckets per tenant.

```go
ctrl := burst.New(10, 5) // 10 req/sec, burst of 5

ctx := context.WithTenantID(context.Background(), "acme")
if ctrl.Allow(ctx) {
	// handle request
}
```

## License

MIT
