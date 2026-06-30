# Wallet Real-Custody — M4 Runbook (enable + verify real funds)

Branch: `wallet-real-custody`. M0–M3 are committed and tested headlessly. M4 is
the live-node + manual milestone: it can only be completed against a Constellation
testnet, in Xcode, and with a security sign-off. This is the checklist.

## State entering M4 (done)
- **Bundle** (`ios/Echo/Resources/wallet-sdk/echo-wallet.bundle.js`): dag4 keygen,
  address derivation, message signing, and **brotli transaction signing** all run
  in JavaScriptCore. `EchoWallet.signTransaction(priv, pub, body)` returns the
  submittable `{value, proofs}`.
- **iOS**: `StargazerSigner` (loads the bundle, injects `process` /
  `crypto.getRandomValues` / `TextEncoder`), `WalletKeyStore` (Keychain custody +
  backup gate), `WalletProvisioner` (real address + `X-Wallet-Proof`), real
  `RecoveryCoordinator`.
- **Backend**: `DagProofVerifier` + `/v3/wallet/challenge`; the **stake handler
  forwards a client-signed `{value, proofs}`** (`StakeEchoSigned` →
  `MetagraphClient.SubmitSignedByType` → `/token-locks`) when
  `ECHO_WALLET_REAL_FUNDS=1` and a `signed` field is present. Backend never signs.
- Cross-validation (`internal/wallet/dag_crossvalidation_test.go`) proves the
  bundle and Go agree on address + signature.

## Step 1 — Xcode: bundle the SDK (required)
Add `ios/Echo/Resources/wallet-sdk/echo-wallet.bundle.js` to the Echo app target's
**Copy Bundle Resources** build phase in `EchoApp.xcodeproj`. Verify at runtime:
`Bundle.main.url(forResource: "echo-wallet.bundle", withExtension: "js")` is
non-nil. Without this, `StargazerSigner` throws `.bundleMissing` and provisioning
falls back to interim mode.

## Step 2 — lastRef / parent (the one design decision left)
TokenLock / DelegatedStake bodies include a `parent` (the signer's last tx
reference: `{hash, ordinal}`). The client must include it in the body it signs.
Pick one:
- **(A, recommended) Backend tx-context endpoint.** Add `GET /v3/wallet/tx-context?type=tokenLock`
  that queries the metagraph L1 for the address's last reference and returns
  `{parent, source}`. Keeps metagraph access server-side. Add a
  `MetagraphClient.QueryLastRef(addr)` (GET `{CurrencyL1URL}/transactions/last-reference/{addr}`
  or the metagraph's equivalent — confirm the path against the testnet).
- **(B) Client-direct.** iOS queries the L1 last-reference endpoint itself.

The `parent` is inspectable (a hash + ordinal), so signing the assembled body is
not blind-signing.

## Step 3 — iOS: build, sign, attach (stake path)
In `HTTPWalletAPIClient.submitTokenLock(amount:tier:)`:
1. `let account = try await WalletKeyStore.shared.ensureWallet()`
2. Fetch tx-context (Step 2) → `parent`, `source` (= account.address).
3. Build the body:
   `["source": source, "amount": EchoDatum.toDatum(amount), "fee": 0, "parent": parent]`
   (omit null `currencyId`/`unlockEpoch`; the bundle/normalize strips nulls).
4. `let signed = try await StargazerSigner.shared.signTransaction(privateKey: account.privateKey, publicKey: account.publicKey, body: body)`
5. Build the `X-Wallet-Proof` header (reuse `WalletProvisioner.buildProofHeaders`
   pattern: GET `/v3/wallet/challenge` → `WalletKeyStore.signChallenge`).
6. POST `/v3/wallet/stake` with body `{amount, tier, signed: {value, proofs}}` and
   the proof header (use `apiClient.post(endpoint:body:headers:)`).
Repeat the pattern for unstake (`/withdraw` → DelegatedStake withdraw, PUT) and
delegate (`/delegated-stakes`). NOTE: withdraw is a **PUT**; add a PUT variant of
`SubmitSignedByType` / `submitTransaction` (current relay is POST-only).

## Step 4 — Backend: extend the flow-flip to unstake/delegate
Mirror the stake change (`internal/api/wallet_handlers.go` +
`internal/wallet/service.go` + `metagraph_querier.go`): add `UnstakeSigned` /
`DelegateSigned` accepting `signed`, calling `SubmitSignedByType` with
`"delegatedStake"` (and a withdraw type once PUT is supported). Add the
`delegatedStake` → Global L0 mapping is already in `SubmitSignedByType`.

## Step 5 — Testnet config
Set on the backend:
- `CURRENCY_L1_URL` = testnet DAG L1, `L0_URL` = testnet Global L0 (confirm the
  exact env var names in `main.go` / `MetagraphConfig`).
- `ECHO_WALLET_REAL_FUNDS=1` (turns on `requireCustody` + signed forwarding).
Confirm startup logs: `Wallet REAL-FUNDS custody mode ON`.

## Step 6 — End-to-end verification (testnet)
1. Fresh install → first run provisions a wallet (24-word phrase, `DAG…` address).
2. `GET /v3/wallet` shows `custodyMode: "real_funds"`.
3. Fund the address from a testnet faucet.
4. Back up the recovery phrase (clears the backup gate) — required before staking.
5. Stake a small amount → confirm the TokenLock appears on the DAG explorer and in
   `GET /v3/wallet`.
6. Reinstall, restore from the phrase → same address; positions reappear.
7. Negative: a hand-crafted `SHA256(did)` address is rejected
   (`SERVER_DERIVABLE_ADDRESS`); a replayed challenge is rejected.

## Step 7 — Security review (gate before mainnet / real value)
- [ ] Supply chain: `echo-wallet.bundle.js` rebuilt via `scripts/build-wallet-sdk.sh`,
      checksum reviewed; `npm audit` shows no high/critical (see `wallet-sdk/SECURITY.md`).
- [ ] Key handling: mnemonic/privkey only in Keychain; never logged; minimal time
      in the JS heap; consider biometric (`kSecAccessControl`) gating on the
      mnemonic item.
- [ ] Proof replay: server nonce TTL (5 min) + single-use confirmed; proof bound
      to DID + address.
- [ ] No blind signing: the client constructs and serializes (brotli) the exact
      body it signs.
- [ ] Brotli byte-exactness: confirm Tessellation accepts the client-produced
      `{value, proofs}` (Step 6.5) — the only thing not verifiable off-node.
- [ ] Recovery: phrase backup gate enforced before first real-value action;
      restore round-trip verified (Step 6.6).

## Rollback
`ECHO_WALLET_REAL_FUNDS=0` returns to interim mode (server-derivable address, no
real funds, signed forwarding disabled) without code changes.
