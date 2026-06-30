# wallet-sdk security & supply-chain notes

The shipped artifact is **`echo-wallet.bundle.js`** (a single self-contained IIFE
loaded into iOS JavaScriptCore). The `node_modules` tree is build-time only.

## Bundle size
- ~1.3 MB (down from ~5.7 MB) after switching from the `@stardust-collective/dag4`
  umbrella to **`@stardust-collective/dag4-keystore`** only (drops the
  network/account/monitor layers) and shimming Node builtins over `@noble/hashes`.

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
The bundle expects the JSContext to provide:
- `process` (minimal `{ env: {} }`) — see gen-vector.js sandbox.
- `crypto.getRandomValues` — Swift must inject one backed by `SecRandomCopyBytes`
  (used for keygen and uuid). Without it, key generation throws by design.

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
