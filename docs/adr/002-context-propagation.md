# ADR-002: Context Propagation Strategy

## Status

Accepted

## Context

The tenant ID must flow through the entire request lifecycle: from HTTP/gRPC ingress, through business logic, into database queries, cache keys, message queue headers, and observability spans. How the tenant ID is stored and propagated affects API ergonomics, performance, and correctness.

## Decision

Use `context.Context` as the sole carrier for tenant IDs, with a private `contextKey` type to prevent collisions:

```go
type contextKey struct{}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
    return context.WithValue(ctx, contextKey{}, TenantID(tenantID))
}
```

All packages extract the tenant ID from `context.Context`:
- Middleware injects it on ingress
- Messaging packages inject it into outbound headers and extract from inbound headers
- Observability packages add it as a span attribute or log field
- `CloneCtx` preserves it when creating background contexts

## Alternatives Considered

1. **Thread-local storage** — Not idiomatic in Go, requires `sync.Pool` or similar hacks.
2. **Explicit parameter passing** — Would require threading `tenantID string` through every function signature, making the API verbose.
3. **Global variable** — Brittle for tests and concurrent workloads.

## Consequences

**Positive:**
- Idiomatic Go — `context.Context` is the standard for request-scoped values
- Composable — works with cancellation, deadlines, and OpenTelemetry
- No API pollution — tenant ID doesn't appear in function signatures
- Testable — easy to inject tenant ID in test contexts

**Negative:**
- Hidden dependency — tenant ID is not visible in function signatures
- Must use `context.Context` everywhere (already a Go best practice)
- Slight overhead from `context.WithValue` (stack-allocated wrapper)
