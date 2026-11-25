#!/usr/bin/env bash
# Checks that the local chain is closed (`make dev` first): a node that is not on the allowlist,
# holding the right genesis and the validators' addresses, and sitting on the same network, is
# refused by every validator and receives no blocks. Takes about 40 seconds.
set -uo pipefail
cd "$(dirname "$0")"
N=.network
BESU_IMAGE="${BESU_IMAGE:-hyperledger/besu:26.11.2}"
rpc() { # rpc <port> <method> -> result as a decimal number, or -1
  local r; r=$(curl -s -m 3 -X POST -H 'Content-Type: application/json' \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"$2\",\"params\":[],\"id\":1}" "http://localhost:$1" 2>/dev/null \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p')
  [ -n "$r" ] && printf '%d' "$r" || printf -- '-1'
}
FAILED=0
expect() { if [ "$2" = "$3" ]; then echo "  ok    $1"; else echo "  FAIL  $1 (got $2, wanted $3)"; FAILED=1; fi; }
cleanup() { docker rm -f echo-besu-outsider >/dev/null 2>&1; }
trap cleanup EXIT; cleanup

echo "== an outsider with the right genesis and the validators' addresses tries to join"
docker run -d --name echo-besu-outsider --network echo-besu_besu --ip 172.63.0.99 -p 8599:8545 -u root -v /data \
  -v "$PWD/$N/genesis.json:/config/genesis.json:ro" -v "$PWD/$N/static-nodes.json:/config/static-nodes.json:ro" \
  "$BESU_IMAGE" --data-path=/data --genesis-file=/config/genesis.json --static-nodes-file=/config/static-nodes.json \
  --discovery-enabled=false --min-gas-price=0 --rpc-http-enabled --rpc-http-host=0.0.0.0 --rpc-http-api=ETH,NET \
  --host-allowlist='*' --p2p-host=172.63.0.99 >/dev/null
for _ in $(seq 1 30); do [ "$(rpc 8599 net_peerCount)" -ge 0 ] && break; sleep 2; done
expect "the outsider started" "$([ "$(rpc 8599 net_peerCount)" -ge 0 ] && echo yes || echo no)" yes
# Long enough for several connection attempts in both directions.
sleep 25
expect "validators the outsider is connected to" "$(rpc 8599 net_peerCount)" 0
expect "blocks the outsider has received" "$(rpc 8599 eth_blockNumber)" 0
for i in 1 2 3 4; do
  expect "validator $i is connected to the other three and no one else" "$(rpc $((8544 + i)) net_peerCount)" 3
done
expect "the chain it could not join has blocks" "$([ "$(rpc 8545 eth_blockNumber)" -gt 0 ] && echo yes || echo no)" yes

[ "$FAILED" = 0 ] && echo "PERMISSION TEST PASSED" || { echo "PERMISSION TEST FAILED"; exit 1; }
