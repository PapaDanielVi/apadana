# ADR-003: JWT Tenant Resolution Requires Explicit Verification

## Status

Accepted

## Context

A common way to carry tenant identity is a claim inside a JWT bearer token.
The v1 resolver exposed `FromJWTClaim`, which decoded the JWT payload and read
the claim without verifying the signature. The doc comment said to use it only
when an upstream layer had already validated the token, but nothing in the API
enforced that. A caller who reached for the obvious-looking function got a
tenant ID straight from an attacker-controlled token: forge a JWT, set any
tenant, and the request runs as that tenant. For a multi-tenant library this is
a cross-tenant access vulnerability waiting to happen.

## Decision

Split the two behaviors and make the safe one the obvious choice.

- `FromVerifiedJWT(claimName, verify)` takes a caller-supplied `Verifier`
  (`func(token string) (map[string]any, error)`). The token is verified before
  the claim is read, so it is safe on untrusted requests. The verifier is kept
  dependency-free so callers plug in `golang-jwt`, a JWKS client, or whatever
  they already use.
- `FromUnsafeJWTClaim(claimName)` keeps the unverified behavior for the genuine
  "already validated upstream" case, but the name now states the risk and the
  doc comment spells out the consequence.

The old `FromJWTClaim` name is removed in v2. There is no signature-checking
code in the library itself; verification is delegated so the library does not
ship or pin a crypto/JWT implementation.

## Alternatives Considered

1. **Keep `FromJWTClaim` and rely on docs.** Rejected. The dangerous path had
   the friendliest name, which is exactly backwards.
2. **Bundle a JWT verifier (e.g. golang-jwt) in the library.** Rejected. It
   would pin a JWT library and key-management opinions onto every consumer.
3. **Drop JWT support entirely.** Rejected. The claim case is common; a safe,
   pluggable resolver is more useful than no resolver.

## Consequences

**Positive:**
- The safe resolver is the one with the plain name; misuse now requires typing
  "Unsafe".
- No JWT/crypto dependency is forced on consumers.
- Verification strategy (HMAC, RSA, JWKS rotation) is the caller's to choose.

**Negative:**
- Breaking change: `FromJWTClaim` callers must move to `FromVerifiedJWT` (and
  supply a verifier) or `FromUnsafeJWTClaim`. This lands in the v2 bump.
- Callers must wire their own verifier, which is slightly more setup than the
  one-liner it replaces.
