# ADR-001: Use Generics for Type-Safe Multi-Tenant Resources

## Status

Accepted

## Context

Multi-tenant SDKs often manage per-tenant instances of arbitrary types: database connections, Redis clients, API clients, configuration structs, etc. Using `any` everywhere works but sacrifices compile-time type safety and requires extensive type assertions at call sites.

Since Go 1.18, generics provide a way to write type-safe abstractions without sacrificing flexibility.

## Decision

Use Go generics (`[T any]`) for the core multi-tenant resource managers:

- `ConfigMgr[T]` — holds per-tenant configs of any type
- `SDKMgr[S, C]` — manages per-tenant SDK instances with a factory function
- `SDKMgrE[S, C]` — variant with error handling during initialization
- `SDKMgrWMet[S, C, M]` — variant with metrics injection

Concrete types are created at call sites:

```go
dbMgr := mt.NewSDKMgr[string, *sql.DB](configs, func(ctx context.Context, connStr string) *sql.DB {
    db, _ := sql.Open("postgres", connStr)
    return db
})
db := dbMgr.Get(ctx) // *sql.DB, no type assertion needed
```

## Consequences

**Positive:**
- Full compile-time type safety for multi-tenant resources
- No type assertions at call sites
- IDE autocompletion works correctly
- Zero runtime overhead from interface boxing

**Negative:**
- Generic type signatures can be verbose (`SDKMgrWMet[S, C, M]`)
- Some developers may find generics harder to read
- Cannot mix different types in a single manager (by design)
