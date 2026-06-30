#!/usr/bin/env bash
# scripts/build-wallet-sdk.sh — reproducibly build the embedded dag4 wallet SDK
# bundle that the iOS app loads into JavaScriptCore for real-custody signing.
#
# Output: ios/Echo/Resources/wallet-sdk/echo-wallet.bundle.js (self-contained
# IIFE; no Node deps at runtime). Also regenerates testvector.json used by the
# Go cross-validation test (internal/wallet/dag_crossvalidation_test.go).
#
# Requires: node + npm. Pinned versions live in package.json / package-lock.json.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SDK_DIR="$ROOT/ios/Echo/Resources/wallet-sdk"
cd "$SDK_DIR"

echo "== install (pinned) =="
if [ -f package-lock.json ]; then
  npm ci
else
  npm install
fi

echo "== bundle =="
node build.js

echo "== regenerate cross-validation vector =="
node gen-vector.js >/dev/null

echo "== checksum =="
shasum -a 256 echo-wallet.bundle.js | tee echo-wallet.bundle.js.sha256

echo "✅ wallet SDK built: $SDK_DIR/echo-wallet.bundle.js"
echo "   Vendor this bundle in the app target; review the checksum on each rebuild."
