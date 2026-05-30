# ADR 0004 — Echo Passport Recovery: social/threshold (Shamir), no server-held key

- **Status:** Accepted
- **Date:** 2026-05-29
- **Deciders:** Product / Platform / Identity / Security
- **Related work order:** WO-292 (mini-design), WO-296 (implementation), WO-303 (hardening)
- **Blocks / unblocks:** WO-296, WO-297 (iOS), WO-294 (key hierarchy)
- **Supersedes:** N/A

## Context

ADR 0003 establishes that Echo Passport never holds a decryption key — the `PassportRootKey` is
derived on-device from the Secure Enclave key plus a recovery secret. That makes **recovery the
make-or-break UX problem**: there is no server backstop, so if the user loses their device the
recovery path is the *only* way back to their credentials.

Two naive options both fail:

- **Pure device-only / single recovery phrase.** Loses the user's data the day they drop their
  phone with no second factor; high real-world data-loss rate.
- **Server-assisted recovery (server holds or escrows a key).** Recreates the honeypot ADR 0003
  exists to prevent, and makes Echo a data controller of recoverable PII.

Echo already ships a partial recovery foundation in Phase 1: a 24-word BIP-39 phrase + SMS OTP
(only `H(phone)` stored on the backend), referenced by WO-234. The question is how to make a
non-custodial vault **feel as easy as a custodial one** without a server-held key.

## Decision

**Recover the Passport via social/threshold recovery: Shamir-split the recovery secret into N
shares with M-of-N reconstruction, distributed across the user's own devices and optional
guardians, layered on the existing BIP-39 + SMS flow. No share or key is ever reconstructable
server-side.**

### Scheme (locked)

- **Shamir M-of-N** split of the recovery secret (which gates `PassportRootKey`).
- **Default parameters (consumer):** **2-of-3** — one share on the primary device, one on a
  second enrolled device (or encrypted backup blob the user controls), one with an optional
  guardian. High-assurance / citizenship tiers may raise to **3-of-5** (WO-303 policy).
- **Field:** GF(256) via `github.com/codahale/shamir` (or equivalent audited library); 32-byte
  recovery secret; shares are 33 bytes each (standard Shamir over bytes).
- **Shareholders (guardian taxonomy):**
  - the user's own enrolled devices (Secure Enclave-held shares),
  - optional **trusted contacts** (other Echo users; a guardian's acceptance is itself a signed VC),
  - optional **Comply org** as an institutional guardian (for enterprise-issued credentials).
- **Layering:** the existing 24-word BIP-39 phrase remains a valid single-holder fallback;
  social recovery is the higher-availability path on top of it. SMS OTP remains a liveness/notice
  channel, never a key holder.

### Server boundary
- The backend stores **share metadata only** (which guardian holds which share index, status) in
  `passport_recovery_share` — never reconstructable share material, never the secret. M-1 shares
  reveal nothing.
- All reconstruction happens client-side. **No server-held key under any path** — this is the
  honeypot line from ADR 0003.

### Threat model

| Threat | Mitigation |
|--------|------------|
| **Collusion** | M-of-N with guardian diversity (device + contact + optional org); default 2-of-3 requires two independent shareholders; optional guardian bonds in WO-303 |
| **Coercion / duress** | Single device share is below threshold; integrate duress-PIN (existing hidden-persona pattern) to surface decoy vault |
| **Total share loss** | BIP-39 phrase remains fallback; pre-loss UX prompts share distribution before high-value credentials |
| **Guardian churn** | Share rotation/re-issuance when guardian removed or device rotated (WO-303); stale share metadata marked `revoked` server-side |
| **Server compromise** | Server stores share **metadata only** (index, guardian DID, status) — never share bytes or recovery secret; M-1 shares reveal nothing |

### Guardian-acceptance VC (sketch)

When a trusted contact accepts guardianship, Echo issues a lightweight VC (native Echo issuer):

```json
{
  "@context": ["https://www.w3.org/ns/credentials/v2"],
  "type": ["VerifiableCredential", "EchoGuardianCredential"],
  "issuer": "did:key:z…(Echo Identity Service)",
  "credentialSubject": {
    "id": "did:key:z…(guardian)",
    "guardianFor": "did:key:z…(holder)",
    "shareIndex": 2,
    "acceptedAt": "2026-05-29T00:00:00Z"
  }
}
```

The guardian never receives the Shamir share until the holder explicitly distributes it (encrypted to the guardian's device key). The VC is proof of consent, not the share itself.

## Consequences

### Positive

- Non-custodial recovery that *feels* custodial-easy — losing one device is recoverable without a
  server key.
- Reuses the Phase-1 BIP-39 + SMS foundation rather than replacing it.
- Guardian-acceptance-as-VC reuses the credential pipeline (no bespoke protocol).

### Negative

- Social recovery has genuine UX complexity (explaining guardians, share distribution) and a
  collusion surface; mitigated by defaults + WO-303 hardening.
- Requires careful client-side crypto; an external audit of the Shamir implementation is scoped
  in WO-303.

### Neutral

- Default M-of-N parameters and whether Comply-org guardianship is enabled per credential tier are
  finalized in WO-292 before WO-296 implementation begins.

## Alternatives considered

### Option B — Single recovery phrase only (no social layer)
Rejected as the *primary* path (kept as fallback). Too high a real-world data-loss rate for a
product meant to hold government ID, payment, and legal credentials.

### Option C — Server-escrowed key (custodial recovery)
Rejected. Violates ADR 0003 and the T0–T7 invariant; makes Echo a data controller of recoverable
PII and a breach target.

### Option D — Passkey/iCloud Keychain sync as the recovery root
Rejected as the sole mechanism: ties recovery to a single platform vendor, weakens
decentralization, and doesn't cover cross-platform (web/glasses/agent) or institutional
guardianship. May be offered as *one* shareholder, not the root.

## Implementation status

- [x] Wave 0 design complete (ADR 0004 Accepted; default 2-of-3, guardian VC sketch, threat model)
- [x] WO-296: `pkg/passport/recovery/` Shamir split/combine, guardian-acceptance VC, `passport_recovery_share`
- [ ] WO-297: iOS recovery setup UI
- [ ] WO-303: guardian bonds, device rotation, third-party audit
