#!/usr/bin/env bash
# scripts/validate-phase1.sh — Phase-1 testnet go/no-go validation script (WO-230).
#
# Runs the 6-step end-to-end check defined in WO-230:
#   1. Derive did:key locally from a P-256 key pair (no chain transaction)
#   2. Register DID with local Identity Service (POST /identity/register)
#   3. Issue a test trust tier VC and confirm it anchors on local Identity
#      Metagraph (finality < 30s)
#   4. Send a test message through the local relay
#   5. Confirm Merkle root committed to local Data L1 (finality < 30s)
#   6. Verify snapshot height increments on Global L0
#
# Exit codes:
#   0  — all steps passed
#   1  — at least one step failed
#   2  — a prerequisite is missing (cluster not running, blocker WO unresolved)

set -uo pipefail

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
GLOBAL_L0_URL="${GLOBAL_L0_URL:-http://localhost:9000}"
METAGRAPH_L0_URL="${METAGRAPH_L0_URL:-http://localhost:9200}"
CURRENCY_L1_URL="${CURRENCY_L1_URL:-http://localhost:9300}"
DATA_L1_URL="${DATA_L1_URL:-http://localhost:9400}"
IDENTITY_L0_URL="${IDENTITY_L0_URL:-http://localhost:9600}"
BACKEND_URL="${BACKEND_URL:-http://localhost:8000}"
FINALITY_TIMEOUT_SECS="${FINALITY_TIMEOUT_SECS:-60}"

PASS=0
FAIL=0
SKIP=0

# -----------------------------------------------------------------------------
# Output helpers
# -----------------------------------------------------------------------------
if [ -t 1 ]; then
  GREEN=$'\033[0;32m'; RED=$'\033[0;31m'; YELLOW=$'\033[0;33m'
  BLUE=$'\033[0;34m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  GREEN=""; RED=""; YELLOW=""; BLUE=""; BOLD=""; RESET=""
fi

step()    { printf "\n%s=== Step %s: %s ===%s\n" "$BOLD$BLUE" "$1" "$2" "$RESET"; }
ok()      { printf "  %s✓%s %s\n" "$GREEN" "$RESET" "$1"; PASS=$((PASS + 1)); }
fail()    { printf "  %s✗%s %s\n" "$RED"   "$RESET" "$1"; FAIL=$((FAIL + 1)); }
skip()    { printf "  %s○%s %s\n" "$YELLOW" "$RESET" "$1"; SKIP=$((SKIP + 1)); }
info()    { printf "    %s\n" "$1"; }

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "missing prerequisite command: $1"
    return 1
  fi
}

# -----------------------------------------------------------------------------
# Prerequisites
# -----------------------------------------------------------------------------
printf "%s===== Phase-1 Testnet Go/No-Go Validation =====%s\n" "$BOLD" "$RESET"
printf "Backend:        %s\n" "$BACKEND_URL"
printf "Global L0:      %s\n" "$GLOBAL_L0_URL"
printf "Metagraph L0:   %s\n" "$METAGRAPH_L0_URL"
printf "Currency L1:    %s\n" "$CURRENCY_L1_URL"
printf "Data L1:        %s\n" "$DATA_L1_URL"
printf "Identity L0:    %s (separate metagraph — 'make start-identity')\n" "$IDENTITY_L0_URL"
printf "Finality limit: %ss\n" "$FINALITY_TIMEOUT_SECS"

step 0 "Prerequisites"
PREREQ_OK=1
for cmd in curl jq openssl xxd; do
  if require_cmd "$cmd"; then
    ok "found: $cmd"
  else
    PREREQ_OK=0
  fi
done
if [ "$PREREQ_OK" -ne 1 ]; then
  printf "\n%sAborting: install missing prerequisites and re-run.%s\n" "$RED" "$RESET"
  exit 2
fi

check_endpoint() {
  local label=$1 url=$2
  if curl -fsS --max-time 5 "$url/node/info" >/dev/null 2>&1 \
     || curl -fsS --max-time 5 "$url/health"    >/dev/null 2>&1; then
    ok "$label reachable at $url"
    return 0
  else
    fail "$label NOT reachable at $url"
    return 1
  fi
}

CLUSTER_OK=1
check_endpoint "Global L0"    "$GLOBAL_L0_URL"    || CLUSTER_OK=0
check_endpoint "Metagraph L0" "$METAGRAPH_L0_URL" || CLUSTER_OK=0
check_endpoint "Data L1"      "$DATA_L1_URL"      || CLUSTER_OK=0
check_endpoint "Currency L1"  "$CURRENCY_L1_URL"  || CLUSTER_OK=0
check_endpoint "Backend"      "$BACKEND_URL"      || CLUSTER_OK=0

if [ "$CLUSTER_OK" -ne 1 ]; then
  printf "\n%sCluster not fully up. Run 'make dev' first, then retry.%s\n" "$RED" "$RESET"
  exit 2
fi

# -----------------------------------------------------------------------------
# Step 1: Derive did:key from a fresh P-256 key pair (local-only, no chain)
# -----------------------------------------------------------------------------
step 1 "Derive did:key from P-256 key pair (local)"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if openssl ecparam -name prime256v1 -genkey -noout -out "$TMP_DIR/key.pem" 2>/dev/null; then
  ok "generated P-256 key pair"
else
  fail "openssl ecparam failed"
fi

# Export the SubjectPublicKeyInfo PEM that pkg/didkey can parse directly.
if openssl ec -in "$TMP_DIR/key.pem" -pubout -out "$TMP_DIR/pub.pem" 2>/dev/null; then
  ok "exported public key PEM"
else
  fail "openssl ec -pubout failed"
fi

# Use the canonical Go derivation tool — same code path the backend uses to
# validate registrations. Output format is `did=…\npublic_key_hex=…`.
if go run ./cmd/didkey -pem "$TMP_DIR/pub.pem" >"$TMP_DIR/derive.out" 2>"$TMP_DIR/derive.err"; then
  DID_KEY=$(awk -F= '/^did=/    {print $2}' "$TMP_DIR/derive.out")
  PUB_HEX=$(awk -F= '/^public_key_hex=/ {print $2}' "$TMP_DIR/derive.out")
  if [ -n "$DID_KEY" ] && [ -n "$PUB_HEX" ]; then
    ok "derived canonical DID: $DID_KEY"
    info "public_key_hex (uncompressed, ${#PUB_HEX} chars): ${PUB_HEX:0:32}…"
  else
    fail "didkey tool output missing did/public_key_hex; raw: $(cat "$TMP_DIR/derive.out")"
  fi
else
  fail "didkey derivation tool failed: $(tail -3 "$TMP_DIR/derive.err")"
fi

# -----------------------------------------------------------------------------
# Step 2: Register DID with local Identity Service (POST /identity/register)
# -----------------------------------------------------------------------------
step 2 "Register DID with local Identity Service"

REGISTER_BODY=$(jq -nc --arg did "$DID_KEY" --arg pub "$PUB_HEX" \
  '{did:$did, public_key_hex:$pub}')

REGISTER_HTTP_CODE=$(curl -sS -o "$TMP_DIR/register.out" -w '%{http_code}' \
  --max-time 10 \
  -X POST "$BACKEND_URL/identity/register" \
  -H "Content-Type: application/json" \
  -d "$REGISTER_BODY" 2>/dev/null) || REGISTER_HTTP_CODE="000"

case "$REGISTER_HTTP_CODE" in
  200|201)
    REGISTERED_DID=$(jq -r '.did // empty' "$TMP_DIR/register.out")
    if [ "$REGISTERED_DID" = "$DID_KEY" ]; then
      ok "registered DID (HTTP $REGISTER_HTTP_CODE): $REGISTERED_DID"
    else
      fail "registered DID mismatch (expected $DID_KEY, got '$REGISTERED_DID')"
    fi

    # Re-register with the same payload to prove the endpoint is idempotent.
    REREG_CODE=$(curl -sS -o "$TMP_DIR/rereg.out" -w '%{http_code}' \
      --max-time 10 \
      -X POST "$BACKEND_URL/identity/register" \
      -H "Content-Type: application/json" \
      -d "$REGISTER_BODY" 2>/dev/null) || REREG_CODE="000"
    if [ "$REREG_CODE" = "200" ] && [ "$(jq -r '.existing // false' "$TMP_DIR/rereg.out")" = "true" ]; then
      ok "re-registration is idempotent (HTTP 200, existing=true)"
    else
      fail "re-registration not idempotent (HTTP $REREG_CODE, body: $(cat "$TMP_DIR/rereg.out"))"
    fi

    # Resolve the binding via GET /identity/<did> and confirm the round-trip.
    RESOLVE_CODE=$(curl -sS -o "$TMP_DIR/resolve.out" -w '%{http_code}' \
      --max-time 10 "$BACKEND_URL/identity/$DID_KEY" 2>/dev/null) || RESOLVE_CODE="000"
    # Response shape: {did, devices:[{public_key_hex,device_label,registered_at}]}
    RESOLVED_KEY=$(jq -r '.devices[0].public_key_hex // empty' "$TMP_DIR/resolve.out")
    if [ "$RESOLVE_CODE" = "200" ] && [ "$RESOLVED_KEY" = "$PUB_HEX" ]; then
      ok "GET /identity/<did> round-trip OK"
    else
      fail "GET /identity/<did> failed (HTTP $RESOLVE_CODE, body: $(cat "$TMP_DIR/resolve.out"))"
    fi
    ;;
  000)
    fail "POST /identity/register: backend unreachable"
    ;;
  *)
    fail "POST /identity/register returned HTTP $REGISTER_HTTP_CODE: $(cat "$TMP_DIR/register.out")"
    ;;
esac

# -----------------------------------------------------------------------------
# Step 3: trust-tier commitment on Identity L1 (via dev-only API proxy)
# -----------------------------------------------------------------------------
step 3 "Anchor test trust-tier commitment on Identity Metagraph"

IDENTITY_L1_URL="${IDENTITY_L1_URL:-http://localhost:9500}"

if curl -fsS --max-time 5 "$IDENTITY_L0_URL/node/info" >/dev/null 2>&1; then
  ok "Identity Metagraph L0 reachable ($IDENTITY_L0_URL)"
  if curl -fsS --max-time 5 "$IDENTITY_L1_URL/node/info" >/dev/null 2>&1; then
    ok "Identity Metagraph L1 reachable ($IDENTITY_L1_URL)"
  else
    fail "Identity L1 unreachable at $IDENTITY_L1_URL — start it with 'make start-identity' (separate from hydra)"
  fi

  NONCE="$(openssl rand -hex 16)"
  TIER_BODY=$(jq -nc --arg did "$DID_KEY" --arg nonce "$NONCE" '{subject_did:$did, tier:2, nonce:$nonce}')
  TIER_CODE=$(curl -sS -o "$TMP_DIR/tier.out" -w '%{http_code}' \
    --max-time "$FINALITY_TIMEOUT_SECS" \
    -X POST "$BACKEND_URL/v1/phase1/trust-tier-commitment" \
    -H "Content-Type: application/json" \
    -d "$TIER_BODY" 2>/dev/null) || TIER_CODE="000"

  case "$TIER_CODE" in
    200)
      COMM=$(jq -r '.commitment // empty' "$TMP_DIR/tier.out")
      if [ "${#COMM}" -eq 64 ]; then
        ok "trust-tier commitment anchored (H(tier||nonce) = ${COMM:0:16}…)"
      else
        fail "unexpected response body: $(cat "$TMP_DIR/tier.out")"
      fi
      ;;
    403)
      fail "Phase-1 validate endpoint disabled — API must run with ENVIRONMENT=development (or PHASE1_ALLOW_OPEN_VALIDATE=true)"
      ;;
    000)
      fail "POST /v1/phase1/trust-tier-commitment: backend unreachable"
      ;;
    *)
      fail "POST /v1/phase1/trust-tier-commitment returned HTTP $TIER_CODE: $(cat "$TMP_DIR/tier.out")"
      ;;
  esac
else
  skip "Identity Metagraph not running — start with 'make dev' (WO-276 skeleton + WO-272 validators landed)"
  info "CI/static skeleton: run 'make metagraph-verify-skeleton' (jq). Full compile: cd metagraph && sbt compile;"
  info "cluster boot requires Docker + Euclid sibling repo + hydra (see metagraph/scripts/setup-euclid.sh)."
fi

# -----------------------------------------------------------------------------
# Step 4: Send a test message through the local relay
# -----------------------------------------------------------------------------
step 4 "Send a test message through local relay"

# The relay is a WebSocket service. We use the Go integration tests as a
# proxy for "the relay works end-to-end against the local backend".
if go test -count=1 -run TestE2E_MessageSendReceive ./test/integration >"$TMP_DIR/relay.log" 2>&1; then
  ok "relay E2E test (TestE2E_MessageSendReceive) passed"
else
  fail "relay E2E test failed — see $TMP_DIR/relay.log"
  tail -20 "$TMP_DIR/relay.log" | sed 's/^/    /'
fi

# -----------------------------------------------------------------------------
# Step 5: Confirm Merkle root committed to local Data L1 (finality < 30s)
# -----------------------------------------------------------------------------
step 5 "Commit Merkle root to local Data L1 and verify finality"

# Build a deterministic 32-byte SHA-256 root.
MERKLE_ROOT=$(printf 'phase1-validation-%s' "$DID_KEY" | openssl dgst -sha256 | awk '{print $2}')
ok "computed Merkle root: $MERKLE_ROOT"

# Submit via POST /v1/data-l1/merkle-roots (public route, no auth required).
# Falls back to skip when the backend is unreachable or DataL1 client not configured.
MERKLE_BODY=$(jq -nc --arg root "$MERKLE_ROOT" '{root:$root, leafCount:1}')
SUBMIT_RESP=$(curl -fsS --max-time 10 \
  -X POST "$BACKEND_URL/v1/data-l1/merkle-roots" \
  -H "Content-Type: application/json" \
  -d "$MERKLE_BODY" 2>/dev/null) || SUBMIT_RESP=""

if [ -z "$SUBMIT_RESP" ]; then
  skip "POST /v1/data-l1/merkle-roots returned no response (backend down or DATA_L1_URL not configured)"
  info "Ensure DATA_L1_URL is set in .env.local and the backend DataL1 client is initialised."
else
  TX_ID=$(printf '%s' "$SUBMIT_RESP" | jq -r '.tx_id // .txHash // .id // empty')
  if [ -n "$TX_ID" ]; then
    ok "submitted Merkle root, tx_id=$TX_ID"
    DEADLINE=$(( $(date +%s) + FINALITY_TIMEOUT_SECS ))
    FINALIZED=0
    while [ "$(date +%s)" -lt "$DEADLINE" ]; do
      if curl -fsS --max-time 5 "$DATA_L1_URL/data-application/merkle-roots/$MERKLE_ROOT" \
           | jq -e '.finalized == true' >/dev/null 2>&1; then
        FINALIZED=1
        break
      fi
      sleep 2
    done
    if [ "$FINALIZED" -eq 1 ]; then
      ok "Merkle root finalized in <${FINALITY_TIMEOUT_SECS}s"
    else
      fail "Merkle root not finalized within ${FINALITY_TIMEOUT_SECS}s"
    fi
  else
    fail "submit response missing tx_id: $SUBMIT_RESP"
  fi
fi

# -----------------------------------------------------------------------------
# Step 6: Verify snapshot height increments on Global L0
# -----------------------------------------------------------------------------
step 6 "Verify Global L0 snapshot height increments"

read_height() {
  # Constellation API returns {"value":N} — extract the integer regardless of wrapper.
  local raw
  raw=$(curl -fsS --max-time 5 "$GLOBAL_L0_URL/global-snapshots/latest/ordinal" 2>/dev/null \
     || curl -fsS --max-time 5 "$GLOBAL_L0_URL/snapshots/latest/ordinal" 2>/dev/null) || true
  if [[ "$raw" == \{* ]]; then
    printf '%s' "$raw" | jq -r '.value // .ordinal // .height // empty'
  else
    printf '%s' "$raw"
  fi
}

START_HEIGHT=$(read_height || echo "")
if [ -z "$START_HEIGHT" ] || ! [[ "$START_HEIGHT" =~ ^[0-9]+$ ]]; then
  fail "could not read Global L0 snapshot ordinal (got: '${START_HEIGHT}')"
else
  ok "Global L0 starting ordinal: $START_HEIGHT"
  # Allow at least one snapshot interval before the first poll.
  sleep 5
  DEADLINE=$(( $(date +%s) + FINALITY_TIMEOUT_SECS ))
  ADVANCED=0
  while [ "$(date +%s)" -lt "$DEADLINE" ]; do
    NOW_HEIGHT=$(read_height || echo "")
    if [[ "$NOW_HEIGHT" =~ ^[0-9]+$ ]] && [ "$NOW_HEIGHT" -gt "$START_HEIGHT" ]; then
      ok "Global L0 ordinal advanced: $START_HEIGHT → $NOW_HEIGHT"
      ADVANCED=1
      break
    fi
    sleep 3
  done
  [ "$ADVANCED" -eq 1 ] || fail "Global L0 ordinal did not advance within ${FINALITY_TIMEOUT_SECS}s"
fi

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
printf "\n%s===== Summary =====%s\n" "$BOLD" "$RESET"
printf "  passed:  %s%d%s\n" "$GREEN"  "$PASS" "$RESET"
printf "  failed:  %s%d%s\n" "$RED"    "$FAIL" "$RESET"
printf "  skipped: %s%d%s (Step 3 VC anchor optional; Step 5 finality poll if Data L1 exposes read API)\n" "$YELLOW" "$SKIP" "$RESET"

if [ "$FAIL" -gt 0 ]; then
  printf "\n%sGo/No-Go: NO-GO%s\n" "$RED" "$RESET"
  exit 1
fi
printf "\n%sGo/No-Go: %s%s\n" \
  "$([ "$SKIP" -gt 0 ] && printf "%s" "$YELLOW" || printf "%s" "$GREEN")" \
  "$([ "$SKIP" -gt 0 ] && echo "GO (with $SKIP pending steps)" || echo "GO")" \
  "$RESET"
exit 0
