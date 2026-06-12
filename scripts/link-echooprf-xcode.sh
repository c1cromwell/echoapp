#!/usr/bin/env bash
# WO-221: Verify EchoOPRF.xcframework is built and print Xcode link steps.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FRAMEWORK="${ROOT}/ios/Echo/Libraries/EchoOPRF.xcframework"
PBXPROJ="${ROOT}/ios/Echo/EchoApp.xcodeproj/project.pbxproj"

if [[ ! -d "$FRAMEWORK" ]]; then
  echo "EchoOPRF.xcframework missing. Run: make echooprf-ios" >&2
  exit 1
fi

if grep -q 'EchoOPRF.xcframework' "$PBXPROJ" 2>/dev/null; then
  echo "EchoOPRF.xcframework already referenced in EchoApp.xcodeproj"
  exit 0
fi

cat <<EOF
Built: $FRAMEWORK

Add to EchoApp in Xcode (one-time):
  1. Open ios/Echo/EchoApp.xcodeproj
  2. EchoApp target → General → Frameworks, Libraries, and Embedded Content
  3. + → Add Other → Add Files → select EchoOPRF.xcframework
  4. Set Embed to "Embed & Sign"

OPRFClientFactory will then use LiveOPRFClient (#if canImport(Echooprf)).
EOF
