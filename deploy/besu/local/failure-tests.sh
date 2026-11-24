#!/usr/bin/env bash
# Failure tests against the local chain and backend (`make dev` first). Each case is something
# that needed an operator, or lost data, on the chain Echo used before Besu. Takes about 5 minutes
# and restarts every validator; do not run it against anything but the local network.
set -uo pipefail
cd "$(dirname "$0")"
BACKEND="${BACKEND_URL:-http://localhost:8000}"
bn() { # block height of validator $1, or -1 if it does not answer
  local r; r=$(curl -s -m 3 -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "http://localhost:$((8544 + $1))" 2>/dev/null \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p')
  [ -n "$r" ] && printf '%d' "$r" || printf -- '-1'
}
heights() { echo "v1=$(bn 1) v2=$(bn 2) v3=$(bn 3) v4=$(bn 4)"; }
# Anchors a random root through the gateway; prints the HTTP status. 201 = anchored and final.
# A refusal's body goes to stderr with a timestamp, so a failure can be explained.
anchor() {
  local body code
  body=$(mktemp)
  code=$(curl -s -o "$body" -w '%{http_code}' -m 100 -X POST "$BACKEND/v1/data-l1/merkle-roots" \
    -H 'Content-Type: application/json' -d "{\"root\":\"$(openssl rand -hex 32)\",\"leafCount\":1}")
  [ "$code" = 201 ] || echo "    $(date -u +%T) anchor refused ($code): $(head -c 400 "$body")" >&2
  rm -f "$body"
  echo "$code"
}
wait_height() { until [ "$(bn "$1")" -ge "$2" ]; do sleep 1; done; }
FAILED=0
expect() { # expect <description> <actual> <wanted>
  if [ "$2" = "$3" ]; then echo "  ok    $1"; else echo "  FAIL  $1 (got $2, wanted $3)"; FAILED=1; fi
}

echo "== baseline: $(heights)"
expect "anchor through the gateway" "$(anchor)" 201

echo "== one of four validators down (v4)"
docker stop echo-besu-4 >/dev/null
expect "anchor with three validators" "$(anchor)" 201

echo "== the validator the gateway tries first is down (v1)"
docker start echo-besu-4 >/dev/null; wait_height 4 1; sleep 5
docker stop echo-besu-1 >/dev/null
expect "anchor fails over to another validator" "$(anchor)" 201
docker start echo-besu-1 >/dev/null; wait_height 1 1

echo "== two of four down: below quorum, the chain must stop, not fork"
sleep 5; docker stop echo-besu-3 echo-besu-4 >/dev/null; before=$(bn 1); sleep 20
expect "no new block while below quorum" "$(bn 1)" "$before"

echo "== both restarted, nothing else done"
t=$(date +%s); docker start echo-besu-3 echo-besu-4 >/dev/null; wait_height 1 $((before + 1))
echo "  blocks resumed $(( $(date +%s) - t ))s after restart"
expect "anchor after recovery" "$(anchor)" 201

echo "== v2 loses its disk and rejoins from nothing"
docker compose rm -sf besu2 >/dev/null 2>&1; docker volume rm echo-besu_data2 >/dev/null
t=$(date +%s); docker compose up -d besu2 >/dev/null 2>&1; target=$(bn 1); wait_height 2 "$target"
echo "  v2 synced to block $target in $(( $(date +%s) - t ))s"

echo "== anchoring every ~5s for 3 minutes while every validator is restarted in turn"
ok=0; bad=0
for i in $(seq 1 36); do
  case $i in 4) docker restart echo-besu-1 >/dev/null & ;; 12) docker restart echo-besu-2 >/dev/null & ;;
             20) docker restart echo-besu-3 >/dev/null & ;; 28) docker restart echo-besu-4 >/dev/null & ;; esac
  if [ "$(anchor)" = 201 ]; then ok=$((ok + 1)); else bad=$((bad + 1)); fi
  sleep 3
done
wait; sleep 8
echo "  anchors accepted=$ok failed=$bad; $(heights)"
expect "no anchor lost during rolling restarts" "$bad" 0
h=$(bn 1); expect "all validators at the same height" "$(heights)" "v1=$h v2=$h v3=$h v4=$h"

[ "$FAILED" = 0 ] && echo "RESULT: all failure tests passed" || { echo "RESULT: FAILED"; exit 1; }
