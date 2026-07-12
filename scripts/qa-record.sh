#!/usr/bin/env bash
# Automated QA/QE: boot a simulator, screen-record a full UI-test run, and pull
# the session video + per-test screenshots out of the .xcresult bundle.
#
# Usage:
#   ./scripts/qa-record.sh                                  # EchoApp scheme, iPhone 17
#   SCHEME=EchoMessaging ./scripts/qa-record.sh
#   SIMULATOR="iPhone 16 Pro" ./scripts/qa-record.sh
#   ONLY_TESTING=EchoUITests/EchoUXFlowUITests ./scripts/qa-record.sh
#
# Prereqs (one-time, in Xcode — see NOTES at the bottom):
#   1. A UI Testing Bundle target (e.g. "EchoUITests") exists.
#   2. That target is included in the scheme's test plan so `xcodebuild test`
#      actually runs it.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCHEME="${SCHEME:-EchoApp}"
SIMULATOR="${SIMULATOR:-iPhone 17}"
OUT="${QA_OUT:-$ROOT/build/qa}"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$OUT/$STAMP"
RESULT_BUNDLE="$RUN_DIR/result.xcresult"
VIDEO="$RUN_DIR/session.mp4"

mkdir -p "$RUN_DIR"
cd "$ROOT/ios/Echo"

echo "==> Booting simulator: $SIMULATOR"
xcrun simctl boot "$SIMULATOR" 2>/dev/null || true
xcrun simctl bootstatus "$SIMULATOR" -b >/dev/null 2>&1 || true
open -a Simulator >/dev/null 2>&1 || true   # visible locally; harmless in CI

echo "==> Recording video → $VIDEO"
xcrun simctl io booted recordVideo --codec=h264 --force "$VIDEO" &
REC_PID=$!
# Always stop the recorder cleanly, however the script exits.
stop_recording() {
  if kill -0 "$REC_PID" 2>/dev/null; then
    kill -INT "$REC_PID" 2>/dev/null || true
    wait "$REC_PID" 2>/dev/null || true
  fi
}
trap stop_recording EXIT

echo "==> Running UI tests (scheme=$SCHEME, sim=$SIMULATOR)"
set +e
xcodebuild test \
  -project EchoApp.xcodeproj \
  -scheme "$SCHEME" \
  -destination "platform=iOS Simulator,name=$SIMULATOR" \
  -resultBundlePath "$RESULT_BUNDLE" \
  ${ONLY_TESTING:+-only-testing:$ONLY_TESTING} \
  2>&1 | tee "$RUN_DIR/xcodebuild.log"
TEST_EXIT=${PIPESTATUS[0]}
set -e

# Stop recording before we read the bundle so the .mp4 is finalized.
stop_recording
trap - EXIT

echo "==> Extracting attachments (screenshots / per-test clips) from result bundle"
mkdir -p "$RUN_DIR/attachments"
# Xcode 16+ : direct attachment export. Older toolchains: open result.xcresult in Xcode.
xcrun xcresulttool export attachments \
  --path "$RESULT_BUNDLE" \
  --output-path "$RUN_DIR/attachments" >/dev/null 2>&1 \
  || echo "    (attachment export unavailable on this Xcode — open result.xcresult in Xcode instead)"

echo ""
echo "==> QA run complete (xcodebuild test exit=$TEST_EXIT)"
echo "    Video:       $VIDEO"
echo "    Result:      $RESULT_BUNDLE   (open in Xcode for the full step timeline)"
echo "    Attachments: $RUN_DIR/attachments"
echo "    Log:         $RUN_DIR/xcodebuild.log"
exit "$TEST_EXIT"

# ------------------------------------------------------------------ NOTES -----
# Creating the UI Testing target (one-time, in Xcode; can't be scripted safely
# while the project is open):
#   File ▸ New ▸ Target… ▸ "UI Testing Bundle" ▸ name it EchoUITests,
#   Target to be Tested = Echo (the app).
#   Then add ios/Echo/UITests/EchoUXFlowUITests.swift to that target, and add
#   EchoUITests to the scheme's test plan (Product ▸ Scheme ▸ Edit Scheme… ▸
#   Test ▸ +).
#
# CI: this script is headless-friendly. On a CI runner, drop `open -a Simulator`
# (it no-ops without a GUI) and archive $RUN_DIR as a build artifact.
