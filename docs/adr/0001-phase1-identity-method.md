# ADR 0001 — Phase 1 Identity Method: `did:key`

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Platform / Identity
- **Related work order:** WO-275
- **Blocks / unblocks:** WO-230, WO-272, WO-276
- **Supersedes:** N/A

## Context

WO-230 ("Set Up Phase 1 Local Development Testnet Infrastructure") mandates
that Phase 1 derive DIDs locally from the Secure Enclave key pair with **no
Cardano dependency** and **no chain transaction** — i.e., the W3C
[`did:key`](https://w3c-ccg.github.io/did-method-key/) method.

The codebase as of this ADR ships a **Cardano + Atala PRISM** identity stack
that contradicts that requirement:

| Layer | Path | Status |
|---|---|---|
| Cmd entry points | `cmd/cardanoidentity/main.go`, `cmd/credentials/main.go` | Cardano/PRISM |
| Cardano integration | `pkg/cardano/*` (9 files) | Cardano/PRISM |
| DID layer | `pkg/did/atala_client.go`, parts of `pkg/did/service.go` | Atala PRISM |
| Credentials | `pkg/credentials/*` (17 files) | did:prism issuance |
| API layer | `pkg/api/handlers/{issuer,schema,credential,transaction}.go` | Cardano-coupled |
| Backend wiring | `internal/auth/service.go`, `internal/api/v3_handlers.go`, `internal/api/enrollment_handlers.go` | did:prism strings |
| Scripts | `scripts/setup-cardano-testnet.sh` | Cardano testnet bootstrap |
| Compiled artifacts | `./cardanoidentity` (12 MB), `./credentials` (12 MB) | Stale binaries in repo root |
| iOS | `ios/Echo/Sources/Features/Onboarding/Recovery/RecoveryService.swift`, `ios/Echo/Tests/Mocks/MockServices.swift` | did:prism string refs |
| OpenAPI | `openapi.yaml` | did:prism examples |
| Placeholder | `pkg/identity/handlers.go` `createIdentity` returns hard-coded `did:example:123` | Not real |

A small amount of `did:key` plumbing already exists in
`internal/services/onboarding/trust_registry.go` and its tests, indicating
prior intent to use `did:key` that was never completed.

This contradiction blocks every Phase-1 build WO that touches identity, auth,
or credentials. Engineers are adding to whichever stack they touched last,
deepening the divergence.

## Decision

**Adopt W3C `did:key` as the sole Phase-1 DID method.** Deprecate and remove
all Cardano / Atala PRISM code paths from the Phase-1 build. Reconsider
chain-anchored identity (or any non-`did:key` method) at the Phase 2/3
boundary as a separate ADR.

### Concretely

1. **DID derivation:** P-256 (`secp256r1` / `prime256v1`) public key →
   compressed 33-byte SEC1 encoding → multicodec prefix `0x1200`
   (P-256 public key) → multibase `base58btc` (`z` prefix) →
   `did:key:z…`. The private key never leaves the Secure Enclave on iOS,
   or `internal/crypto` on backend test fixtures.

2. **Resolution:** purely deterministic — the DID *is* the public key.
   No network call, no on-chain anchor, no Atala PRISM dependency.
   Resolution lives in a new `pkg/did/keymethod` package or replaces the
   current `pkg/did/resolver.go`.

3. **VC issuance & revocation:** owned by WO-272 (Identity Metagraph).
   Issuance uses VCs **signed by an `did:key` issuer**; revocation uses
   StatusList2021 anchored on the Constellation Identity Metagraph.
   The Cardano metadata-anchored revocation list is removed.

4. **Backend `POST /identity/register`:** accepts `{did, public_key_hex}`,
   verifies the DID matches the canonical derivation from `public_key_hex`,
   persists the binding, returns `201 Created`. No chain transaction, no
   Atala PRISM call.

5. **Compiled binaries removed from repo:** `./cardanoidentity`,
   `./credentials` — these are leftover build outputs that should never have
   been committed.

## Consequences

### Positive

- Aligns Phase-1 reality with WO-230's "no Cardano dependency" guarantee.
- Eliminates an entire external dependency (Atala PRISM service + Cardano node)
  from the Phase-1 development environment, simplifying onboarding and
  shrinking the testnet attack surface.
- Removes ~30 files and two compiled binaries from the repo, reducing carrying
  cost for code that no one is shipping.
- WO-230 go/no-go Step 1 ("Derive `did:key` locally from a P-256 key pair")
  becomes implementable without a strategy detour.
- WO-272 (VC issuance) gets an unambiguous issuer-DID method to design against.

### Negative

- Loses the on-chain anchor property of `did:prism`. Mitigated because:
  - Phase 1 explicitly does not require chain-anchored identity (per blueprint).
  - VC issuance integrity comes from the Identity Metagraph in WO-272.
  - Trust-tier commitments and Digital Evidence fingerprints continue to
    anchor on Data L1 — they don't need a chain-anchored *DID* to do so.
- `did:key` cannot be rotated (the DID *is* the key). Multi-device support
  must use a wrapper (e.g., a controller DID document linking multiple
  device `did:key` identifiers) rather than DID rotation. Captured as
  scope for `pkg/did/multidevice.go` rewrite.
- Existing `did:prism`-bearing tests (43 occurrences across `internal/auth/*`
  and `internal/services/onboarding/*`) need updating.

### Neutral

- `pkg/credentials/oidc4vc/*` is largely method-agnostic and may be retained.
  The `oidc4vc/verifier.go` file does have a `did:prism` string reference
  that needs updating, but the OID4VC protocol itself is DID-method-neutral.
- The OpenAPI spec needs updates but the route shapes stay the same.

## Alternatives considered

### Option B — Keep `did:prism`

Rejected. Would require:
- Rewriting WO-230 to drop the "no Cardano dependency" line.
- Operating an Atala PRISM testnet locally (additional 2+ containers,
  Cardano node sync ~hours on first run).
- Reverting the existing `did:key` work in `trust_registry.go`.

The Phase-1 testnet's purpose is fast local iteration. PRISM's chain
dependency is incompatible with the < 30s finality go/no-go criterion.

### Option C — Dual-method (`did:key` + `did:prism` during Phase 1)

Rejected. Doubles the surface area, doubles the test matrix, doubles the
config complexity, and forces every downstream WO (WO-272, WO-276) to
support two issuer DID methods. The point of WO-275 is to *resolve* the
contradiction, not paper over it.

## Implementation Plan

Tracked as three follow-up work orders, each `blocked_by` this WO:

| WO | Title | Type | Priority |
|---|---|---|---|
| (created) | Implement `did:key` derivation library and `POST /identity/register` handler | build | urgent |
| (created) | Migrate `pkg/credentials` and `pkg/did` to `did:key`; remove Atala PRISM client | build | high |
| (created) | Remove Cardano/PRISM stack and update OpenAPI + iOS clients to `did:key` | fix | medium |

## Acceptance criteria for this ADR

- [x] Decision recorded under `docs/adr/0001-phase1-identity-method.md`
- [x] Inventory of impacted code paths captured (Context section)
- [x] Three follow-up migration/cleanup WOs created and linked as blockers
- [x] WO-230 description updated to reference `did:key` and this ADR
- [x] Decentralized Identity and Authentication blueprint edits drafted
      under `docs/adr/0001-blueprint-edits-decentralized-identity.md` for
      manual application *(software-factory-echo MCP does not expose
      `edit_blueprint`; the edits doc is ready to copy/paste into the
      blueprint document on the next blueprint-owner pass)*

## References

- [W3C `did:key` Method Specification (CCG)](https://w3c-ccg.github.io/did-method-key/)
- [Multicodec table (P-256 public key = `0x1200`)](https://github.com/multiformats/multicodec/blob/master/table.csv)
- WO-230 — Set Up Phase 1 Local Development Testnet Infrastructure
- WO-272 — VC issuance, trust tier commitments, StatusList2021 revocation (Identity Metagraph)
- WO-276 — Add Constellation Identity Metagraph Module Skeleton
