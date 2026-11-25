#!/usr/bin/env bash
# Restore drill on the local chain (`make dev` first): proves the chain can be rebuilt from
# backups alone. It
#   1. anchors a root and exports its proof,
#   2. backs up each validator in turn while the chain keeps running, plus the keys and genesis,
#   3. destroys every validator's data and the network's configuration,
#   4. restores from the backups and nothing else,
#   5. checks that the old proof still verifies, the old root is still on the chain, the reserve is
#      intact, the chain is producing blocks past where it was, and a new root can be anchored.
# Takes about three minutes. Local network only: step 3 deletes the chain.
set -uo pipefail
cd "$(dirname "$0")"
ROOT="$(cd ../../.. && pwd)"
N=.network
BACKEND="${BACKEND_URL:-http://localhost:8000}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
# shellcheck disable=SC1091
. "$N/chain.env"
chain() { (cd "$ROOT" && go run ./cmd/echochain "$@"); }
bn() {
  local r; r=$(curl -s -m 3 -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "http://localhost:$((8544 + $1))" 2>/dev/null \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p')
  [ -n "$r" ] && printf '%d' "$r" || printf -- '-1'
}
anchor() { # anchors root $1 through the gateway; prints the HTTP status
  curl -s -o /dev/null -w '%{http_code}' -m 100 -X POST "$BACKEND/v1/data-l1/merkle-roots" \
    -H 'Content-Type: application/json' -d "{\"root\":\"$1\",\"leafCount\":1}"
}
FAILED=0
expect() { if [ "$2" = "$3" ]; then echo "  ok    $1"; else echo "  FAIL  $1 (got $2, wanted $3)"; FAILED=1; fi; }

echo "== before"
ROOT_HEX="$(openssl rand -hex 32)"
expect "anchor a root through the gateway" "$(anchor "$ROOT_HEX")" 201
chain prove -rpc http://localhost:8545 -contract "$BESU_ANCHOR_CONTRACT" -root "$ROOT_HEX" -out "$WORK/proof.json" >/dev/null 2>&1
expect "export its proof" "$([ -s "$WORK/proof.json" ] && echo yes || echo no)" yes
RESERVE_BEFORE="$(chain reserve -rpc http://localhost:8545 -contract "$BESU_RESERVE_CONTRACT" 2>/dev/null | tr '\n' ' ')"

echo "== back up each validator in turn; the chain keeps running on the other three"
for i in 1 2 3 4; do
  ../ops/backup.sh "echo-besu-$i" "echo-besu_data$i" "$WORK/backup/node$i.tar.gz" >/dev/null &
  pid=$!
  # Anchor while this validator is stopped for its backup.
  until [ "$(docker inspect -f '{{.State.Running}}' "echo-besu-$i")" = false ] || ! kill -0 $pid 2>/dev/null; do sleep 0.2; done
  code="$(anchor "$(openssl rand -hex 32)")"
  wait $pid
  expect "validator $i backed up" "$([ -s "$WORK/backup/node$i.tar.gz" ] && echo yes || echo no)" yes
  expect "a root was anchored while validator $i was down" "$code" 201
  until [ "$(bn $i)" -ge 1 ]; do sleep 1; done
  sleep 5
done
tar -C "$N" -czf "$WORK/backup/config.tar.gz" .
HEIGHT_BEFORE=$(bn 1)
echo "   height before the loss: $HEIGHT_BEFORE"

echo "== lose everything: all four validators' data, keys, genesis"
docker compose -f docker-compose.yml down -v >/dev/null 2>&1
rm -rf "$N"
expect "chain data is gone" "$(docker volume ls -q | grep -c '^echo-besu_data' || true)" 0
expect "configuration is gone" "$([ -e "$N" ] && echo present || echo gone)" gone

echo "== restore from the backups alone"
mkdir -p "$N" && tar -C "$N" -xzf "$WORK/backup/config.tar.gz"
docker compose -f docker-compose.yml create >/dev/null 2>&1
for i in 1 2 3 4; do
  ../ops/restore.sh "echo-besu-$i" "echo-besu_data$i" "$WORK/backup/node$i.tar.gz" >/dev/null
done
# A tampered backup must be refused.
cp "$WORK/backup/node1.tar.gz" "$WORK/backup/bad.tar.gz"; cp "$WORK/backup/node1.tar.gz.sha256" "$WORK/backup/bad.tar.gz.sha256"
sed -i.bak 's/node1.tar.gz/bad.tar.gz/' "$WORK/backup/bad.tar.gz.sha256"; printf 'x' >> "$WORK/backup/bad.tar.gz"
if ../ops/restore.sh echo-besu-1 echo-besu_data1 "$WORK/backup/bad.tar.gz" >/dev/null 2>&1; then r=accepted; else r=refused; fi
expect "a backup that does not match its checksum" "$r" refused
docker compose -f docker-compose.yml up -d >/dev/null 2>&1
for _ in $(seq 1 90); do [ "$(bn 1)" -gt "$HEIGHT_BEFORE" ] && break; sleep 2; done

echo "== after"
expect "chain is producing blocks past where it was" "$([ "$(bn 1)" -gt "$HEIGHT_BEFORE" ] && echo yes || echo no)" yes
sleep 4
h1=$(bn 1); ok=yes; for i in 2 3 4; do [ "$(bn $i)" -ge $((h1 - 3)) ] || ok=no; done
expect "all four validators are following" "$ok" yes
if (cd "$ROOT" && go run ./cmd/verifyproof -bundle "$WORK/proof.json" -validators "$BESU_VALIDATORS") >/dev/null 2>&1; then r=valid; else r=invalid; fi
expect "the proof exported before the loss still verifies offline" "$r" valid
if chain prove -rpc http://localhost:8547 -contract "$BESU_ANCHOR_CONTRACT" -root "$ROOT_HEX" -out "$WORK/proof2.json" >/dev/null 2>&1; then r=found; else r=missing; fi
expect "the root anchored before the loss is on the restored chain" "$r" found
expect "the reserve is unchanged" "$(chain reserve -rpc http://localhost:8546 -contract "$BESU_RESERVE_CONTRACT" 2>/dev/null | tr '\n' ' ')" "$RESERVE_BEFORE"
code=000; for _ in $(seq 1 20); do code="$(anchor "$(openssl rand -hex 32)")"; [ "$code" = 201 ] && break; sleep 3; done
expect "a new root can be anchored after the restore" "$code" 201

[ "$FAILED" = 0 ] && echo "RESTORE DRILL PASSED" || { echo "RESTORE DRILL FAILED"; exit 1; }
