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

### Go idioms
- Use `any` instead of `interface{}` (Go 1.18+)
- Maps are not comparable - use pointers (`new(int)`) in tests for equality checks
- `new(struct{})` returns same pointer for zero-sized types - use `new(int)` for unique pointers

### Dependencies
- Run `go mod tidy` after adding dependencies
- Missing go.sum entries cause build failures - `go mod tidy` fixes them
- Some deps need transitive deps manually added (prometheus)

### Testing
- Unused imports cause build failures in test files too
- Use `httptest.NewServer` and `httptest.NewRequest` for HTTP tests
- Skip integration tests when servers aren't available: `t.Skip("no server")`

### Lint rules (golangci-lint v2)
- `paralleltest`: Add `t.Parallel()` to tests
- `testpackage`: Use `package foo_test` not `package foo`
- `canonicalheader`: HTTP headers use canonical form (`X-Tenant-Id` not `X-Tenant-ID`)
- `gochecknoglobals`: Avoid package-level globals (use `sync.Map` or init functions)
- `modernize`: Use modern Go constructs (`any` over `interface{}`)

### Multi-tenancy patterns
- All packages extract tenant ID from `context.Context` using `pkg/context`
- Header injection pattern: `X-Tenant-ID` for HTTP, RabbitMQ, Kafka, NATS
- Per-tenant singletons: Use `sync.Map` with `LoadOrStore` for thread safety
