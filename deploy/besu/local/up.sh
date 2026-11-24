#!/usr/bin/env bash
# Brings up Echo's local chain and leaves everything the gateway needs in .network/chain.env.
# Safe to re-run: existing keys, genesis, chain data and contract are kept.
#   up.sh           start (creating the network on first run)
#   up.sh --reset   delete the chain's data, keys and contract first, then start afresh
#   up.sh --destroy delete them and stop
#
# The keys made here are for local development only.
set -euo pipefail
cd "$(dirname "$0")"
ROOT="$(cd ../../.. && pwd)"
BESU_IMAGE="${BESU_IMAGE:-hyperledger/besu:26.11.2}"
N=.network
compose="docker compose -f docker-compose.yml"

if [ "${1:-}" = --reset ] || [ "${1:-}" = --destroy ]; then
  [ -f $N/genesis.json ] && $compose down -v >/dev/null 2>&1 || true
  rm -rf $N
fi

[ "${1:-}" != --destroy ] || { echo "  local chain deleted"; exit 0; }

if [ ! -f $N/genesis.json ]; then
  mkdir -p $N
  docker run --rm -u "$(id -u):$(id -g)" -v "$PWD/..:/in:ro" -v "$PWD/$N:/out" "$BESU_IMAGE" \
    operator generate-blockchain-config --config-file=/in/qbft-config.json --to=/out/gen --private-key-file-name=key >/dev/null
  mv $N/gen/genesis.json $N/genesis.json
  i=0; nodes=""
  for d in $N/gen/keys/*/; do
    i=$((i + 1)); mkdir -p "$N/node$i"; mv "$d/key" "$d/key.pub" "$N/node$i/"
    basename "$d" > "$N/node$i/address"
    nodes="$nodes\"enode://$(sed 's/^0x//' "$N/node$i/key.pub")@172.63.0.1$i:30303\","
  done
  echo "[${nodes%,}]" > $N/static-nodes.json
  rm -rf $N/gen
  echo "  created 4 local validators"
fi

$compose up -d >/dev/null 2>&1
height() {
  curl -s -m 3 -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "http://localhost:$1" 2>/dev/null \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p' || true
}
printf "  waiting for blocks "
for _ in $(seq 1 60); do
  h="$(height 8545)"; [ -n "$h" ] && [ "$((h))" -ge 1 ] && break
  printf "."; sleep 2
done
[ -n "${h:-}" ] && [ "$((h))" -ge 1 ] || { echo " chain did not start; see: docker logs echo-besu-1"; exit 1; }
echo " block $((h))"

chain() { (cd "$ROOT" && go run ./cmd/echochain "$@"); }
# The gateway signs with the key it already has; on this chain that key's address is the writer.
# shellcheck disable=SC1091
ECHOAPP_ROOT="$ROOT" source "$ROOT/scripts/identity/ensure-identity-service-key.sh" >/dev/null
signer="$(chain address -key "$IDENTITY_SERVICE_KEY_PEM")"

[ -f $N/owner.key ] || chain newkey -out "$PWD/$N/owner.key" >/dev/null
if [ ! -s $N/anchor.address ]; then
  chain deploy -rpc http://localhost:8545 -owner "$PWD/$N/owner.key" -signer "$signer" > $N/anchor.address
  echo "  deployed anchor contract $(cat $N/anchor.address)"
fi

cat > $N/chain.env <<ENV
# Written by deploy/besu/local/up.sh. Local development only.
export ECHO_CHAIN_NETWORK=besu
export BESU_RPC_URLS=http://localhost:8545,http://localhost:8546,http://localhost:8547,http://localhost:8548
export BESU_RPC_URLS_DOCKER=http://host.docker.internal:8545,http://host.docker.internal:8546,http://host.docker.internal:8547,http://host.docker.internal:8548
export BESU_ANCHOR_CONTRACT=$(cat $N/anchor.address)
export BESU_VALIDATORS=$(cat $N/node*/address | paste -sd, -)
export BESU_GATEWAY_ADDRESS=$signer
ENV
echo "  chain ready: contract $(cat $N/anchor.address), gateway writes as $signer"
