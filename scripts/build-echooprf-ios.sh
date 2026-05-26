#!/usr/bin/env bash
# Build EchoOPRF.xcframework for iOS (WO-221). Requires full Xcode + gomobile.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/ios/Echo/Libraries/EchoOPRF.xcframework"

if [[ -d "$OUT" ]]; then
  echo "EchoOPRF.xcframework already exists at $OUT"
  exit 0
fi

if ! xcode-select -p 2>/dev/null | grep -q 'Xcode.app'; then
  echo "error: full Xcode is required (xcode-select -s /Applications/Xcode.app/Contents/Developer)" >&2
  exit 1
fi

export PATH="${PATH}:$(go env GOPATH)/bin"
if ! command -v gomobile >/dev/null; then
  go install golang.org/x/mobile/cmd/gomobile@latest
  go install golang.org/x/mobile/cmd/gobind@latest
fi

cd "$ROOT"
go get -tool golang.org/x/mobile/cmd/gobind 2>/dev/null || true
gomobile init
mkdir -p "$(dirname "$OUT")"
gomobile bind -target=ios -o "$OUT" ./mobile/echooprf

echo "Built $OUT"
echo "In Xcode: EchoApp target → General → Frameworks → + → Add EchoOPRF.xcframework → Embed & Sign"
