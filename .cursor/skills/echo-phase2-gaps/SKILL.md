---
name: echo-phase2-gaps
description: >-
  Phase 2 onboarding, credentials, contacts, and VC gaps for Echo. Use when
  implementing WO-14, WO-100, WO-287, PSI discovery, OIDC4VC iOS, or when
  tempted to implement Cardano DID WOs 180/182.
---

# Echo Phase 2 gaps

**Source of truth:** [`docs/PHASE2_GAP_AUDIT.md`](../../docs/PHASE2_GAP_AUDIT.md) (2026-05-24).

Architecture: **`did:key` + Constellation Identity Metagraph**. **No Cardano in Phase 1–2** ([ADR-0001](../../docs/adr/0001-phase1-identity-method.md)).

## Stop — obsolete WOs

| WO | Do NOT implement | Use instead |
|----|------------------|-------------|
| **180** | Atala PRISM / Cardano DID | **WO-273** → `pkg/did/` ✅ |
| **182** | Cardano VC | **WO-274** → `pkg/credentials/` ✅ |

## Already built (do not rebuild)

| Capability | Location |
|------------|----------|
| did:key register/resolve | `pkg/did/`, `pkg/didkey/` |
| VC 2.0 + StatusList2021 | `pkg/credentials/` |
| OIDC4VC backend | `pkg/credentials/oidc4vc/` (gated `OIDC4VC_ENABLED`) |
| Passkey signed REST | `internal/api/passkey_auth.go` |
| BIP-39 recovery UI | `ios/.../Onboarding/Recovery/` (WO-234 ✅) |
| Glacial first-run | `FirstRunCoordinator` (WO-292 ✅) |

**Frozen UX:** Do not redesign iOS onboarding/login from the React prototype. WO-203/204 (universal phone-first orchestration) are **backend/product backlog**, not a mandate to replace `FirstRunCoordinator` UI.

## Real backend gaps (thin integration)

| Gap | Evidence | WO |
|-----|----------|-----|
| Refresh/revoke HTTP routes | Logic in `internal/auth/token.go`; no `/v3/auth/refresh` | **287** |
| Refresh tokens in-memory | Not durable Postgres | **287** |
| Username check unreachable | `GET /v1/users/check-username` not routed | **14** |
| PSI discovery stub | `contacts/service.go` PSIDiscovery | **220** |
| Enrollment stubs | `handleEnrollmentVC/IDV/mDL` → `{"status":"ok"}` | **199** |
| Universal phone-first orchestrator (backend) | WO-203 service — **not** iOS flow replacement | **203** |

## Real iOS gaps

| WO | Gap |
|----|-----|
| **100** | No OIDC4VC client — `RegisterWithVerifiableCredentialUseCase`, wallet flow |
| **14** | Named WO-14 views vs Phase 1 `FirstRunCoordinator` — credential path backlog |
| **221** | No PSI/Argon2id client |
| **39** / **187** | Contact system use-cases not wired to real backend |
| **228** | Privacy settings screen incomplete |

## Before closing a Phase 2 WO

1. Read gap audit row for that WO — status may be Partial/Stub, not Missing.
2. Grep codebase for existing implementation (avoid duplicate engines).
3. Confirm not superseded by WO-273/274.
4. Sync status via skill **`echo-work-order-sync`** + Software Factory MCP.

## Consolidation traps

- Two trust/verification engines: `pkg/credentials/verifier.go` vs `internal/services/onboarding/credentials.go` — prefer real ECDSA path.
- Two discovery impls: `contacts/service.go` vs `trustnet/discovery.go` — consolidate before shipping WO-222.

## Active work (SF status)

- **WO-100** — in_progress (OIDC4VC iOS)
- **WO-14** — backlog (credential-path onboarding after Phase 1)
- **WO-287** — completed (verify routes/storage match audit before assuming done)

Re-sync SF if audit and tickets diverge.

## Related

- `docs/phase-2-work-orders.md`
- Skill: `echo-auth-contracts`, `echo-work-order-sync`
