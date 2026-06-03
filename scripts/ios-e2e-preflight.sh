#!/usr/bin/env bash
# scripts/ios-e2e-preflight.sh — Checklist before Xcode manual E2E (simulator or device).
#
# Agents: run via `make ios-preflight` or MCP `run_ios_preflight`.
# Humans: fix anything marked FAIL, then follow docs/E2E_QUICK_START.md § "You in Xcode".
#
# Usage:
#   ./scripts/ios-e2e-preflight.sh           # checks only
#   ./scripts/ios-e2e-preflight.sh --build   # + xcodebuild EchoApp for simulator
#   ./scripts/ios-e2e-preflight.sh --tests   # + EchoSecurityTests + EchoPhase3Tests + Phase2

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DO_BUILD=0
DO_TESTS=0
API_BASE="${API_URL:-http://localhost:8000}"
API_BASE="${API_BASE%/}"
SIM_DEST="${IOS_SIM_DEST:-}"

for arg in "$@"; do
  case "$arg" in
    --build) DO_BUILD=1 ;;
    --tests) DO_TESTS=1 ;;
    -h|--help)
      sed -n '2,12p' "$0"
      exit 0
      ;;
    *) echo "unknown flag: $arg" >&2; exit 2 ;;
  esac
done

pick_simulator_dest() {
  if [ -n "$SIM_DEST" ]; then
    echo "$SIM_DEST"
    return
  fi
  local devdir="${DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}"
  local name
  for name in "iPhone 17 Pro" "iPhone 17" "iPhone 16 Pro" "iPhone 16e"; do
    if xcrun simctl list devices available 2>/dev/null | grep -F "$name (" >/dev/null 2>&1; then
      echo "platform=iOS Simulator,name=$name"
      return
    fi
  done
  echo "generic/platform=iOS Simulator"
}

if [ -t 1 ]; then
  GREEN=$'\033[0;32m'; YELLOW=$'\033[1;33m'; RED=$'\033[0;31m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  GREEN=""; YELLOW=""; RED=""; BOLD=""; RESET=""
fi

ok()   { printf "  %s✓%s %s\n" "$GREEN" "$RESET" "$1"; }
warn() { printf "  %s!%s %s\n" "$YELLOW" "$RESET" "$1"; WARNINGS=$((WARNINGS + 1)); }
fail() { printf "  %s✗%s %s\n" "$RED" "$RESET" "$1"; FAILURES=$((FAILURES + 1)); }

FAILURES=0
WARNINGS=0

section() { printf "\n%s=== %s ===%s\n" "$BOLD" "$1" "$RESET"; }

section "Xcode toolchain"
XCODE_DEV=""
if [ -d /Applications/Xcode.app/Contents/Developer ]; then
  XCODE_DEV="/Applications/Xcode.app/Contents/Developer"
elif [ -d /Applications/Xcode_16.2.app/Contents/Developer ]; then
  XCODE_DEV="/Applications/Xcode_16.2.app/Contents/Developer"
fi

if [ -z "$XCODE_DEV" ]; then
  fail "Xcode.app not found — install from App Store (Command Line Tools alone is not enough)"
else
  ok "Xcode.app at $XCODE_DEV"
  export DEVELOPER_DIR="$XCODE_DEV"
  if xcode-select -p 2>/dev/null | grep -q CommandLineTools; then
    warn "xcode-select points at CLT — run: sudo xcode-select -s $XCODE_DEV"
  else
    ok "xcode-select OK ($(xcode-select -p))"
  fi
fi

section "Backend ($API_BASE)"
health_ok=0
if curl -sf --max-time 5 "$API_BASE/health" >/dev/null 2>&1; then
  ok "GET /health → operational"
  health_ok=1
else
  fail "Backend unreachable at $API_BASE — run: make dev"
fi

if [ "$health_ok" -eq 1 ]; then
  sms_code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 \
    -X POST "$API_BASE/v1/auth/sms-recovery/register" \
    -H 'Content-Type: application/json' \
    -d '{"phone_hash":"sha256:00","phone_raw":"+15550001111","did":"did:key:zTest"}' 2>/dev/null || echo "000")
  if [ "$sms_code" = "000" ] || [ -z "$sms_code" ]; then
    fail "SMS recovery endpoint unreachable"
  else
    ok "SMS recovery endpoint responds (HTTP $sms_code)"
  fi

  vc_code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 \
    -X POST "$API_BASE/v1/enrollment/vc/start" \
    -H 'Content-Type: application/json' \
    -d '{"requested_claims":{"givenName":true,"familyName":true,"ageOver18":true}}' 2>/dev/null || echo "000")
  if [ "$vc_code" = "200" ]; then
    ok "OIDC4VC enrollment enabled (POST /v1/enrollment/vc/start → 200)"
  elif [ "$vc_code" = "503" ] || [ "$vc_code" = "404" ]; then
    warn "OIDC4VC not enabled (HTTP $vc_code) — set OIDC4VC_ENABLED=true in .env for WO-100"
  elif [ "$vc_code" = "000" ] || [ -z "$vc_code" ]; then
    fail "OIDC4VC start unreachable"
  else
    warn "OIDC4VC start returned HTTP $vc_code"
  fi
fi

section "iOS client wiring"
SCHEME="$ROOT/ios/Echo/EchoApp.xcodeproj/xcshareddata/xcschemes/EchoApp.xcscheme"
if [ -f "$SCHEME" ]; then
  ok "Scheme EchoApp found"
  if grep -q 'key = "API_URL"' "$SCHEME" && grep -q 'localhost:8000' "$SCHEME"; then
    ok "Scheme sets API_URL=http://localhost:8000 (simulator default)"
  else
    warn "Check EchoApp scheme → Run → Environment → API_URL"
  fi
else
  fail "Missing EchoApp.xcscheme"
fi

OPRF="$ROOT/ios/Echo/Libraries/EchoOPRF.xcframework"
if [ -d "$OPRF" ]; then
  ok "EchoOPRF.xcframework present (live PSI)"
else
  warn "EchoOPRF.xcframework missing — PSI uses MockOPRFClient (dev only). Run: ./scripts/build-echooprf-ios.sh"
fi

LAN_IP=""
if command -v ipconfig >/dev/null; then
  LAN_IP=$(ipconfig getifaddr en0 2>/dev/null || true)
fi
if [ -n "$LAN_IP" ]; then
  ok "Mac LAN IP for physical device: $LAN_IP"
  printf "      → Xcode scheme API_URL = http://%s:8000 (phone cannot use localhost)\n" "$LAN_IP"
else
  warn "Could not detect en0 IP — set API_URL to your Mac LAN address for device testing"
fi

if [ -f "$ROOT/.env" ] && grep -q '^DEV_MODE=true' "$ROOT/.env" 2>/dev/null; then
  ok "DEV_MODE=true — SMS OTP in X-Dev-OTP response header"
else
  warn "DEV_MODE not true — SMS OTP needs Twilio or enable DEV_MODE=true for local E2E"
fi

if [ "$DO_BUILD" -eq 1 ]; then
  section "EchoApp simulator build"
  if [ -z "$XCODE_DEV" ]; then
    fail "Skipping build — no Xcode.app"
  else
    DEST=$(pick_simulator_dest)
    ok "Using destination: $DEST"
    if xcodebuild -project "$ROOT/ios/Echo/EchoApp.xcodeproj" \
      -scheme EchoApp \
      -destination "$DEST" \
      -configuration Debug build -quiet; then
      ok "xcodebuild EchoApp (simulator) succeeded"
    else
      fail "xcodebuild EchoApp failed — open Xcode and fix compile errors"
    fi
  fi
fi

if [ "$DO_TESTS" -eq 1 ]; then
  section "iOS SPM tests (Security + Phase 3)"
  if [ -z "$XCODE_DEV" ]; then
    fail "Skipping tests — no Xcode.app"
  else
    warn "SPM full Echo library may fail on macOS host — Xcode EchoApp build is the compile gate"
    (
      cd "$ROOT/ios/Echo"
      if swift build --target EchoSecurityTests 2>/dev/null; then
        swift test --filter EchoSecurityTests && ok "EchoSecurityTests"
      else
        warn "EchoSecurityTests skipped (SPM host build)"
      fi
      if swift build --target EchoPhase3Tests 2>/dev/null; then
        swift test --filter EchoPhase3Tests && ok "EchoPhase3Tests"
      else
        warn "EchoPhase3Tests skipped (SPM host build)"
      fi
    )
    go test ./mobile/echooprf/... -count=1 >/dev/null && ok "mobile/echooprf interop"
  fi
fi

section "Manual E2E (you in Xcode — ~15 min)"
cat <<'EOF'
  1. make dev && make ios-preflight          ← agents can run this
  2. Open ios/Echo/EchoApp.xcodeproj
  3. Scheme EchoApp → iPhone 16 Pro simulator → Run
  4. Simulator: Features → Face ID → Enrolled
  5. Walkthrough: Welcome → username → Face ID → Recovery (SMS) → Messages
  6. Settings → Privacy → Contact discovery ON → Contacts on ECHO → Scan → Add
  Detail: docs/E2E_QUICK_START.md
EOF

section "Summary"
if [ "$FAILURES" -gt 0 ]; then
  printf "  %s✗%s %d blocker(s), %d warning(s) — fix FAIL items before Xcode E2E\n" "$RED" "$RESET" "$FAILURES" "$WARNINGS"
  exit 1
fi
if [ "$WARNINGS" -gt 0 ]; then
  printf "  %s!%s Ready with %d warning(s) — simulator onboarding should work; device/PSI/OIDC may need fixes\n" "$YELLOW" "$RESET" "$WARNINGS"
  exit 0
fi
printf "  %s✓%s All checks passed — open EchoApp in Xcode\n" "$GREEN" "$RESET"
exit 0
