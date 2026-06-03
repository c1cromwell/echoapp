#!/usr/bin/env bash
# scripts/phase3-signals-proof.sh — Headless proof for WO-192 / WO-10 signal relay rules.
#
# Usage: ./scripts/phase3-signals-proof.sh
# Manual two-client UI proof: docs/E2E_QUICK_START.md § Phase 3 signals

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== Phase 3 backend WS relay (privacy routing) =="
go test ./internal/api/ -run 'TestRouteInbound_' -count=1

echo ""
echo "== Phase 3 iOS unit tests (codec, ViewModel, service) =="
if command -v xcrun >/dev/null 2>&1 && xcode-select -p 2>/dev/null | grep -q Xcode.app; then
  (
    cd ios/Echo
    DEVELOPER_DIR="${DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}"
    swift test --filter EchoPhase3Tests 2>&1
  ) || {
    echo "⚠️  SPM Phase3 tests failed (known macOS cross-target noise)."
    echo "    Run in Xcode: Product → Test → EchoPhase3Tests"
    exit 1
  }
else
  echo "⚠️  Full Xcode required for EchoPhase3Tests; backend tests passed."
fi

echo ""
echo "✅ Headless Phase 3 proof complete."
echo "Next: two simulators — typing, read receipts, reactions (E2E_QUICK_START.md)."
