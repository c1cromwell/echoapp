# wallet-sdk security & supply-chain notes

The shipped artifact is **`echo-wallet.bundle.js`** (a single self-contained IIFE
loaded into iOS JavaScriptCore). The `node_modules` tree is build-time only.

## Bundle size
- ~1.3 MB for keygen/address/message-signing (down from ~5.7 MB) after switching
  from the `@stardust-collective/dag4` umbrella to **`@stardust-collective/dag4-keystore`**
  only and shimming Node builtins over `@noble/hashes`.
- ~3.6 MB once the **brotli-wasm** (~1.38 MB, base64-embedded) is included for
  transaction signing (TokenLock / DelegatedStake / WithdrawDelegatedStake).
  The wasm is instantiated from embedded bytes (`init(bytes)`), so there is no
  runtime fetch/`import.meta`.

## Vulnerability posture
- **No high/critical** advisories. The earlier critical (`form-data`) and high
  (`ws`) came from `eth-lib`'s unused networking subtree (`servify`/`request`);
  they are pinned via `overrides` AND tree-shaken out of the bundle.
- What the runtime bundle actually contains: `dag4-keystore` crypto,
  `@noble/hashes`, `@noble/secp256k1`/`elliptic`, `bip39`, `bs58`, `bignumber.js`,
  `buffer`, and the local shims. Verified absent from the bundle: `form-data`,
  `ws`, `request`, `servify`, `uuid` (shimmed via `src/shims/uuid-v4.js`).
- `overrides` (package.json) force-patch leaf deps: `form-data`, `ws`, `qs`,
  `tough-cookie`, `elliptic`, `secp256k1`.
- Any residual `npm audit` entries are **dev-tree transitive deps not present in
  the shipped artifact** (confirm with `grep -c <pkg> echo-wallet.bundle.js`).

## Runtime host requirements (Swift, M1)
The bundle expects the JSContext to provide (all injected by `StargazerSigner`):
- `process` (minimal `{ env: {} }`).
- `crypto.getRandomValues` — backed by `SecRandomCopyBytes` (keygen, uuid).
  Without it, key generation throws by design.
- `TextEncoder` / `TextDecoder` — bare JSC lacks them; minimal UTF-8 polyfills
  are injected (dag4 + the brotli glue need them).
- `WebAssembly` is built into JSC (no injection needed) — used for brotli.
- `self` is deliberately NOT defined: it would route `@noble` through an absent
  `crypto.subtle`. dag4's brotli takes the Node branch, where our shim exposes an
  async `compress`.

## iOS integration (required Xcode step)
`StargazerSigner` loads the bundle via `Bundle.main.url(forResource:
"echo-wallet.bundle", withExtension: "js")`. **Add `echo-wallet.bundle.js` to the
Echo app target's "Copy Bundle Resources" build phase in `EchoApp.xcodeproj`**
(the app is built by the Xcode project, not SPM). Without it, `StargazerSigner`
throws `.bundleMissing` and provisioning falls back to interim mode.

Runtime verification (Xcode / device): first run generates a 24-word mnemonic,
derives a `DAG…` (40-char) address, links it with an `X-Wallet-Proof` header, and
Settings → recovery reveals the phrase; restoring it yields the same address.

## Reproducible build
`scripts/build-wallet-sdk.sh` → `npm ci` (pinned `package-lock.json`) →
`node build.js` → checksum. Review `echo-wallet.bundle.js.sha256` on each rebuild.
`gen-vector.js` regenerates `testvector.json`, consumed by the Go cross-validation
test (`internal/wallet/dag_crossvalidation_test.go`) which proves the bundle and
backend agree on address derivation + signature verification.
