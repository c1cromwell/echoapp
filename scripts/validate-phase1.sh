#!/usr/bin/env bash
# scripts/validate-phase1.sh — local go/no-go validation script (WO-230).
#
# Runs the end-to-end check against the backend and Echo's chain (`make dev`):
#   1. Derive did:key locally from a P-256 key pair (no chain transaction)
#   2. Register DID with local Identity Service (POST /identity/register)
#   3. Anchor a trust-tier commitment through the gateway and read it back from a validator
#   4. Send a test message through the local relay
#   5. Anchor a Merkle root through the gateway, export its proof from a
#      validator, and verify that proof offline, signatures included
#   6. Every validator keeps producing blocks and they agree on the chain
#   7. The reserve contract holds the whole fixed supply at one address
#
# Exit codes:
#   0  — all steps passed
#   1  — at least one step failed
#   2  — a prerequisite is missing (cluster not running, blocker WO unresolved)

set -uo pipefail

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# Echo's chain: a private Besu network. `make dev` writes its settings here.
CHAIN_ENV="${CHAIN_ENV:-$REPO_ROOT/deploy/besu/local/.network/chain.env}"
# shellcheck disable=SC1090
[ -f "$CHAIN_ENV" ] && . "$CHAIN_ENV"
BESU_RPC_URLS="${BESU_RPC_URLS:-http://localhost:8545,http://localhost:8546,http://localhost:8547,http://localhost:8548}"
BESU_ANCHOR_CONTRACT="${BESU_ANCHOR_CONTRACT:-}"
BESU_VALIDATORS="${BESU_VALIDATORS:-}"
IFS=',' read -r -a CHAIN_NODES <<< "$BESU_RPC_URLS"
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
printf "Chain:          %s validators (%s)\n" "${#CHAIN_NODES[@]}" "$BESU_RPC_URLS"
printf "Anchor contract: %s\n" "${BESU_ANCHOR_CONTRACT:-(not set)}"
printf "Finality limit: %ss\n" "$FINALITY_TIMEOUT_SECS"

step 0 "Prerequisites"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/identity/ensure-identity-service-key.sh
# Identity L1 POST /data requires a stable service DID + signing key on backend and L1/L0.
source "$SCRIPT_DIR/identity/ensure-identity-service-key.sh"
ok "identity service signing key ready (${IDENTITY_SERVICE_DID:0:24}…)"

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

# Block height of one validator, or empty if it does not answer.
chain_height() {
  local hex
  hex=$(curl -s --max-time 3 -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "$1" 2>/dev/null \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p')
  [ -n "$hex" ] && printf '%d' "$hex"
}

CLUSTER_OK=1
for node in "${CHAIN_NODES[@]}"; do
  h=$(chain_height "$node" || true)
  if [ -n "$h" ]; then ok "validator $node at block $h"; else fail "validator $node NOT reachable"; CLUSTER_OK=0; fi
done
if [ -n "$BESU_ANCHOR_CONTRACT" ] && [ -n "$BESU_VALIDATORS" ]; then
  ok "anchor contract $BESU_ANCHOR_CONTRACT"
else
  fail "chain settings missing (BESU_ANCHOR_CONTRACT, BESU_VALIDATORS); run 'make dev'"; CLUSTER_OK=0
fi
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
# Step 3: identity commitments on-chain
# -----------------------------------------------------------------------------
step 3 "Anchor a trust-tier commitment on-chain"

if [ -z "${BESU_IDENTITY_CONTRACT:-}" ]; then
  fail "BESU_IDENTITY_CONTRACT is not set; run 'make dev'"
else
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
      # The commitment is SHA-256(tier byte || nonce); recompute it here so the check does not
      # rest on what the gateway says it anchored.
      EXPECTED=$( { printf '\x02'; printf '%s' "$NONCE"; } | openssl dgst -sha256 2>/dev/null | sed -E 's/^.*[= ]//')
      if [ "${#COMM}" -ne 64 ]; then
        fail "response has no 64-character commitment: $(cat "$TMP_DIR/tier.out")"
      elif [ "$COMM" != "$EXPECTED" ]; then
        fail "gateway returned commitment $COMM; H(tier||nonce) is $EXPECTED"
      else
        ok "gateway anchored H(tier||nonce) = ${COMM:0:16}…"
        # Read it back from a validator the gateway does not use first.
        READER="${CHAIN_NODES[$(( ${#CHAIN_NODES[@]} - 1 ))]}"
        if BLOCK=$(cd "$REPO_ROOT" && go run ./cmd/echochain trust-tier -rpc "$READER" -identity "$BESU_IDENTITY_CONTRACT" -commitment "$COMM" 2>"$TMP_DIR/tier.err"); then
          ok "commitment is on-chain in block $BLOCK (read from $READER)"
        else
          fail "commitment not found on-chain: $(tr '\n' ' ' < "$TMP_DIR/tier.err")"
        fi
        # The record is stored under a hash of the DID. Check the block itself: the DID's bytes
        # must not appear anywhere in the transactions it carries.
        if [[ "${BLOCK:-}" =~ ^[0-9]+$ ]]; then
          DID_HEX=$(printf '%s' "$DID_KEY" | xxd -p | tr -d '\n')
          BLOCK_JSON=$(curl -s --max-time 5 -X POST -H 'Content-Type: application/json' \
            --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$(printf '0x%x' "$BLOCK")\",true],\"id\":1}" "$READER" 2>/dev/null)
          if ! printf '%s' "$BLOCK_JSON" | jq -e '.result.transactions | length > 0' >/dev/null 2>&1; then
            fail "could not read block $BLOCK to check its contents"
          elif printf '%s' "$BLOCK_JSON" | grep -qi "$DID_HEX"; then
            fail "block $BLOCK contains the DID in clear text"
          else
            ok "block $BLOCK does not contain the DID: it is stored as a hash"
          fi
        fi
      fi
      ;;
    503)
      fail "the backend has no identity contract configured: $(cat "$TMP_DIR/tier.out")"
      ;;
    000)
      fail "POST /v1/phase1/trust-tier-commitment: backend unreachable"
      ;;
    *)
      fail "POST /v1/phase1/trust-tier-commitment returned HTTP $TIER_CODE: $(cat "$TMP_DIR/tier.out")"
      ;;
  esac
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
# Step 5: Anchor a Merkle root, then prove it without trusting the gateway or the chain
# -----------------------------------------------------------------------------
step 5 "Anchor a Merkle root and verify its proof offline"

# Build a deterministic 32-byte SHA-256 root.
# macOS LibreSSL prints "(stdin)= <hex>"; GNU openssl uses "<hex>".
MERKLE_ROOT=$(printf 'phase1-validation-%s-%s' "$DID_KEY" "$(date +%s)" | openssl dgst -sha256 2>/dev/null | sed -E 's/^.*[= ]//')
if [ -z "$MERKLE_ROOT" ] || [ "${#MERKLE_ROOT}" -ne 64 ]; then
  MERKLE_ROOT=$(printf 'phase1-validation-%s-%s' "$DID_KEY" "$(date +%s)" | shasum -a 256 2>/dev/null | awk '{print $1}')
fi
ok "computed Merkle root: $MERKLE_ROOT"

MERKLE_BODY=$(jq -nc --arg root "$MERKLE_ROOT" '{root:$root, leafCount:1}')
SUBMIT_STARTED=$(date +%s)
SUBMIT_CODE=$(curl -sS -o "$TMP_DIR/merkle.out" -w '%{http_code}' --max-time "$FINALITY_TIMEOUT_SECS" \
  -X POST "$BACKEND_URL/v1/data-l1/merkle-roots" \
  -H "Content-Type: application/json" \
  -d "$MERKLE_BODY" 2>/dev/null) || SUBMIT_CODE="000"
SUBMIT_RESP=$(cat "$TMP_DIR/merkle.out" 2>/dev/null || true)

case "$SUBMIT_CODE" in
  200|201)
    TX_ID=$(printf '%s' "$SUBMIT_RESP" | jq -r '.tx_id // .txHash // .id // empty')
    if [ -z "$TX_ID" ]; then
      fail "submit response missing tx_id: $SUBMIT_RESP"
    else
      # The gateway answers only once the write is final: QBFT has no forks, so a
      # transaction that is in a block cannot be undone.
      ok "gateway anchored the root, final after $(( $(date +%s) - SUBMIT_STARTED ))s (tx $TX_ID)"

      # Ask a validator the gateway does not use first, and build the proof from
      # the chain alone. Nothing from the gateway is trusted from here on.
      PROVER="${CHAIN_NODES[$(( ${#CHAIN_NODES[@]} - 1 ))]}"
      if (cd "$REPO_ROOT" && go run ./cmd/echochain prove -rpc "$PROVER" -contract "$BESU_ANCHOR_CONTRACT" \
            -root "$MERKLE_ROOT" -out "$TMP_DIR/bundle.json") 2>"$TMP_DIR/prove.err"; then
        ok "proof exported from validator $PROVER"
        # Offline: no network access. Checks the validators' signatures on the
        # block, the state proof, and the stored entry.
        if (cd "$REPO_ROOT" && go run ./cmd/verifyproof -bundle "$TMP_DIR/bundle.json" -validators "$BESU_VALIDATORS") >"$TMP_DIR/verify.out" 2>&1; then
          ok "proof verified offline: $(grep -m1 'carries valid signatures' "$TMP_DIR/verify.out" | sed 's/^OK *//')"
        else
          fail "offline verification failed: $(tr '\n' ' ' < "$TMP_DIR/verify.out")"
        fi
        # The same proof must be refused by a verifier that trusts other validators.
        if (cd "$REPO_ROOT" && go run ./cmd/verifyproof -bundle "$TMP_DIR/bundle.json" \
              -validators 0x0000000000000000000000000000000000000001,0x0000000000000000000000000000000000000002,0x0000000000000000000000000000000000000003,0x0000000000000000000000000000000000000004) >/dev/null 2>&1; then
          fail "proof was accepted against validators that did not sign it"
        else
          ok "proof is refused when checked against other validators"
        fi
      else
        fail "could not export a proof from $PROVER: $(tr '\n' ' ' < "$TMP_DIR/prove.err")"
      fi
    fi
    ;;
  503)
    fail "the backend has no chain configured (BESU_RPC_URLS / BESU_ANCHOR_CONTRACT): $SUBMIT_RESP"
    ;;
  000)
    fail "POST /v1/data-l1/merkle-roots: backend unreachable"
    ;;
  *)
    fail "POST /v1/data-l1/merkle-roots returned HTTP $SUBMIT_CODE: $SUBMIT_RESP"
    ;;
esac

# -----------------------------------------------------------------------------
# Step 6: every validator advances, and they agree on the chain
# -----------------------------------------------------------------------------
step 6 "Verify validators keep producing blocks and agree"

block_hash() { # block_hash <node> <decimal height>
  curl -s --max-time 3 -X POST -H 'Content-Type: application/json' \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$(printf '0x%x' "$2")\",false],\"id\":1}" "$1" 2>/dev/null \
    | jq -r '.result.hash // empty'
}

START_HEIGHT=$(chain_height "${CHAIN_NODES[0]}" || true)
if ! [[ "$START_HEIGHT" =~ ^[0-9]+$ ]]; then
  fail "could not read the block height"
else
  ok "starting block: $START_HEIGHT"
  # Blocks arrive every 2 seconds under load and every 30 when idle; a write forces one.
  curl -s -o /dev/null --max-time "$FINALITY_TIMEOUT_SECS" -X POST "$BACKEND_URL/v1/data-l1/merkle-roots" \
    -H "Content-Type: application/json" -d "$(jq -nc --arg root "$(openssl rand -hex 32)" '{root:$root, leafCount:1}')" || true
  DEADLINE=$(( $(date +%s) + FINALITY_TIMEOUT_SECS ))
  BEHIND=""
  while [ "$(date +%s)" -lt "$DEADLINE" ]; do
    BEHIND=""
    for node in "${CHAIN_NODES[@]}"; do
      h=$(chain_height "$node" || true)
      { [[ "$h" =~ ^[0-9]+$ ]] && [ "$h" -gt "$START_HEIGHT" ]; } || BEHIND="$BEHIND $node"
    done
    [ -z "$BEHIND" ] && break
    sleep 2
  done
  if [ -z "$BEHIND" ]; then
    ok "all ${#CHAIN_NODES[@]} validators advanced past block $START_HEIGHT"
  else
    fail "validators did not advance within ${FINALITY_TIMEOUT_SECS}s:$BEHIND"
  fi
  # One hash per validator for the same height: more than one distinct value would be a fork.
  CHECK=$(( START_HEIGHT + 1 ))
  HASHES=$(for node in "${CHAIN_NODES[@]}"; do block_hash "$node" "$CHECK"; done | sort -u | grep -c .)
  if [ "$HASHES" -eq 1 ]; then
    ok "all validators hold the same block $CHECK"
  else
    fail "validators disagree on block $CHECK ($HASHES distinct hashes)"
  fi
fi

# -----------------------------------------------------------------------------
# Step 7: the reserve holds the whole fixed supply at one address
# -----------------------------------------------------------------------------
step 7 "Verify the reserve: fixed supply, held at one address"

if [ -z "${BESU_RESERVE_CONTRACT:-}" ]; then
  fail "BESU_RESERVE_CONTRACT is not set; run 'make dev'"
else
  READER="${CHAIN_NODES[$(( ${#CHAIN_NODES[@]} - 1 ))]}"
  if (cd "$REPO_ROOT" && go run ./cmd/echochain reserve -rpc "$READER" -contract "$BESU_RESERVE_CONTRACT") >"$TMP_DIR/reserve.out" 2>"$TMP_DIR/reserve.err"; then
    SUPPLY=$(sed -n 's/^total_supply=//p' "$TMP_DIR/reserve.out")
    HELD=$(sed -n 's/^balance_of_reserved=//p' "$TMP_DIR/reserve.out")
    RESERVED=$(sed -n 's/^reserved=//p' "$TMP_DIR/reserve.out")
    # 1,000,000,000 units with 8 decimal places.
    if [ "$SUPPLY" = "100000000000000000" ]; then
      ok "total supply is exactly 1,000,000,000 (8 decimals)"
    else
      fail "total supply is $SUPPLY, expected 100000000000000000"
    fi
    if [ -n "$HELD" ] && [ "$HELD" = "$SUPPLY" ]; then
      ok "all of it is held at the reserved address $RESERVED"
    else
      fail "the reserved address holds $HELD of $SUPPLY"
    fi
  else
    fail "could not read the reserve: $(tr '\n' ' ' < "$TMP_DIR/reserve.err")"
  fi
fi

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
printf "\n%s===== Summary =====%s\n" "$BOLD" "$RESET"
printf "  passed:  %s%d%s\n" "$GREEN"  "$PASS" "$RESET"
printf "  failed:  %s%d%s\n" "$RED"    "$FAIL" "$RESET"
printf "  skipped: %s%d%s\n" "$YELLOW" "$SKIP" "$RESET"

if [ "$FAIL" -gt 0 ]; then
  printf "\n%sGo/No-Go: NO-GO%s\n" "$RED" "$RESET"
  exit 1
fi
printf "\n%sGo/No-Go: %s%s\n" \
  "$([ "$SKIP" -gt 0 ] && printf "%s" "$YELLOW" || printf "%s" "$GREEN")" \
  "$([ "$SKIP" -gt 0 ] && echo "GO (with $SKIP pending steps)" || echo "GO")" \
  "$RESET"
exit 0
