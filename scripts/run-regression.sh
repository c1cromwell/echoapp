#!/usr/bin/env bash
# scripts/run-regression.sh — Headless regression gate (Go + optional iOS SPM).
#
# Usage:
#   ./scripts/run-regression.sh              # default: Go release-check + targeted suites + iOS SPM if Xcode
#   ./scripts/run-regression.sh --quick        # Go only (no validate-phase1, no iOS)
#   ./scripts/run-regression.sh --with-phase1  # also run make validate-phase1 (needs Docker + Euclid)
#   ./scripts/run-regression.sh --ios-only     # iOS SPM tests only
#
# Manual Xcode/device E2E is NOT run here — see docs/TESTING.md §5–6.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

QUICK=0
PHASE1=0
IOS_ONLY=0

for arg in "$@"; do
  case "$arg" in
    --quick) QUICK=1 ;;
    --with-phase1) PHASE1=1 ;;
    --ios-only) IOS_ONLY=1 ;;
    -h|--help)
      sed -n '2,10p' "$0"
      exit 0
      ;;
    *) echo "unknown flag: $arg" >&2; exit 2 ;;
  esac
done

if [ -t 1 ]; then
  GREEN=$'\033[0;32m'; RED=$'\033[0;31m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  GREEN=""; RED=""; BOLD=""; RESET=""
fi

section() { printf "\n%s=== %s ===%s\n" "$BOLD" "$1" "$RESET"; }
pass()    { printf "  %s✓%s %s\n" "$GREEN" "$RESET" "$1"; }
fail()    { printf "  %s✗%s %s\n" "$RED" "$RESET" "$1"; exit 1; }

run_ios_spm() {
  section "iOS SPM (EchoSecurityTests + EchoPhase3Tests)"
  if ! command -v swift >/dev/null; then
    fail "swift not found — install Xcode"
  fi
  local devdir=""
  if [ -d /Applications/Xcode.app/Contents/Developer ]; then
    devdir="/Applications/Xcode.app/Contents/Developer"
  elif [ -d /Applications/Xcode_16.2.app/Contents/Developer ]; then
    devdir="/Applications/Xcode_16.2.app/Contents/Developer"
  fi
  if [ -z "$devdir" ]; then
    fail "Xcode.app not found — iOS SPM tests require full Xcode (see docs/TESTING.md §4)"
  fi
  export DEVELOPER_DIR="$devdir"
  (
    cd ios/Echo
    swift build --target Echo
    pass "Echo library build"
    swift build --target EchoSecurityTests
    swift test --filter EchoSecurityTests
    pass "EchoSecurityTests"
    swift build --target EchoPhase3Tests
    swift test --filter EchoPhase3Tests
    pass "EchoPhase3Tests"
  )
  go test ./mobile/echooprf/... -count=1
  pass "mobile/echooprf interop"
}

if [ "$IOS_ONLY" -eq 1 ]; then
  run_ios_spm
  section "Done"
  pass "iOS regression complete"
  exit 0
fi

if [ "$QUICK" -eq 0 ]; then
  section "Go release-check"
  make release-check
  pass "make release-check"

  section "Targeted API / crypto suites"
  go test -count=1 ./internal/api/ -run 'EnrollmentVC|WS|Reaction|VIP|StepUp' ./internal/services/contacts/... ./mobile/echooprf/...
  pass "targeted Go suites"

  if [ "$PHASE1" -eq 1 ]; then
    section "Phase 1 metagraph go/no-go (WO-230)"
    make validate-phase1
    pass "validate-phase1"
  fi

  run_ios_spm
else
  section "Quick Go regression"
  go test -race -count=1 ./internal/... ./pkg/... ./test/...
  pass "Go tests (race)"
fi

section "Done"
printf "%sRegression gate passed.%s\n" "$GREEN" "$RESET"
printf "Next: run manual Xcode E2E checklists in docs/TESTING.md §5–7 before TestFlight.\n"
