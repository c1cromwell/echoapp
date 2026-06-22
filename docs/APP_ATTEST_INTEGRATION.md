# App Attest Device Integrity — Integration

## Why

`X-Device-Info` was plaintext-trusted: a client could send `jailbroken: false` or any
device claim and the server believed it (`ValidateDeviceInfo` only checked field
presence). Apple **App Attest** replaces that trust with hardware-backed proof — the
device's attestation key is generated in the Secure Enclave and certified by Apple, and
each request can carry an assertion signed by that key.

## What shipped (server-side, tested)

`internal/auth/appattest` — a complete, unit-tested verifier:

- **`Verifier.VerifyAttestation(attestation, challenge, keyID)`** runs Apple's checks:
  CBOR-decode → verify the `x5c` chain (credCert → intermediate → trusted root) →
  recompute `nonce = SHA256(authData ‖ SHA256(challenge))` and match it against the
  credCert's Apple nonce extension (OID `1.2.840.113635.100.8.2`) → confirm
  `keyID == SHA256(pubkey)` → check the app-ID hash in `authData`. Returns the attested
  P-256 public key.
- **`Verifier.VerifyAssertion(assertion, clientData, pub, lastSignCount)`** verifies the
  per-request signature over `SHA256(authenticatorData ‖ SHA256(clientData))` and enforces
  a **strictly increasing sign count** (clone/replay protection). Returns the new count.
- **`KeyStore`** (`MemoryKeyStore` provided) maps `keyID → {publicKey, signCount}`.
- **Injectable trust root**: `NewVerifier(appID, pool)` / `NewVerifierFromPEM(appID, pem)`.
  Production pins Apple's *App Attest Root CA* from an operator-managed PEM (env / mounted
  secret) — never hardcoded. Tests inject a synthetic root and synthesize a full
  attestation + assertion, so the entire chain/nonce/assertion logic is verified in CI
  without real-device blobs (`appattest_test.go`).

## Integration plan (gated until the iOS client ships)

1. **Enrollment** (new handler): client calls `DCAppAttestService.attestKey` and POSTs
   `{keyID, attestationObject}` against a server-issued challenge. Server runs
   `VerifyAttestation`, then `KeyStore.Put(keyID, {pubkey, signCount})` bound to the
   device/DID. Back the store with Postgres (mirror `MemoryKeyStore`).
2. **Per-request middleware**: client attaches `{keyID, assertion}` (assertion over the
   request's canonical bytes — reuse the existing signing canonical). Middleware loads the
   key, runs `VerifyAssertion`, persists the new sign count, and sets a *verified*
   integrity flag that supersedes the self-reported `X-Device-Info`.
3. **Config gating**: `APP_ATTEST_REQUIRED` (off by default). While off, assertions are
   verified opportunistically and recorded; flipping it on makes a valid assertion
   mandatory for sensitive endpoints — only after the iOS client is shipping attestations.

## Remaining production requirements

- **Apple App Attest Root CA PEM** provided to the server (operator secret).
- **iOS client**: `DCAppAttestService` key generation + attestation at enrollment and
  assertion generation per request (the canonical it signs must match the server's).
- **DB-backed `KeyStore`** (sign count must persist and be updated atomically).
- **Real-device E2E**: App Attest only runs on physical devices (not the Simulator), so
  the final gate is a device run — the server logic above is already CI-verified.
