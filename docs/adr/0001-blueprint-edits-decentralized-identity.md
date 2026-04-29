# Blueprint Edits — Decentralized Identity and Authentication

**Companion to:** `docs/adr/0001-phase1-identity-method.md`
**Target blueprint:** Decentralized Identity and Authentication (`282795a9-7bbd-41b3-a1db-13f54d05ff03`)
**Purpose:** Apply the Phase-1 `did:key` decision to the blueprint document.

The MCP server does not expose `edit_blueprint`, so these edits are
hand-applied. Each edit below shows **anchor text** to locate the section,
followed by **REPLACE WITH** content. Apply in order, top to bottom. After
applying all edits, the blueprint should contain zero `did:prism:cardano:`
or "Atala PRISM" references in the canonical (first) copy.

---

## Pre-flight: Delete the duplicate second copy

The blueprint contains the entire body twice. The second copy starts after
the first `## Security Principles` block with another:

```markdown
# Decentralized Identity and Authentication

## Overview

User authentication and identity verification form the foundation of the Echo messaging app. The system combines device-based authentication using iOS Secure Enclave passkeys with optional high-assurance identity verification through third-party services and verifiable credentials stored on the Cardano blockchain.
```

**Action:** delete from that second `# Decentralized Identity and Authentication`
heading through the end of the document. The first copy is the canonical version
and is the only one that needs editing below.

---

## Edit 1 — Section: "Decentralized Identifier (DID) Management" (intro paragraph)

**Find:**

> The system uses Decentralized Identifiers (DIDs) as the foundation for self-sovereign identity, enabling users to maintain complete control over their identity data while establishing verifiable credentials on the Cardano blockchain. DIDs are created using the Atala PRISM infrastructure, which implements the W3C DID specification and KERI standards for interoperability.

**Replace with:**

> The system uses Decentralized Identifiers (DIDs) as the foundation for self-sovereign identity, enabling users to maintain complete control over their identity data. ECHO uses the **W3C [`did:key`](https://w3c-ccg.github.io/did-method-key/) method** for the user identity primitive: the DID is derived deterministically from a P-256 key pair generated in the iOS Secure Enclave. Resolution is purely local — the public key is embedded in the DID identifier itself — so no network round-trip and no chain transaction are required to create or resolve a DID.
>
> Verifiable Credentials (trust tier commitments, KYC, professional credentials) are issued and anchored on the **Constellation Identity Metagraph** as W3C VC 2.0 records, with revocation handled via StatusList2021 entries on the same metagraph. There is no Cardano dependency in Phase 1–2. Cardano and Midnight remain Phase 3+ evaluation candidates for ZK proof circuits only — see ADR-0001 for the decision record.

---

## Edit 2 — Subsection: "DID Creation and Storage" (intro paragraph)

**Find:**

> When a user creates an account, the system generates a unique DID anchored to the Cardano blockchain. The DID follows the format `did:prism:cardano:<unique-identifier>` and serves as the user's immutable identity anchor across the platform and potentially other applications.

**Replace with:**

> When a user creates an account, the system derives a `did:key` locally from the iOS Secure Enclave's P-256 key pair. The DID follows the format `did:key:z<base58btc-of-multicodec-prefixed-public-key>` (multicodec `0x1200` for P-256). It is derived once at account creation and re-derivable on demand from the same key pair. Because the DID *is* the public key, no chain transaction or network call is required, and resolution is deterministic and offline-capable.

---

## Edit 3 — Subsection: "DID Creation Process" (numbered list)

**Find:**

> 1. User completes initial onboarding with username and passkey
> 2. Go backend generates a new DID using Atala PRISM infrastructure
> 3. DID is anchored to Cardano blockchain through a transaction that records the DID document
> 4. DID document includes the user's public key, verification methods, and service endpoints
> 5. DID is stored locally on the iOS device in the Secure Enclave alongside the passkey
> 6. Backend maintains a mapping between the user's DID and their account for quick lookup

**Replace with:**

> 1. User completes initial onboarding with username and passkey
> 2. iOS app generates a P-256 key pair in the Secure Enclave (private key never leaves the device)
> 3. iOS app derives the `did:key` deterministically from the P-256 public key (no chain transaction)
> 4. iOS app submits `POST /identity/register` to the Go backend with `{ did, public_key_hex }`
> 5. Backend verifies the supplied DID matches the canonical derivation from `public_key_hex` and persists the `(did, public_key, registered_at)` binding
> 6. The DID document is computed on demand from the registered public key — there is no on-chain DID document; service endpoints (e.g. `MessagingService`) are returned by the resolver as a synthetic document built at request time

---

## Edit 4 — Subsection: "DID Document Structure" (JSON example)

**Find** the entire fenced JSON block beginning with:

```plaintext
{
  "@context": "https://www.w3.org/ns/did/v1",
  "id": "did:prism:cardano:abc123def456",
  ...
}
```

**Replace with:**

```plaintext
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/multikey/v1"
  ],
  "id": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R",
  "verificationMethod": [
    {
      "id": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R#z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R",
      "type": "Multikey",
      "controller": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R",
      "publicKeyMultibase": "z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R"
    }
  ],
  "authentication": [
    "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R#z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R"
  ],
  "assertionMethod": [
    "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R#z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R"
  ],
  "service": [
    {
      "id": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R#messaging",
      "type": "MessagingService",
      "serviceEndpoint": "https://backend.echo.app/messages"
    }
  ]
}
```

> Note: Verifiable credentials are no longer embedded in the DID document.
> They are stored on the **Constellation Identity Metagraph** and queried at
> request time using the holder's `did:key` as the lookup key. This keeps
> the DID document static and key-derivable while making credential
> issuance/revocation independently mutable.

---

## Edit 5 — Subsection: "DID Resolution and Verification"

**Find:**

> When the backend receives a request from a user, it resolves their DID to verify their identity and retrieve their public key for signature validation. The DID resolution process queries the Cardano blockchain to retrieve the authoritative DID document.
>
> **DID Resolution Flow**:
>
> 1. Backend receives authenticated request with user's DID
> 2. Backend queries Cardano blockchain for the DID document
> 3. DID document is retrieved and cached locally for performance
> 4. Public key is extracted from the DID document
> 5. Request signature is validated against the public key
> 6. If valid, request is processed; if invalid, request is rejected
>
> The backend maintains a local cache of recently resolved DIDs to reduce blockchain queries. Cache entries are invalidated after 24 hours or when the user updates their DID document.

**Replace with:**

> When the backend receives a request from a user, it resolves their DID to verify their identity and retrieve their public key for signature validation. Because `did:key` embeds the public key in the identifier itself, resolution is **purely deterministic and offline** — no blockchain query is required.
>
> **DID Resolution Flow**:
>
> 1. Backend receives authenticated request with user's DID
> 2. Backend extracts the multibase-encoded public key from the DID string and decodes it (`did:key` parser library)
> 3. Public key is used directly for signature validation against the request
> 4. If valid, request is processed; if invalid, request is rejected
> 5. (Optional) Backend looks up trust-tier VCs on the Identity Metagraph for the resolved DID, with caching and circuit-breaker fallback to last-known-good per the Backend blueprint's "Circuit Breakers Per Chain" rules
>
> Because resolution is local, no blockchain cache is needed for the DID document itself. The Identity Metagraph VC cache is the only network-bound cache and follows the standard Identity Metagraph circuit-breaker policy (5-failure threshold, 30s reset).

---

## Edit 6 — Subsection: "Multi-Device DID Support"

**Find:**

> Users can register multiple devices with their DID, with each device maintaining a separate passkey in its Secure Enclave. The DID document includes multiple public keys, one for each registered device, enabling the user to authenticate from any registered device.
>
> **Multi-Device Registration**:
>
> 1. User authenticates on primary device with passkey
> 2. User initiates device registration on secondary device
> 3. Primary device displays QR code containing registration token
> 4. Secondary device scans QR code and generates new passkey
> 5. Secondary device submits registration request with new public key
> 6. Backend verifies request is from authenticated user
> 7. New public key is added to DID document on Cardano
> 8. Secondary device can now authenticate independently

**Replace with:**

> Because `did:key` binds the DID to a single public key, multi-device support uses a **controller pattern** rather than DID rotation. The user's primary device DID is the *controller*, and each additional device gets its own `did:key` that is authorized by a signed device-attestation credential issued by the controller. The Identity Metagraph stores the device-attestation set so the backend can resolve "all DIDs controlled by user X" without exposing the device list to the network at signing time.
>
> **Multi-Device Registration**:
>
> 1. User authenticates on primary device with its `did:key` (the controller)
> 2. User initiates device registration on the secondary device
> 3. Primary device displays a QR code containing a one-time registration nonce signed by the controller key
> 4. Secondary device scans the QR code, generates its own P-256 key pair in its Secure Enclave, and derives a new `did:key`
> 5. Secondary device submits a registration request with `{ controller_did, device_did, nonce, controller_signature }` to the backend
> 6. Backend verifies the controller signature, the nonce freshness, and the public-key-to-DID derivation for the new device
> 7. Backend issues a `DeviceAttestationCredential` (W3C VC 2.0) signed by the platform issuer DID and submits it to the Identity Metagraph
> 8. Secondary device can now authenticate independently using its own `did:key`; revocation is handled by adding the device DID to a per-controller StatusList2021 entry on the Identity Metagraph

---

## Edit 7 — Subsection: "Trust Score Tiers and Feature Access" (intro line)

**Find:**

> Trust scores map to 5 tiers (Tier 1–5). Tier commitments are stored on Cardano; raw scores are never on-chain.

**Replace with:**

> Trust scores map to 5 tiers (Tier 1–5). Tier commitments (`H(score || nonce)`) are stored on the Constellation Identity Metagraph; raw scores are never on-chain.

---

## Edit 8 — Subsection: "Credential Types" → "Proof of Humanity Credential" (JSON example)

**Find** the JSON block whose `issuer` is `"did:prism:cardano:prove-issuer"`.

**Replace with:**

```plaintext
{
  "@context": [
    "https://www.w3.org/ns/credentials/v2",
    "https://w3id.org/security/multikey/v1"
  ],
  "type": ["VerifiableCredential", "ProofOfHumanity"],
  "issuer": "did:key:z6MkrJVnaZkeFzdQyMZu1cgjg7k1pZZ6pvBQ7XJPt4swbTQ2",
  "validFrom": "2026-01-15T10:30:00Z",
  "validUntil": "2027-01-15T10:30:00Z",
  "credentialSubject": {
    "id": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R",
    "humanityProof": true,
    "verificationMethod": "liveness_check"
  },
  "credentialStatus": {
    "id": "https://identity-metagraph.echo.app/status/0#42",
    "type": "StatusList2021Entry",
    "statusPurpose": "revocation",
    "statusListIndex": "42",
    "statusListCredential": "https://identity-metagraph.echo.app/status/0"
  },
  "proof": {
    "type": "DataIntegrityProof",
    "cryptosuite": "ecdsa-2019",
    "created": "2026-01-15T10:30:00Z",
    "verificationMethod": "did:key:z6MkrJVnaZkeFzdQyMZu1cgjg7k1pZZ6pvBQ7XJPt4swbTQ2#z6MkrJVnaZkeFzdQyMZu1cgjg7k1pZZ6pvBQ7XJPt4swbTQ2",
    "proofPurpose": "assertionMethod",
    "proofValue": "<signature>"
  }
}
```

> Note: this example moves to **W3C VC 2.0** (`v2` context, `validFrom`/`validUntil`, `DataIntegrityProof` with `ecdsa-2019` cryptosuite, `StatusList2021Entry` for revocation). Same pattern applies to the KYC-Lite, High-Assurance, and Professional credential examples — issuer DIDs and subject DIDs become `did:key:…`, expiration uses `validUntil`, revocation uses `StatusList2021Entry` rather than embedding inside the DID document.

---

## Edit 9 — Subsection: "Credential Storage and Revocation"

**Find:**

> Verifiable credentials are stored on the Cardano blockchain as part of the user's DID document. This ensures credentials are immutable and portable across applications. Credentials can be revoked by the issuer if the user's status changes (e.g., professional certification expires).
>
> **Credential Revocation Process**:
>
> 1. Issuer determines credential should be revoked
> 2. Issuer submits revocation transaction to Cardano
> 3. Revocation is recorded in the credential revocation registry
> 4. Backend queries revocation registry during credential verification
> 5. If credential is revoked, it is no longer considered valid
> 6. User's trust score is recalculated without the revoked credential

**Replace with:**

> Verifiable credentials are issued by the platform issuer DID and anchored on the **Constellation Identity Metagraph**. Each issued credential includes a `credentialStatus` field pointing to a StatusList2021 credential maintained by the issuer. The status list itself is a single VC anchored on the Identity Metagraph and updated by the issuer when revocations occur.
>
> **Credential Revocation Process**:
>
> 1. Issuer determines a credential should be revoked
> 2. Issuer flips the corresponding bit in its StatusList2021 bitstring
> 3. Issuer submits the updated StatusList2021 credential as an Identity Metagraph update; consensus finalizes the new status (< 30s target)
> 4. Backend queries the StatusList2021 credential during verification (cached with circuit-breaker fallback)
> 5. If the bit is set, the credential is no longer considered valid
> 6. User's trust score is recalculated without the revoked credential

---

## Edit 10 — Section: "Zero-Knowledge Proof Integration" → "Phase 1–2: Standard Credential Verification"

**Find:**

> Credentials are verified directly via Cardano. The backend resolves the user's DID, checks credential status bits, and validates the trust tier UTXO datum. Trust tier is confirmed, but the verification method and issuer are visible to the backend.

**Replace with:**

> Credentials are verified directly via the **Constellation Identity Metagraph**. The backend resolves the user's `did:key` locally (no network call), then queries the Identity Metagraph for the corresponding trust-tier VC and its StatusList2021 entry. Trust tier is confirmed; the verification method and issuer are visible to the backend (this is acceptable in Phase 1–2 because no end-user privacy promise is made for the issuer field at this stage).

---

## Edit 11 — Section: "Zero-Knowledge Proof Integration" → final paragraph (right after the ZK Use Cases table)

**Find:**

> Midnight uses Compact (TypeScript DSL) for contracts — not Scala. The Scala requirement applies only to Constellation metagraph L1 validation. DID registration and credential issuance remain on Cardano permanently.

**Replace with:**

> Midnight uses Compact (TypeScript DSL) for contracts — not Scala. The Scala requirement applies only to Constellation metagraph L1 validation. **DID derivation is local (`did:key`); credential issuance is anchored on the Constellation Identity Metagraph. There is no permanent Cardano dependency.** Cardano and Midnight remain Phase 3+ evaluation candidates exclusively for ZK circuits, and only if their privacy / interop properties cannot be achieved natively on the Identity Metagraph.

---

## Edit 12 — Section: "Component Breakdown" → "Streamlined Onboarding with Verifiable Credentials" (Key Features list)

**Find:**

> * DID generation and Cardano anchoring

**Replace with:**

> * `did:key` derivation from Secure Enclave P-256 key (local, no chain transaction)
> * `POST /identity/register` binding the DID to the user account

---

## Edit 13 — Section: "Component Breakdown" → "In-App High-Assurance Identity Verification" (Key Features list)

**Find:**

> * Verifiable credential issuance on Cardano

**Replace with:**

> * Verifiable credential issuance and anchoring on the Constellation Identity Metagraph

---

## Edit 14 — Section: "Component Breakdown" → "Device Passkey Management" (Key Features list)

**Find:**

> * DID document updates for multi-device registration

**Replace with:**

> * Device-attestation credentials issued on the Identity Metagraph for multi-device registration (controller-pattern; no DID document mutation since `did:key` is immutable)

---

## Edit 15 — Subsection: "ECHO Reward Coordination" (Reward Transaction Structure JSON)

**Find** the JSON block:

```plaintext
{
  "transaction_type": "verification_reward",
  "user_did": "did:prism:cardano:abc123def456",
  ...
  "issuer_did": "did:prism:cardano:verification-service",
  ...
}
```

**Replace with:**

```plaintext
{
  "transaction_type": "verification_reward",
  "user_did": "did:key:z2DA8x9XyAEWfJUg5FctK46tFZ6oEJ4nBM5Cv6fMA8DGo7R",
  "reward_amount": 100000000000000000,
  "verification_type": "high_assurance",
  "verification_timestamp": "2026-01-15T10:30:00Z",
  "issuer_did": "did:key:z6MkrJVnaZkeFzdQyMZu1cgjg7k1pZZ6pvBQ7XJPt4swbTQ2",
  "signature": "<backend-signature>",
  "nonce": 12345
}
```

---

## Edit 16 — Section: "Security Principles"

**Find:**

> * Trust levels are immutably recorded on Cardano and referenced for access control

**Replace with:**

> * Trust-tier commitments (`H(score || nonce)`) are anchored on the Constellation Identity Metagraph and referenced for access control; the raw score never leaves the user's device or the backend's trust service
> * `did:key` is permanent and key-derived: rotation is performed by replacing the controller relationship in the device-attestation credential set, not by mutating an on-chain DID document

---

## Verification checklist

After applying all edits, verify the blueprint:

- [ ] Search for `did:prism` returns zero matches
- [ ] Search for `Atala PRISM` returns zero matches
- [ ] Search for "Cardano blockchain" returns zero matches in DID/credential context (mentions in ZK Phase-3+ evaluation context are acceptable)
- [ ] The duplicate second copy is removed (the blueprint should have exactly one `# Decentralized Identity and Authentication` heading)
- [ ] References to "Constellation Identity Metagraph" appear in the DID Management, Credentials, ZK, and Security Principles sections
- [ ] The `Updated: <date>` metadata reflects today's edit pass

## Cross-references

After this blueprint update, the following sibling blueprints **also reference `did:prism:cardano:`** and should get a follow-up review (out of scope for WO-275, file a small fix WO if you want):

- `Privacy-Preserving Contact Discovery` (1 ref, line 110)
- `Privacy-Preserving Blockchain Data Model` (1 ref, line 163)
- `Zero-Knowledge Proofs and Midnight Integration` (1 ref, line 133 — `approvedIssuers`)
- `Portable Social Graph and Protocol Layer` (3 refs, lines 13/46/52)
- `Post-Quantum Cryptography Mode` (1 ref, line 96)
- `Data Sovereignty Layer` (1 ref, line 35)
- `Universal Onboarding and Identity Creation` (3 refs, lines 172/227/241/263)
- `Data Layer` (1 ref, line 166 — Cardano Atala PRISM cost-table row)

These can each be addressed via a one-line search-and-replace; the
substantive doctrine change happens only in the Decentralized Identity
blueprint above.
