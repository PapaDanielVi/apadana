## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

## 5. Implementation Learnings

**Patterns and gotchas discovered during development.**

### Package naming
- `tenantctx` was too verbose, renamed to `tctx` for brevity
- Import alias convention: `tctx "github.com/PapaDanielVi/apadana/pkg/context"`
- New packages added: `pkg/grpc`, `pkg/resolver`

### Go idioms
- Use `any` instead of `interface{}` (Go 1.18+)
- Maps are not comparable - use pointers (`new(int)`) in tests for equality checks
- `new(struct{})` returns same pointer for zero-sized types - use `new(int)` for unique pointers
- `context.Context` cancellation: always `select` on `<-ctx.Done()` in consumer loops, return `ctx.Err()`
- `strings.Builder` has no `ReadFrom` method — use `io.ReadAll(reader)` instead
- `maps.Copy(dst, src)` replaces manual `for k, v := range` map copy loops (modernize linter)
- `base64.RawURLEncoding` for JWT payload decoding (not `StdEncoding`)

### Dependencies
- Run `go mod tidy` after adding dependencies
- Missing go.sum entries cause build failures - `go mod tidy` fixes them
- Some deps need transitive deps manually added (prometheus)
- `google.golang.org/grpc` is a direct dependency (added for `pkg/grpc`)

### Testing
- Unused imports cause build failures in test files too
- Use `httptest.NewServer` and `httptest.NewRequest` for HTTP tests
- Skip integration tests when servers aren't available: `t.Skip("no server")`
- `fatcontext` linter: don't store `context.Context` in variables outside function scope — use pointer pattern (`gotCtx := new(context.Context)` / `*gotCtx = ctx`)
- Test helper types (e.g. `eagerConfig`, `centralizedConfig`) must implement the right interfaces for the code path being tested
- `lazyInit` returns `true` by default when config doesn't implement `lazyIniter` — plain types are lazy by default
- `httptest.NewRequest` sets `Host` to `example.com` by default — account for subdomain resolver in tests

### Lint rules (golangci-lint v2)
- `paralleltest`: Add `t.Parallel()` to tests
- `testpackage`: Use `package foo_test` not `package foo`
- `canonicalheader`: HTTP headers use canonical form (`X-Tenant-Id` not `X-Tenant-ID`)
- `gochecknoglobals`: Avoid package-level globals (use `sync.Map` or init functions)
- `modernize`: Use modern Go constructs (`any` over `interface{}`, `maps.Copy` over manual loops)
- `sloglint`: Use `slog.String("key", val)` instead of raw `"key", val` pairs
- `gosec G114`: Use `&http.Server{ReadTimeout: ..., WriteTimeout: ...}` instead of `http.ListenAndServe`
- `embeddedstructfieldcheck`: Embedding types that contain `context.Context` triggers this — add `//nolint:embeddedstructfieldcheck` if intentional
- `fatcontext`: Don't store `context.Context` in closure-captured variables — use pointer indirection
- `golines`: Break long function calls across multiple lines
- `goimports`: `sync` belongs in stdlib group; `maps` is a stdlib package

### Multi-tenancy patterns
- All packages extract tenant ID from `context.Context` using `pkg/context`
- Header injection pattern: `X-Tenant-Id` for HTTP, RabbitMQ, Kafka, NATS
- gRPC metadata key: `x-tenant-id` (lowercase, as gRPC metadata is case-insensitive)
- Per-tenant singletons: Use `sync.Map` with `LoadOrStore` for thread safety
- Consumer loops: use `for { select { case <-ctx.Done(): ... case msg := <-ch: ... } }` pattern
- `pkg/resolver` provides chain-of-responsibility tenant resolution (header, query, subdomain, cookie, JWT)
- `pkg/grpc` provides unary/stream server/client interceptors for tenant metadata propagation
- `pkg/resolver.Registry` provides thread-safe tenant metadata caching with `sync.RWMutex`

### CI/CD
- Workflows split into separate files: `test.yml`, `lint.yml`, `security.yml`
- `security.yml` runs `govulncheck ./...` on a weekly schedule and on PRs
- Test matrix includes Go 1.24 and 1.26
- Badge URLs in README must match workflow file names
